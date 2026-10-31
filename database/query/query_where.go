package query

import (
	"fmt"
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

// escapeLikePattern escapes LIKE wildcards (%, _) and the escape
// character itself (!) so the value is matched literally. A non-backslash
// escape character is used because backslash literals are
// dialect-dependent (e.g. MySQL string escaping).
func escapeLikePattern(value string) string {
	value = strings.ReplaceAll(value, `!`, `!!`)
	value = strings.ReplaceAll(value, `%`, `!%`)
	return strings.ReplaceAll(value, `_`, `!_`)
}

// whereLike appends a LIKE-based where clause. When escape is true the
// clause gets an ESCAPE '!' suffix and the pattern is expected to have
// its wildcards pre-escaped by the caller.
func (q *Query) whereLike(_type, column, pattern string, not, escape bool) orm.Query {
	if !isValidColumnReference(column) {
		if q.buildError == nil {
			q.buildError = fmt.Errorf("invalid column name in WhereLike: column names must be valid identifiers")
		}
		return q
	}
	op := "LIKE"
	if not {
		op = "NOT LIKE"
	}
	clause := fmt.Sprintf("%s %s ?", NewBuilder(q).quoteIdentifier(column), op)
	if escape {
		clause += ` ESCAPE '!'`
	}
	q.wheres = append(q.wheres, whereClause{_type: _type, query: clause, args: []any{pattern}})
	return q
}

// WhereLike adds a where like clause to the query.
// The pattern may contain LIKE wildcards (% and _).
func (q *Query) WhereLike(column, pattern string) orm.Query {
	return q.whereLike("and", column, pattern, false, false)
}

// OrWhereLike adds an or where like clause to the query.
// The pattern may contain LIKE wildcards (% and _).
func (q *Query) OrWhereLike(column, pattern string) orm.Query {
	return q.whereLike("or", column, pattern, false, false)
}

// WhereNotLike adds a where not like clause to the query.
// The pattern may contain LIKE wildcards (% and _).
func (q *Query) WhereNotLike(column, pattern string) orm.Query {
	return q.whereLike("and", column, pattern, true, false)
}

// OrWhereNotLike adds an or where not like clause to the query.
// The pattern may contain LIKE wildcards (% and _).
func (q *Query) OrWhereNotLike(column, pattern string) orm.Query {
	return q.whereLike("or", column, pattern, true, false)
}

// WhereStartsWith adds a where clause matching values that start with
// the given prefix. LIKE wildcards in the value are escaped, so the
// value is matched literally.
func (q *Query) WhereStartsWith(column, value string) orm.Query {
	return q.whereLike("and", column, escapeLikePattern(value)+"%", false, true)
}

// OrWhereStartsWith adds an or where clause matching values that start
// with the given prefix.
func (q *Query) OrWhereStartsWith(column, value string) orm.Query {
	return q.whereLike("or", column, escapeLikePattern(value)+"%", false, true)
}

// WhereEndsWith adds a where clause matching values that end with the
// given suffix. LIKE wildcards in the value are escaped, so the value
// is matched literally.
func (q *Query) WhereEndsWith(column, value string) orm.Query {
	return q.whereLike("and", column, "%"+escapeLikePattern(value), false, true)
}

// OrWhereEndsWith adds an or where clause matching values that end with
// the given suffix.
func (q *Query) OrWhereEndsWith(column, value string) orm.Query {
	return q.whereLike("or", column, "%"+escapeLikePattern(value), false, true)
}

// WhereContains adds a where clause matching values that contain the
// given substring. LIKE wildcards in the value are escaped, so the
// value is matched literally.
func (q *Query) WhereContains(column, value string) orm.Query {
	return q.whereLike("and", column, "%"+escapeLikePattern(value)+"%", false, true)
}

// OrWhereContains adds an or where clause matching values that contain
// the given substring.
func (q *Query) OrWhereContains(column, value string) orm.Query {
	return q.whereLike("or", column, "%"+escapeLikePattern(value)+"%", false, true)
}

// WhereNotContains adds a where clause matching values that do not
// contain the given substring. LIKE wildcards in the value are escaped,
// so the value is matched literally.
func (q *Query) WhereNotContains(column, value string) orm.Query {
	return q.whereLike("and", column, "%"+escapeLikePattern(value)+"%", true, true)
}

// OrWhereNotContains adds an or where clause matching values that do
// not contain the given substring.
func (q *Query) OrWhereNotContains(column, value string) orm.Query {
	return q.whereLike("or", column, "%"+escapeLikePattern(value)+"%", true, true)
}

// WhereColumn adds a where column clause to the query.
func (q *Query) WhereColumn(first, operator, second string) orm.Query {
	// Validate column names (allow table.column dotted references)
	if !isValidColumnReference(first) || !isValidColumnReference(second) {
		if q.buildError == nil {
			q.buildError = fmt.Errorf("invalid column name in WhereColumn: column names must be valid identifiers")
		}
		return q
	}
	// Validate operator against allowed whitelist
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
	}
	if !allowedOperators[operator] {
		if q.buildError == nil {
			q.buildError = fmt.Errorf("invalid operator in WhereColumn: %s", operator)
		}
		return q
	}
	// Quote both sides at construction: the right-hand column would not be
	// quoted by quoteWhereIdentifiers (it only quotes the word before the
	// operator), and a reserved-word column needs quoting to be valid SQL.
	builder := NewBuilder(q)
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s %s %s", builder.quoteIdentifier(first), operator, builder.quoteIdentifier(second)), args: nil})
	return q
}

