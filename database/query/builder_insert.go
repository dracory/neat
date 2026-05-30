package query

import (
	"fmt"
	"reflect"
	"strings"
)

// BuildInsert builds an INSERT query from the query state.
func (b *Builder) BuildInsert(value any) (string, []any) {
	var parts []string
	var args []any

	// INSERT clause
	parts = append(parts, "INSERT")

	// INTO clause
	if b.query.table != "" {
		parts = append(parts, fmt.Sprintf("INTO %s", b.quoteIdentifier(b.query.table)))
		args = append(args, b.query.tableArgs...)
	}

	// Extract columns and values from the value
	columns, values, err := b.extractColumnsAndValues(value)
	if err != nil {
		return "", nil
	}

	if len(columns) > 0 {
		// Quote column names
		quotedColumns := make([]string, len(columns))
		for i, col := range columns {
			quotedColumns[i] = b.quoteIdentifier(col)
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(quotedColumns, ", ")))
		parts = append(parts, "VALUES")

		// Get placeholder function for the dialect
		placeholderFunc := func(n int) string { return "?" }
		if b.query.driver != nil {
			placeholderFunc = b.query.driver.Placeholder
		}

		// Handle bulk insert
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			rowPlaceholders := make([]string, v.Len())
			placeholderIndex := 1
			for i := 0; i < v.Len(); i++ {
				placeholders := make([]string, len(columns))
				for j := range placeholders {
					// Check if this value is a RawExpression
					valIndex := i*len(columns) + j
					if valIndex < len(values) {
						if rawExpr, ok := values[valIndex].(RawExpression); ok {
							// Use raw SQL directly with placeholder replacement
							// Replace ? placeholders in raw SQL with dialect-specific placeholders
							rawSQL := rawExpr.SQL
							for _, arg := range rawExpr.Args {
								rawSQL = strings.Replace(rawSQL, "?", placeholderFunc(placeholderIndex), 1)
								placeholderIndex++
								args = append(args, arg)
							}
							placeholders[j] = rawSQL
						} else {
							placeholders[j] = placeholderFunc(placeholderIndex)
							placeholderIndex++
							args = append(args, values[valIndex])
						}
					} else {
						placeholders[j] = placeholderFunc(placeholderIndex)
						placeholderIndex++
					}
				}
				rowPlaceholders[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
			}
			parts = append(parts, strings.Join(rowPlaceholders, ", "))
		} else {
			// Single insert
			placeholders := make([]string, len(columns))
			placeholderIndex := 1
			for i := range placeholders {
				// Check if this value is a RawExpression
				if i < len(values) {
					if rawExpr, ok := values[i].(RawExpression); ok {
						// Use raw SQL directly with placeholder replacement
						rawSQL := rawExpr.SQL
						for _, arg := range rawExpr.Args {
							rawSQL = strings.Replace(rawSQL, "?", placeholderFunc(placeholderIndex), 1)
							placeholderIndex++
							args = append(args, arg)
						}
						placeholders[i] = rawSQL
					} else {
						placeholders[i] = placeholderFunc(placeholderIndex)
						placeholderIndex++
						args = append(args, values[i])
					}
				} else {
					placeholders[i] = placeholderFunc(placeholderIndex)
					placeholderIndex++
				}
			}
			parts = append(parts, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
		}
	}

	return strings.Join(parts, " "), args
}
