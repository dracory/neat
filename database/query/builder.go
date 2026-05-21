package query

import (
	"fmt"
	"strings"
)

// Builder handles SQL query building.
type Builder struct {
	query *Query
}

// NewBuilder creates a new Builder instance.
func NewBuilder(q *Query) *Builder {
	return &Builder{query: q}
}

// BuildSelect builds a SELECT query from the query state.
func (b *Builder) BuildSelect() (string, []any) {
	var parts []string
	var args []any

	// SELECT clause
	if len(b.query.selects) > 0 {
		var selectParts []string
		for _, s := range b.query.selects {
			selectParts = append(selectParts, fmt.Sprintf("%v", s))
		}
		parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(selectParts, ", ")))
	} else if b.query.aggregate != "" {
		parts = append(parts, fmt.Sprintf("SELECT %s(%s)", b.query.aggregate, b.query.aggregateCol))
	} else {
		parts = append(parts, "SELECT *")
	}

	// FROM clause
	if b.query.table != "" {
		parts = append(parts, fmt.Sprintf("FROM %s", b.query.table))
	}

	// JOIN clauses
	for _, join := range b.query.joins {
		parts = append(parts, fmt.Sprintf("%s %s", join._type, join.query))
		args = append(args, join.args...)
	}

	// WHERE clauses
	if len(b.query.wheres) > 0 {
		whereParts, whereArgs := b.buildWheres()
		parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
		args = append(args, whereArgs...)
	}

	// GROUP BY clauses
	if len(b.query.groups) > 0 {
		parts = append(parts, fmt.Sprintf("GROUP BY %s", strings.Join(b.query.groups, ", ")))
	}

	// HAVING clauses
	if len(b.query.havings) > 0 {
		for _, having := range b.query.havings {
			parts = append(parts, fmt.Sprintf("HAVING %s", having.query))
			args = append(args, having.args...)
		}
	}

	// ORDER BY clauses
	if len(b.query.orders) > 0 {
		var orderParts []string
		for _, order := range b.query.orders {
			orderParts = append(orderParts, fmt.Sprintf("%s %s", order.column, order.direction))
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(orderParts, ", ")))
	}

	// LIMIT clause
	if b.query.limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
	}

	// OFFSET clause
	if b.query.offset != nil {
		parts = append(parts, fmt.Sprintf("OFFSET %d", *b.query.offset))
	}

	return strings.Join(parts, " "), args
}

// BuildInsert builds an INSERT query from the query state.
func (b *Builder) BuildInsert(value any) (string, []any) {
	// TODO: Implement INSERT query building
	return "", nil
}

// BuildUpdate builds an UPDATE query from the query state.
func (b *Builder) BuildUpdate(column any, values ...any) (string, []any) {
	// TODO: Implement UPDATE query building
	return "", nil
}

// BuildDelete builds a DELETE query from the query state.
func (b *Builder) BuildDelete() (string, []any) {
	var parts []string
	var args []any

	// DELETE clause
	parts = append(parts, "DELETE")

	// FROM clause
	if b.query.table != "" {
		parts = append(parts, fmt.Sprintf("FROM %s", b.query.table))
	}

	// WHERE clauses
	if len(b.query.wheres) > 0 {
		whereParts, whereArgs := b.buildWheres()
		parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
		args = append(args, whereArgs...)
	}

	return strings.Join(parts, " "), args
}

// buildWheres builds the WHERE clause from where clauses.
func (b *Builder) buildWheres() (string, []any) {
	var parts []string
	var args []any

	for i, where := range b.query.wheres {
		if i > 0 {
			parts = append(parts, strings.ToUpper(where._type))
		}
		parts = append(parts, where.query)
		args = append(args, where.args...)
	}

	return strings.Join(parts, " "), args
}
