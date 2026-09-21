package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/contracts/database/orm"
	neatErrors "github.com/dracory/neat/errors"
)

// Table sets the table for the query.
func (q *Query) Table(name string, args ...any) orm.Query {
	// Validate table name is a table identifier (unless it's a subquery or has an alias).
	// Reserved SQL keywords are allowed: the builder always quotes the table name.
	// Invalid names set buildError instead of silently keeping a stale table.
	if !strings.Contains(name, "(") {
		parts := strings.Fields(name)
		valid := false
		// Allow "table", "table alias", or "table AS alias"
		switch len(parts) {
		case 1:
			valid = isTableIdentifier(parts[0])
		case 2:
			valid = isTableIdentifier(parts[0]) && isTableIdentifier(parts[1])
		case 3:
			valid = strings.ToUpper(parts[1]) == "AS" && isTableIdentifier(parts[0]) && isTableIdentifier(parts[2])
		}
		if !valid {
			if q.buildError == nil {
				q.buildError = fmt.Errorf("invalid table name: %q", name)
			}
			return q
		}
	}
	q.table = name
	q.tableArgs = nil
	// If it's a subquery callback
	if strings.Contains(name, "(") && strings.Contains(name, ")") && len(args) > 0 {
		for _, arg := range args {
			if fn, ok := arg.(func(orm.Query) orm.Query); ok {
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
	conn := q.writeConn()
	if conn == nil {
		return nil, neatErrors.ErrNilDatabase
	}
	return conn, nil
}

// ReadDB returns the read-replica connection (falls back to primary if none configured).
func (q *Query) ReadDB() (*sql.DB, error) {
	if q.tx != nil {
		return nil, fmt.Errorf("cannot get ReadDB during transaction")
	}
	conn := q.readConn()
	if conn == nil {
		return nil, neatErrors.ErrNilDatabase
	}
	return conn, nil
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
	if q.queryLog != nil {
		*q.queryLog = make([]orm.QueryLog, 0)
	}
}

// GetQueryLog returns the query log.
func (q *Query) GetQueryLog() []orm.QueryLog {
	if q.queryLog == nil {
		return nil
	}
	return *q.queryLog
}

// WithContext returns a new Query instance with the specified context.
func (q *Query) WithContext(ctx context.Context) orm.Query {
	newQuery := q.Clone().(*Query)
	newQuery.ctx = ctx
	return newQuery
}

// Observe registers an observer for the given model.
func (q *Query) Observe(model any, observer orm.Observer) {
	q.modelToObserver = append(q.modelToObserver, orm.ModelToObserver{
		Model:    model,
		Observer: observer,
	})
}

// WithoutEvents disables event firing for the query.
func (q *Query) WithoutEvents() orm.Query {
	newQuery := q.Clone().(*Query)
	newQuery.withoutEvents = true
	return newQuery
}
