package orm

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/dracory/neat/contracts/config"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/query"
)

type Orm struct {
	ctx             context.Context
	config          config.Config
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
	config config.Config,
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
		config:          config,
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

func BuildOrm(ctx context.Context, config config.Config, connection string, log log.Log, refresh func()) (*Orm, error) {
	// Initialize drivers and connections
	drivers := make(map[string]driver.Driver)
	dbConnections := make(map[string]*sql.DB)

	// Get default connection if not specified
	if connection == "" {
		connection = config.GetString("database.default")
	}

	// Build query for the connection
	query, err := buildQuery(ctx, config, connection, log, drivers, dbConnections)
	if err != nil {
		return NewOrm(ctx, config, connection, nil, nil, log, nil, refresh, drivers, dbConnections), err
	}

	queries := map[string]contractsorm.Query{
		connection: query,
	}

	return NewOrm(ctx, config, connection, query, queries, log, nil, refresh, drivers, dbConnections), nil
}

func buildQuery(ctx context.Context, config config.Config, connection string, log log.Log, drivers map[string]driver.Driver, dbConnections map[string]*sql.DB) (contractsorm.Query, error) {
	driverName := config.GetString(fmt.Sprintf("database.connections.%s.driver", connection))
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

	// Build DSN
	var err error
	dsn := config.GetString(fmt.Sprintf("database.connections.%s.dsn", connection))
	if dsn == "" {
		// Build DSN from config
		host := config.GetString(fmt.Sprintf("database.connections.%s.host", connection))
		port := config.GetInt(fmt.Sprintf("database.connections.%s.port", connection))
		database := config.GetString(fmt.Sprintf("database.connections.%s.database", connection))
		username := config.GetString(fmt.Sprintf("database.connections.%s.username", connection))
		password := config.GetString(fmt.Sprintf("database.connections.%s.password", connection))
		charset := config.GetString(fmt.Sprintf("database.connections.%s.charset", connection))
		schema := config.GetString(fmt.Sprintf("database.connections.%s.schema", connection))
		sslmode := config.GetString(fmt.Sprintf("database.connections.%s.sslmode", connection))
		loc := config.GetString(fmt.Sprintf("database.connections.%s.loc", connection))
		timezone := config.GetString(fmt.Sprintf("database.connections.%s.timezone", connection))

		connConfig := db.ConnectionConfig{
			Driver:   driverName,
			Host:     host,
			Port:     port,
			Database: database,
			Username: username,
			Password: password,
			Charset:  charset,
			Schema:   schema,
			SSLMode:  sslmode,
			Loc:      loc,
			Timezone: timezone,
		}

		builder := db.NewConfigBuilder(connConfig)
		dsn, err = builder.BuildDSN()
		if err != nil {
			return nil, err
		}
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

		// Configure connection pool
		maxIdleConns := config.GetInt("database.pool.max_idle_conns", 10)
		maxOpenConns := config.GetInt("database.pool.max_open_conns", 100)
		connMaxLifetime := config.GetInt("database.pool.conn_max_lifetime", 3600)
		connMaxIdleTime := config.GetInt("database.pool.conn_max_idletime", 3600)

		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
		sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Second)

		dbConnections[connection] = sqlDB
	}

	// Create query instance
	return query.NewQuery(ctx, sqlDB, dbDriver, connection, config, log), nil
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
		name = r.config.GetString("database.default")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if instance, exist := r.queries[name]; exist {
		return NewOrm(r.ctx, r.config, name, instance, r.queries, r.log, r.modelToObserver, r.refresh, r.drivers, r.dbConnections)
	}

	query, err := buildQuery(r.ctx, r.config, name, r.log, r.drivers, r.dbConnections)
	if err != nil || query == nil {
		r.log.Errorf("[Orm] Init %s connection error: %v", name, err)
		return NewOrm(r.ctx, r.config, name, nil, r.queries, r.log, r.modelToObserver, r.refresh, r.drivers, r.dbConnections)
	}

	r.queries[name] = query

	return NewOrm(r.ctx, r.config, name, query, r.queries, r.log, r.modelToObserver, r.refresh, r.drivers, r.dbConnections)
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
	return r.config.GetString(fmt.Sprintf("database.connections.%s.database", r.connection))
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
	if r.ctx != context.Background() && r.query != nil {
		if queryWithContext, ok := r.query.(contractsorm.QueryWithContext); ok {
			return queryWithContext.WithContext(r.ctx)
		}
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
	return NewOrm(ctx, r.config, r.connection, r.query, r.queries, r.log, r.modelToObserver, r.refresh, r.drivers, r.dbConnections)
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
