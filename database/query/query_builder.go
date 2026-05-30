package query

import (
	"fmt"
	"strings"

	"github.com/dracory/neat/contracts/database/orm"
)

// Select adds columns to the select clause.
func (q *Query) Select(query any, args ...any) orm.Query {
	var queryStr string
	var processedArgs []any

	if fn, ok := query.(func(orm.Query) orm.Query); ok {
		subQuery := fn(q.newQuery())
		builder := NewBuilder(subQuery.(*Query))
		subSQL, subArgs := builder.BuildSelect()
		if len(args) > 0 {
			queryStr = fmt.Sprintf("(%s) as %s", subSQL, args[0])
		} else {
			queryStr = fmt.Sprintf("(%s)", subSQL)
		}
		processedArgs = append(processedArgs, subArgs...)
	} else {
		if slice, ok := query.([]string); ok {
			queryStr = strings.Join(slice, ", ")
		} else {
			queryStr = fmt.Sprintf("%v", query)
		}

		processedArgs = make([]any, 0, len(args))
		for _, arg := range args {
			if fn, ok := arg.(func(orm.Query) orm.Query); ok {
				subQuery := fn(q.Clone().(*Query))
				// Temporarily remove driver to build subquery with ? placeholders
				originalDriver := subQuery.(*Query).driver
				subQuery.(*Query).driver = nil
				builder := NewBuilder(subQuery.(*Query))
				subSQL, subArgs := builder.BuildSelect()
				// Restore driver
				subQuery.(*Query).driver = originalDriver
				if strings.Contains(queryStr, "?") {
					queryStr = strings.Replace(queryStr, "?", fmt.Sprintf("(%s)", subSQL), 1)
				} else {
					processedArgs = append(processedArgs, fmt.Sprintf("(%s)", subSQL))
				}
				processedArgs = append(processedArgs, subArgs...)
			} else {
				processedArgs = append(processedArgs, arg)
			}
		}
	}
	q.selects = append(q.selects, selectClause{expr: queryStr, args: processedArgs})
	return q
}

// Where adds a where clause to the query.
// Supports Laravel-style syntax: Where("column", "value") automatically uses = operator.
func (q *Query) Where(query any, args ...any) orm.Query {
	queryStr := fmt.Sprintf("%v", query)
	// Laravel-style: Where("column", "value") -> Where("column = ?", "value")
	// Only transform if exactly 1 arg and no operator is present
	if len(args) == 1 && !containsOperator(queryStr) {
		queryStr = fmt.Sprintf("%s = ?", queryStr)
	}
	q.wheres = append(q.wheres, whereClause{_type: "and", query: queryStr, args: args})
	return q
}

// containsOperator checks if a string contains SQL comparison operators
// Uses string matching to avoid false positives from column names containing operator-like strings
func containsOperator(s string) bool {
	upper := strings.ToUpper(s)

	// Check for comparison operators with optional whitespace
	// These patterns ensure operators are properly spaced to avoid matching column names
	operators := []string{
		" = ", " != ", " <> ", " > ", " < ", " >= ", " <= ",
		" LIKE ", " NOT LIKE ", " IN ", " NOT IN ", " BETWEEN ", " NOT BETWEEN ",
	}

	for _, op := range operators {
		if strings.Contains(upper, op) {
			return true
		}
	}

	// Also check for operators at start or end (e.g., "= value" or "value =")
	if strings.HasPrefix(upper, "= ") || strings.HasPrefix(upper, "!= ") ||
		strings.HasPrefix(upper, "<> ") || strings.HasPrefix(upper, "> ") ||
		strings.HasPrefix(upper, "< ") || strings.HasPrefix(upper, ">= ") ||
		strings.HasPrefix(upper, "<= ") {
		return true
	}

	if strings.HasSuffix(upper, " =") || strings.HasSuffix(upper, " !=") ||
		strings.HasSuffix(upper, " <>") || strings.HasSuffix(upper, " >") ||
		strings.HasSuffix(upper, " <") || strings.HasSuffix(upper, " >=") ||
		strings.HasSuffix(upper, " <=") {
		return true
	}

	// Check for operators without spaces (e.g., "name=?", "age>5")
	// Must be preceded/followed by non-operator character or ? to avoid false positives
	// Pattern: operator must be between word boundaries or next to ?
	noSpaceOps := []string{"=", "!=", "<>", ">", "<", ">=", "<="}
	for _, op := range noSpaceOps {
		// Check if operator exists in string
		idx := strings.Index(upper, op)
		if idx == -1 {
			continue
		}

		// Check character before operator (if any)
		before := ""
		if idx > 0 {
			before = string(upper[idx-1])
		}

		// Check character after operator (if any)
		after := ""
		if idx+len(op) < len(upper) {
			after = string(upper[idx+len(op)])
		}

		// Valid if:
		// - Before is empty or alphanumeric/underscore (column name)
		// - After is empty or ? or space or alphanumeric/underscore
		beforeValid := before == "" || isWordChar(before)
		afterValid := after == "" || after == "?" || after == " " || isWordChar(after)

		if beforeValid && afterValid {
			return true
		}
	}

	return false
}

