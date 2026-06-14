package query

import (
	"database/sql"
	"fmt"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/observer"
)

// hasSoftDeleteCapability checks if the model implements SoftDeleteColumnNamer,
// which is the interface used to detect soft delete support. Models that embed
// SoftDeletes or SoftDeletedAt satisfy this interface automatically via promoted methods.
func hasSoftDeleteCapability(model any) bool {
	if model == nil {
		return false
	}
	_, ok := model.(contractsorm.SoftDeleteColumnNamer)
	return ok
}

// getSoftDeleteColumn returns the soft delete column name for the given model.
// Falls back to "soft_deleted_at" if the model does not implement SoftDeleteColumnNamer.
func getSoftDeleteColumn(model any) string {
	if namer, ok := model.(contractsorm.SoftDeleteColumnNamer); ok {
		return namer.SoftDeletedAtColumn()
	}
	return "soft_deleted_at"
}

// Delete deletes records from the database.
func (q *Query) Delete(value ...any) (*contractsorm.Result, error) {
	// Validate common conditions (build errors, nil DB, empty table)
	if err := q.validate(); err != nil {
		return nil, err
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

	if useSoftDelete && !q.includeSoftDeleted && !q.onlySoftDeleted {
		// Use UPDATE to set the soft delete column instead of DELETE
		// Clone the query to preserve WHERE clauses
		clone := q.Clone().(*Query)
		clone.includeSoftDeleted = true
		builder := NewBuilder(clone)
		now := time.Now()
		col := getSoftDeleteColumn(q.model)
		deleteSQL, args = builder.BuildUpdate(map[string]any{col: now})
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
		return nil, q.sanitizeError(fmt.Errorf("failed to execute DELETE query: %w", err))
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
