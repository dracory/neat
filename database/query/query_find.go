package query

import (
	"fmt"
	"reflect"
	"time"
)

// Find retrieves records matching the given conditions.
func (q *Query) Find(dest any, conds ...any) error {
	q = q.applyScopes()
	// Validate common conditions (build errors, nil DB, empty table)
	if err := q.validate(); err != nil {
		return err
	}
	// Use a clone to avoid mutating the original query state
	clone := q.Clone().(*Query)

	// Add conditions to where clause
	for _, cond := range conds {
		clone.wheres = append(clone.wheres, whereClause{_type: "and", query: fmt.Sprintf("%v", cond), args: nil})
	}

	// Build SELECT query
	builder := NewBuilder(clone)
	sql, args := builder.BuildSelect()

	start := time.Now()
	ctx, cancel := q.timeoutContext()
	defer cancel()
	// Execute query
	if q.tx != nil {
		rows, err := q.tx.QueryContext(ctx, sql, args...)
		if err != nil {
			return q.sanitizeError(fmt.Errorf("failed to execute query: %w", err))
		}
		defer func() { _ = rows.Close() }()
		q.logQuery(sql, args, start)
		return clone.scanRows(rows, dest)
	}

	dbConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(ctx, sql, args...)
	if err != nil {
		return q.sanitizeError(fmt.Errorf("failed to execute query: %w", err))
	}
	defer func() { _ = rows.Close() }()
	q.logQuery(sql, args, start)
	return clone.scanRows(rows, dest)
}

// FindOrFail retrieves records matching the given conditions or returns an error if not found.
func (q *Query) FindOrFail(dest any, conds ...any) error {
	if err := q.Find(dest, conds...); err != nil {
		return err
	}
	// For slice destinations, empty result is a failure
	v := reflect.Indirect(reflect.ValueOf(dest))
	if v.Kind() == reflect.Slice && v.Len() == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}
