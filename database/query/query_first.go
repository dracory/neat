package query

import (
	"fmt"
	"reflect"
	"time"
)

// First retrieves the first record matching the query.
func (q *Query) First(dest any) error {
	q = q.applyScopes()
	// Use a clone to avoid mutating the original query state
	clone := q.Clone().(*Query)

	// Set limit to 1
	limit := 1
	clone.limit = &limit

	// Build SELECT query
	builder := NewBuilder(clone)
	sql, args := builder.BuildSelect()

	start := time.Now()
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		q.logQuery(sql, args, start)
		if err := q.scanRows(rows, dest); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

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

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	q.logQuery(sql, args, start)
	if err := q.scanRows(rows, dest); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

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

// FirstOrFail retrieves the first record or returns an error if not found.
func (q *Query) FirstOrFail(dest any) error {
	if err := q.First(dest); err != nil {
		return err
	}
	// If the dest is a struct and still zero, nothing was found
	v := reflect.Indirect(reflect.ValueOf(dest))
	if v.Kind() == reflect.Struct && v.IsZero() {
		return fmt.Errorf("record not found")
	}
	return nil
}
