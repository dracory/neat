package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dracory/neat/contracts/database"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Table sets the table for the query.
func (q *Query) Table(name string, args ...any) contractsorm.Query {
	q.table = name
	q.tableArgs = nil
	// If it's a subquery callback
	if strings.Contains(name, "(") && strings.Contains(name, ")") && len(args) > 0 {
		for _, arg := range args {
			if fn, ok := arg.(func(contractsorm.Query) contractsorm.Query); ok {
				subQuery := fn(q.newQuery())
				builder := NewBuilder(subQuery.(*Query))
				subSQL, subArgs := builder.BuildSelect()
				q.table = strings.Replace(q.table, "?", fmt.Sprintf("(%s)", subSQL), 1)
				q.tableArgs = append(q.tableArgs, subArgs...)
			} else {
				q.tableArgs = append(q.tableArgs, arg)
			}
		}
	} else {
		q.tableArgs = args
	}
	return q
}

// DB returns the write (primary) database connection.
func (q *Query) DB() (*sql.DB, error) {
	if q.tx != nil {
		return nil, fmt.Errorf("cannot get DB during transaction, use transaction methods instead")
	}
	return q.writeConn(), nil
}

// ReadDB returns the read-replica connection (falls back to primary if none configured).
func (q *Query) ReadDB() (*sql.DB, error) {
	if q.tx != nil {
		return nil, fmt.Errorf("cannot get ReadDB during transaction")
	}
	return q.readConn(), nil
}

// InTransaction returns true if the query is in a transaction.
func (q *Query) InTransaction() bool {
	return q.inTransaction
}

// Driver returns the database driver.
func (q *Query) Driver() database.Driver {
	return database.Driver(q.driver.Dialect())
}

// EnableQueryLog enables query logging.
func (q *Query) EnableQueryLog() {
	q.enableLog = true
}

// DisableQueryLog disables query logging.
func (q *Query) DisableQueryLog() {
	q.enableLog = false
}

// FlushQueryLog clears the query log.
func (q *Query) FlushQueryLog() {
	q.queryLog = make([]contractsorm.QueryLog, 0)
}

// GetQueryLog returns the query log.
func (q *Query) GetQueryLog() []contractsorm.QueryLog {
	return q.queryLog
}

// WithContext returns a new Query instance with the specified context.
func (q *Query) WithContext(ctx context.Context) contractsorm.Query {
	newQuery := *q
	newQuery.ctx = ctx
	return &newQuery
}

// Observe registers an observer for the given model.
func (q *Query) Observe(model any, observer contractsorm.Observer) {
	q.modelToObserver = append(q.modelToObserver, contractsorm.ModelToObserver{
		Model:    model,
		Observer: observer,
	})
}

// WithoutEvents disables event firing for the query.
func (q *Query) WithoutEvents() contractsorm.Query {
	newQuery := *q
	newQuery.withoutEvents = true
	return &newQuery
}
