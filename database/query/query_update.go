package query

import (
	"database/sql"
	"fmt"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/observer"
)

// Update updates records in the database.
func (q *Query) Update(column any, value ...any) (*contractsorm.Result, error) {
	// Fire Updating event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchUpdating(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("updating event error: %w", err)
		}
	}

	// Build UPDATE query
	builder := NewBuilder(q)
	sqlStr, args := builder.BuildUpdate(column, value...)
	if sqlStr == "" {
		return nil, fmt.Errorf("failed to build UPDATE query")
	}

	// Execute query
	var err error
	var result sql.Result
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sqlStr, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, sqlStr, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute UPDATE query: %w", err)
	}
	q.logQuery(sqlStr, args, start)

	// Fire Updated event if not disabled
	if !q.withoutEvents && q.model != nil {
		attributes := observer.ExtractModelAttributes(q.model)
		if err := q.dispatcher.DispatchUpdated(q.ctx, q.model, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return nil, fmt.Errorf("updated event error: %w", err)
		}
	}

	// Get affected rows
	rowsAffected, _ := result.RowsAffected()
	return &contractsorm.Result{
		RowsAffected: rowsAffected,
	}, nil
}

// UpdateOrCreate updates a record if it exists, or creates it if it doesn't.
func (q *Query) UpdateOrCreate(dest any, attributes any, values any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		// Record exists, update it
		return q.Save(values)
	}

	// Record doesn't exist, create it
	return q.Create(values)
}
