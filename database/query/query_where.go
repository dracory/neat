package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dracory/neat/contracts/database/orm"
)

// WhereIn adds a where in clause to the query.
func (q *Query) WhereIn(column string, values []any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IN (?)", NewBuilder(q).quoteIdentifier(column)), args: []any{values}})
	return q
}

// WhereNotIn adds a where not in clause to the query.
func (q *Query) WhereNotIn(column string, values []any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s NOT IN (?)", NewBuilder(q).quoteIdentifier(column)), args: []any{values}})
	return q
}

// OrWhereIn adds an or where in clause to the query.
func (q *Query) OrWhereIn(column string, values []any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s IN (?)", NewBuilder(q).quoteIdentifier(column)), args: []any{values}})
	return q
}

// OrWhereNotIn adds an or where not in clause to the query.
func (q *Query) OrWhereNotIn(column string, values []any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s NOT IN (?)", NewBuilder(q).quoteIdentifier(column)), args: []any{values}})
	return q
}

// WhereBetween adds a where between clause to the query.
func (q *Query) WhereBetween(column string, x, y any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s BETWEEN ? AND ?", NewBuilder(q).quoteIdentifier(column)), args: []any{x, y}})
	return q
}

// WhereNotBetween adds a where not between clause to the query.
func (q *Query) WhereNotBetween(column string, x, y any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s NOT BETWEEN ? AND ?", NewBuilder(q).quoteIdentifier(column)), args: []any{x, y}})
	return q
}

// OrWhereBetween adds an or where between clause to the query.
func (q *Query) OrWhereBetween(column string, x, y any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s BETWEEN ? AND ?", NewBuilder(q).quoteIdentifier(column)), args: []any{x, y}})
	return q
}

// OrWhereNotBetween adds an or where not between clause to the query.
func (q *Query) OrWhereNotBetween(column string, x, y any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s NOT BETWEEN ? AND ?", NewBuilder(q).quoteIdentifier(column)), args: []any{x, y}})
	return q
}

// WhereNull adds a where null clause to the query.
func (q *Query) WhereNull(column string) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IS NULL", NewBuilder(q).quoteIdentifier(column)), args: nil})
	return q
}

// WhereNotNull adds a where not null clause to the query.
func (q *Query) WhereNotNull(column string) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IS NOT NULL", NewBuilder(q).quoteIdentifier(column)), args: nil})
	return q
}

// OrWhereNull adds an or where null clause to the query.
func (q *Query) OrWhereNull(column string) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s IS NULL", NewBuilder(q).quoteIdentifier(column)), args: nil})
	return q
}

// WhereColumn adds a where column clause to the query.
func (q *Query) WhereColumn(first, operator, second string) orm.Query {
	// Validate column names
	if !isSimpleIdentifier(first) || !isSimpleIdentifier(second) {
		return q
	}
	// Validate operator against allowed whitelist
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
	}
	if !allowedOperators[operator] {
		return q
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s %s %s", first, operator, second), args: nil})
	return q
}

// OrWhereColumn adds an or where column clause to the query.
func (q *Query) OrWhereColumn(first, operator, second string) orm.Query {
	// Validate column names
	if !isSimpleIdentifier(first) || !isSimpleIdentifier(second) {
		return q
	}
	// Validate operator against allowed whitelist
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
	}
	if !allowedOperators[operator] {
		return q
	}
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s %s %s", first, operator, second), args: nil})
	return q
}

// WhereExists adds a where exists clause to the query.
func (q *Query) WhereExists(callback func(orm.Query) orm.Query) orm.Query {
	subQ := q.Clone().(*Query)
	subQ = callback(subQ).(*Query)
	subSQL, subArgs := NewBuilder(subQ).BuildSelect()
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("EXISTS (%s)", subSQL), args: subArgs})
	return q
}

