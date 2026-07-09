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
	// Work on a clone to avoid mutating the original query and to apply
	// any variadic value arguments as additional WHERE conditions.
	query := q.Clone().(*Query)
	if len(value) > 0 {
		applyConditions(query, value)
	}

	// Validate common conditions (build errors, nil DB, empty table)
	if err := query.validate(); err != nil {
		return nil, err
	}
	// Fire Deleting event if not disabled
	if !query.withoutEvents && query.model != nil {
		attributes := observer.ExtractModelAttributes(query.model)
		if err := query.dispatcher.DispatchDeleting(query.ctx, query.model, query.modelToObserver, nil, attributes, nil, query); err != nil {
			return nil, fmt.Errorf("deleting event error: %w", err)
		}
	}

	// Check if model has soft delete capability
	useSoftDelete := hasSoftDeleteCapability(query.model)

	var deleteSQL string
	var args []any
	var err error

	if useSoftDelete && !query.includeSoftDeleted && !query.onlySoftDeleted {
		// Use UPDATE to set the soft delete column instead of DELETE
		query.includeSoftDeleted = true
		builder := NewBuilder(query)
		col := getSoftDeleteColumn(query.model)
		// Check if model implements SoftDeleteStrategy for custom delete value
		var deleteValue any = time.Now()
		if strat, ok := query.model.(contractsorm.SoftDeleteStrategy); ok {
			deleteValue = strat.SoftDeleteValue()
		}
		deleteSQL, args = builder.BuildUpdate(map[string]any{col: deleteValue})
		if deleteSQL == "" {
			return nil, fmt.Errorf("failed to build SOFT DELETE query")
		}
		// Log the soft delete SQL for debugging
		query.logQuery(deleteSQL, args, time.Now())
	} else {
		// Build DELETE query
		builder := NewBuilder(query)
		deleteSQL, args = builder.BuildDelete()
		if deleteSQL == "" {
			return nil, fmt.Errorf("failed to build DELETE query")
		}
	}

	// Execute query
	ctx, cancel := query.timeoutContext()
	defer cancel()
	var result interface{ RowsAffected() (int64, error) }
	start := time.Now()
	if query.tx != nil {
		result, err = query.tx.ExecContext(ctx, deleteSQL, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = query.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(ctx, deleteSQL, args...)
	}

	if err != nil {
		return nil, query.sanitizeError(fmt.Errorf("failed to execute DELETE query: %w", err))
	}
	query.logQuery(deleteSQL, args, start)

	// Fire Deleted event if not disabled
	if !query.withoutEvents && query.model != nil {
		attributes := observer.ExtractModelAttributes(query.model)
		if err := query.dispatcher.DispatchDeleted(query.ctx, query.model, query.modelToObserver, nil, attributes, nil, query); err != nil {
			return nil, fmt.Errorf("deleted event error: %w", err)
		}
	}

	// Get affected rows
	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

// Destroy is an alias for Delete, providing Sequelize-style syntax.
// Deletes records from the database.
func (q *Query) Destroy(value ...any) (*contractsorm.Result, error) {
	return q.Delete(value...)
}