// OrWhereColumn adds an or where column clause to the query.
func (q *Query) OrWhereColumn(first, operator, second string) orm.Query {
	// Validate column names (allow table.column dotted references)
	if !isValidColumnReference(first) || !isValidColumnReference(second) {
		q.buildError = fmt.Errorf("invalid column name in OrWhereColumn: column names must be valid identifiers")
		return q
	}
	// Validate operator against allowed whitelist
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
	}
	if !allowedOperators[operator] {
		q.buildError = fmt.Errorf("invalid operator in OrWhereColumn: %s", operator)
		return q
	}
	builder := NewBuilder(q)
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s %s %s", builder.quoteIdentifier(first), operator, builder.quoteIdentifier(second)), args: nil})
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

// Exclude is an alias for WhereNot, providing Django-style syntax.
// Supports Django-style syntax: Exclude("column", "value") automatically uses = operator.
func (q *Query) Exclude(query any, args ...any) orm.Query {
	return q.WhereNot(query, args...)
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
	// Validate operator against allowed whitelist (case-insensitive)
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
		"LIKE": true, "NOT LIKE": true, "ILIKE": true, "NOT ILIKE": true,
	}
	normOp := strings.ToUpper(strings.TrimSpace(operator))
	if !allowedOperators[normOp] {
		if q.buildError == nil {
			q.buildError = fmt.Errorf("invalid operator in WhereAny: %s", operator)
		}
		return q
	}
	var parts []string
	var args []any
	builder := NewBuilder(q)
	for _, col := range columns {
		if !isTableIdentifier(col) {
			if q.buildError == nil {
				q.buildError = fmt.Errorf("invalid column name in WhereAny: %s", col)
			}
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		parts = append(parts, fmt.Sprintf("%s %s ?", quotedCol, normOp))
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
	// Validate operator against allowed whitelist (case-insensitive)
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
		"LIKE": true, "NOT LIKE": true, "ILIKE": true, "NOT ILIKE": true,
	}
	normOp := strings.ToUpper(strings.TrimSpace(operator))
	if !allowedOperators[normOp] {
		if q.buildError == nil {
			q.buildError = fmt.Errorf("invalid operator in WhereAll: %s", operator)
		}
		return q
	}
	var parts []string
	var args []any
	builder := NewBuilder(q)
	for _, col := range columns {
		if !isTableIdentifier(col) {
			if q.buildError == nil {
				q.buildError = fmt.Errorf("invalid column name in WhereAll: %s", col)
			}
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		parts = append(parts, fmt.Sprintf("%s %s ?", quotedCol, normOp))
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
	// Validate operator against allowed whitelist (case-insensitive)
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
		"LIKE": true, "NOT LIKE": true, "ILIKE": true, "NOT ILIKE": true,
	}
	normOp := strings.ToUpper(strings.TrimSpace(operator))
	if !allowedOperators[normOp] {
		if q.buildError == nil {
			q.buildError = fmt.Errorf("invalid operator in WhereNone: %s", operator)
		}
		return q
	}
	var parts []string
	var args []any
	builder := NewBuilder(q)
	for _, col := range columns {
		if !isTableIdentifier(col) {
			if q.buildError == nil {
				q.buildError = fmt.Errorf("invalid column name in WhereNone: %s", col)
			}
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		parts = append(parts, fmt.Sprintf("%s %s ?", quotedCol, normOp))
		args = append(args, value)
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: "NOT (" + strings.Join(parts, " OR ") + ")", args: args})
	return q
}

// isValidJSONPathSegment reports whether s is a safe JSON path segment:
// a simple identifier optionally followed by [n] array indexes ("items",
// "items[0]", "items[0][2]") or a pure array index ("0"). The admitted
// character set contains no quotes or separators, so a validated segment
// can be embedded in a JSON path literal safely.
func isValidJSONPathSegment(s string) bool {
	if s == "" {
		return false
	}
	// Pure numeric array index
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	head, indexes := s, ""
	if i := strings.IndexByte(s, '['); i != -1 {
		head, indexes = s[:i], s[i:]
	}
	if !isTableIdentifier(head) {
		return false
	}
	// indexes must be a sequence of [n] groups
	for indexes != "" {
		if indexes[0] != '[' {
			return false
		}
		end := strings.IndexByte(indexes, ']')
		if end < 2 {
			return false
		}
		if _, err := strconv.Atoi(indexes[1:end]); err != nil {
			return false
		}
		indexes = indexes[end+1:]
	}
	return true
}

// jsonPathSegmentParts splits a validated segment into its identifier head
// and its array indexes: "items[0][2]" -> ("items", ["0","2"]) and a pure
// index "0" -> ("", ["0"]). Input must satisfy isValidJSONPathSegment.
func jsonPathSegmentParts(segment string) (head string, indexes []string) {
	if i := strings.IndexByte(segment, '['); i != -1 {
		head = segment[:i]
		rest := segment[i:]
		for rest != "" {
			end := strings.IndexByte(rest, ']')
			indexes = append(indexes, rest[1:end])
			rest = rest[end+1:]
		}
		return head, indexes
	}
	if _, err := strconv.Atoi(segment); err == nil {
		return "", []string{segment}
	}
	return segment, nil
}

// jsonPathSegments splits a JSON column ("data->meta->active") into the
// column reference and the path segments, validating each: the column must
// be a valid column reference and every segment a valid JSON path segment.
// On invalid input it sets q.buildError and returns ok=false; callers must
// emit no clause then.
func (q *Query) jsonPathSegments(column string) (col string, segments []string, ok bool) {
	parts := strings.Split(column, "->")
	if !isValidColumnReference(parts[0]) {
		if q.buildError == nil {
			q.buildError = fmt.Errorf("invalid JSON column name: %q", parts[0])
		}
		return "", nil, false
	}
	for _, segment := range parts[1:] {
		if !isValidJSONPathSegment(segment) {
			if q.buildError == nil {
				q.buildError = fmt.Errorf("invalid JSON path segment: %q", segment)
			}
			return "", nil, false
		}
	}
	return parts[0], parts[1:], true
}

// splitJsonColumn splits a JSON column name into a validated column and a
// SQLite JSON path ("$.a.b"). Returns empty strings on invalid input
// (q.buildError is set); callers must emit no clause then.
func (q *Query) splitJsonColumn(column string) (string, string) {
	col, segments, ok := q.jsonPathSegments(column)
	if !ok {
		return "", ""
	}
	if len(segments) == 0 {
		return col, ""
	}
	return col, "$." + strings.Join(segments, ".")
}

// splitJsonColumnForMySQL splits a JSON column name into a validated column
// and a MySQL JSON path. Numeric segments and [n] indexes emit bracket
// syntax ("$[0]", "$.a[0]"). Returns empty strings on invalid input
// (q.buildError is set); callers must emit no clause then.
func (q *Query) splitJsonColumnForMySQL(column string) (string, string) {
	col, segments, ok := q.jsonPathSegments(column)
	if !ok {
		return "", ""
	}
	if len(segments) == 0 {
		return col, ""
	}
	var path strings.Builder
	path.WriteString("$")
	for _, segment := range segments {
		head, indexes := jsonPathSegmentParts(segment)
		if head != "" {
			path.WriteString("." + head)
		}
		for _, index := range indexes {
			path.WriteString("[" + index + "]")
		}
	}
	return col, path.String()
}

// splitJsonColumnForPostgreSQL splits a JSON column name into a validated
// column and a PostgreSQL JSON path ("->'a'->0"). Returns empty strings on
// invalid input (q.buildError is set); callers must emit no clause then.
func (q *Query) splitJsonColumnForPostgreSQL(column string) (string, string) {
	col, segments, ok := q.jsonPathSegments(column)
	if !ok {
		return "", ""
	}
	var pathBuilder strings.Builder
	for _, segment := range segments {
		head, indexes := jsonPathSegmentParts(segment)
		if head != "" {
			pathBuilder.WriteString("->'")
			pathBuilder.WriteString(head)
			pathBuilder.WriteString("'")
		}
		for _, index := range indexes {
			pathBuilder.WriteString("->")
			pathBuilder.WriteString(index)
		}
	}
	return col, pathBuilder.String()
}

// WhereJsonContains adds a where json contains clause to the query.
func (q *Query) WhereJsonContains(column string, value any) orm.Query {
	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_extract(%s, '%s') = ?", builder.quoteIdentifier(col), path), args: []any{value}})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		// MySQL's JSON_CONTAINS requires the value to be valid JSON
		// If it's a string, wrap it in quotes to make it a JSON string
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// MySQL: JSON_CONTAINS(column, value, path)
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?, '%s')", quotedCol, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", quotedCol), args: []any{jsonValue}})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
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
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?, '%s')", quotedCol, path), args: []any{value}})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", quotedCol), args: []any{value}})
		}
	}
	return q
}