// WhereNot adds a where not clause to the query.
func (q *Query) WhereNot(query any, args ...any) orm.Query {
	// Handle closure callback - wrap conditions in NOT
	if fn, ok := query.(func(orm.Query) orm.Query); ok {
		subQ := q.Clone().(*Query)
		subQ = fn(subQ).(*Query)
		// Wrap each where clause in NOT and add to main query
		for i := range subQ.wheres {
			wc := &subQ.wheres[i]
			wc.query = fmt.Sprintf("NOT (%s)", wc.query)
		}
		q.wheres = append(q.wheres, subQ.wheres...)
	} else {
		// For raw query strings, ensure proper parameterization
		// The query string should use ? placeholders for parameters
		queryStr := fmt.Sprintf("%v", query)
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT (%s)", queryStr), args: args})
	}
	return q
}

// OrWhereNot adds an or where not clause to the query.
func (q *Query) OrWhereNot(query any, args ...any) orm.Query {
	// Handle closure callback - wrap conditions in NOT
	if fn, ok := query.(func(orm.Query) orm.Query); ok {
		subQ := q.Clone().(*Query)
		subQ = fn(subQ).(*Query)
		// Wrap each where clause in NOT and add to main query as OR
		for i := range subQ.wheres {
			wc := &subQ.wheres[i]
			wc.query = fmt.Sprintf("NOT (%s)", wc.query)
			wc._type = "or"
		}
		q.wheres = append(q.wheres, subQ.wheres...)
	} else {
		// For raw query strings, ensure proper parameterization
		// The query string should use ? placeholders for parameters
		queryStr := fmt.Sprintf("%v", query)
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT (%s)", queryStr), args: args})
	}
	return q
}

// WhereAny adds a where any clause to the query.
func (q *Query) WhereAny(columns []string, operator string, value any) orm.Query {
	if len(columns) == 0 {
		// Return without adding a clause for empty columns
		return q
	}
	var parts []string
	var args []any
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("%s %s ?", col, operator))
		args = append(args, value)
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: "(" + strings.Join(parts, " OR ") + ")", args: args})
	return q
}

// WhereAll adds a where all clause to the query.
func (q *Query) WhereAll(columns []string, operator string, value any) orm.Query {
	if len(columns) == 0 {
		// Return without adding a clause for empty columns
		return q
	}
	var parts []string
	var args []any
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("%s %s ?", col, operator))
		args = append(args, value)
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: "(" + strings.Join(parts, " AND ") + ")", args: args})
	return q
}

// WhereNone adds a where none clause to the query.
func (q *Query) WhereNone(columns []string, operator string, value any) orm.Query {
	if len(columns) == 0 {
		// Return without adding a clause for empty columns
		return q
	}
	var parts []string
	var args []any
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("%s %s ?", col, operator))
		args = append(args, value)
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: "NOT (" + strings.Join(parts, " OR ") + ")", args: args})
	return q
}

// splitJsonColumn splits a JSON column name into column and path.
func (q *Query) splitJsonColumn(column string) (string, string) {
	if !strings.Contains(column, "->") {
		return column, ""
	}
	parts := strings.SplitN(column, "->", 2)
	if len(parts) < 2 {
		return column, ""
	}
	// Replace all -> with . for SQLite JSON path syntax
	path := "." + strings.ReplaceAll(parts[1], "->", ".")
	return parts[0], path
}

// splitJsonColumnForMySQL splits a JSON column name into column and path for MySQL.
// MySQL uses $.path for object paths and $[index] for array indexing.
func (q *Query) splitJsonColumnForMySQL(column string) (string, string) {
	if !strings.Contains(column, "->") {
		return column, ""
	}
	parts := strings.SplitN(column, "->", 2)
	if len(parts) < 2 {
		return column, ""
	}
	// Convert -> to MySQL JSON path syntax
	// For array indices like tags->0, convert to $.tags[0]
	// For object paths like meta->id, convert to $.meta.id
	path := strings.ReplaceAll(parts[1], "->", ".")
	// Handle array indexing: replace .0. with [0]., .1. with [1]., etc.
	re := regexp.MustCompile(`\.(\d+)\.`)
	path = re.ReplaceAllString(path, "[$1].")
	// Handle trailing array index: replace .0 with [0]
	re = regexp.MustCompile(`\.(\d+)$`)
	path = re.ReplaceAllString(path, "[$1]")
	return parts[0], "$." + path
}

