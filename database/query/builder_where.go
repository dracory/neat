package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dracory/neat/contracts/database/orm"
)

// convertTimeArgs normalizes every time bind argument to UTC (see
// normalizeTimeArg), so comparisons match the format the write paths store.
// It is idempotent, so re-applying it over already-converted args is safe.
func (b *Builder) convertTimeArgs(args []any) []any {
	converted := make([]any, len(args))
	for i, arg := range args {
		converted[i] = b.normalizeTimeArg(arg)
	}
	return converted
}

// toAnySlice converts a slice of any type to []any.
// Returns ok=false if the value is not a slice.
func toAnySlice(v any) ([]any, bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

// expandInPlaceholders replaces the placeholder in an IN/NOT IN clause with
// multiple placeholders, one per slice element. It handles both "IN (?)" and
// "IN ?" patterns (with or without parentheses).
// Returns the original query unchanged if no IN pattern is found.
func expandInPlaceholders(query string, placeholders []string) string {
	expanded := "(" + strings.Join(placeholders, ", ") + ")"
	// First try the "(?)" pattern (used by WhereIn/WhereNotIn)
	if strings.Contains(query, "(?)") {
		return strings.Replace(query, "(?)", expanded, 1)
	}
	// Then try the " IN ?" or " NOT IN ?" pattern (used by raw Where)
	upper := strings.ToUpper(query)
	if idx := strings.Index(upper, " IN ?"); idx >= 0 {
		// idx+4 is the position of "?"; replace it with the expanded placeholders
		return query[:idx+4] + expanded + query[idx+5:]
	}
	if idx := strings.Index(upper, " NOT IN ?"); idx >= 0 {
		return query[:idx+8] + expanded + query[idx+9:]
	}
	// No IN pattern found — return unchanged to avoid corrupting unrelated placeholders
	return query
}

// buildWheresWithSoftDelete prepends the soft-delete condition when the model implements
// SoftDeleteColumnNamer and neither includeSoftDeleted nor onlySoftDeleted is set.
func (b *Builder) buildWheresWithSoftDelete() (string, []any) {
	return b.buildWheresWithSoftDeleteIndex(1)
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

		// Expand IN (?) / NOT IN (?) when the single arg is a slice of any type.
		clauseQuery := where.query
		clauseArgs := where.args
		if len(clauseArgs) == 1 {
			if slice, ok := toAnySlice(clauseArgs[0]); ok {
				placeholders := make([]string, len(slice))
				for j := range slice {
					placeholders[j] = placeholderFunc(placeholderIndex)
					placeholderIndex++
				}
				clauseQuery = expandInPlaceholders(clauseQuery, placeholders)
				clauseArgs = slice
			}
		}

		// Replace remaining placeholders with dialect-specific ones
		// Count placeholders first to avoid infinite loop if placeholderFunc returns "?"
		placeholderCount := strings.Count(clauseQuery, "?")
		for i := 0; i < placeholderCount; i++ {
			clauseQuery = strings.Replace(clauseQuery, "?", placeholderFunc(placeholderIndex), 1)
			placeholderIndex++
		}

		// Quote identifiers in the WHERE clause
		clauseQuery = b.quoteWhereIdentifiers(clauseQuery)

		parts = append(parts, clauseQuery)
		args = append(args, clauseArgs...)
	}

	return strings.Join(parts, " "), b.convertTimeArgs(args)
}

// buildWheresWithIndex builds the WHERE clause from where clauses with a starting placeholder index.
func (b *Builder) buildWheresWithIndex(startIndex int) (string, []any) {
	var parts []string
	args := make([]any, 0)

	// Get placeholder function for the dialect
	placeholderFunc := func(n int) string { return "?" }
	if b.query.driver != nil {
		placeholderFunc = b.query.driver.Placeholder
	}
	placeholderIndex := startIndex

	for i, where := range b.query.wheres {
		if i > 0 {
			parts = append(parts, strings.ToUpper(where._type))
		}

		// Expand IN (?) / NOT IN (?) when the single arg is a slice of any type.
		clauseQuery := where.query
		clauseArgs := where.args
		if len(clauseArgs) == 1 {
			if slice, ok := toAnySlice(clauseArgs[0]); ok {
				placeholders := make([]string, len(slice))
				for j := range slice {
					placeholders[j] = placeholderFunc(placeholderIndex)
					placeholderIndex++
				}
				clauseQuery = expandInPlaceholders(clauseQuery, placeholders)
				clauseArgs = slice
			}
		}

		// Replace remaining placeholders with dialect-specific ones
		// Count placeholders first to avoid infinite loop if placeholderFunc returns "?"
		placeholderCount := strings.Count(clauseQuery, "?")
		for i := 0; i < placeholderCount; i++ {
			clauseQuery = strings.Replace(clauseQuery, "?", placeholderFunc(placeholderIndex), 1)
			placeholderIndex++
		}

		// Quote identifiers in the WHERE clause
		clauseQuery = b.quoteWhereIdentifiers(clauseQuery)

		parts = append(parts, clauseQuery)
		args = append(args, clauseArgs...)
	}

	return strings.Join(parts, " "), b.convertTimeArgs(args)
}

// buildWheresWithSoftDeleteIndex prepends the soft-delete condition when the model implements
// SoftDeleteColumnNamer and neither includeSoftDeleted nor onlySoftDeleted is set, with a starting placeholder index.
func (b *Builder) buildWheresWithSoftDeleteIndex(startIndex int) (string, []any) {
	var prefix string
	var prefixArgs []any

	if hasSoftDeleteCapability(b.query.model) {
		// Check if model implements SoftDeleteStrategy for custom WHERE conditions
		if strat, ok := b.query.model.(orm.SoftDeleteStrategy); ok {
			switch {
			case b.query.onlySoftDeleted:
				prefix, prefixArgs = strat.SoftDeletedCondition(b.quoteIdentifier)
			case b.query.includeSoftDeleted:
				// include all rows — no filter
			default:
				prefix, prefixArgs = strat.NotSoftDeletedCondition(b.quoteIdentifier)
			}
			// The strategy condition binds one parameter per prefix arg, so startIndex advances accordingly
			if prefix != "" {
				placeholderFunc := func(n int) string { return "?" }
				if b.query.driver != nil {
					placeholderFunc = b.query.driver.Placeholder
				}
				for range prefixArgs {
					prefix = strings.Replace(prefix, "?", placeholderFunc(startIndex), 1)
					startIndex++
				}
			}
		} else {
			// NULL-based strategy (default)
			col := b.quoteIdentifier(getSoftDeleteColumn(b.query.model))
			switch {
			case b.query.onlySoftDeleted:
				prefix = fmt.Sprintf("%s IS NOT NULL", col)
			case b.query.includeSoftDeleted:
				// include all rows — no filter
			default:
				prefix = fmt.Sprintf("%s IS NULL", col)
			}
		}
	}

	if len(b.query.wheres) == 0 {
		return prefix, b.convertTimeArgs(prefixArgs)
	}

	base, args := b.buildWheresWithIndex(startIndex)
	if prefix == "" {
		return base, args
	}
	return prefix + " AND " + base, append(b.convertTimeArgs(prefixArgs), args...)
}