// OrWhereJsonContains adds an or where json contains clause to the query.
func (q *Query) OrWhereJsonContains(column string, value any) orm.Query {
	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_extract(%s, '%s') = ?", builder.quoteIdentifier(col), path), args: []any{value}})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		// MySQL's JSON_CONTAINS requires the value to be valid JSON
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// MySQL: JSON_CONTAINS(column, value, path)
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?, '%s')", quotedCol, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", quotedCol), args: []any{jsonValue}})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
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
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?, '%s')", quotedCol, path), args: []any{value}})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS(%s, ?)", quotedCol), args: []any{value}})
		}
	}
	return q
}

// WhereJsonDoesntContain adds a where json doesnt contain clause to the query.
func (q *Query) WhereJsonDoesntContain(column string, value any) orm.Query {
	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_extract(%s, '%s') != ?", builder.quoteIdentifier(col), path), args: []any{value}})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		// MySQL's JSON_CONTAINS requires the value to be valid JSON
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// MySQL: NOT JSON_CONTAINS(column, value, path)
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?, '%s')", quotedCol, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", quotedCol), args: []any{jsonValue}})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
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
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?, '%s')", quotedCol, path), args: []any{value}})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", quotedCol), args: []any{value}})
		}
	}
	return q
}

// OrWhereJsonDoesntContain adds an or where json doesnt contain clause to the query.
func (q *Query) OrWhereJsonDoesntContain(column string, value any) orm.Query {
	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_extract(%s, '%s') != ?", builder.quoteIdentifier(col), path), args: []any{value}})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		// MySQL's JSON_CONTAINS requires the value to be valid JSON
		jsonValue := value
		if strVal, ok := value.(string); ok {
			jsonValue = fmt.Sprintf(`"%s"`, strVal)
		}
		if path != "" {
			// MySQL: NOT JSON_CONTAINS(column, value, path)
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?, '%s')", quotedCol, path), args: []any{jsonValue}})
		} else {
			// No path specified, search entire document
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", quotedCol), args: []any{jsonValue}})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
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
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?, '%s')", quotedCol, path), args: []any{value}})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS(%s, ?)", quotedCol), args: []any{value}})
		}
	}
	return q
}

