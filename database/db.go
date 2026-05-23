package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	databaseorm "github.com/dracory/neat/database/orm"
	"github.com/dracory/neat/database/schema"
)

// Database is the main entry point for the neat package.
type Database struct {
	ctx         context.Context
	config      *db.DBConfig
	logger      log.Log
	eventBus    *databaseorm.EventBus
	ormInstance orm.Orm
	schema      *schema.Schema
}

// Option is a functional option for configuring the Database.
type Option func(*options)

type options struct {
	ctx      context.Context
	logger   log.Log
	eventBus *databaseorm.EventBus
	pool     *db.PoolConfig
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

	// Ensure default connection is set
	if cfg.Default == "" && len(cfg.Connections) > 0 {
		for name := range cfg.Connections {
			cfg.Default = name
			break
		}
	}

	// Create database instance
	database := &Database{
		ctx:      o.ctx,
		config:   &cfg,
		logger:   o.logger,
		eventBus: o.eventBus,
	}

	// Initialize ORM
	ormInstance, err := databaseorm.BuildOrm(database.ctx, database.config, database.config.Default, database.logger, func() {
		// Refresh function - for now a no-op
	})
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
		Debug: false,
	}

	return New(cfg, opts...)
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
