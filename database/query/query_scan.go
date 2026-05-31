package query

import (
	"fmt"
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
