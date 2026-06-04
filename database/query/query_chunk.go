package query

import (
	"fmt"
)

// Chunk processes the query results in chunks and calls the callback for each chunk.
func (q *Query) Chunk(size int, callback any) error {
	// Build SELECT query without limit (we chunk in memory)
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	ctx, cancel := q.timeoutContext()
	defer cancel()
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute CHUNK query: %w", err)
		}
		defer func() { _ = rows.Close() }()

		return q.chunkRows(rows, size, callback)
	}

	databaseConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute CHUNK query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return q.chunkRows(rows, size, callback)
}
