package query

import (
	"fmt"
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
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s %s %s", first, operator, second), args: nil})
	return q
}

// OrWhereColumn adds an or where column clause to the query.
func (q *Query) OrWhereColumn(first, operator, second string) orm.Query {
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
	q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT (%v)", query), args: args})
	return q
}

// OrWhereNot adds an or where not clause to the query.
func (q *Query) OrWhereNot(query any, args ...any) orm.Query {
	q.wheres = append(q.wheres, whereClause{_type: "or", query: fmt.Sprintf("NOT (%v)", query), args: args})
	return q
}

// WhereAny adds a where any clause to the query.
func (q *Query) WhereAny(columns []string, operator string, value any) orm.Query {
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
	return parts[0], "." + parts[1]
}

// WhereJsonContains adds a where json contains clause to the query.
func (q *Query) WhereJsonContains(column string, value any) orm.Query {
	if q.driver != nil && q.driver.Dialect() == "sqlite" {
		col, path := q.splitJsonColumn(column)
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("json_extract(%s, '$%s') = ?", col, path), args: []any{value}})
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
	} else {
		q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("JSON_LENGTH(%s) %s ?", column, operator), args: []any{value}})
	}
	return q
}
