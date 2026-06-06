package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dracory/neat/contracts/database/orm"
	contractsseeder "github.com/dracory/neat/contracts/database/seeder"
	"github.com/dracory/neat/contracts/log"
	contractsMigration "github.com/dracory/neat/contracts/migration"
	"github.com/dracory/neat/database/db"
	databaseMigration "github.com/dracory/neat/database/migration"
	databaseorm "github.com/dracory/neat/database/orm"
	"github.com/dracory/neat/database/schema"
	databaseseeder "github.com/dracory/neat/database/seeder"
)

// Database is the main entry point for the neat package.
type Database struct {
	ctx         context.Context
	config      *db.DBConfig
	logger      log.Log
	eventBus    *databaseorm.EventBus
	ormInstance orm.Orm
	schema      *schema.Schema
	seeder      *databaseseeder.Runner
}

// Option is a functional option for configuring the Database.
type Option func(*options)

type options struct {
	ctx      context.Context
	logger   log.Log
	eventBus *databaseorm.EventBus
	pool     *db.PoolConfig
	skipPing bool
	debug    bool
}

// WithContext sets the context for the database.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithLogger sets the logger for the database.
func WithLogger(logger log.Log) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithEventBus sets the event bus for the database.
func WithEventBus(eventBus *databaseorm.EventBus) Option {
	return func(o *options) {
		o.eventBus = eventBus
	}
}

// WithPool sets the connection pool configuration for the database.
func WithPool(pool db.PoolConfig) Option {
	return func(o *options) {
		o.pool = &pool
	}
}

// SkipPing skips the initial database ping during connection.
func SkipPing() Option {
	return func(o *options) {
		o.skipPing = true
	}
}

// WithDebug enables debug mode for the database.
func WithDebug() Option {
	return func(o *options) {
		o.debug = true
	}
}

// New creates a new Database instance from a DBConfig.
func New(cfg db.DBConfig, opts ...Option) (*Database, error) {
	// Apply options
	o := &options{
		ctx:      context.Background(),
		logger:   log.NewStdLogger(),
		eventBus: databaseorm.NewEventBus(),
	}
	for _, opt := range opts {
		opt(o)
	}

	// Handle nil context by using background context
	if o.ctx == nil {
		o.ctx = context.Background()
	}

	// Ensure default connection is set
	if cfg.Default == "" && len(cfg.Connections) > 0 {
		for name := range cfg.Connections {
			cfg.Default = name
			break
		}
	}

	// Apply pool configuration from options if provided
	if o.pool != nil {
		cfg.Pool = *o.pool
	}

	// Create database instance
	database := &Database{
		ctx:      o.ctx,
		config:   &cfg,
		logger:   o.logger,
		eventBus: o.eventBus,
		seeder:   databaseseeder.NewRunner(),
	}

	// Initialize ORM
	ormInstance, err := databaseorm.BuildOrm(database.ctx, database.config, database.config.Default, database.logger, func() {
		// Refresh function - for now a no-op
	}, databaseorm.WithSkipPing(o.skipPing))
	if err != nil {
		return nil, fmt.Errorf("failed to build ORM: %w", err)
	}
	database.ormInstance = ormInstance

	// Initialize Schema
	s, err := schema.NewSchema(database.config, database.logger, database.ormInstance, nil)
	if err != nil {
		return nil, err
	}
	database.schema = s

	return database, nil
}

// NewFromDSN creates a new Database instance from a DSN string.
// Supported DSN formats:
// - postgres://user:pass@localhost:5432/mydb?sslmode=disable
// - mysql://user:pass@localhost:3306/mydb
// - sqlite://path/to/database.db
func NewFromDSN(dsn string, opts ...Option) (*Database, error) {
	// Apply options
	o := &options{
		ctx:      context.Background(),
		logger:   log.NewStdLogger(),
		eventBus: databaseorm.NewEventBus(),
	}
	for _, opt := range opts {
		opt(o)
	}

	// Parse DSN
	_, config, err := parseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN %s: %w", redactDSN(dsn), err)
	}

	// Create DBConfig
	poolConfig := db.PoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: int(time.Hour.Seconds()),
		ConnMaxIdleTime: int(time.Hour.Seconds()),
	}
	if o.pool != nil {
		poolConfig = *o.pool
	}

	cfg := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": config,
		},
		Pool:  poolConfig,
		Debug: o.debug,
	}

	return New(cfg, opts...)
}

