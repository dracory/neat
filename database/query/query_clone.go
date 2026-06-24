package query

import (
	"sync"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
)

// Clone returns a new Query with shared connection state but empty query-builder state.
func (q *Query) Clone() contractsorm.Query {
	clone := q.newQuery()
	clone.table = q.table
	clone.tableArgs = append([]any{}, q.tableArgs...)
	clone.model = q.model
	clone.omitColumns = append([]string{}, q.omitColumns...)
	clone.distinct = q.distinct
	clone.distinctCols = append([]string{}, q.distinctCols...)

	clone.wheres = make([]whereClause, len(q.wheres))
	for i, w := range q.wheres {
		clone.wheres[i] = w
		clone.wheres[i].args = append([]any{}, w.args...)
	}

	clone.selects = make([]selectClause, len(q.selects))
	for i, s := range q.selects {
		clone.selects[i] = s
		clone.selects[i].args = append([]any{}, s.args...)
	}

	clone.joins = make([]joinClause, len(q.joins))
	for i, j := range q.joins {
		clone.joins[i] = j
		clone.joins[i].args = append([]any{}, j.args...)
	}

	clone.havings = make([]havingClause, len(q.havings))
	for i, h := range q.havings {
		clone.havings[i] = h
		clone.havings[i].args = append([]any{}, h.args...)
	}

	clone.groups = append([]string{}, q.groups...)
	clone.orders = append([]orderClause{}, q.orders...)

	if q.limit != nil {
		limit := *q.limit
		clone.limit = &limit
	}
	if q.offset != nil {
		offset := *q.offset
		clone.offset = &offset
	}

	clone.includeSoftDeleted = q.includeSoftDeleted
	clone.onlySoftDeleted = q.onlySoftDeleted
	clone.excludeSoftDeleted = q.excludeSoftDeleted
	clone.queryLog = q.queryLog
	clone.rawSQL = q.rawSQL
	clone.rawArgs = append([]any{}, q.rawArgs...)
	clone.lockForUpdate = q.lockForUpdate
	clone.sharedLock = q.sharedLock
	clone.scopes = append([]func(contractsorm.Query) contractsorm.Query{}, q.scopes...)
	clone.withRelations = append([]string{}, q.withRelations...)
	if q.relationConstraints != nil {
		clone.relationConstraints = make(map[string]func(contractsorm.Query) contractsorm.Query)
		for k, v := range q.relationConstraints {
			clone.relationConstraints[k] = v
		}
	}
	clone.withCountQueries = append([]countQuery{}, q.withCountQueries...)
	clone.withExistsQueries = append([]existsQuery{}, q.withExistsQueries...)
	clone.buildError = q.buildError

	// Per-query population cache
	if q.populatedTables != nil {
		clone.populatedTables = make(map[string]bool)
		for k, v := range q.populatedTables {
			clone.populatedTables[k] = v
		}
	}

	// Transaction state
	clone.inTransaction = q.inTransaction
	clone.tx = q.tx
	clone.savepointLevel = q.savepointLevel
	clone.savepointName = q.savepointName

	// Observer state
	clone.modelToObserver = append([]contractsorm.ModelToObserver{}, q.modelToObserver...)
	clone.withoutEvents = q.withoutEvents
	clone.dispatcher = q.dispatcher

	// Transaction lifecycle hooks
	clone.beforeCommit = append([]func() error{}, q.beforeCommit...)
	clone.afterCommit = append([]func() error{}, q.afterCommit...)
	clone.beforeRollback = append([]func() error{}, q.beforeRollback...)
	clone.afterRollback = append([]func() error{}, q.afterRollback...)

	// Debug state
	clone.debugState = q.debugState
	clone.debugMu = sync.RWMutex{}

	return clone
}

// Connection returns a new Query instance scoped to the named connection.
func (q *Query) Connection(name string) contractsorm.Query {
	if name == "" || q.dbConfig == nil {
		return q
	}
	connCfg, ok := q.dbConfig.Connections[name]
	if !ok {
		return q
	}
	drv := newDriverForDialect(connCfg.Driver)
	dsn, err := db.NewConfigBuilder(connCfg).BuildDSN()
	if err != nil {
		if q.log != nil {
			q.log.Errorf("failed to build DSN for connection '%s': %v", name, err)
		}
		return q
	}
	sqlDB, err := drv.Open(dsn)
	if err != nil {
		if q.log != nil {
			q.log.Errorf("failed to open connection '%s': %v", name, err)
		}
		return q
	}
	return NewQuery(q.ctx, sqlDB, drv, name, q.dbConfig, q.log)
}

// newDriverForDialect returns a Driver for the given dialect name.
func newDriverForDialect(dialect string) driver.Driver {
	switch dialect {
	case "mysql":
		return driver.NewMySQL()
	case "postgres":
		return driver.NewPostgreSQL()
	case "sqlserver":
		return driver.NewSQLServer()
	case "turso":
		return driver.NewTurso()
	case "oracle":
		return driver.NewOracle()
	case "array":
		return driver.NewArray()
	default:
		return driver.NewSQLite()
	}
}
