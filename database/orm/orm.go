package orm

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/query"
)

// Orm represents the ORM instance for database operations.
type Orm struct {
	ctx             context.Context
	dbConfig        *db.DBConfig
	connection      string
	log             log.Log
	modelToObserver []contractsorm.ModelToObserver
	mutex           sync.Mutex
	query           contractsorm.Query
	queries         map[string]contractsorm.Query
	refresh         func()
	drivers         map[string]driver.Driver
	dbConnections   map[string]*sql.DB
}

// NewOrm creates a new Orm instance.
func NewOrm(
	ctx context.Context,
	dbConfig *db.DBConfig,
	connection string,
	query contractsorm.Query,
	queries map[string]contractsorm.Query,
	log log.Log,
	modelToObserver []contractsorm.ModelToObserver,
	refresh func(),
	drivers map[string]driver.Driver,
	dbConnections map[string]*sql.DB,
) *Orm {
	return &Orm{
		ctx:             ctx,
		dbConfig:        dbConfig,
		connection:      connection,
		log:             log,
		modelToObserver: modelToObserver,
		query:           query,
		queries:         queries,
		refresh:         refresh,
		drivers:         drivers,
		dbConnections:   dbConnections,
	}
}

// OrmOption is a function that configures Orm options.
type OrmOption func(*ormOptions)

// ormOptions holds configuration options for Orm.
type ormOptions struct {
	skipPing bool
}

// WithSkipPing sets whether to skip the database ping during initialization.
func WithSkipPing(skip bool) OrmOption {
	return func(o *ormOptions) {
		o.skipPing = skip
	}
}

// BuildOrm builds and initializes a new Orm instance with the given configuration.
func BuildOrm(ctx context.Context, dbConfig *db.DBConfig, connection string, log log.Log, refresh func(), opts ...OrmOption) (*Orm, error) {
	// Apply options
	o := &ormOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Initialize drivers and connections
	drivers := make(map[string]driver.Driver)
	dbConnections := make(map[string]*sql.DB)

	// Get default connection if not specified
	if connection == "" {
		connection = dbConfig.Default
	}

	// Build query for the connection
	query, err := buildQuery(ctx, dbConfig, connection, log, drivers, dbConnections, o.skipPing)
	if err != nil {
		return NewOrm(ctx, dbConfig, connection, nil, nil, log, nil, refresh, drivers, dbConnections), err
	}

	queries := map[string]contractsorm.Query{
		connection: query,
	}

	return NewOrm(ctx, dbConfig, connection, query, queries, log, nil, refresh, drivers, dbConnections), nil
}

