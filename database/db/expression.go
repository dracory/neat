package db

import (
	"fmt"
	"strings"
)

// Expression represents a database expression.
type Expression interface {
	// String returns the string representation of the expression.
	String() string
	// Values returns the values for the expression.
	Values() []any
}

// RawExpression represents a raw SQL expression.
type RawExpression struct {
	expr string
	args []any
}

// NewRawExpression creates a new raw expression.
func NewRawExpression(expr string, args ...any) *RawExpression {
	return &RawExpression{expr: expr, args: args}
}

// String returns the string representation of the raw expression.
func (e *RawExpression) String() string {
	return e.expr
}

// Values returns the values for the raw expression.
func (e *RawExpression) Values() []any {
	return e.args
}

// ColumnExpression represents a column expression.
type ColumnExpression struct {
	column string
}

// NewColumnExpression creates a new column expression.
func NewColumnExpression(column string) *ColumnExpression {
	return &ColumnExpression{column: column}
}

// String returns the string representation of the column expression.
func (e *ColumnExpression) String() string {
	return e.column
}

// Values returns the values for the column expression.
func (e *ColumnExpression) Values() []any {
	return nil
}

// TableExpression represents a table expression.
type TableExpression struct {
	table string
	alias string
}

// NewTableExpression creates a new table expression.
func NewTableExpression(table string, alias ...string) *TableExpression {
	t := &TableExpression{table: table}
	if len(alias) > 0 {
		t.alias = alias[0]
	}
	return t
}

// String returns the string representation of the table expression.
func (e *TableExpression) String() string {
	if e.alias != "" {
		return fmt.Sprintf("%s AS %s", e.table, e.alias)
	}
	return e.table
}

// Values returns the values for the table expression.
func (e *TableExpression) Values() []any {
	return nil
}

// WhereExpression represents a where clause expression.
type WhereExpression struct {
	column   string
	operator string
	value    any
	boolean  string // AND, OR
}

// NewWhereExpression creates a new where expression.
func NewWhereExpression(column, operator string, value any, boolean ...string) *WhereExpression {
	boolType := "AND"
	if len(boolean) > 0 {
		boolType = boolean[0]
	}
	return &WhereExpression{
		column:   column,
		operator: operator,
		value:    value,
		boolean:  boolType,
	}
}

// String returns the string representation of the where expression.
func (e *WhereExpression) String() string {
	return fmt.Sprintf("%s %s ?", e.column, e.operator)
}

// Values returns the values for the where expression.
func (e *WhereExpression) Values() []any {
	return []any{e.value}
}

// Boolean returns the boolean operator (AND/OR).
func (e *WhereExpression) Boolean() string {
	return e.boolean
}

// QuoteIdentifier quotes an identifier for SQL.
func QuoteIdentifier(identifier string, dialect string) string {
	switch dialect {
	case "mysql":
		return fmt.Sprintf("`%s`", identifier)
	case "postgres", "sqlite", "turso":
		return fmt.Sprintf(`"%s"`, identifier)
	case "sqlserver":
		return fmt.Sprintf("[%s]", identifier)
	default:
		return identifier
	}
}

// QuoteValue quotes a value for SQL.
func QuoteValue(value any, dialect string) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", v)
	}
}
