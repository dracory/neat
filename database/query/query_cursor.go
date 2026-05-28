package query

import (
	"database/sql"
	"fmt"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Cursor returns a cursor for streaming query results.
func (q *Query) Cursor() (chan contractsorm.Cursor, error) {
	// Build SELECT query
	builder := NewBuilder(q)
	querySQL, args := builder.BuildSelect()

	// Execute query
	var rows *sql.Rows
	var err error
	if q.tx != nil {
		rows, err = q.tx.QueryContext(q.ctx, querySQL, args...)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return nil, err
		}
		rows, err = databaseConn.QueryContext(q.ctx, querySQL, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute cursor query: %w", err)
	}

	// Create cursor channel
	cursorChan := make(chan contractsorm.Cursor, 10)

	go func() {
		defer rows.Close()
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
	// This is a simplified implementation
	// In a full implementation, this would properly map the data to the destination
	return nil
}