// WhereJsonContainsKey adds a where json contains key clause to the query.
func (q *Query) WhereJsonContainsKey(column string) orm.Query {
	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_type(%s, '%s') IS NOT NULL", builder.quoteIdentifier(col), path), args: nil})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// MySQL: JSON_CONTAINS_PATH(column, 'one', path)
			// splitJsonColumnForMySQL already includes the $ prefix
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '%s')", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.*')", quotedCol), args: nil})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: check if path exists using -> operator with IS NOT NULL
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s%s IS NOT NULL", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IS NOT NULL AND %s::text != ''", quotedCol, quotedCol), args: nil})
		}
	} else {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '%s')", quotedCol, path), args: nil})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.*')", quotedCol), args: nil})
		}
	}
	return q
}

// OrWhereJsonContainsKey adds an or where json contains key clause to the query.
func (q *Query) OrWhereJsonContainsKey(column string) orm.Query {
	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_type(%s, '%s') IS NOT NULL", builder.quoteIdentifier(col), path), args: nil})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// MySQL: JSON_CONTAINS_PATH(column, 'one', path)
			// splitJsonColumnForMySQL already includes the $ prefix
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '%s')", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.*')", quotedCol), args: nil})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: check if path exists using -> operator with IS NOT NULL
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s%s IS NOT NULL", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s IS NOT NULL AND %s::text != ''", quotedCol, quotedCol), args: nil})
		}
	} else {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '%s')", quotedCol, path), args: nil})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.*')", quotedCol), args: nil})
		}
	}
	return q
}

