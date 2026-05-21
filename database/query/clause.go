package query

import (
	"fmt"
	"strings"
)

// ClauseType represents the type of a query clause.
type ClauseType int

const (
	WhereClause ClauseType = iota
	SelectClause
	JoinClause
	GroupClause
	HavingClause
	OrderClause
	LimitClause
	OffsetClause
)

// Clause represents a query clause.
type Clause interface {
	Type() ClauseType
	ToSQL() (string, []any)
}

// WhereClauseBuilder builds WHERE clauses.
type WhereClauseBuilder struct {
	_type string
	query string
	args  []any
}

func (w *WhereClauseBuilder) Type() ClauseType { return WhereClause }
func (w *WhereClauseBuilder) ToSQL() (string, []any) {
	return w.query, w.args
}

// SelectClauseBuilder builds SELECT clauses.
type SelectClauseBuilder struct {
	columns []any
}

func (s *SelectClauseBuilder) Type() ClauseType { return SelectClause }
func (s *SelectClauseBuilder) ToSQL() (string, []any) {
	var parts []string
	for _, col := range s.columns {
		parts = append(parts, fmt.Sprintf("%v", col))
	}
	return strings.Join(parts, ", "), nil
}

// JoinClauseBuilder builds JOIN clauses.
type JoinClauseBuilder struct {
	_type string
	query string
	args  []any
}

func (j *JoinClauseBuilder) Type() ClauseType { return JoinClause }
func (j *JoinClauseBuilder) ToSQL() (string, []any) {
	return fmt.Sprintf("%s %s", j._type, j.query), j.args
}

// OrderClauseBuilder builds ORDER BY clauses.
type OrderClauseBuilder struct {
	column    string
	direction string
}

func (o *OrderClauseBuilder) Type() ClauseType { return OrderClause }
func (o *OrderClauseBuilder) ToSQL() (string, []any) {
	return fmt.Sprintf("%s %s", o.column, o.direction), nil
}

// LimitClauseBuilder builds LIMIT clauses.
type LimitClauseBuilder struct {
	limit int
}

func (l *LimitClauseBuilder) Type() ClauseType { return LimitClause }
func (l *LimitClauseBuilder) ToSQL() (string, []any) {
	return fmt.Sprintf("LIMIT %d", l.limit), nil
}

// OffsetClauseBuilder builds OFFSET clauses.
type OffsetClauseBuilder struct {
	offset int
}

func (o *OffsetClauseBuilder) Type() ClauseType { return OffsetClause }
func (o *OffsetClauseBuilder) ToSQL() (string, []any) {
	return fmt.Sprintf("OFFSET %d", o.offset), nil
}
