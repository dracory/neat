package query

import (
	"fmt"
	"strings"
)

// buildWheresWithSoftDelete prepends the soft-delete condition when the model has a
// DeletedAt field and neither WithTrashed nor OnlyTrashed is set.
func (b *Builder) buildWheresWithSoftDelete() (string, []any) {
	var prefix string
	if hasSoftDeleteCapability(b.query.model) {
		switch {
		case b.query.onlyTrashed:
			prefix = fmt.Sprintf("%s IS NOT NULL", b.quoteIdentifier("deleted_at"))
		case b.query.withTrashed:
			// include all rows — no filter
		default:
			prefix = fmt.Sprintf("%s IS NULL", b.quoteIdentifier("deleted_at"))
		}
	}

	if len(b.query.wheres) == 0 {
		return prefix, []any{}
	}

	base, args := b.buildWheres()
	if prefix == "" {
		return base, args
	}
	return prefix + " AND " + base, args
}

// buildWheres builds the WHERE clause from where clauses.
func (b *Builder) buildWheres() (string, []any) {
	var parts []string
	args := make([]any, 0)

	// Get placeholder function for the dialect
	placeholderFunc := func(n int) string { return "?" }
	if b.query.driver != nil {
		placeholderFunc = b.query.driver.Placeholder
	}
	placeholderIndex := 1

	for i, where := range b.query.wheres {
		if i > 0 {
			parts = append(parts, strings.ToUpper(where._type))
		}

		// Expand IN (?) / NOT IN (?) when the single arg is a []any slice.
		clauseQuery := where.query
		clauseArgs := where.args
		if len(clauseArgs) == 1 {
			if slice, ok := clauseArgs[0].([]any); ok {
				placeholders := make([]string, len(slice))
				for j := range slice {
					placeholders[j] = placeholderFunc(placeholderIndex)
					placeholderIndex++
				}
				clauseQuery = strings.Replace(clauseQuery, "(?)", "("+strings.Join(placeholders, ", ")+")", 1)
				clauseArgs = slice
			}
		}

		// Replace remaining placeholders with dialect-specific ones
		if strings.Contains(clauseQuery, "?") {
			clauseQuery = strings.Replace(clauseQuery, "?", placeholderFunc(placeholderIndex), 1)
			placeholderIndex++
		}

		// Quote identifiers in the WHERE clause
		clauseQuery = b.quoteWhereIdentifiers(clauseQuery)

		parts = append(parts, clauseQuery)
		args = append(args, clauseArgs...)
	}

	return strings.Join(parts, " "), args
}
