package query

import (
	"fmt"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// ToSql implements the ToSql interface for generating SQL without execution.
type ToSql struct {
	query     *Query
	useValues bool
}

// NewToSql creates a new ToSql instance.
func NewToSql(q *Query) *ToSql {
	return &ToSql{query: q}
}

// Count generates the SQL for a COUNT query.
func (t *ToSql) Count() string {
	t.query.aggregate = "COUNT"
	t.query.aggregateCol = "*"
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Create generates the SQL for an INSERT query.
func (t *ToSql) Create(value any) string {
	builder := NewBuilder(t.query)
	sql, args := builder.BuildInsert(value)
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// InsertGetId generates the SQL for an INSERT query with ID retrieval.
func (t *ToSql) InsertGetId(values any) string {
	return t.Create(values)
}

// Delete generates the SQL for a DELETE query.
func (t *ToSql) Delete(value ...any) string {
	builder := NewBuilder(t.query)
	sql, args := builder.BuildDelete()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Find generates the SQL for a SELECT query.
func (t *ToSql) Find(dest any, conds ...any) string {
	// Add conditions to where clause
	for _, cond := range conds {
		t.query.wheres = append(t.query.wheres, whereClause{_type: "and", query: fmt.Sprintf("%v", cond), args: nil})
	}
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// First generates the SQL for a SELECT query with LIMIT 1.
func (t *ToSql) First(dest any) string {
	limit := 1
	t.query.limit = &limit
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// ForceDelete generates the SQL for a DELETE query (same as Delete for now).
func (t *ToSql) ForceDelete(value ...any) string {
	return t.Delete(value...)
}

// Get generates the SQL for a SELECT query.
func (t *ToSql) Get(dest any) string {
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Pluck generates the SQL for a SELECT query with a single column.
func (t *ToSql) Pluck(column string, dest any) string {
	t.query.selects = []selectClause{{expr: column}}
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Value generates the SQL for a SELECT query with a single column and LIMIT 1.
func (t *ToSql) Value(column string, dest any) string {
	t.query.selects = []selectClause{{expr: column}}
	limit := 1
	t.query.limit = &limit
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Save generates the SQL for an INSERT or UPDATE query based on primary key.
func (t *ToSql) Save(value any) string {
	// Determine if it's an insert or update based on primary key
	id := getPrimaryKeyValue(value)
	var sql string
	var args []any

	if id != 0 {
		// UPDATE: set WHERE id = <id> on a clone, then generate UPDATE query
		clone := *t.query
		clone.wheres = append(clone.wheres, whereClause{_type: "and", query: "id = ?", args: []any{id}})
		builder := NewBuilder(&clone)
		sql, args = builder.BuildUpdate(value)
	} else {
		// INSERT: generate INSERT query
		builder := NewBuilder(t.query)
		sql, args = builder.BuildInsert(value)
	}

	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Avg generates the SQL for an AVG aggregation query.
func (t *ToSql) Avg(column string, dest any) string {
	t.query.aggregate = "AVG"
	t.query.aggregateCol = column
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Max generates the SQL for a MAX aggregation query.
func (t *ToSql) Max(column string, dest any) string {
	t.query.aggregate = "MAX"
	t.query.aggregateCol = column
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Min generates the SQL for a MIN aggregation query.
func (t *ToSql) Min(column string, dest any) string {
	t.query.aggregate = "MIN"
	t.query.aggregateCol = column
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Sum generates the SQL for a SUM aggregation query.
func (t *ToSql) Sum(column string, dest any) string {
	t.query.aggregate = "SUM"
	t.query.aggregateCol = column
	builder := NewBuilder(t.query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Update generates the SQL for an UPDATE query.
func (t *ToSql) Update(column any, value ...any) string {
	builder := NewBuilder(t.query)
	sql, args := builder.BuildUpdate(column, value...)
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Increment generates the SQL for an INCREMENT query.
func (t *ToSql) Increment(column string, amount ...any) string {
	incAmount := int64(1)
	if len(amount) > 0 {
		if val, ok := amount[0].(int64); ok {
			incAmount = val
		}
	}
	updateQuery := fmt.Sprintf("%s = %s + ?", column, column)
	sql, args := NewBuilder(t.query).BuildUpdate(updateQuery, incAmount)
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Decrement generates the SQL for a DECREMENT query.
func (t *ToSql) Decrement(column string, amount ...any) string {
	decAmount := int64(1)
	if len(amount) > 0 {
		if val, ok := amount[0].(int64); ok {
			decAmount = val
		}
	}
	updateQuery := fmt.Sprintf("%s = %s - ?", column, column)
	sql, args := NewBuilder(t.query).BuildUpdate(updateQuery, decAmount)
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// replacePlaceholders replaces ? placeholders with actual values for display.
func (t *ToSql) replacePlaceholders(sql string, _ []any) string {
	// For ToSql, keep placeholders as-is
	return sql
}

// replacePlaceholdersWithValues replaces ? placeholders with actual values for display.
func (t *ToSql) replacePlaceholdersWithValues(sql string, args []any) string {
	// Determine the placeholder pattern based on the dialect
	var placeholderPattern string
	if t.query.driver != nil {
		dialect := t.query.driver.Dialect()
		switch dialect {
		case "oracle":
			placeholderPattern = ":%d"
		case "postgres":
			placeholderPattern = "$%d"
		case "sqlserver":
			placeholderPattern = "@p%d"
		default:
			placeholderPattern = "?"
		}
	} else {
		placeholderPattern = "?"
	}

	// Replace placeholders with actual values
	for i, arg := range args {
		var val string
		switch v := arg.(type) {
		case string:
			val = fmt.Sprintf("'%s'", v)
		case int, int64, uint, uint64:
			val = fmt.Sprintf("%d", v)
		case float32, float64:
			val = fmt.Sprintf("%f", v)
		case nil:
			val = "NULL"
		default:
			val = fmt.Sprintf("'%v'", v)
		}

		if placeholderPattern == "?" {
			sql = strings.Replace(sql, "?", val, 1)
		} else {
			placeholder := fmt.Sprintf(placeholderPattern, i+1)
			sql = strings.Replace(sql, placeholder, val, 1)
		}
	}
	return sql
}

// ToRawSql returns the raw SQL with placeholders replaced by values.
func (q *Query) ToRawSql() contractsorm.ToSql {
	return &ToSql{query: q, useValues: true}
}

// ToSql returns a ToSql instance for generating SQL without execution.
func (q *Query) ToSql() contractsorm.ToSql {
	return NewToSql(q)
}
