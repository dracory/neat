package query

import (
	"fmt"
	"strings"
)

// BuildUpdate builds an UPDATE query from the query state.
func (b *Builder) BuildUpdate(column any, values ...any) (string, []any) {
	var parts []string
	var args []any
	var setArgs []any // Store SET args separately to add them after WHERE args

	// UPDATE clause
	parts = append(parts, "UPDATE")

	// Table name
	if b.query.table != "" {
		parts = append(parts, b.quoteIdentifier(b.query.table))
	}

	// SET clause
	var setParts []string

	// Get placeholder function for the dialect
	placeholderFunc := func(n int) string { return "?" }
	if b.query.driver != nil {
		placeholderFunc = b.query.driver.Placeholder
	}
	placeholderIndex := 1

	// Handle map[string]any for column/value pairs
	if m, ok := column.(map[string]any); ok {
		for col, val := range m {
			// Skip omitted columns
			omitted := false
			for _, omit := range b.query.omitColumns {
				if omit == col {
					omitted = true
					break
				}
			}
			if omitted {
				continue
			}
			setParts = append(setParts, fmt.Sprintf("%s = %s", b.quoteIdentifier(col), placeholderFunc(placeholderIndex)))
			placeholderIndex++
			setArgs = append(setArgs, val)
		}
	} else if len(values) > 0 {
		// Handle single column with value
		if colStr, ok := column.(string); ok {
			// Check if the column string is already a complete SET expression (contains =)
			if strings.Contains(colStr, "=") {
				// Use the expression as-is (for Increment/Decrement), but replace ? with dialect-specific placeholder
				replacedExpr := strings.Replace(colStr, "?", placeholderFunc(placeholderIndex), 1)
				setParts = append(setParts, replacedExpr)
				placeholderIndex++
				setArgs = append(setArgs, values...)
			} else if strings.Contains(colStr, "->") {
				// Handle JSON path updates for MySQL and SQLite
				parts := strings.Split(colStr, "->")
				if len(parts) >= 2 {
					jsonColumn := b.quoteIdentifier(parts[0])
					// Build JSON path: $.name for "data->name", $.meta.active for "data->meta->active"
					jsonPath := "$." + strings.Join(parts[1:], ".")

					if b.query.driver != nil && b.query.driver.Dialect() == "mysql" {
						// MySQL: JSON_SET(column, '$.path', value)
						setParts = append(setParts, fmt.Sprintf("%s = JSON_SET(%s, '%s', %s)", jsonColumn, jsonColumn, jsonPath, placeholderFunc(placeholderIndex)))
						placeholderIndex++
						setArgs = append(setArgs, values[0])
					} else if b.query.driver != nil && b.query.driver.Dialect() == "sqlite" {
						// SQLite: json_set(column, '$.path', value)
						setParts = append(setParts, fmt.Sprintf("%s = json_set(%s, '%s', %s)", jsonColumn, jsonColumn, jsonPath, placeholderFunc(placeholderIndex)))
						placeholderIndex++
						setArgs = append(setArgs, values[0])
					} else {
						// Fallback to normal behavior for other databases
						setParts = append(setParts, fmt.Sprintf("%s = %s", b.quoteIdentifier(colStr), placeholderFunc(placeholderIndex)))
						placeholderIndex++
						setArgs = append(setArgs, values[0])
					}
				} else {
					// Fallback to normal behavior
					setParts = append(setParts, fmt.Sprintf("%s = %s", b.quoteIdentifier(colStr), placeholderFunc(placeholderIndex)))
					placeholderIndex++
					setArgs = append(setArgs, values[0])
				}
			} else {
				setParts = append(setParts, fmt.Sprintf("%s = %s", b.quoteIdentifier(colStr), placeholderFunc(placeholderIndex)))
				placeholderIndex++
				setArgs = append(setArgs, values[0])
			}
		}
	} else {
		// Handle struct or pointer-to-struct: extract fields as col=? pairs
		cols, vals, err := b.extractColumnsAndValues(column)
		if err == nil && vals != nil {
			for i, col := range cols {
				// Skip omitted columns
				omitted := false
				for _, omit := range b.query.omitColumns {
					if omit == col {
						omitted = true
						break
					}
				}
				if omitted {
					continue
				}
				setParts = append(setParts, fmt.Sprintf("%s = %s", b.quoteIdentifier(col), placeholderFunc(placeholderIndex)))
				placeholderIndex++
				setArgs = append(setArgs, vals[i])
			}
		}
	}

	if len(setParts) > 0 {
		parts = append(parts, fmt.Sprintf("SET %s", strings.Join(setParts, ", ")))
	}

	// WHERE clauses (with automatic soft-delete filter)
	// Skip soft-delete filter if we're updating deleted_at column (for soft delete operations)
	isSoftDeleteOperation := false
	if m, ok := column.(map[string]any); ok {
		if _, hasDeletedAt := m["deleted_at"]; hasDeletedAt {
			isSoftDeleteOperation = true
		}
	}

	// Add SET args first (they appear first in the SQL: SET ... WHERE ...)
	args = append(args, setArgs...)

	// Build WHERE clause (will be used for both normal WHERE and LIMIT workaround)
	var whereParts string
	var whereArgs []any
	if !isSoftDeleteOperation {
		whereParts, whereArgs = b.buildWheresWithSoftDeleteIndex(placeholderIndex)
	} else {
		// For soft delete operations, use regular WHERE without soft-delete filter
		whereParts, whereArgs = b.buildWheresWithIndex(placeholderIndex)
	}

	// LIMIT clause
	// MySQL supports LIMIT directly in UPDATE
	// SQLite requires a subquery workaround: UPDATE ... WHERE rowid IN (SELECT rowid FROM ... ORDER BY ... LIMIT N)
	// PostgreSQL supports LIMIT directly in UPDATE
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
			// PostgreSQL supports LIMIT directly in UPDATE
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
			parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
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
