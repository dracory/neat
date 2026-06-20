package query

import (
	"fmt"
	"strings"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// convertTimeArgs passes time.Time / *time.Time values as-is to the database driver.
// The driver handles time.Time natively, ensuring consistent formatting between
// INSERT and WHERE comparisons. Converting to a string here (e.g. via carbon)
// produces "2006-01-02 15:04:05" which does not match the RFC3339 format
// ("2006-01-02T15:04:05Z") that SQLite stores when time.Time is passed directly,
// causing lexicographic comparison failures in soft-delete filters.
func convertTimeArgs(args []any) []any {
	converted := make([]any, len(args))
	for i, arg := range args {
		if ptr, ok := arg.(*time.Time); ok && ptr != nil {
			converted[i] = *ptr
		} else {
			converted[i] = arg
		}
	}
	return converted
}

// buildWheresWithSoftDelete prepends the soft-delete condition when the model implements
// SoftDeleteColumnNamer and neither includeSoftDeleted nor onlySoftDeleted is set.
func (b *Builder) buildWheresWithSoftDelete() (string, []any) {
	var prefix string
	var prefixArgs []any

	if hasSoftDeleteCapability(b.query.model) {
		// Check if model implements SoftDeleteStrategy for custom WHERE conditions
		if strat, ok := b.query.model.(contractsorm.SoftDeleteStrategy); ok {
			switch {
			case b.query.onlySoftDeleted:
				prefix, prefixArgs = strat.SoftDeletedCondition(b.quoteIdentifier)
			case b.query.includeSoftDeleted:
				// include all rows — no filter
			default:
				prefix, prefixArgs = strat.NotSoftDeletedCondition(b.quoteIdentifier)
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
		return prefix, convertTimeArgs(prefixArgs)
	}

	base, args := b.buildWheres()
	if prefix == "" {
		return base, args
	}
	return prefix + " AND " + base, append(convertTimeArgs(prefixArgs), args...)
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

	return strings.Join(parts, " "), convertTimeArgs(args)
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

	return strings.Join(parts, " "), convertTimeArgs(args)
}

// buildWheresWithSoftDeleteIndex prepends the soft-delete condition when the model implements
// SoftDeleteColumnNamer and neither includeSoftDeleted nor onlySoftDeleted is set, with a starting placeholder index.
func (b *Builder) buildWheresWithSoftDeleteIndex(startIndex int) (string, []any) {
	var prefix string
	var prefixArgs []any

	if hasSoftDeleteCapability(b.query.model) {
		// Check if model implements SoftDeleteStrategy for custom WHERE conditions
		if strat, ok := b.query.model.(contractsorm.SoftDeleteStrategy); ok {
			switch {
			case b.query.onlySoftDeleted:
				prefix, prefixArgs = strat.SoftDeletedCondition(b.quoteIdentifier)
			case b.query.includeSoftDeleted:
				// include all rows — no filter
			default:
				prefix, prefixArgs = strat.NotSoftDeletedCondition(b.quoteIdentifier)
			}
			// For max-date strategy, we have 1 bind parameter, so startIndex needs adjustment
			if prefix != "" {
				// Replace ? with proper placeholder for the soft delete condition
				placeholderFunc := func(n int) string { return "?" }
				if b.query.driver != nil {
					placeholderFunc = b.query.driver.Placeholder
				}
				prefix = strings.Replace(prefix, "?", placeholderFunc(startIndex), 1)
				startIndex++
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
		return prefix, convertTimeArgs(prefixArgs)
	}

	base, args := b.buildWheresWithIndex(startIndex)
	if prefix == "" {
		return base, args
	}
	return prefix + " AND " + base, append(convertTimeArgs(prefixArgs), args...)
}
