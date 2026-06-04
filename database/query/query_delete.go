package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/observer"
)

// hasSoftDeleteCapability checks if the model has soft delete capability.
func hasSoftDeleteCapability(model any) bool {
	if model == nil {
		return false
	}

	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false
	}

	// Check for DeletedAt field (including embedded fields)
	deletedAtField := val.FieldByName("DeletedAt")
	if deletedAtField.IsValid() && deletedAtField.Type() == reflect.TypeOf(&time.Time{}) {
		return true
	}

	// Check embedded structs for DeletedAt
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embeddedVal := val.Field(i)
			embeddedDeletedAt := embeddedVal.FieldByName("DeletedAt")
			if embeddedDeletedAt.IsValid() && embeddedDeletedAt.Type() == reflect.TypeOf(&time.Time{}) {
				return true
			}
		}
	}

	return false
}

// Delete deletes records from the database.
func (q *Query) Delete(value ...any) (*contractsorm.Result, error) {
	if q.buildError != nil {
		return nil, q.buildError
	}
	// Fire Deleting event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchDeleting(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("deleting event error: %w", err)
		}
	}

	// Check if model has soft delete capability
	useSoftDelete := hasSoftDeleteCapability(q.model)

	var deleteSQL string
	var args []any
	var err error

	if useSoftDelete && !q.withTrashed && !q.onlyTrashed {
		// Use UPDATE to set deleted_at instead of DELETE
		// Clone the query to preserve WHERE clauses
		clone := q.Clone().(*Query)
		clone.withTrashed = true
		builder := NewBuilder(clone)
		now := time.Now()
		deleteSQL, args = builder.BuildUpdate(map[string]any{"deleted_at": now})
		if deleteSQL == "" {
			return nil, fmt.Errorf("failed to build SOFT DELETE query")
		}
		// Log the soft delete SQL for debugging
		q.logQuery(deleteSQL, args, time.Now())
	} else {
		// Build DELETE query
		builder := NewBuilder(q)
		deleteSQL, args = builder.BuildDelete()
		if deleteSQL == "" {
			return nil, fmt.Errorf("failed to build DELETE query")
		}
	}

	// Execute query
	ctx, cancel := q.timeoutContext()
	defer cancel()
	var result interface{ RowsAffected() (int64, error) }
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(ctx, deleteSQL, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(ctx, deleteSQL, args...)
	}

	if err != nil {
		return nil, sanitizeError(fmt.Errorf("failed to execute DELETE query: %w", err), q.isProduction())
	}
	q.logQuery(deleteSQL, args, start)

	// Fire Deleted event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchDeleted(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("deleted event error: %w", err)
		}
	}

	// Get affected rows
	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}