// splitJsonColumnForPostgreSQL splits a JSON column name into column and path for PostgreSQL.
// PostgreSQL uses -> for JSON object access and ->> for text extraction.
func (q *Query) splitJsonColumnForPostgreSQL(column string) (string, string) {
	if !strings.Contains(column, "->") {
		return column, ""
	}
	parts := strings.SplitN(column, "->", 2)
	if len(parts) < 2 {
		return column, ""
	}
	// Convert -> to PostgreSQL JSON path syntax
	// For array indices like tags->0, convert to ->0
	// For object paths like meta->id, convert to ->'meta'->'id'
	pathParts := strings.Split(parts[1], "->")
	var pathBuilder strings.Builder
	for _, part := range pathParts {
		// Check if it's a numeric index (array)
		if _, err := strconv.Atoi(part); err == nil {
			// It's a number, don't quote it
			pathBuilder.WriteString("->")
			pathBuilder.WriteString(part)
		} else {
			// It's a string key, quote it
			pathBuilder.WriteString("->'")
			pathBuilder.WriteString(part)
			pathBuilder.WriteString("'")
		}
	}
	return parts[0], pathBuilder.String()
}

// WhereJsonContains adds a where json contains clause to the query.
func (q *Query) WhereJsonContains(column string, value any) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_extract(%s, '$%s') = ?", col, path), args: []any{value}})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumnForMySQL(column)
		// MySQL's JSON_CONTAINS requires the value to be valid JSON
		// If it's a string, wrap it in quotes to make it a JSON string
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// MySQL: JSON_CONTAINS(column, value, path)
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?, '%s')", col, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", col), args: []any{jsonValue}})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		// PostgreSQL uses @> operator for JSONB containment
		// Convert value to JSON format
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// PostgreSQL: column->path @> value (use ->> for text extraction if needed)
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s%s @> ?::jsonb", quotedCol, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s @> ?::jsonb", quotedCol), args: []any{jsonValue}})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	}
	return q
}

// OrWhereJsonContains adds an or where json contains clause to the query.
func (q *Query) OrWhereJsonContains(column string, value any) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_extract(%s, '$%s') = ?", col, path), args: []any{value}})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumnForMySQL(column)
		// MySQL's JSON_CONTAINS requires the value to be valid JSON
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// MySQL: JSON_CONTAINS(column, value, path)
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?, '%s')", col, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", col), args: []any{jsonValue}})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		// PostgreSQL uses @> operator for JSONB containment
		// Convert value to JSON format
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// PostgreSQL: column->path @> value (use ->> for text extraction if needed)
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s%s @> ?::jsonb", quotedCol, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s @> ?::jsonb", quotedCol), args: []any{jsonValue}})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	}
	return q
}

// WhereJsonDoesntContain adds a where json doesnt contain clause to the query.
func (q *Query) WhereJsonDoesntContain(column string, value any) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_extract(%s, '$%s') != ?", col, path), args: []any{value}})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumnForMySQL(column)
		// MySQL's JSON_CONTAINS requires the value to be valid JSON
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// MySQL: NOT JSON_CONTAINS(column, value, path)
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?, '%s')", col, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", col), args: []any{jsonValue}})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		// PostgreSQL uses @> operator for JSONB containment
		// Convert value to JSON format
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// PostgreSQL: NOT (column->path @> value)
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT (%s%s @> ?::jsonb)", quotedCol, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT (%s @> ?::jsonb)", quotedCol), args: []any{jsonValue}})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	}
	return q
}

// OrWhereJsonDoesntContain adds an or where json doesnt contain clause to the query.
func (q *Query) OrWhereJsonDoesntContain(column string, value any) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_extract(%s, '$%s') != ?", col, path), args: []any{value}})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumnForMySQL(column)
		// MySQL's JSON_CONTAINS requires the value to be valid JSON
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// MySQL: NOT JSON_CONTAINS(column, value, path)
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?, '%s')", col, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", col), args: []any{jsonValue}})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		// PostgreSQL uses @> operator for JSONB containment
		// Convert value to JSON format
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// PostgreSQL: NOT (column->path @> value)
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT (%s%s @> ?::jsonb)", quotedCol, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT (%s @> ?::jsonb)", quotedCol), args: []any{jsonValue}})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", column), args: []any{value}})
	}
	return q
}