// isWordChar checks if a character is alphanumeric or underscore
func isWordChar(c string) bool {
	return (c >= "A" && c <= "Z") || (c >= "0" && c <= "9") || c == "_"
}

// OrWhere adds an or where clause to the query.
// Supports Laravel-style syntax: OrWhere("column", "value") automatically uses = operator.
func (q *Query) OrWhere(query any, args ...any) orm.Query {
	queryStr := fmt.Sprintf("%v", query)
	// Laravel-style: OrWhere("column", "value") -> OrWhere("column = ?", "value")
	// Only transform if exactly 1 arg and no operator is present
	if len(args) == 1 && !containsOperator(queryStr) {
		queryStr = fmt.Sprintf("%s = ?", queryStr)
	}
	q.wheres = append(q.wheres, whereClause{_type: "or", query: queryStr, args: args})
	return q
}

// Order adds an order by clause to the query.
func (q *Query) Order(value any) orm.Query {
	expr := fmt.Sprintf("%v", value)
	upperExpr := strings.ToUpper(expr)
	if strings.Contains(upperExpr, " DESC") {
		expr = strings.TrimSuffix(expr, " DESC")
		expr = strings.TrimSuffix(expr, " desc")
		q.orders = append(q.orders, orderClause{column: expr, direction: "desc"})
	} else if strings.Contains(upperExpr, " ASC") {
		expr = strings.TrimSuffix(expr, " ASC")
		expr = strings.TrimSuffix(expr, " asc")
		q.orders = append(q.orders, orderClause{column: expr, direction: "asc"})
	} else {
		q.orders = append(q.orders, orderClause{column: expr, direction: "asc"})
	}
	return q
}

// OrderBy adds an order by clause with direction.
func (q *Query) OrderBy(column string, direction ...string) orm.Query {
	dir := "asc"
	if len(direction) > 0 {
		dir = direction[0]
	}
	q.orders = append(q.orders, orderClause{column: column, direction: dir})
	return q
}

// OrderByDesc adds an order by clause with desc direction.
func (q *Query) OrderByDesc(column string) orm.Query {
	q.orders = append(q.orders, orderClause{column: column, direction: "desc"})
	return q
}

// Limit adds a limit clause to the query.
func (q *Query) Limit(limit int) orm.Query {
	q.limit = &limit
	return q
}

// Offset adds an offset clause to the query.
func (q *Query) Offset(offset int) orm.Query {
	q.offset = &offset
	return q
}

// Distinct adds distinct to the query.
func (q *Query) Distinct(args ...any) orm.Query {
	q.distinct = true
	if len(args) > 0 {
		q.distinctCols = make([]string, 0)
		for _, arg := range args {
			q.distinctCols = append(q.distinctCols, fmt.Sprintf("%v", arg))
		}
	}
	return q
}

// Join adds a join clause to the query.
func (q *Query) Join(query string, args ...any) orm.Query {
	q.joins = append(q.joins, joinClause{_type: "JOIN", query: query, args: args})
	return q
}

// LeftJoin adds a left join clause to the query.
func (q *Query) LeftJoin(query string, args ...any) orm.Query {
	q.joins = append(q.joins, joinClause{_type: "LEFT JOIN", query: query, args: args})
	return q
}

// RightJoin adds a right join clause to the query.
func (q *Query) RightJoin(query string, args ...any) orm.Query {
	q.joins = append(q.joins, joinClause{_type: "RIGHT JOIN", query: query, args: args})
	return q
}

// CrossJoin adds a cross join clause to the query.
func (q *Query) CrossJoin(query string, args ...any) orm.Query {
	q.joins = append(q.joins, joinClause{_type: "CROSS JOIN", query: query, args: args})
	return q
}

// Group adds a group by clause to the query.
func (q *Query) Group(name string) orm.Query {
	q.groups = append(q.groups, name)
	return q
}

// Having adds a having clause to the query.
func (q *Query) Having(query any, args ...any) orm.Query {
	queryStr := fmt.Sprintf("%v", query)
	processedArgs := make([]any, 0, len(args))
	for _, arg := range args {
		if fn, ok := arg.(func(orm.Query) orm.Query); ok {
			subQuery := fn(q.Clone().(*Query))
			// Temporarily remove driver to build subquery with ? placeholders
			originalDriver := subQuery.(*Query).driver
			subQuery.(*Query).driver = nil
			builder := NewBuilder(subQuery.(*Query))
			subSQL, subArgs := builder.BuildSelect()
			// Restore driver
			subQuery.(*Query).driver = originalDriver
			if strings.Contains(queryStr, "?") {
				queryStr = strings.Replace(queryStr, "?", fmt.Sprintf("(%s)", subSQL), 1)
			} else {
				processedArgs = append(processedArgs, fmt.Sprintf("(%s)", subSQL))
			}
			processedArgs = append(processedArgs, subArgs...)
		} else {
			processedArgs = append(processedArgs, arg)
		}
	}
	q.havings = append(q.havings, havingClause{query: queryStr, args: processedArgs})
	return q
}
