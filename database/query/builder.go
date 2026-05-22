package query

import (
	"fmt"
	"reflect"
	"strings"
	"time"
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

	// WHERE clauses (with automatic soft-delete filter)
	whereParts, whereArgs := b.buildWheresWithSoftDelete()
	if whereParts != "" {
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

	// Locking clauses
	if b.query.lockForUpdate {
		parts = append(parts, "FOR UPDATE")
	} else if b.query.sharedLock {
		parts = append(parts, "LOCK IN SHARE MODE")
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

		// Handle bulk insert
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			rowPlaceholders := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				placeholders := make([]string, len(columns))
				for j := range placeholders {
					placeholders[j] = "?"
				}
				rowPlaceholders[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
			}
			parts = append(parts, strings.Join(rowPlaceholders, ", "))
		} else {
			// Single insert
			placeholders := make([]string, len(columns))
			for i := range placeholders {
				placeholders[i] = "?"
			}
			parts = append(parts, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
		}

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
	} else {
		// Handle struct or pointer-to-struct: extract fields as col=? pairs
		cols, vals, err := b.extractColumnsAndValues(column)
		if err == nil {
			for i, col := range cols {
				setParts = append(setParts, fmt.Sprintf("%s = ?", col))
				args = append(args, vals[i])
			}
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

	// WHERE clauses (with automatic soft-delete filter)
	whereParts, whereArgs := b.buildWheresWithSoftDelete()
	if whereParts != "" {
		parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
		args = append(args, whereArgs...)
	}

	return strings.Join(parts, " "), args
}

// buildWheresWithSoftDelete prepends the soft-delete condition when the model has a
// DeletedAt field and neither WithTrashed nor OnlyTrashed is set.
func (b *Builder) buildWheresWithSoftDelete() (string, []any) {
	var prefix string
	if hasSoftDeleteCapability(b.query.model) {
		switch {
		case b.query.onlyTrashed:
			prefix = "deleted_at IS NOT NULL"
		case b.query.withTrashed:
			// include all rows — no filter
		default:
			prefix = "deleted_at IS NULL"
		}
	}

	if len(b.query.wheres) == 0 {
		return prefix, nil
	}

	base, args := b.buildWheres()
	if prefix == "" {
		return base, args
	}
	return prefix + " AND " + base, args
}

// buildWheres builds the WHERE clause from where clauses.
func (b *Builder) buildWheres() (string, []any) {
	var parts []string
	var args []any

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
					placeholders[j] = "?"
				}
				clauseQuery = strings.Replace(clauseQuery, "(?)", "("+strings.Join(placeholders, ", ")+")", 1)
				clauseArgs = slice
			}
		}

		parts = append(parts, clauseQuery)
		args = append(args, clauseArgs...)
	}

	return strings.Join(parts, " "), args
}

// extractColumnsAndValues extracts column names and values from a struct, map, or slice.
func (b *Builder) extractColumnsAndValues(value any) ([]string, []any, error) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle slice/array for bulk insert
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() == 0 {
			return nil, nil, nil
		}

		var allColumns []string
		var allValues []any

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i).Interface()
			cols, vals, err := b.extractSingleColumnsAndValues(elem)
			if err != nil {
				return nil, nil, err
			}

			if i == 0 {
				allColumns = cols
			}
			allValues = append(allValues, vals...)
		}

		return allColumns, allValues, nil
	}

	return b.extractSingleColumnsAndValues(value)
}

// extractSingleColumnsAndValues extracts column names and values from a single struct or map.
func (b *Builder) extractSingleColumnsAndValues(value any) ([]string, []any, error) {
	var columns []string
	var values []any

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle map[string]any
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			columns = append(columns, key.String())
			values = append(values, v.MapIndex(key).Interface())
		}
		return columns, values, nil
	}

	// Handle struct using reflection
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
			if tag := field.Tag.Get("db"); tag != "" {
				if tag == "-" {
					continue
				}
				columnName = tag
			} else if tag := field.Tag.Get("neat"); tag != "" {
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

			// Skip slice/struct fields that are not handled as basic types
			if (fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Ptr) &&
				fieldValue.Type() != reflect.TypeOf(time.Time{}) {
				// Special case: if it's a pointer to a basic type, we might want it, but for associations we skip
				if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
					elem := fieldValue.Elem()
					if elem.Kind() == reflect.Struct {
						continue
					}
				} else if fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Struct {
					continue
				}
			}

			// Skip omitted columns
			omitted := false
			for _, omit := range b.query.omitColumns {
				if omit == columnName {
					omitted = true
					break
				}
			}
			if omitted {
				continue
			}

			// Skip zero values except for boolean and explicit zero values
			if fieldValue.IsZero() && fieldValue.Kind() != reflect.Bool && fieldValue.Type() != reflect.TypeOf(time.Time{}) {
				continue
			}

			columns = append(columns, columnName)
			values = append(values, fieldValue.Interface())
		}
		return columns, values, nil
	}

	return nil, nil, fmt.Errorf("unsupported value type for INSERT: %T", value)
}