// NewFromSQLDB creates a new Database instance from an already-open *sql.DB.
// The driverName parameter specifies the database driver ("mysql", "postgres",
// "sqlite", "sqlserver", "oracle", "turso"). Pass an empty string to let Neat
// auto-detect the driver by inspecting db.Driver() via reflection; an error is
// returned when detection fails.
// The caller retains full ownership of sqlDB — Neat will not close it or alter
// its connection-pool settings.
func NewFromSQLDB(sqlDB *sql.DB, driverName string, opts ...Option) (*Database, error) {
	if sqlDB == nil {
		return nil, fmt.Errorf("sqlDB cannot be nil")
	}

	if driverName == "" {
		driverName = detectDriverName(sqlDB)
	}
	if driverName == "" {
		return nil, fmt.Errorf("cannot detect database driver from *sql.DB; pass the driver name explicitly")
	}

	o := &options{
		ctx:      context.Background(),
		logger:   log.NewStdLogger(),
		eventBus: databaseorm.NewEventBus(),
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.ctx == nil {
		o.ctx = context.Background()
	}

	const connName = "default"
	cfg := db.DBConfig{
		Default: connName,
		Connections: map[string]db.ConnectionConfig{
			connName: {Driver: driverName},
		},
		Debug: o.debug,
	}
	if o.pool != nil {
		cfg.Pool = *o.pool
	}

	database := &Database{
		ctx:      o.ctx,
		config:   &cfg,
		logger:   o.logger,
		eventBus: o.eventBus,
		seeder:   databaseseeder.NewRunner(),
	}

	ormInstance, err := databaseorm.BuildOrmFromDB(o.ctx, sqlDB, driverName, connName, &cfg, o.logger, func() {})
	if err != nil {
		return nil, fmt.Errorf("failed to build ORM from *sql.DB: %w", err)
	}
	database.ormInstance = ormInstance

	s, err := schema.NewSchema(database.config, database.logger, database.ormInstance, nil)
	if err != nil {
		return nil, err
	}
	database.schema = s

	return database, nil
}

// detectDriverName returns the Neat driver name for the given *sql.DB by
// inspecting the type name of db.Driver() via reflection.
func detectDriverName(sqlDB *sql.DB) string {
	name := reflect.ValueOf(sqlDB.Driver()).Type().String()
	switch {
	case strings.Contains(name, "mysql"):
		return "mysql"
	case strings.Contains(name, "postgres"), strings.Contains(name, "pq"), strings.Contains(name, "pgx"):
		return "postgres"
	case strings.Contains(name, "sqlite"):
		return "sqlite"
	case strings.Contains(name, "mssql"), strings.Contains(name, "sqlserver"):
		return "sqlserver"
	case strings.Contains(name, "oracle"):
		return "oracle"
	default:
		return ""
	}
}

// redactDSN removes credentials from a DSN string for safe logging/error messages.
func redactDSN(dsn string) string {
	if dsn == "" {
		return dsn
	}
	// Handle URL-style DSNs (e.g., postgres://user:pass@host/db)
	if strings.Contains(dsn, "://") {
		parts := strings.SplitN(dsn, "://", 2)
		if len(parts) == 2 {
			scheme := parts[0]
			rest := parts[1]
			// Remove user:password@ part if present
			if atIdx := strings.Index(rest, "@"); atIdx != -1 {
				return scheme + "://[REDACTED]@" + rest[atIdx+1:]
			}
			return dsn
		}
	}
	// Handle mysql:// DSNs with user:pass@tcp(host:port)/db format
	if strings.HasPrefix(dsn, "mysql://") {
		rest := strings.TrimPrefix(dsn, "mysql://")
		if atIdx := strings.Index(rest, "@"); atIdx != -1 {
			return "mysql://[REDACTED]@" + rest[atIdx+1:]
		}
	}
	return dsn
}

// parseDSN parses a DSN string and returns the driver and connection config.
func parseDSN(dsn string) (string, db.ConnectionConfig, error) {
	if dsn == "" {
		return "", db.ConnectionConfig{}, fmt.Errorf("DSN cannot be empty")
	}
	if len(dsn) > 4096 {
		return "", db.ConnectionConfig{}, fmt.Errorf("DSN exceeds maximum length of 4096 characters")
	}

	// Special handling for sqlite:// which url.Parse might fail on if it contains colons like :memory:
	if strings.HasPrefix(dsn, "sqlite://") {
		rawPath := strings.TrimPrefix(dsn, "sqlite://")
		dbPath := rawPath
		if questionIndex := strings.Index(rawPath, "?"); questionIndex != -1 {
			dbPath = rawPath[:questionIndex]
		}
		return "sqlite", db.ConnectionConfig{
			Driver:   "sqlite",
			Database: dbPath,
		}, nil
	}

	if strings.HasPrefix(dsn, "turso://") {
		rawPath := strings.TrimPrefix(dsn, "turso://")
		dbPath := rawPath
		return "turso", db.ConnectionConfig{
			Driver:   "turso",
			Database: dbPath,
		}, nil
	}

	// Detect driver from scheme
	u, err := url.Parse(dsn)
	if err != nil {
		// Fallback for mysql://user:pass@tcp(host:port)/db
		if strings.HasPrefix(dsn, "mysql://") {
			rawDsn := strings.TrimPrefix(dsn, "mysql://")
			config := db.ConnectionConfig{
				Driver: "mysql",
				Dsn:    rawDsn,
			}

			// Extract database name: find last '/' and first '?' after it
			lastSlash := strings.LastIndex(rawDsn, "/")
			if lastSlash != -1 {
				dbName := rawDsn[lastSlash+1:]
				if firstQuestion := strings.Index(dbName, "?"); firstQuestion != -1 {
					dbName = dbName[:firstQuestion]
				}
				config.Database = dbName
			}

			// Try to extract username/password for metadata if possible
			if atIndex := strings.LastIndex(rawDsn, "@"); atIndex != -1 {
				userPass := rawDsn[:atIndex]
				if colonIndex := strings.Index(userPass, ":"); colonIndex != -1 {
					config.Username = userPass[:colonIndex]
					config.Password = userPass[colonIndex+1:]
				} else {
					config.Username = userPass
				}
			}

			return "mysql", config, nil
		}
		return "", db.ConnectionConfig{}, fmt.Errorf("invalid DSN %s: %w", redactDSN(dsn), err)
	}

	driver := u.Scheme
	config := db.ConnectionConfig{
		Driver:   driver,
		Host:     u.Hostname(),
		Username: u.User.Username(),
	}

	if password, ok := u.User.Password(); ok {
		config.Password = password
	}

	// Parse database name from path
	if u.Path != "" {
		config.Database = strings.TrimPrefix(u.Path, "/")
	} else if u.Host != "" && driver == "sqlite" {
		config.Database = u.Host
	}

	// Parse port
	if portStr := u.Port(); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}

	// Parse query parameters for driver-specific options
	query := u.Query()
	switch driver {
	case "postgres":
		config.SSLMode = query.Get("sslmode")
		if config.SSLMode == "" {
			config.SSLMode = "require"
		}
		config.Schema = query.Get("search_path")
		if config.Schema == "" {
			config.Schema = "public"
		}
		config.Timezone = query.Get("timezone")
		if config.Timezone == "" {
			config.Timezone = "UTC"
		}
	case "mysql":
		config.Charset = query.Get("charset")
		config.Loc = query.Get("loc")
	}

	return driver, config, nil
}

