package query

import (
	"database/sql"
	"fmt"
	"reflect"
)

// Scan executes the query and scans the results into the destination.
func (q *Query) Scan(dest any) error {
	// Check for build errors from query construction
	if q.buildError != nil {
		return q.buildError
	}
	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	ctx, cancel := q.timeoutContext()
	defer cancel()
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(ctx, sql, args...)
		if err != nil {
			return q.sanitizeError(fmt.Errorf("failed to execute SCAN query: %w", err))
		}
		defer func() { _ = rows.Close() }()

		return q.scanRows(rows, dest)
	}

	databaseConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(ctx, sql, args...)
	if err != nil {
		return q.sanitizeError(fmt.Errorf("failed to execute SCAN query: %w", err))
	}
	defer func() { _ = rows.Close() }()

	return q.scanRows(rows, dest)
}

// scanRows scans database rows into the destination.
func (q *Query) scanRows(rows *sql.Rows, dest any) error {
	// Use reflection to handle different destination types
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Pointer {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()

	// Handle *interface{} destination (for AsVar methods)
	if destValue.Kind() == reflect.Interface {
		if !rows.Next() {
			return fmt.Errorf("no rows found")
		}

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		m := make(map[string]any, len(columns))
		for i, col := range columns {
			m[col] = values[i]
		}
		destValue.Set(reflect.ValueOf(m))

		return rows.Err()
	}

	// Handle slice destination
	if destValue.Kind() == reflect.Slice {
		sliceType := destValue.Type()
		elemType := sliceType.Elem()

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		for rows.Next() {
			// Create new element
			elemPtr := reflect.New(elemType)
			elem := elemPtr.Elem()

			// Scan into element
			values := make([]any, len(columns))

			if elem.Kind() == reflect.Interface {
				// Scan into a map and set as interface value
				ptrs := make([]any, len(columns))
				for i := range values {
					ptrs[i] = &values[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
				m := make(map[string]any, len(columns))
				for i, col := range columns {
					m[col] = values[i]
				}
				elem.Set(reflect.ValueOf(m))
			} else if elem.Kind() == reflect.Map {
				// Scan into a temporary []any then build the map
				ptrs := make([]any, len(columns))
				for i := range values {
					ptrs[i] = &values[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
				m := reflect.MakeMap(elemType)
				keyType := elemType.Key()
				for i, col := range columns {
					m.SetMapIndex(reflect.ValueOf(col).Convert(keyType), reflect.ValueOf(values[i]))
				}
				elem.Set(m)
			} else if elem.Kind() == reflect.Struct {
				values = structScanDests(elem, columns)
				if err := rows.Scan(values...); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
				copyScanResults(elem, columns, values)
			} else {
				// Scalar slice element (e.g. []string, []int)
				if err := rows.Scan(elem.Addr().Interface()); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
			}

			// Note: Relations are loaded after rows are closed to avoid SQLite deadlock
			// This is handled by the calling methods (First, Get, etc.)

			// Append to slice
			destValue.Set(reflect.Append(destValue, elem))
		}

		return rows.Err()
	}

	// Handle single struct destination
	if destValue.Kind() == reflect.Struct {
		if !rows.Next() {
			return fmt.Errorf("no rows found")
		}

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		values := structScanDests(destValue, columns)
		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		copyScanResults(destValue, columns, values)

		// Note: Relations are loaded after rows are closed to avoid SQLite deadlock
		// This is handled by the calling methods (First, Get, etc.)

		return rows.Err()
	}

	// Handle single map destination (*map[string]any)
	if destValue.Kind() == reflect.Map {
		if !rows.Next() {
			return nil
		}

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		m := make(map[string]any, len(columns))
		for i, col := range columns {
			m[col] = values[i]
		}
		destValue.Set(reflect.ValueOf(m))

		return rows.Err()
	}

	return fmt.Errorf("unsupported destination type: %T", dest)
}

// chunkRows processes rows in chunks and calls the callback for each chunk.
func (q *Query) chunkRows(rows *sql.Rows, size int, callback any) error {
	// Use reflection to call the callback
	callbackValue := reflect.ValueOf(callback)
	if callbackValue.Kind() != reflect.Func {
		return fmt.Errorf("callback must be a function")
	}

	// Get callback parameter type
	callbackType := callbackValue.Type()
	if callbackType.NumIn() != 1 {
		return fmt.Errorf("callback must accept exactly one parameter")
	}
	if callbackType.NumOut() > 1 {
		return fmt.Errorf("callback must return at most one value")
	}
	if callbackType.NumOut() == 1 {
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if !callbackType.Out(0).Implements(errorType) {
			return fmt.Errorf("callback return type must be error")
		}
	}

	paramType := callbackType.In(0)
	if paramType.Kind() != reflect.Slice {
		return fmt.Errorf("callback parameter must be a slice")
	}

	elemType := paramType.Elem()
	realElemType := elemType
	isPtr := elemType.Kind() == reflect.Pointer
	if isPtr {
		realElemType = elemType.Elem()
	}

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Process rows in chunks
	chunk := reflect.MakeSlice(paramType, 0, size)

	for rows.Next() {
		// Create a new element
		var elem reflect.Value
		if isPtr {
			elem = reflect.New(realElemType)
		} else {
			elem = reflect.New(elemType).Elem()
		}

		// Scan into element
		if realElemType.Kind() == reflect.Struct {
			var scanElem reflect.Value
			if isPtr {
				scanElem = elem.Elem()
			} else {
				scanElem = elem
			}
			values := structScanDests(scanElem, columns)
			if err := rows.Scan(values...); err != nil {
				return fmt.Errorf("failed to scan row into struct: %w", err)
			}
			copyScanResults(scanElem, columns, values)
		} else if realElemType.Kind() == reflect.Map {
			// Scan into a map
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				return fmt.Errorf("failed to scan row into map: %w", err)
			}
			m := reflect.MakeMap(realElemType)
			keyType := realElemType.Key()
			for i, col := range columns {
				m.SetMapIndex(reflect.ValueOf(col).Convert(keyType), reflect.ValueOf(values[i]))
			}
			if isPtr {
				elem.Elem().Set(m)
			} else {
				elem.Set(m)
			}
		} else {
			// Scalar element
			var scanDest reflect.Value
			if isPtr {
				scanDest = elem
			} else {
				scanDest = elem.Addr()
			}
			if err := rows.Scan(scanDest.Interface()); err != nil {
				return fmt.Errorf("failed to scan row into scalar: %w", err)
			}
		}

		chunk = reflect.Append(chunk, elem)

		// Call callback when chunk is full
		if chunk.Len() >= size {
			results := callbackValue.Call([]reflect.Value{chunk})
			if len(results) > 0 {
				if err, ok := results[0].Interface().(error); ok && err != nil {
					return err
				}
			}
			chunk = reflect.MakeSlice(paramType, 0, size)
		}
	}

	// Process remaining rows in the last chunk
	if chunk.Len() > 0 {
		results := callbackValue.Call([]reflect.Value{chunk})
		if len(results) > 0 {
			if err, ok := results[0].Interface().(error); ok && err != nil {
				return err
			}
		}
	}

	return rows.Err()
}
