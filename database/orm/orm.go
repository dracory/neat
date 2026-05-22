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

func BuildOrm(ctx context.Context, dbConfig *db.DBConfig, connection string, log log.Log, refresh func()) (*Orm, error) {
	// Initialize drivers and connections
	drivers := make(map[string]driver.Driver)
	dbConnections := make(map[string]*sql.DB)

	// Get default connection if not specified
	if connection == "" {
		connection = dbConfig.Default
	}

	// Build query for the connection
	query, err := buildQuery(ctx, dbConfig, connection, log, drivers, dbConnections)
	if err != nil {
		return NewOrm(ctx, dbConfig, connection, nil, nil, log, nil, refresh, drivers, dbConnections), err
	}

	queries := map[string]contractsorm.Query{
		connection: query,
	}

	return NewOrm(ctx, dbConfig, connection, query, queries, log, nil, refresh, drivers, dbConnections), nil
}

func buildQuery(ctx context.Context, dbConfig *db.DBConfig, connection string, log log.Log, drivers map[string]driver.Driver, dbConnections map[string]*sql.DB) (contractsorm.Query, error) {
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

		// Configure connection pool from DBConfig.
		// In-memory SQLite: each physical connection has its own isolated database,
		// so pin to a single connection so DDL and DML always share the same DB.
		if connConfig.Driver == "sqlite" && (dsn == ":memory:" || dsn == "") {
			sqlDB.SetMaxOpenConns(1)
			sqlDB.SetMaxIdleConns(1)
		} else {
			sqlDB.SetMaxIdleConns(dbConfig.Pool.MaxIdleConns)
			sqlDB.SetMaxOpenConns(dbConfig.Pool.MaxOpenConns)
		}
		sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.Pool.ConnMaxLifetime) * time.Second)
		sqlDB.SetConnMaxIdleTime(time.Duration(dbConfig.Pool.ConnMaxIdleTime) * time.Second)

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
		if err == nil {
			readSQLDB, _ = dbDriver.Open(replicaDSN)
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
		if err == nil {
			writeSQLDB, _ = dbDriver.Open(primaryDSN)
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

	query, err := buildQuery(r.ctx, r.dbConfig, name, r.log, r.drivers, r.dbConnections)
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

func (r *Orm) Factory() contractsorm.Factory {
	// TODO: Implement factory when needed
	return nil
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

	for _, q := range r.queries {
		if queryWithObserver, ok := q.(contractsorm.QueryWithObserver); ok {
			queryWithObserver.Observe(model, observer)
		}
	}

	if r.query != nil {
		if queryWithObserver, ok := r.query.(contractsorm.QueryWithObserver); ok {
			queryWithObserver.Observe(model, observer)
		}
	}
}

func (r *Orm) Query() contractsorm.Query {
	if r.query == nil {
		return nil
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
	return r.Query().Transaction(txFunc, opts...)
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
		if err := db.Close(); err != nil {
			r.log.Errorf("[Orm] Failed to close connection %s: %v", name, err)
			lastErr = err
		}
	}
	return lastErr
}
