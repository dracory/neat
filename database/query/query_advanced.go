package query

import (
	"fmt"
	"reflect"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/observer"
)

// Increment increments a column's value by a specified amount.
func (q *Query) Increment(column string, amount ...any) (*contractsorm.Result, error) {
	if err := q.validateAggregate(column, "*"); err != nil {
		return nil, err
	}

	incAmount := int64(1)
	if len(amount) > 0 {
		switch v := amount[0].(type) {
		case int64:
			incAmount = v
		case int:
			incAmount = int64(v)
		case int32:
			incAmount = int64(v)
		}
	}

	// Use ? placeholder - the builder will replace it with dialect-specific placeholder
	updateQuery := fmt.Sprintf("%s = %s + ?", column, column)
	return q.Update(updateQuery, incAmount)
}

// Decrement decrements a column's value by a specified amount.
func (q *Query) Decrement(column string, amount ...any) (*contractsorm.Result, error) {
	if err := q.validateAggregate(column, "*"); err != nil {
		return nil, err
	}

	decAmount := int64(1)
	if len(amount) > 0 {
		switch v := amount[0].(type) {
		case int64:
			decAmount = v
		case int:
			decAmount = int64(v)
		case int32:
			decAmount = int64(v)
		}
	}

	// Use ? placeholder - the builder will replace it with dialect-specific placeholder
	updateQuery := fmt.Sprintf("%s = %s - ?", column, column)
	return q.Update(updateQuery, decAmount)
}

// InRandomOrder orders the query results randomly.
func (q *Query) InRandomOrder() contractsorm.Query {
	var order string
	if q.driver != nil && q.driver.Dialect() == "mysql" {
		order = "RAND()"
	} else {
		order = "RANDOM()"
	}
	newQ := *q
	newQ.orders = append(newQ.orders, orderClause{column: order, direction: ""})
	return &newQ
}

// LockForUpdate locks the selected rows for update.
func (q *Query) LockForUpdate() contractsorm.Query {
	newQ := q.Clone().(*Query)
	newQ.lockForUpdate = true
	return newQ
}

// SharedLock locks the selected rows with a shared lock.
func (q *Query) SharedLock() contractsorm.Query {
	newQ := q.Clone().(*Query)
	newQ.sharedLock = true
	return newQ
}

// Raw sets a raw SQL query to be executed.
func (q *Query) Raw(sql string, values ...any) contractsorm.Query {
	if q == nil {
		return &Query{}
	}
	// Store raw SQL for later use
	newQ := *q
	newQ.rawSQL = sql
	newQ.rawArgs = values
	return &newQ
}

// Exec executes a raw SQL query.
func (q *Query) Exec(sql string, values ...any) (*contractsorm.Result, error) {
	// Execute raw SQL
	var err error
	var result interface{ RowsAffected() (int64, error) }
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, values...)
	} else {
		databaseConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		result, err = databaseConn.ExecContext(q.ctx, sql, values...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute raw SQL: %w", err)
	}

	if result == nil {
		return &contractsorm.Result{
			RowsAffected: 0,
		}, nil
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &contractsorm.Result{
			RowsAffected: 0,
		}, nil
	}
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

// Omit specifies columns to omit from INSERT/UPDATE operations.
func (q *Query) Omit(columns ...string) contractsorm.Query {
	newQ := *q
	newQ.omitColumns = append(newQ.omitColumns, columns...)
	return &newQ
}

// Restore restores a soft-deleted record.
func (q *Query) Restore(model ...any) (*contractsorm.Result, error) {
	// Fire Restoring event if not disabled
	if !q.withoutEvents && len(model) > 0 {
		attributes := observer.ExtractModelAttributes(model[0])
		if err := q.dispatcher.DispatchRestoring(q.ctx, model[0], q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("restoring event error: %w", err)
		}
	}

	// Build UPDATE query to set deleted_at to NULL
	// Clone the query to preserve WHERE clauses
	clone := q.Clone().(*Query)
	clone.withTrashed = true

	// If a model instance is provided, extract its ID and add WHERE clause
	if len(model) > 0 && model[0] != nil {
		v := reflect.ValueOf(model[0])
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			// Try to get ID field
			idField := v.FieldByName("ID")
			if idField.IsValid() {
				idValue := idField.Uint()
				if idValue > 0 {
					clone.wheres = append(clone.wheres, whereClause{_type: "and", query: "id = ?", args: []any{idValue}})
				}
			}
		}
	}

	builder := NewBuilder(clone)
	sql, args := builder.BuildUpdate(map[string]any{"deleted_at": nil})
	if sql == "" {
		return nil, fmt.Errorf("failed to build RESTORE query")
	}

	// Execute query
	var err error
	var result interface{ RowsAffected() (int64, error) }
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, args...)
	} else {
		dbConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, sql, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute RESTORE query: %w", err)
	}
	q.logQuery(sql, args, start)

	// Fire Restored event if not disabled
	if !q.withoutEvents && len(model) > 0 {
		attributes := observer.ExtractModelAttributes(model[0])
		if err := q.dispatcher.DispatchRestored(q.ctx, model[0], q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("restored event error: %w", err)
		}
	}

	// Get affected rows
	var rowsAffected int64
	if result != nil {
		rowsAffected, _ = result.RowsAffected()
	}
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

// ForceDelete permanently deletes a record (bypasses soft delete).
func (q *Query) ForceDelete(value ...any) (*contractsorm.Result, error) {
	// Fire ForceDeleting event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchForceDeleting(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("force_deleting event error: %w", err)
		}
	}

	// Build DELETE query (permanent delete, not soft delete)
	// Use WithTrashed to include soft-deleted records
	clone := q.Clone().(*Query)
	clone.withTrashed = true
	builder := NewBuilder(clone)
	sql, args := builder.BuildDelete()
	if sql == "" {
		return nil, fmt.Errorf("failed to build FORCE DELETE query")
	}

	// Execute query
	var err error
	var result interface{ RowsAffected() (int64, error) }
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sql, args...)
	} else {
		dbConn, err := q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, sql, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute FORCE DELETE query: %w", err)
	}
	q.logQuery(sql, args, start)

	// Fire ForceDeleted event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchForceDeleted(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("force_deleted event error: %w", err)
		}
	}

	// Get affected rows
	var rowsAffected int64
	if result != nil {
		rowsAffected, _ = result.RowsAffected()
	}
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}
