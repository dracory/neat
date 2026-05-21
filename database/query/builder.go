package query

import (
	"fmt"
	"reflect"
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
	var parts []string
	var args []any

	// INSERT clause
	parts = append(parts, "INSERT")

	// INTO clause
	if b.query.table != "" {
		parts = append(parts, fmt.Sprintf("INTO %s", b.query.table))
	}

	// Extract columns and values from the value
	columns, values, err := b.extractColumnsAndValues(value)
	if err != nil {
		return "", nil
	}

	if len(columns) > 0 {
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(columns, ", ")))
		parts = append(parts, "VALUES")

		// Build placeholders
		placeholders := make([]string, len(values))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
		args = append(args, values...)
	}

	return strings.Join(parts, " "), args
}

// BuildUpdate builds an UPDATE query from the query state.
func (b *Builder) BuildUpdate(column any, values ...any) (string, []any) {
	var parts []string
	var args []any

	// UPDATE clause
	parts = append(parts, "UPDATE")

	// Table name
	if b.query.table != "" {
		parts = append(parts, b.query.table)
	}

	// SET clause
	var setParts []string

	// Handle map[string]any for column/value pairs
	if m, ok := column.(map[string]any); ok {
		for col, val := range m {
			setParts = append(setParts, fmt.Sprintf("%s = ?", col))
			args = append(args, val)
		}
	} else if len(values) > 0 {
		// Handle single column with value
		if colStr, ok := column.(string); ok {
			setParts = append(setParts, fmt.Sprintf("%s = ?", colStr))
			args = append(args, values[0])
		}
	}

	if len(setParts) > 0 {
		parts = append(parts, fmt.Sprintf("SET %s", strings.Join(setParts, ", ")))
	}

	// WHERE clauses
	if len(b.query.wheres) > 0 {
		whereParts, whereArgs := b.buildWheres()
		parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
		args = append(args, whereArgs...)
	}

	return strings.Join(parts, " "), args
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

// extractColumnsAndValues extracts column names and values from a struct or map.
func (b *Builder) extractColumnsAndValues(value any) ([]string, []any, error) {
	var columns []string
	var values []any

	// Handle map[string]any
	if m, ok := value.(map[string]any); ok {
		for col, val := range m {
			columns = append(columns, col)
			values = append(values, val)
		}
		return columns, values, nil
	}

	// Handle struct using reflection
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			// Skip unexported fields
			if !fieldValue.CanInterface() {
				continue
			}

			// Get column name from tag or field name
			columnName := field.Name
			if tag := field.Tag.Get("neat"); tag != "" {
				if parts := strings.Split(tag, ";"); len(parts) > 0 {
					if colPart := strings.Split(parts[0], ":"); len(colPart) > 1 {
						columnName = colPart[1]
					}
				}
			} else if tag := field.Tag.Get("gorm"); tag != "" {
				if parts := strings.Split(tag, ";"); len(parts) > 0 {
					if colPart := strings.Split(parts[0], ":"); len(colPart) > 1 {
						columnName = colPart[1]
					}
				}
			}

			// Skip zero values except for boolean and explicit zero values
			if fieldValue.IsZero() && fieldValue.Kind() != reflect.Bool {
				continue
			}

			columns = append(columns, columnName)
			values = append(values, fieldValue.Interface())
		}
		return columns, values, nil
	}

	return nil, nil, fmt.Errorf("unsupported value type for INSERT: %T", value)
}