// WhereJsonContainsKey adds a where json contains key clause to the query.
func (q *Query) WhereJsonContainsKey(column string) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_type(%s, '$%s') IS NOT NULL", col, path), args: nil})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumn(column)
		if path != "" {
			// MySQL: JSON_CONTAINS_PATH(column, 'one', path)
			jsonPath := "$" + path
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '%s')", col, jsonPath), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.*')", col), args: nil})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: check if path exists using -> operator with IS NOT NULL
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s%s IS NOT NULL", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IS NOT NULL AND %s::text != ''", quotedCol, quotedCol), args: nil})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	}
	return q
}

// OrWhereJsonContainsKey adds an or where json contains key clause to the query.
func (q *Query) OrWhereJsonContainsKey(column string) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_type(%s, '$%s') IS NOT NULL", col, path), args: nil})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumn(column)
		if path != "" {
			// MySQL: JSON_CONTAINS_PATH(column, 'one', path)
			jsonPath := "$" + path
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '%s')", col, jsonPath), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.*')", col), args: nil})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: check if path exists using -> operator with IS NOT NULL
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s%s IS NOT NULL", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s IS NOT NULL AND %s::text != ''", quotedCol, quotedCol), args: nil})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	}
	return q
}

// WhereJsonDoesntContainKey adds a where json doesnt contain key clause to the query.
func (q *Query) WhereJsonDoesntContainKey(column string) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_type(%s, '$%s') IS NULL", col, path), args: nil})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumn(column)
		if path != "" {
			// MySQL: NOT JSON_CONTAINS_PATH(column, 'one', path)
			jsonPath := "$" + path
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '%s')", col, jsonPath), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '$.*')", col), args: nil})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: check if path doesn't exist using -> operator with IS NULL
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s%s IS NULL", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IS NULL OR %s::text = ''", quotedCol, quotedCol), args: nil})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	}
	return q
}

// OrWhereJsonDoesntContainKey adds an or where json doesnt contain key clause to the query.
func (q *Query) OrWhereJsonDoesntContainKey(column string) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_type(%s, '$%s') IS NULL", col, path), args: nil})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumn(column)
		if path != "" {
			// MySQL: NOT JSON_CONTAINS_PATH(column, 'one', path)
			jsonPath := "$" + path
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '%s')", col, jsonPath), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '$.*')", col), args: nil})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: check if path doesn't exist using -> operator with IS NULL
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s%s IS NULL", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s IS NULL OR %s::text = ''", quotedCol, quotedCol), args: nil})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, '$.%s')", column, column), args: nil})
	}
	return q
}

// WhereJsonLength adds a where json length clause to the query.
func (q *Query) WhereJsonLength(column string, operator string, value any) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_array_length(%s, '$%s') %s ?", col, path, operator), args: []any{value}})
	} else if q.driver != nil && q.driver.Dialect() == "mysql" {
		col, path := q.splitJsonColumn(column)
		if path != "" {
			// MySQL: JSON_LENGTH(column, path)
			jsonPath := "$" + path
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s, '%s') %s ?", col, jsonPath, operator), args: []any{value}})
		} else {
			// No path specified, get length of entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s) %s ?", col, operator), args: []any{value}})
		}
	} else if q.driver != nil && q.driver.Dialect() == "postgres" {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		builder := NewBuilder(q)
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: jsonb_array_length(column->path)
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("jsonb_array_length(%s%s) %s ?", quotedCol, path, operator), args: []any{value}})
		} else {
			// No path specified, get length of entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("jsonb_array_length(%s) %s ?", quotedCol, operator), args: []any{value}})
		}
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s) %s ?", column, operator), args: []any{value}})
	}
	return q
}
