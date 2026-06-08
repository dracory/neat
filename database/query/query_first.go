package query

import (
	"fmt"
	"reflect"
	"time"
)

// First retrieves the first record matching the query.
func (q *Query) First(dest any) error {
	q = q.applyScopes()
	// Validate common conditions (build errors, nil DB, empty table)
	if err := q.validate(); err != nil {
		return err
	}
	// Use a clone to avoid mutating the original query state
	clone := q.Clone().(*Query)

	// Set limit to 1
	limit := 1
	clone.limit = &limit

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

// FirstOr retrieves the first record or executes a callback if not found.
func (q *Query) FirstOr(dest any, callback func() error) error {
	err := q.First(dest)
	if err != nil {
		return callback()
	}
	return nil
}

// FirstOrCreate retrieves the first record or creates it if not found.
func (q *Query) FirstOrCreate(dest any, conds ...any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, create it
	return q.Create(dest)
}

// FirstOrNew retrieves the first record or prepares a new instance if not found.
func (q *Query) FirstOrNew(dest any, attributes any, values ...any) error {
	// Clone the query to avoid modifying the original
	query := q.Clone().(*Query)

	// Apply attributes as WHERE conditions
	if attributes != nil {
		if err := applyWhereConditions(query, attributes); err != nil {
			return err
		}
	}

	// Try to find the record first
	err := query.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, prepare new instance (without saving)
	// Apply attributes to the destination
	if attributes != nil {
		if err := applyAttributes(dest, attributes); err != nil {
			return err
		}
	}

	// Apply values if provided
	if len(values) > 0 && values[0] != nil {
		if err := applyAttributes(dest, values[0]); err != nil {
			return err
		}
	}

	return nil
}
