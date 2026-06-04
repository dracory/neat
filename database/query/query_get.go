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
			return sanitizeError(fmt.Errorf("failed to execute query: %w", err), q.isProduction())
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
		return sanitizeError(fmt.Errorf("failed to execute query: %w", err), q.isProduction())
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
