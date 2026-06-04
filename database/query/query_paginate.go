package query

import (
	"fmt"
)

// Paginate paginates the query results.
func (q *Query) Paginate(page, limit int, dest any, total *int64) error {
	// Calculate offset
	offset := (page - 1) * limit
	q.offset = &offset
	q.limit = &limit

	// Get total count first
	countQuery := q.Clone().(*Query)
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
	ctx, cancel := q.timeoutContext()
	defer cancel()
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute PAGINATE query: %w", err)
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
		return fmt.Errorf("failed to execute PAGINATE query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return q.scanRows(rows, dest)
}
