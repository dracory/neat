package cursor

import (
	"database/sql"
	"fmt"
)

// Cursor represents a database cursor for streaming results.
type Cursor struct {
	rows   *sql.Rows
	err    error
	closed bool
}

// NewCursor creates a new Cursor instance.
func NewCursor(rows *sql.Rows) *Cursor {
	return &Cursor{
		rows:   rows,
		err:    nil,
		closed: false,
	}
}

// Scan scans the current row into dest.
func (c *Cursor) Scan(dest any) error {
	if c.closed {
		return fmt.Errorf("cursor is closed")
	}
	if c.err != nil {
		return c.err
	}
	return c.rows.Scan(dest)
}

// Next advances the cursor to the next row.
func (c *Cursor) Next() bool {
	if c.closed || c.err != nil {
		return false
	}
	return c.rows.Next()
}

// Close closes the cursor.
func (c *Cursor) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.rows.Close()
}

// Err returns any error encountered by the cursor.
func (c *Cursor) Err() error {
	if c.err != nil {
		return c.err
	}
	return c.rows.Err()
}

// Columns returns the column names.
func (c *Cursor) Columns() ([]string, error) {
	if c.closed {
		return nil, fmt.Errorf("cursor is closed")
	}
	return c.rows.Columns()
}
