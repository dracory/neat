package query

import (
	"context"
	"fmt"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	neaterrors "github.com/dracory/neat/errors"
)

// logQuery appends a QueryLog entry with the actual execution duration.
// It also emits a warning via the logger when SlowThreshold is configured and exceeded.
func (q *Query) logQuery(sql string, bindings []any, start time.Time) {
	elapsed := float64(time.Since(start).Milliseconds())
	if q.enableLog && q.queryLog != nil {
		*q.queryLog = append(*q.queryLog, contractsorm.QueryLog{
			Query:    sql,
			Bindings: bindings,
			Time:     elapsed,
		})
	}
	// Slow-query warning
	if q.dbConfig != nil && q.dbConfig.SlowThreshold > 0 && elapsed >= float64(q.dbConfig.SlowThreshold) {
		if q.log != nil {
			q.log.Warningf("[slow query %.1fms] %s [%d bindings redacted]", elapsed, sql, len(bindings))
		}
	}
}

func (q *Query) validateAggregate(column string) error {
	// Validate column name: alphanumeric, underscores, dots or *
	for _, r := range column {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '*') {
			return fmt.Errorf("invalid column name: %s", column)
		}
	}

	return nil
}

// validate checks for common validation errors before executing terminal methods.
// It validates that the database connection is not nil and that a table name is set.
// This provides clear, structured errors instead of relying on database driver errors.
func (q *Query) validate() error {
	// Check for build errors from query construction
	if q.buildError != nil {
		return q.buildError
	}
	// Check for nil database connection
	if q.db == nil && q.readDB == nil && q.writeDB == nil {
		return neaterrors.ErrNilDatabase.SetModule("query")
	}

	// Check for empty table name (only for operations that require a table)
	if q.table == "" && q.rawSQL == "" {
		return neaterrors.NewValidationError("table name cannot be empty").SetModule("query")
	}

	return nil
}

// timeoutContext returns a context derived from q.ctx with a QueryTimeout deadline
// applied when one is configured. The caller must invoke the returned cancel func
// (e.g. via defer) to release resources.
func (q *Query) timeoutContext() (context.Context, context.CancelFunc) {
	base := q.ctx
	if base == nil {
		base = context.Background()
	}
	if q.dbConfig != nil && q.dbConfig.Pool.QueryTimeout > 0 {
		return context.WithTimeout(base, time.Duration(q.dbConfig.Pool.QueryTimeout)*time.Second)
	}
	return base, func() {}
}

// UpdateOrInsert updates a record if it exists, or creates it if it doesn't.
// The operation is performed atomically within a transaction to prevent race conditions.
func (q *Query) UpdateOrInsert(attributes any, values any) error {
	// If already in a transaction, proceed without nesting
	if q.inTransaction {
		return q.updateOrInsertInTransaction(attributes, values)
	}

	// Wrap the entire operation in a transaction for atomicity
	return q.Transaction(func(tx contractsorm.Query) error {
		txQ, ok := tx.(*Query)
		if !ok {
			return fmt.Errorf("unexpected transaction type: %T", tx)
		}
		return txQ.updateOrInsertInTransaction(attributes, values)
	})
}

// updateOrInsertInTransaction performs the actual UpdateOrInsert logic within a transaction.
func (q *Query) updateOrInsertInTransaction(attributes any, values any) error {
	// Build WHERE conditions from attributes
	clone := q.Clone().(*Query)

	// Handle nil attributes - just create with values
	if attributes == nil {
		return q.Create(values)
	}

	// Handle map[string]any for attributes
	if attrs, ok := attributes.(map[string]any); ok {
		for col, val := range attrs {
			clone.Where(col, val)
		}
	} else {
		// For structs, extract fields and build WHERE conditions
		cols, vals, err := NewBuilder(q).extractSingleColumnsAndValues(attributes)
		if err != nil {
			return fmt.Errorf("failed to extract columns and values from attributes: %w", err)
		}
		for i, col := range cols {
			clone.Where(col, vals[i])
		}
	}

	// Try to find the record first
	count := int64(0)
	if err := clone.Count(&count); err != nil {
		return err
	}

	if count > 0 {
		// Record exists, update it
		// Use the original query with WHERE conditions from attributes
		updateQ := q.Clone().(*Query)
		if attrs, ok := attributes.(map[string]any); ok {
			for col, val := range attrs {
				updateQ.Where(col, val)
			}
		} else {
			cols, vals, err := NewBuilder(q).extractSingleColumnsAndValues(attributes)
			if err != nil {
				return fmt.Errorf("failed to extract columns and values from attributes: %w", err)
			}
			for i, col := range cols {
				updateQ.Where(col, vals[i])
			}
		}
		_, err := updateQ.Update(values)
		return err
	}

	// Record doesn't exist, create it
	// Merge attributes and values for the insert
	if attrsMap, ok := attributes.(map[string]any); ok {
		if valsMap, ok := values.(map[string]any); ok {
			// Merge both maps
			merged := make(map[string]any)
			for k, v := range attrsMap {
				merged[k] = v
			}
			for k, v := range valsMap {
				merged[k] = v
			}
			return q.Create(merged)
		} else {
			// Attributes is map, values is struct - merge them
			merged := make(map[string]any)
			for k, v := range attrsMap {
				merged[k] = v
			}
			// Extract struct fields and add to merged map
			cols, vals, err := NewBuilder(q).extractSingleColumnsAndValues(values)
			if err == nil {
				for i, col := range cols {
					merged[col] = vals[i]
				}
			}
			return q.Create(merged)
		}
	}

	// For struct values or mixed types, just create with values
	return q.Create(values)
}
