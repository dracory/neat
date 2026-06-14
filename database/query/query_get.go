package query

import (
	"fmt"
	"reflect"
	"time"
)

// Get retrieves all records matching the query.
func (q *Query) Get(dest any) error {
	q = q.applyScopes()
	// Check for build errors from query construction
	if q.buildError != nil {
		return q.buildError
	}
	// Use a clone to avoid mutating the original query state
	clone := q.Clone().(*Query)

	// Build SELECT query
	builder := NewBuilder(clone)
	sql, args := builder.BuildSelect()

	start := time.Now()
	ctx, cancel := q.timeoutContext()
	defer cancel()
	if q.tx != nil {
		rows, err := q.tx.QueryContext(ctx, sql, args...)
		if err != nil {
			return q.sanitizeError(fmt.Errorf("failed to execute query: %w", err))
		}
		q.logQuery(sql, args, start)
		if err := q.scanRows(rows, dest); err != nil {
			_ = rows.Close()
			return err
		}
		_ = rows.Close()

		// Load relations after rows are closed to avoid SQLite deadlock
		if len(q.withRelations) > 0 {
			destValue := reflect.ValueOf(dest)
			if destValue.Kind() == reflect.Ptr {
				destValue = destValue.Elem()
			}
			q.initializeRelations(destValue)
			if err := q.loadRelations(destValue); err != nil {
				return err
			}
		}

		return nil
	}

	dbConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(ctx, sql, args...)
	if err != nil {
		return q.sanitizeError(fmt.Errorf("failed to execute query: %w", err))
	}
	q.logQuery(sql, args, start)
	if err := q.scanRows(rows, dest); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()

	// Load relations after rows are closed to avoid SQLite deadlock
	if len(q.withRelations) > 0 {
		destValue := reflect.ValueOf(dest)
		if destValue.Kind() == reflect.Ptr {
			destValue = destValue.Elem()
		}
		q.initializeRelations(destValue)
		if err := q.loadRelations(destValue); err != nil {
			return err
		}
	}

	return nil
}

// All is an alias for Get, providing Django-style syntax.
// Retrieves all records matching the query.
func (q *Query) All(dest any) error {
	return q.Get(dest)
}

// FindAll is an alias for All, providing Sequelize-style syntax.
// Retrieves all records matching the query.
func (q *Query) FindAll(dest any) error {
	return q.All(dest)
}
