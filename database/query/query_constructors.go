package query

import (
	"context"
	"database/sql"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/observer"
)

// NewQuery creates a new Query instance.
func NewQuery(ctx context.Context, db *sql.DB, drv driver.Driver, connection string, dbConfig *db.DBConfig, log log.Log) *Query {
	debugState := false
	if dbConfig != nil {
		debugState = dbConfig.Debug
	}

	return &Query{
		ctx:             ctx,
		db:              db,
		readDB:          nil,
		writeDB:         nil,
		driver:          drv,
		connection:      connection,
		dbConfig:        dbConfig,
		log:             log,
		enableLog:       false,
		queryLog:        &[]contractsorm.QueryLog{},
		modelToObserver: make([]contractsorm.ModelToObserver, 0),
		withoutEvents:   false,
		dispatcher:      observer.NewDispatcher(log),
		debugState:      debugState,
	}
}

// NewQueryWithReplicas creates a Query with separate read and write sql.DB connections.
func NewQueryWithReplicas(ctx context.Context, writeConn, readConn *sql.DB, drv driver.Driver, connection string, dbConfig *db.DBConfig, lg log.Log) *Query {
	q := NewQuery(ctx, writeConn, drv, connection, dbConfig, lg)
	q.readDB = readConn
	q.writeDB = writeConn
	return q
}

// readConn returns the connection to use for read (SELECT) queries.
func (q *Query) readConn() *sql.DB {
	if q.readDB != nil {
		return q.readDB
	}
	return q.db
}

// writeConn returns the connection to use for write (INSERT/UPDATE/DELETE) queries.
func (q *Query) writeConn() *sql.DB {
	if q.writeDB != nil {
		return q.writeDB
	}
	return q.db
}

// isPostgres returns true if the driver dialect is PostgreSQL.
func (q *Query) isPostgres() bool {
	return q.driver != nil && q.driver.Dialect() == "postgres"
}

// isSQLServer returns true if the driver dialect is SQL Server.
func (q *Query) isSQLServer() bool {
	return q.driver != nil && q.driver.Dialect() == "sqlserver"
}

// isMySQL returns true if the driver dialect is MySQL.
func (q *Query) isMySQL() bool {
	return q.driver != nil && q.driver.Dialect() == "mysql"
}

// isSQLite returns true if the driver dialect is SQLite.
func (q *Query) isSQLite() bool {
	return q.driver != nil && (q.driver.Dialect() == "sqlite" || q.driver.Dialect() == "array")
}

// isOracle returns true if the driver dialect is Oracle.
func (q *Query) isOracle() bool {
	return q.driver != nil && q.driver.Dialect() == "oracle"
}

// newQuery creates a new Query instance with shared connection state.
func (q *Query) newQuery() *Query {
	newQ := NewQuery(q.ctx, q.db, q.driver, q.connection, q.dbConfig, q.log)
	newQ.readDB = q.readDB
	newQ.writeDB = q.writeDB
	newQ.enableLog = q.enableLog
	newQ.queryLog = q.queryLog
	newQ.withRelations = nil
	newQ.relationConstraints = nil
	// Note: buildError is intentionally NOT copied to newQuery()
	// newQuery() creates a fresh query without inheriting build errors
	// Use Clone() to preserve buildError across query copies
	return newQ
}
