package query

import (
	"database/sql"
	"fmt"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Scan executes the query and scans the results into the destination.
func (q *Query) Scan(dest any) error {
	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute SCAN query: %w", err)
		}
		defer rows.Close()

		return q.scanRows(rows, dest)
	}

	databaseConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute SCAN query: %w", err)
	}
	defer rows.Close()

	return q.scanRows(rows, dest)
}

// Chunk processes the query results in chunks and calls the callback for each chunk.
func (q *Query) Chunk(size int, callback any) error {
	// Build SELECT query without limit (we chunk in memory)
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute CHUNK query: %w", err)
		}
		defer rows.Close()

		return q.chunkRows(rows, size, callback)
	}

	databaseConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute CHUNK query: %w", err)
	}
	defer rows.Close()

	return q.chunkRows(rows, size, callback)
}

// Paginate paginates the query results.
func (q *Query) Paginate(page, limit int, dest any, total *int64) error {
	// Calculate offset
	offset := (page - 1) * limit
	q.offset = &offset
	q.limit = &limit

	// Get total count first
	countQuery := *q
	countQuery.limit = nil
	countQuery.offset = nil
	var count int64
	if err := countQuery.Count(&count); err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}
	if total != nil {
		*total = count
	}

	// Build SELECT query for paginated results
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute PAGINATE query: %w", err)
		}
		defer rows.Close()

		return q.scanRows(rows, dest)
	}

	databaseConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute PAGINATE query: %w", err)
	}
	defer rows.Close()

	return q.scanRows(rows, dest)
}

// FirstOr retrieves the first record or executes a callback if not found.
func (q *Query) FirstOr(dest any, callback func() error) error {
	err := q.First(dest)
	if err != nil {
		return callback()
	}
	return nil
}

// FirstOrCreate retrieves the first record or creates it if not found.
func (q *Query) FirstOrCreate(dest any, conds ...any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, create it
	return q.Create(dest)
}

// FirstOrNew retrieves the first record or prepares a new instance if not found.
func (q *Query) FirstOrNew(dest any, attributes any, values ...any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, prepare new instance (without saving)
	// This is a simplified implementation
	return nil
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