// WhereJsonDoesntContainKey adds a where json doesnt contain key clause to the query.
func (q *Query) WhereJsonDoesntContainKey(column string) orm.Query {
	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_type(%s, '%s') IS NULL", builder.quoteIdentifier(col), path), args: nil})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// MySQL: NOT JSON_CONTAINS_PATH(column, 'one', path)
			// splitJsonColumnForMySQL already includes the $ prefix
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '%s')", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '$.*')", quotedCol), args: nil})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: check if path doesn't exist using -> operator with IS NULL
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s%s IS NULL", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s IS NULL OR %s::text = ''", quotedCol, quotedCol), args: nil})
		}
	} else {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '%s')", quotedCol, path), args: nil})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '$.*')", quotedCol), args: nil})
		}
	}
	return q
}

// OrWhereJsonDoesntContainKey adds an or where json doesnt contain key clause to the query.
func (q *Query) OrWhereJsonDoesntContainKey(column string) orm.Query {
	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("json_type(%s, '%s') IS NULL", builder.quoteIdentifier(col), path), args: nil})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// MySQL: NOT JSON_CONTAINS_PATH(column, 'one', path)
			// splitJsonColumnForMySQL already includes the $ prefix
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '%s')", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '$.*')", quotedCol), args: nil})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: check if path doesn't exist using -> operator with IS NULL
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s%s IS NULL", quotedCol, path), args: nil})
		} else {
			// No path specified, check if column has any JSON data
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("%s IS NULL OR %s::text = ''", quotedCol, quotedCol), args: nil})
		}
	} else {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '%s')", quotedCol, path), args: nil})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '$.*')", quotedCol), args: nil})
		}
	}
	return q
}

// WhereJsonLength adds a where json length clause to the query.
func (q *Query) WhereJsonLength(column string, operator string, value any) orm.Query {
	// Validate operator against allowed whitelist (case-insensitive)
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
	}
	normOp := strings.ToUpper(strings.TrimSpace(operator))
	if !allowedOperators[normOp] {
		if q.buildError == nil {
			q.buildError = fmt.Errorf("invalid operator in WhereJsonLength: %s", operator)
		}
		return q
	}

	builder := NewBuilder(q)
	if q.isSQLite() {
		col, path := q.splitJsonColumn(column)
		if col == "" {
			return q
		}
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_array_length(%s, '%s') %s ?", builder.quoteIdentifier(col), path, normOp), args: []any{value}})
	} else if q.isMySQL() {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// MySQL: JSON_LENGTH(column, path)
			// splitJsonColumnForMySQL already includes the $ prefix
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s, '%s') %s ?", quotedCol, path, normOp), args: []any{value}})
		} else {
			// No path specified, get length of entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s) %s ?", quotedCol, normOp), args: []any{value}})
		}
	} else if q.isPostgres() {
		col, path := q.splitJsonColumnForPostgreSQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			// PostgreSQL: jsonb_array_length(column->path)
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("jsonb_array_length(%s%s) %s ?", quotedCol, path, normOp), args: []any{value}})
		} else {
			// No path specified, get length of entire document
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("jsonb_array_length(%s) %s ?", quotedCol, normOp), args: []any{value}})
		}
	} else {
		col, path := q.splitJsonColumnForMySQL(column)
		if col == "" {
			return q
		}
		quotedCol := builder.quoteIdentifier(col)
		if path != "" {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s, '%s') %s ?", quotedCol, path, normOp), args: []any{value}})
		} else {
			q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s) %s ?", quotedCol, normOp), args: []any{value}})
		}
	}
	return q
}