// buildQuery builds a query instance for the given connection.
func buildQuery(ctx context.Context, dbConfig *db.DBConfig, connection string, log log.Log, drivers map[string]driver.Driver, dbConnections map[string]*sql.DB, skipPing bool) (contractsorm.Query, error) {
	connConfig, ok := dbConfig.Connections[connection]
	if !ok {
		return nil, fmt.Errorf("connection %s not found in configuration", connection)
	}

	driverName := connConfig.Driver
	if driverName == "" {
		return nil, fmt.Errorf("driver not specified for connection: %s", connection)
	}

	// Get or create driver
	var dbDriver driver.Driver
	if d, ok := drivers[connection]; ok {
		dbDriver = d
	} else {
		dbDriver = createDriver(driverName)
		drivers[connection] = dbDriver
	}

	// Build DSN using the config builder
	builder := db.NewConfigBuilder(connConfig)
	dsn, err := builder.BuildDSN()
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	// Open database connection
	var sqlDB *sql.DB
	if conn, ok := dbConnections[connection]; ok {
		sqlDB = conn
	} else {
		sqlDB, err = dbDriver.Open(dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection: %w", err)
		}

		// Ping to validate the connection and credentials (unless skipped)
		if !skipPing {
			if err := dbDriver.Ping(ctx, sqlDB); err != nil {
				_ = sqlDB.Close()
				return nil, fmt.Errorf("failed to ping database: %w", err)
			}
		}

		// Configure connection pool from DBConfig.
		// SQLite does not support concurrent writers; always pin to a single
		// connection regardless of PoolConfig to prevent "database is locked" errors.
		// WAL mode (applied below) allows concurrent readers alongside the single writer.
		if connConfig.Driver == "sqlite" || connConfig.Driver == "array" {
			sqlDB.SetMaxOpenConns(1)
			sqlDB.SetMaxIdleConns(1)
		} else {
			sqlDB.SetMaxIdleConns(dbConfig.Pool.MaxIdleConns)
			sqlDB.SetMaxOpenConns(dbConfig.Pool.MaxOpenConns)
		}
		sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.Pool.ConnMaxLifetime) * time.Second)
		sqlDB.SetConnMaxIdleTime(time.Duration(dbConfig.Pool.ConnMaxIdleTime) * time.Second)

		// Apply SQLite-specific optimizations. Errors are intentionally ignored —
		// these are performance/safety hints, not requirements for a valid connection.
		if connConfig.Driver == "sqlite" || connConfig.Driver == "array" {
			_, _ = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
			_, _ = sqlDB.ExecContext(ctx, "PRAGMA synchronous=NORMAL;")
			_, _ = sqlDB.ExecContext(ctx, "PRAGMA foreign_keys=ON;")
			_, _ = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout=5000;")
		}

		dbConnections[connection] = sqlDB
	}

	// Open read-replica connection if configured (use first entry, round-robin could be added later)
	var readSQLDB *sql.DB
	if len(connConfig.Read) > 0 {
		replica := connConfig.Read[0]
		replicaCfg := connConfig
		replicaCfg.Host = replica.Host
		replicaCfg.Port = replica.Port
		replicaCfg.Database = replica.Database
		replicaCfg.Username = replica.Username
		replicaCfg.Password = replica.Password
		replicaDSN, err := db.NewConfigBuilder(replicaCfg).BuildDSN()
		if err != nil {
			log.Warningf("[Orm] Failed to build read-replica DSN for connection %s: %v", connection, err)
		} else {
			readSQLDB, err = dbDriver.Open(replicaDSN)
			if err != nil {
				log.Warningf("[Orm] Failed to open read-replica connection for %s: %v", connection, err)
				readSQLDB = nil
			} else if !skipPing {
				// Ping to validate the replica connection
				if err := dbDriver.Ping(ctx, readSQLDB); err != nil {
					log.Warningf("[Orm] Failed to ping read-replica connection for %s: %v", connection, err)
					_ = readSQLDB.Close()
					readSQLDB = nil
				}
			}
		}
	}

	// Open write-primary connection if explicitly configured
	var writeSQLDB *sql.DB
	if len(connConfig.Write) > 0 {
		primary := connConfig.Write[0]
		primaryCfg := connConfig
		primaryCfg.Host = primary.Host
		primaryCfg.Port = primary.Port
		primaryCfg.Database = primary.Database
		primaryCfg.Username = primary.Username
		primaryCfg.Password = primary.Password
		primaryDSN, err := db.NewConfigBuilder(primaryCfg).BuildDSN()
		if err != nil {
			log.Warningf("[Orm] Failed to build write-primary DSN for connection %s: %v", connection, err)
		} else {
			writeSQLDB, err = dbDriver.Open(primaryDSN)
			if err != nil {
				log.Warningf("[Orm] Failed to open write-primary connection for %s: %v", connection, err)
				writeSQLDB = nil
			} else if !skipPing {
				// Ping to validate the primary connection
				if err := dbDriver.Ping(ctx, writeSQLDB); err != nil {
					log.Warningf("[Orm] Failed to ping write-primary connection for %s: %v", connection, err)
					_ = writeSQLDB.Close()
					writeSQLDB = nil
				}
			}
		}
	}

	// Create query instance — use replica routing if replicas are configured
	if readSQLDB != nil || writeSQLDB != nil {
		writeConn := sqlDB
		if writeSQLDB != nil {
			writeConn = writeSQLDB
		}
		readConn := writeConn
		if readSQLDB != nil {
			readConn = readSQLDB
		}
		return query.NewQueryWithReplicas(ctx, writeConn, readConn, dbDriver, connection, dbConfig, log), nil
	}
	return query.NewQuery(ctx, sqlDB, dbDriver, connection, dbConfig, log), nil
}

// BuildOrmFromDB builds an Orm instance from an already-open *sql.DB.
// The caller retains ownership of sqlDB; connection pool settings are not modified.
func BuildOrmFromDB(ctx context.Context, sqlDB *sql.DB, driverName string, connection string, dbConfig *db.DBConfig, log log.Log, refresh func()) (*Orm, error) {
	dbDriver := createDriver(driverName)
	drivers := map[string]driver.Driver{
		connection: dbDriver,
	}
	dbConnections := map[string]*sql.DB{
		connection: sqlDB,
	}

	q := query.NewQuery(ctx, sqlDB, dbDriver, connection, dbConfig, log)
	queries := map[string]contractsorm.Query{
		connection: q,
	}

	return NewOrm(ctx, dbConfig, connection, q, queries, log, nil, refresh, drivers, dbConnections), nil
}

func createDriver(driverName string) driver.Driver {
	switch driverName {
	case "mysql":
		return driver.NewMySQL()
	case "postgres":
		return driver.NewPostgreSQL()
	case "sqlite":
		return driver.NewSQLite()
	case "sqlserver":
		return driver.NewSQLServer()
	case "turso":
		return driver.NewTurso()
	case "oracle":
		return driver.NewOracle()
	case "array":
		return driver.NewArray()
	default:
		return driver.NewMySQL() // Default to MySQL
	}
}

