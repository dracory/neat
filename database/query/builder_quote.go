package query

import (
	"fmt"
	"strings"
)

// quoteIdentifier wraps an identifier in the appropriate quotes for the dialect.
func (b *Builder) quoteIdentifier(name string) string {
	if b.query.driver == nil || name == "*" || name == "" {
		return name
	}

	dialect := b.query.driver.Dialect()
	quoteChar := "\""
	if dialect == "mysql" {
		quoteChar = "`"
	}

	// If already quoted, return as is
	if (strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"")) ||
		(strings.HasPrefix(name, "`") && strings.HasSuffix(name, "`")) {
		return name
	}

	// Handle dotted names (e.g., table.column)
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		for i, part := range parts {
			parts[i] = b.quoteIdentifier(part)
		}
		return strings.Join(parts, ".")
	}

	// Handle " AS " alias (case insensitive)
	upperName := strings.ToUpper(name)
	if idx := strings.Index(upperName, " AS "); idx != -1 {
		identifier := strings.TrimSpace(name[:idx])
		alias := strings.TrimSpace(name[idx+4:])
		return fmt.Sprintf("%s AS %s", b.quoteIdentifier(identifier), b.quoteIdentifier(alias))
	}

	// Handle space alias (e.g., "users u")
	// Only apply if not already handled by AS check above
	if idx := strings.Index(name, " "); idx != -1 {
		upperAfterSpace := strings.ToUpper(name[idx+1:])
		// If the part after space starts with "AS ", it should have been handled above
		if !strings.HasPrefix(upperAfterSpace, "AS ") {
			identifier := strings.TrimSpace(name[:idx])
			alias := strings.TrimSpace(name[idx+1:])
			return fmt.Sprintf("%s %s", b.quoteIdentifier(identifier), b.quoteIdentifier(alias))
		}
	}

	return fmt.Sprintf("%s%s%s", quoteChar, name, quoteChar)
}

// quoteWhereIdentifiers quotes column names in WHERE clauses.
// It uses a conservative approach that only quotes simple identifiers
// to avoid breaking complex expressions, function calls, or subqueries.
func (b *Builder) quoteWhereIdentifiers(query string) string {
	// SQL keywords that should never be quoted
	sqlKeywords := map[string]bool{
		"AND": true, "OR": true, "NOT": true, "NULL": true,
		"TRUE": true, "FALSE": true, "IS": true, "IN": true,
		"LIKE": true, "BETWEEN": true, "SELECT": true, "FROM": true,
		"WHERE": true, "JOIN": true, "ON": true, "AS": true,
		"GROUP": true, "ORDER": true, "BY": true, "HAVING": true,
		"LIMIT": true, "OFFSET": true, "CASE": true, "WHEN": true,
		"THEN": true, "ELSE": true, "END": true, "EXISTS": true,
	}

	// Collect all replacements first, then apply them in reverse order
	type replacement struct {
		start int
		end   int
		value string
	}
	var replacements []replacement

	// Tokenize the query to identify potential column names
	// We look for simple identifiers that appear before operators
	// Only use operators with leading space to avoid matching inside quoted identifiers
	operators := []string{" != ", " <> ", " >= ", " <= ", " LIKE ", " NOT LIKE ", " = ", " > ", " < ", " IS ", " IS NOT ", " IN ", " NOT IN ", " BETWEEN ", " NOT BETWEEN "}

	for _, op := range operators {
		start := 0
		for {
			idx := strings.Index(strings.ToUpper(query[start:]), strings.ToUpper(op))
			if idx == -1 {
				break
			}
			idx += start

			// Get text before operator
			beforeOp := query[:idx]
			trimmed := strings.TrimSpace(beforeOp)

			// Get last word (potential column name)
			lastSpace := strings.LastIndex(trimmed, " ")
			var colName string
			if lastSpace == -1 {
				colName = trimmed
			} else {
				colName = trimmed[lastSpace+1:]
			}

			// Only quote if it's a simple identifier:
			// - Not already quoted
			// - Not a SQL keyword
			// - Contains only alphanumeric characters and underscores
			// - Doesn't contain dots (table.column handled separately)
			// - Doesn't contain parentheses (function calls)
			// - Doesn't start with a number
			if colName != "" &&
				!strings.HasPrefix(colName, "\"") && !strings.HasPrefix(colName, "`") &&
				!sqlKeywords[strings.ToUpper(colName)] &&
				isSimpleIdentifier(colName) {
				quotedCol := b.quoteIdentifier(colName)
				// Find the last occurrence of colName in beforeOp
				colIdx := strings.LastIndex(beforeOp, colName)
				if colIdx != -1 {
					replacements = append(replacements, replacement{
						start: colIdx,
						end:   colIdx + len(colName),
						value: quotedCol,
					})
				}
			}
			start = idx + len(op)
		}
	}

	// Sort replacements by start position descending
	for i := 0; i < len(replacements); i++ {
		for j := i + 1; j < len(replacements); j++ {
			if replacements[i].start < replacements[j].start {
				replacements[i], replacements[j] = replacements[j], replacements[i]
			}
		}
	}

	// Apply replacements from end to start
	result := query
	for _, r := range replacements {
		result = result[:r.start] + r.value + result[r.end:]
	}

	return result
}
