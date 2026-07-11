package query

import (
	"database/sql"
	"fmt"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/observer"
)

// WithSoftDeleted includes soft-deleted records in the query results.
func (q *Query) WithSoftDeleted() contractsorm.Query {
	newQuery := q.Clone().(*Query)
	newQuery.includeSoftDeleted = true
	newQuery.onlySoftDeleted = false
	newQuery.excludeSoftDeleted = false
	return newQuery
}

// WithTrashed includes soft-deleted records in the query results.
//
// Deprecated: Use WithSoftDeleted() instead.
func (q *Query) WithTrashed() contractsorm.Query {
	return q.WithSoftDeleted()
}

// OnlySoftDeleted returns only soft-deleted records.
func (q *Query) OnlySoftDeleted() contractsorm.Query {
	newQuery := q.Clone().(*Query)
	newQuery.includeSoftDeleted = false
	newQuery.onlySoftDeleted = true
	newQuery.excludeSoftDeleted = false
	return newQuery
}

// OnlyTrashed returns only soft-deleted records.
//
// Deprecated: Use OnlySoftDeleted() instead.
func (q *Query) OnlyTrashed() contractsorm.Query {
	return q.OnlySoftDeleted()
}

// WithoutSoftDeleted excludes soft-deleted records from the query results (default behavior).
func (q *Query) WithoutSoftDeleted() contractsorm.Query {
	newQuery := q.Clone().(*Query)
	newQuery.includeSoftDeleted = false
	newQuery.onlySoftDeleted = false
	newQuery.excludeSoftDeleted = true
	return newQuery
}

// WithoutTrashed excludes soft-deleted records from the query results (default behavior).
//
// Deprecated: Use WithoutSoftDeleted() instead.
func (q *Query) WithoutTrashed() contractsorm.Query {
	return q.WithoutSoftDeleted()
}

// SoftDelete soft-deletes records by setting the soft-delete timestamp column.
// Returns an error if the model does not implement SoftDeleteColumnNamer
// (i.e., does not support soft deletes).
func (q *Query) SoftDelete(value ...any) (*contractsorm.Result, error) {
	query := q.Clone().(*Query)
	if len(value) > 0 {
		applyConditions(query, value)
	}

	if err := query.validate(); err != nil {
		return nil, err
	}

	if !hasSoftDeleteCapability(query.model) {
		return nil, fmt.Errorf(
			"SoftDelete() requires a model that implements SoftDeleteColumnNamer; " +
				"use Delete() or HardDelete() instead",
		)
	}

	if !query.withoutEvents && query.model != nil {
		attributes := observer.ExtractModelAttributes(query.model)
		if err := query.dispatcher.DispatchDeleting(query.ctx, query.model, query.modelToObserver, nil, attributes, nil, query); err != nil {
			return nil, fmt.Errorf("deleting event error: %w", err)
		}
	}

	query.includeSoftDeleted = true
	builder := NewBuilder(query)
	col := getSoftDeleteColumn(query.model)

	var deleteValue any = time.Now()
	if strat, ok := query.model.(contractsorm.SoftDeleteStrategy); ok {
		deleteValue = strat.SoftDeleteValue()
	}

	deleteSQL, args := builder.BuildUpdate(map[string]any{col: deleteValue})
	if deleteSQL == "" {
		return nil, fmt.Errorf("failed to build SOFT DELETE query")
	}

	ctx, cancel := query.timeoutContext()
	defer cancel()

	start := time.Now()
	var result interface{ RowsAffected() (int64, error) }
	var err error

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
		return nil, query.sanitizeError(fmt.Errorf("failed to execute SOFT DELETE query: %w", err))
	}
	query.logQuery(deleteSQL, args, start)

	if !query.withoutEvents && query.model != nil {
		attributes := observer.ExtractModelAttributes(query.model)
		if err := query.dispatcher.DispatchDeleted(query.ctx, query.model, query.modelToObserver, nil, attributes, nil, query); err != nil {
			return nil, fmt.Errorf("deleted event error: %w", err)
		}
	}

	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{RowsAffected: rowsAffected}, nil
}
