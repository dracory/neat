package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Cursor returns a cursor for streaming query results.
func (q *Query) Cursor() (chan contractsorm.Cursor, error) {
	// Build SELECT query
	builder := NewBuilder(q)
	querySQL, args := builder.BuildSelect()

	// Execute query
	ctx, cancel := q.timeoutContext()
	var rows *sql.Rows
	var err error
	if q.tx != nil {
		rows, err = q.tx.QueryContext(ctx, querySQL, args...)
	} else {
		var databaseConn *sql.DB
		databaseConn, err = q.ReadDB()
		if err != nil {
			cancel()
			return nil, err
		}
		rows, err = databaseConn.QueryContext(ctx, querySQL, args...)
	}

	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to execute cursor query: %w", err)
	}

	// Create cursor channel
	cursorChan := make(chan contractsorm.Cursor, 10)

	go func() {
		defer cancel()
		defer func() { _ = rows.Close() }()
		defer close(cursorChan)

		for rows.Next() {
			columns, err := rows.Columns()
			if err != nil {
				return
			}

			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return
			}

			// Create map from column names to values
			result := make(map[string]any)
			for i, col := range columns {
				result[col] = values[i]
			}

			// Create a cursor wrapper that implements orm.Cursor
			cursorChan <- &cursorWrapper{data: result}
		}
	}()

	return cursorChan, nil
}

// cursorWrapper wraps a map to implement orm.Cursor
type cursorWrapper struct {
	data map[string]any
}

// Scan implements the orm.Cursor interface
func (c *cursorWrapper) Scan(dest any) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}
	destValue = destValue.Elem()

	// Handle struct destination
	if destValue.Kind() == reflect.Struct {
		colToPath := getColumnToIndexPath(destValue.Type())
		for col, val := range c.data {
			key := strings.ToLower(col)
			if path, ok := colToPath[key]; ok {
				field := destValue.FieldByIndex(path)
				if field.CanSet() {
					valValue := reflect.ValueOf(val)
					if valValue.Type().AssignableTo(field.Type()) {
						field.Set(valValue)
					} else if valValue.Type().ConvertibleTo(field.Type()) {
						field.Set(valValue.Convert(field.Type()))
					}
				}
			}
		}
		return nil
	}

	// Handle map destination
	if destValue.Kind() == reflect.Map {
		if destValue.IsNil() {
			destValue.Set(reflect.MakeMap(destValue.Type()))
		}
		keyType := destValue.Type().Key()
		for col, val := range c.data {
			destValue.SetMapIndex(reflect.ValueOf(col).Convert(keyType), reflect.ValueOf(val))
		}
		return nil
	}

	return fmt.Errorf("unsupported destination type: %T", dest)
}