// Query returns the ORM query builder.
func (d *Database) Query() orm.Query {
	return d.ormInstance.Query()
}

// Schema returns the schema builder.
func (d *Database) Schema() *schema.Schema {
	return d.schema
}

// DB returns the underlying database connection.
func (d *Database) DB() (*sql.DB, error) {
	return d.ormInstance.DB()
}

// Close closes the database connection.
func (d *Database) Close() error {
	sqlDB, err := d.ormInstance.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Connection returns a new Database instance for a different connection.
func (d *Database) Connection(name string) (*Database, error) {
	if name == "" {
		name = d.config.Default
	}

	// Check if connection exists in config
	if _, ok := d.config.Connections[name]; !ok {
		return nil, fmt.Errorf("connection %s not found in configuration", name)
	}

	ormConn := d.ormInstance.Connection(name)
	if ormConn == nil {
		return nil, fmt.Errorf("connection %s not found", name)
	}

	schemaConn := d.schema.Connection(name)
	if schemaConn == nil {
		return nil, fmt.Errorf("schema connection %s not found", name)
	}

	// Type assert from interface to concrete type
	schemaInstance, ok := schemaConn.(*schema.Schema)
	if !ok {
		return nil, fmt.Errorf("failed to convert connection to Schema instance")
	}

	newDB := &Database{
		ctx:         d.ctx,
		config:      d.config,
		logger:      d.logger,
		eventBus:    d.eventBus,
		ormInstance: ormConn,
		schema:      schemaInstance,
		seeder:      databaseseeder.NewRunner(),
	}

	return newDB, nil
}

// Transaction executes a function within a database transaction.
func (d *Database) Transaction(txFunc func(tx orm.Query) error, opts ...*sql.TxOptions) error {
	return d.ormInstance.Transaction(txFunc, opts...)
}

// DatabaseName returns the name of the current database.
func (d *Database) DatabaseName() string {
	return d.ormInstance.DatabaseName()
}

// Name returns the name of the current connection.
func (d *Database) Name() string {
	return d.ormInstance.Name()
}

// DisableQueryLog disables the capturing of executed queries.
func (d *Database) DisableQueryLog() {
	d.ormInstance.DisableQueryLog()
}

// EnableQueryLog enables the capturing of executed queries.
func (d *Database) EnableQueryLog() {
	d.ormInstance.EnableQueryLog()
}

// FlushQueryLog clears the captured queries from the log.
func (d *Database) FlushQueryLog() {
	d.ormInstance.FlushQueryLog()
}

// GetQueryLog retrieves the captured queries from the log.
func (d *Database) GetQueryLog() []orm.QueryLog {
	return d.ormInstance.GetQueryLog()
}

// Factory returns the ORM factory for creating test data.
func (d *Database) Factory() orm.Factory {
	return d.ormInstance.Factory()
}

// Observe registers an observer for the given model.
func (d *Database) Observe(model any, observer orm.Observer) {
	d.ormInstance.Observe(model, observer)
}

// Migrate runs all pending migrations.
func (d *Database) Migrate(paths ...string) error {
	migrator := d.getMigrator(paths)
	return migrator.Run()
}

// MigrateDown rolls back the last migration batch.
func (d *Database) MigrateDown(step int, paths ...string) error {
	migrator := d.getMigrator(paths)
	return migrator.Rollback(step, 0)
}

// MigrateFresh drops all tables and re-runs all migrations.
func (d *Database) MigrateFresh(paths ...string) error {
	migrator := d.getMigrator(paths)
	return migrator.Fresh()
}

// MigrateReset rolls back all migrations and re-runs them.
func (d *Database) MigrateReset(paths ...string) error {
	migrator := d.getMigrator(paths)
	return migrator.Reset()
}

// MigrationStatus returns the status of all migrations.
func (d *Database) MigrationStatus(paths ...string) ([]contractsMigration.Status, error) {
	migrator := d.getMigrator(paths)
	return migrator.Status()
}

func (d *Database) getMigrator(paths []string) contractsMigration.Migrator {
	if len(paths) == 0 {
		paths = []string{"./migrations"}
	}
	return databaseMigration.NewMigrator(d.config, d.ormInstance, d.schema, paths)
}

// Seed runs the specified seeders.
func (d *Database) Seed(seeders []contractsseeder.Seeder) error {
	return d.seeder.Call(seeders)
}

// SeedOnce runs the specified seeders only once.
func (d *Database) SeedOnce(seeders []contractsseeder.Seeder) error {
	return d.seeder.CallOnce(seeders)
}

// Seeder returns a seeder facade for advanced seeder operations.
func (d *Database) Seeder() contractsseeder.Facade {
	return d.seeder
}
