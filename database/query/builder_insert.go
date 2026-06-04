package query

import (
	"fmt"
	"reflect"
	"strings"
)

// BuildInsert builds an INSERT query from the query state.
func (b *Builder) BuildInsert(value any) (string, []any) {
	var parts []string
	var args []any = []any{}

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
		return "", []any{}
	}

	// Return empty if no columns were extracted
	if len(columns) == 0 {
		return "", []any{}
	}

	// Check if this is SQL Server and we need OUTPUT clause for LastInsertId
	isSQLServer := b.query.isSQLServer()
	var idColumn string
	if isSQLServer {
		// Find the ID column for OUTPUT clause (check for common primary key names)
		for _, col := range columns {
			lowerCol := strings.ToLower(col)
			if lowerCol == "id" || lowerCol == "user_id" || lowerCol == "post_id" || lowerCol == "uuid" {
				idColumn = b.quoteIdentifier(col)
				break
			}
		}
		// If no common ID column found, default to "id" (may fail if table doesn't have it)
		if idColumn == "" {
			idColumn = b.quoteIdentifier("id")
		}
	}

	if len(columns) > 0 {
		// Quote column names
		quotedColumns := make([]string, len(columns))
		for i, col := range columns {
			quotedColumns[i] = b.quoteIdentifier(col)
		}

		// Add column list
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(quotedColumns, ", ")))

		// Add OUTPUT clause for SQL Server if we have an ID column (must come after column list)
		if isSQLServer && idColumn != "" {
			parts = append(parts, fmt.Sprintf("OUTPUT INSERTED.%s", idColumn))
		}

		// Get placeholder function for the dialect
		placeholderFunc := func(n int) string { return "?" }
		if b.query.driver != nil {
			placeholderFunc = b.query.driver.Placeholder
		}

		// Check if this is Oracle for INSERT ALL syntax
		isOracle := b.query.isOracle()

		// Handle bulk insert
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			// Oracle doesn't support multi-row INSERT VALUES syntax with identity columns
			// Fall back to single inserts for Oracle to avoid sequence issues
			if isOracle {
				// For Oracle, return empty string to signal that we should handle bulk inserts differently
				// The Create method will iterate and insert one by one
				return "", []any{}
			} else {
				// Standard multi-row INSERT
				parts = append(parts, "VALUES")
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
			}
		} else {
			// Single insert
			parts = append(parts, "VALUES")
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
