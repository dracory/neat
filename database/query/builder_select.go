package query

import (
	"fmt"
	"strings"
)

// BuildSelect builds a SELECT query from the query state.
func (b *Builder) BuildSelect() (string, []any) {
	if b.query.rawSQL != "" {
		return b.query.rawSQL, b.query.rawArgs
	}

	var parts []string
	args := make([]any, 0)

	// SELECT clause
	if b.query.aggregate != "" {
		// When aggregate is set, ignore SELECT list and use aggregate function
		// Handle COUNT with DISTINCT
		if b.query.aggregate == "COUNT" && b.query.distinct {
			if len(b.query.distinctCols) > 0 {
				parts = append(parts, fmt.Sprintf("SELECT COUNT(DISTINCT %s)", strings.Join(b.query.distinctCols, ", ")))
			} else if len(b.query.selects) > 0 {
				var selectParts []string
				for _, s := range b.query.selects {
					selectParts = append(selectParts, s.expr)
					args = append(args, s.args...)
				}
				parts = append(parts, fmt.Sprintf("SELECT COUNT(DISTINCT %s)", strings.Join(selectParts, ", ")))
			} else {
				parts = append(parts, fmt.Sprintf("SELECT %s(%s)", b.query.aggregate, b.query.aggregateCol))
			}
		} else {
			parts = append(parts, fmt.Sprintf("SELECT %s(%s)", b.query.aggregate, b.query.aggregateCol))
		}
	} else if len(b.query.selects) > 0 {
		var selectParts []string
		for _, s := range b.query.selects {
			selectParts = append(selectParts, s.expr)
			args = append(args, s.args...)
		}
		// Prepend DISTINCT if set
		if b.query.distinct {
			parts = append(parts, fmt.Sprintf("SELECT DISTINCT %s", strings.Join(selectParts, ", ")))
		} else {
			parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(selectParts, ", ")))
		}
	} else {
		// No explicit SELECT, derive from model
		if b.query.model != nil {
			cols := b.extractColumnNames(b.query.model)
			if len(cols) > 0 {
				// Filter out omitted columns
				var filteredCols []string
				for _, col := range cols {
					omitted := false
					for _, omit := range b.query.omitColumns {
						if omit == col {
							omitted = true
							break
						}
					}
					if !omitted {
						filteredCols = append(filteredCols, col)
					}
				}
				if len(filteredCols) > 0 {
					parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(filteredCols, ", ")))
				} else {
					parts = append(parts, "SELECT *")
				}
			} else {
				parts = append(parts, "SELECT *")
			}
		} else {
			parts = append(parts, "SELECT *")
		}
	}

	// FROM clause
	if b.query.table != "" {
		if strings.Contains(b.query.table, "(") && strings.Contains(b.query.table, ")") {
			// Subquery in FROM, don't quote
			parts = append(parts, fmt.Sprintf("FROM %s", b.query.table))
		} else {
			parts = append(parts, fmt.Sprintf("FROM %s", b.quoteIdentifier(b.query.table)))
		}
		args = append(args, b.query.tableArgs...)
	}

	// JOIN clauses
	for _, join := range b.query.joins {
		parts = append(parts, fmt.Sprintf("%s %s", join._type, join.query))
		args = append(args, join.args...)
	}

	// WHERE clauses (with automatic soft-delete filter)
	whereParts, whereArgs := b.buildWheresWithSoftDelete()
	if whereParts != "" {
		parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
		args = append(args, whereArgs...)
	}

	// GROUP BY clauses
	if len(b.query.groups) > 0 {
		parts = append(parts, fmt.Sprintf("GROUP BY %s", strings.Join(b.query.groups, ", ")))
	}

	// HAVING clauses
	if len(b.query.havings) > 0 {
		var havingParts []string
		for _, having := range b.query.havings {
			havingParts = append(havingParts, having.query)
			args = append(args, having.args...)
		}
		parts = append(parts, fmt.Sprintf("HAVING %s", strings.Join(havingParts, " AND ")))
	}

	// ORDER BY clauses
	if len(b.query.orders) > 0 {
		var orderParts []string
		for _, order := range b.query.orders {
			orderParts = append(orderParts, fmt.Sprintf("%s %s", order.column, order.direction))
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(orderParts, ", ")))
	}

	// LIMIT clause
	if b.query.limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
	}

	// OFFSET clause
	if b.query.offset != nil {
		// SQLite requires LIMIT when using OFFSET
		if b.query.limit == nil && b.query.driver != nil && b.query.driver.Dialect() == "sqlite" {
			parts = append(parts, "LIMIT -1")
		}
		parts = append(parts, fmt.Sprintf("OFFSET %d", *b.query.offset))
	}

	// Locking clauses
	// Skip lock clauses for SQLite as it doesn't support them
	if b.query.driver == nil || b.query.driver.Dialect() != "sqlite" {
		if b.query.lockForUpdate {
			parts = append(parts, "FOR UPDATE")
		} else if b.query.sharedLock {
			parts = append(parts, "LOCK IN SHARE MODE")
		}
	}

	return strings.Join(parts, " "), args
}
