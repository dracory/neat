package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/observer"
)

// Update updates records in the database.
func (q *Query) Update(column any, value ...any) (*contractsorm.Result, error) {
	// Validate common conditions (build errors, nil DB, empty table)
	if err := q.validate(); err != nil {
		return nil, err
	}
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
	ctx, cancel := q.timeoutContext()
	defer cancel()
	var err error
	var result sql.Result
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(ctx, sqlStr, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(ctx, sqlStr, args...)
	}

	if err != nil {
		return nil, q.sanitizeError(fmt.Errorf("failed to execute UPDATE query: %w", err))
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
// The operation is performed atomically within a transaction to prevent race conditions.
func (q *Query) UpdateOrCreate(dest any, attributes any, values any) error {
	// If already in a transaction, proceed without nesting
	if q.inTransaction {
		return q.updateOrCreateInTransaction(dest, attributes, values)
	}

	// Wrap the entire operation in a transaction for atomicity
	return q.Transaction(func(tx contractsorm.Query) error {
		txQ, ok := tx.(*Query)
		if !ok {
			return fmt.Errorf("unexpected transaction type: %T", tx)
		}
		return txQ.updateOrCreateInTransaction(dest, attributes, values)
	})
}

// updateOrCreateInTransaction performs the actual UpdateOrCreate logic within a transaction.
func (q *Query) updateOrCreateInTransaction(dest any, attributes any, values any) error {
	// Use a clone to avoid mutating the original query
	clone := q.Clone().(*Query)

	// Build WHERE clause from attributes
	attrsMap, err := extractAttributes(attributes)
	if err != nil {
		return fmt.Errorf("failed to extract attributes: %w", err)
	}

	for col, val := range attrsMap {
		clone = clone.Where(col+" = ?", val).(*Query)
	}

	// Try to find the record first
	err = clone.First(dest)
	if err == nil {
		// Record exists, update it
		// Merge attributes into values for the update
		merged := mergeAttributes(values, attributes)
		// Set the ID from the found record to ensure we update the correct record
		setModelPrimaryKey(merged, getModelPrimaryKey(dest))
		if err := q.Save(merged); err != nil {
			return err
		}
		// Re-fetch the updated record into dest
		return clone.First(dest)
	}

	// Record doesn't exist, create it
	// Merge attributes and values for the create
	merged := mergeAttributes(values, attributes)
	if err := q.Create(merged); err != nil {
		return err
	}
	// Fetch the created record into dest
	return clone.First(dest)
}

// extractAttributes extracts key-value pairs from a struct or map
func extractAttributes(value any) (map[string]any, error) {
	result := make(map[string]any)

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			result[key.String()] = v.MapIndex(key).Interface()
		}
		return result, nil
	}

	if v.Kind() != reflect.Struct {
		return result, nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		// Get column name from tag or field name
		colName := structFieldColumnName(field)
		if colName == "" {
			continue
		}

		// Skip zero values
		if fieldValue.IsZero() && fieldValue.Kind() != reflect.Bool {
			continue
		}

		result[colName] = fieldValue.Interface()
	}

	return result, nil
}

// mergeAttributes merges two attribute maps, with values taking precedence
func mergeAttributes(values any, attributes any) any {
	// If both are structs, we need to merge them
	vVal := reflect.ValueOf(values)
	if vVal.Kind() == reflect.Ptr {
		vVal = vVal.Elem()
	}

	vAttr := reflect.ValueOf(attributes)
	if vAttr.Kind() == reflect.Ptr {
		vAttr = vAttr.Elem()
	}

	// If both are maps, merge them with values taking precedence
	if vVal.Kind() == reflect.Map && vAttr.Kind() == reflect.Map {
		merged := make(map[string]any)
		// Copy all attributes first
		for _, key := range vAttr.MapKeys() {
			merged[key.String()] = vAttr.MapIndex(key).Interface()
		}
		// Then overwrite with values
		for _, key := range vVal.MapKeys() {
			merged[key.String()] = vVal.MapIndex(key).Interface()
		}
		return merged
	}

	// If values is a struct and attributes is a struct, create a new struct with merged fields
	if vVal.Kind() == reflect.Struct && vAttr.Kind() == reflect.Struct {
		// Create a new instance of the values type
		merged := reflect.New(vVal.Type()).Elem()

		// Copy all fields from attributes first
		for i := 0; i < vAttr.NumField(); i++ {
			attrField := vAttr.Field(i)
			mergedField := merged.Field(i)
			if mergedField.CanSet() {
				mergedField.Set(attrField)
			}
		}

		// Then copy/overwrite with values (only non-zero values)
		for i := 0; i < vVal.NumField(); i++ {
			valField := vVal.Field(i)
			mergedField := merged.Field(i)
			if mergedField.CanSet() && !valField.IsZero() {
				mergedField.Set(valField)
			}
		}

		return merged.Addr().Interface()
	}

	// Default: return values
	return values
}