func (r *Orm) Connection(name string) contractsorm.Orm {
	if name == "" {
		name = r.dbConfig.Default
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if instance, exist := r.queries[name]; exist {
		return NewOrm(r.ctx, r.dbConfig, name, instance, r.queries, r.log, r.modelToObserver, r.refresh, r.drivers, r.dbConnections)
	}

	query, err := buildQuery(r.ctx, r.dbConfig, name, r.log, r.drivers, r.dbConnections, false)
	if err != nil || query == nil {
		r.log.Errorf("[Orm] Init %s connection error: %v", name, err)
		return NewOrm(r.ctx, r.dbConfig, name, nil, r.queries, r.log, r.modelToObserver, r.refresh, r.drivers, r.dbConnections)
	}

	r.queries[name] = query

	return NewOrm(r.ctx, r.dbConfig, name, query, r.queries, r.log, r.modelToObserver, r.refresh, r.drivers, r.dbConnections)
}

func (r *Orm) DB() (*sql.DB, error) {
	if r.query == nil {
		return nil, fmt.Errorf("query not initialized")
	}
	return r.query.DB()
}

func (r *Orm) DisableQueryLog() {
	if r.query != nil {
		r.query.DisableQueryLog()
	}
}

func (r *Orm) EnableQueryLog() {
	if r.query != nil {
		r.query.EnableQueryLog()
	}
}

func (r *Orm) FlushQueryLog() {
	if r.query != nil {
		r.query.FlushQueryLog()
	}
}

func (r *Orm) GetQueryLog() []contractsorm.QueryLog {
	if r.query != nil {
		return r.query.GetQueryLog()
	}
	return nil
}

// EnableDebug enables debug mode at runtime for all queries.
func (r *Orm) EnableDebug() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, q := range r.queries {
		if queryWithDebug, ok := q.(interface{ EnableDebug() }); ok {
			queryWithDebug.EnableDebug()
		}
	}
}

// DisableDebug disables debug mode at runtime for all queries.
func (r *Orm) DisableDebug() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.query != nil {
		if queryWithDebug, ok := r.query.(interface{ DisableDebug() }); ok {
			queryWithDebug.DisableDebug()
		}
	}
	for _, q := range r.queries {
		if queryWithDebug, ok := q.(interface{ DisableDebug() }); ok {
			queryWithDebug.DisableDebug()
		}
	}
}

// IsDebug returns true if debug mode is enabled.
func (r *Orm) IsDebug() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.query != nil {
		if queryWithDebug, ok := r.query.(interface{ IsDebug() bool }); ok {
			return queryWithDebug.IsDebug()
		}
	}
	return false
}

func (r *Orm) Factory() contractsorm.Factory {
	return NewFactory(r)
}

func (r *Orm) DatabaseName() string {
	if conn, ok := r.dbConfig.Connections[r.connection]; ok {
		return conn.Database
	}
	return ""
}

func (r *Orm) Name() string {
	return r.connection
}

func (r *Orm) Observe(model any, observer contractsorm.Observer) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.modelToObserver = append(r.modelToObserver, contractsorm.ModelToObserver{
		Model:    model,
		Observer: observer,
	})

	// r.query is also stored in r.queries, so iterating r.queries is sufficient.
	// Calling Observe on r.query separately would register the observer twice.
	for _, q := range r.queries {
		if queryWithObserver, ok := q.(contractsorm.QueryWithObserver); ok {
			queryWithObserver.Observe(model, observer)
		}
	}
}

func (r *Orm) Query() contractsorm.Query {
	if r.query == nil {
		r.log.Errorf("[Orm] Query called but query not initialized for connection: %s", r.connection)
		return &query.Query{}
	}
	if r.ctx != context.Background() {
		if queryWithContext, ok := r.query.(contractsorm.QueryWithContext); ok {
			return queryWithContext.WithContext(r.ctx)
		}
	}
	// Return a fresh clone so each call chain starts with clean state.
	if cloneable, ok := r.query.(interface{ Clone() contractsorm.Query }); ok {
		return cloneable.Clone()
	}
	return r.query
}

func (r *Orm) SetQuery(query contractsorm.Query) {
	r.query = query
}

func (r *Orm) Refresh() {
	if r.refresh != nil {
		r.refresh()
	}
}

// Transaction runs a callback wrapped in a database transaction.
// It automatically commits the transaction if the callback returns nil.
// If the callback returns an error or a panic occurs, the transaction is rolled back.
func (r *Orm) Transaction(txFunc func(tx contractsorm.Query) error, opts ...*sql.TxOptions) error {
	if r.query == nil {
		return fmt.Errorf("query not initialized")
	}
	// Use r.query directly to avoid nil return from Query()
	query := r.query
	if r.ctx != context.Background() {
		if queryWithContext, ok := query.(contractsorm.QueryWithContext); ok {
			query = queryWithContext.WithContext(r.ctx)
		}
	}
	return query.Transaction(txFunc, opts...)
}

func (r *Orm) WithContext(ctx context.Context) contractsorm.Orm {
	return NewOrm(ctx, r.dbConfig, r.connection, r.query, r.queries, r.log, r.modelToObserver, r.refresh, r.drivers, r.dbConnections)
}

// Close closes all database connections.
func (r *Orm) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var lastErr error
	for name, db := range r.dbConnections {
		if drv, ok := r.drivers[name]; ok {
			if arrayDrv, ok := drv.(*driver.Array); ok {
				arrayDrv.Cleanup(db)
			}
		}
		if err := db.Close(); err != nil {
			r.log.Errorf("[Orm] Failed to close connection %s: %v", name, err)
			lastErr = err
		}
	}
	return lastErr
}
