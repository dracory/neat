package query

import (
	"fmt"
	"strings"
)

// BuildDelete builds a DELETE query from the query state.
func (b *Builder) BuildDelete() (string, []any) {
	var parts []string
	var args []any

	// DELETE clause
	parts = append(parts, "DELETE")

	// FROM clause
	if b.query.table != "" {
		parts = append(parts, fmt.Sprintf("FROM %s", b.quoteIdentifier(b.query.table)))
	}

	// Build WHERE clause (will be used for both normal WHERE and LIMIT workaround)
	whereParts, whereArgs := b.buildWheresWithSoftDelete()

	// LIMIT clause
	// MySQL supports LIMIT directly in DELETE
	// SQLite requires a subquery workaround: DELETE FROM ... WHERE rowid IN (SELECT rowid FROM ... ORDER BY ... LIMIT N)
	// PostgreSQL supports LIMIT directly in DELETE
	// SQL Server uses TOP instead of LIMIT
	if b.query.limit != nil {
		if b.query.driver != nil && b.query.driver.Dialect() == "mysql" {
			// Add WHERE clause if it exists
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
			parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
		} else if b.query.driver != nil && b.query.driver.Dialect() == "sqlite" {
			// SQLite workaround: wrap in subquery with rowid
			if whereParts == "" {
				whereParts = "1=1"
			}
			// Build ORDER BY clause for deterministic row selection
			var orderClause string
			if len(b.query.orders) > 0 {
				var orderParts []string
				for _, order := range b.query.orders {
					orderParts = append(orderParts, fmt.Sprintf("%s %s", order.column, order.direction))
				}
				orderClause = fmt.Sprintf(" ORDER BY %s", strings.Join(orderParts, ", "))
			}
			// Add WHERE clause with rowid subquery including ORDER BY
			parts = append(parts, fmt.Sprintf("WHERE rowid IN (SELECT rowid FROM %s WHERE %s%s LIMIT %d)", b.quoteIdentifier(b.query.table), whereParts, orderClause, *b.query.limit))
			args = append(args, whereArgs...)
		} else if b.query.driver != nil && b.query.driver.Dialect() == "postgres" {
			// PostgreSQL supports LIMIT directly in DELETE
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
			parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
		} else if b.query.driver != nil && b.query.driver.Dialect() == "sqlserver" {
			// SQL Server uses TOP instead of LIMIT
			// Insert TOP after DELETE
			for i, part := range parts {
				if strings.HasPrefix(part, "DELETE") {
					parts[i] = fmt.Sprintf("DELETE TOP (%d)%s", *b.query.limit, strings.TrimPrefix(part, "DELETE"))
					break
				}
			}
			// Add WHERE clause if it exists
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
		} else {
			// Other databases: add WHERE clause normally (LIMIT may or may not be supported)
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
		}
	} else {
		// No LIMIT: add WHERE clause normally
		if whereParts != "" {
			parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
			args = append(args, whereArgs...)
		}
	}

	return strings.Join(parts, " "), args
}
