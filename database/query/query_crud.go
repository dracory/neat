package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/observer"
)

// Find retrieves records matching the given conditions.
func (q *Query) Find(dest any, conds ...any) error {
	q = q.applyScopes()
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
	// Execute query
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()
		q.logQuery(sql, args, start)
		return clone.scanRows(rows, dest)
	}

	dbConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
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

// First retrieves the first record matching the query.
func (q *Query) First(dest any) error {
	q = q.applyScopes()
	// Use a clone to avoid mutating the original query state
	clone := q.Clone().(*Query)

	// Set limit to 1
	limit := 1
	clone.limit = &limit

	// Build SELECT query
	builder := NewBuilder(clone)
	sql, args := builder.BuildSelect()

	start := time.Now()
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		q.logQuery(sql, args, start)
		if err := q.scanRows(rows, dest); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

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

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	q.logQuery(sql, args, start)
	if err := q.scanRows(rows, dest); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

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

// Get retrieves all records matching the query.
func (q *Query) Get(dest any) error {
	q = q.applyScopes()
	// Use a clone to avoid mutating the original query state
	clone := q.Clone().(*Query)

	// Build SELECT query
	builder := NewBuilder(clone)
	sql, args := builder.BuildSelect()

	start := time.Now()
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		q.logQuery(sql, args, start)
		if err := q.scanRows(rows, dest); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

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

	rows, err := dbConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	q.logQuery(sql, args, start)
	if err := q.scanRows(rows, dest); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

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

// Create inserts a new record into the database.
func (q *Query) Create(value any) error {
	// Fire Creating event if not disabled
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchCreating(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("creating event error: %w", err)
		}
	}

	// Build INSERT query
	builder := NewBuilder(q)
	sqlStr, args := builder.BuildInsert(value)
	if sqlStr == "" {
		return fmt.Errorf("failed to build INSERT query")
	}

	// Execute query
	var result sql.Result
	var err error
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sqlStr, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return err
		}
		result, err = dbConn.ExecContext(q.ctx, sqlStr, args...)
	}

	if err != nil {
		return fmt.Errorf("failed to execute INSERT query: %w", err)
	}
	q.logQuery(sqlStr, args, start)

	// Populate last insert ID back into the model's primary key field
	if lastID, err := result.LastInsertId(); err == nil && lastID > 0 {
		// Handle bulk insert by setting IDs for each element
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			// For bulk inserts, set sequential IDs starting from lastID - len + 1
			startID := lastID - int64(v.Len()) + 1
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				if elem.CanAddr() {
					setModelPrimaryKey(elem.Addr().Interface(), startID+int64(i))
				}
			}
		} else {
			setModelPrimaryKey(value, lastID)
		}
	}

	// Fire Created event if not disabled
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchCreated(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("created event error: %w", err)
		}
	}

	return nil
}

// Save saves the model to the database (INSERT if no primary key, UPDATE otherwise).
func (q *Query) Save(value any) error {
	// Fire Saving event
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchSaving(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("saving event error: %w", err)
		}
	}

	id := getPrimaryKeyValue(value)
	var saveErr error
	if id != 0 {
		// UPDATE: set WHERE id = <id> on a clone, then call Update with the value
		clone := q.Clone().(*Query)
		clone.wheres = append(clone.wheres, whereClause{_type: "and", query: "id = ?", args: []any{id}})
		_, saveErr = clone.Update(value)
	} else {
		saveErr = q.Create(value)
	}

	if saveErr != nil {
		return saveErr
	}

	// Fire Saved event
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchSaved(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("saved event error: %w", err)
		}
	}
	return nil
}

// SaveQuietly saves the model without firing events.
func (q *Query) SaveQuietly(value any) error {
	return q.WithoutEvents().(*Query).Save(value)
}

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
	var result interface{ RowsAffected() (int64, error) }
	start := time.Now()
	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, deleteSQL, args...)
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return nil, err
		}
		result, err = dbConn.ExecContext(q.ctx, deleteSQL, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute DELETE query: %w", err)
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
