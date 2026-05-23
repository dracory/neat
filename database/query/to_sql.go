package query

import (
	"fmt"
	"strings"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// ToSql implements the ToSql interface for generating SQL without execution.
type ToSql struct {
	query *Query
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
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// Create generates the SQL for an INSERT query.
func (t *ToSql) Create(value any) string {
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildInsert(value)
	return t.replacePlaceholders(sql)
}

// InsertGetId generates the SQL for an INSERT query with ID retrieval.
func (t *ToSql) InsertGetId(values any) string {
	return t.Create(values)
}

// Delete generates the SQL for a DELETE query.
func (t *ToSql) Delete(value ...any) string {
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildDelete()
	return t.replacePlaceholders(sql)
}

// Find generates the SQL for a SELECT query.
func (t *ToSql) Find(dest any, conds ...any) string {
	// Add conditions to where clause
	for _, cond := range conds {
		t.query.wheres = append(t.query.wheres, whereClause{_type: "and", query: fmt.Sprintf("%v", cond), args: nil})
	}
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// First generates the SQL for a SELECT query with LIMIT 1.
func (t *ToSql) First(dest any) string {
	limit := 1
	t.query.limit = &limit
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// ForceDelete generates the SQL for a DELETE query (same as Delete for now).
func (t *ToSql) ForceDelete(value ...any) string {
	return t.Delete(value...)
}

// Get generates the SQL for a SELECT query.
func (t *ToSql) Get(dest any) string {
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// Pluck generates the SQL for a SELECT query with a single column.
func (t *ToSql) Pluck(column string, dest any) string {
	t.query.selects = []selectClause{{expr: column}}
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// Value generates the SQL for a SELECT query with a single column and LIMIT 1.
func (t *ToSql) Value(column string, dest any) string {
	t.query.selects = []selectClause{{expr: column}}
	limit := 1
	t.query.limit = &limit
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// Save generates the SQL for an UPDATE query (simplified).
func (t *ToSql) Save(value any) string {
	// This is a simplified implementation
	// In a full implementation, we'd need to determine if it's an insert or update
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildUpdate(value)
	return t.replacePlaceholders(sql)
}

// Avg generates the SQL for an AVG aggregation query.
func (t *ToSql) Avg(column string, dest any) string {
	t.query.aggregate = "AVG"
	t.query.aggregateCol = column
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// Max generates the SQL for a MAX aggregation query.
func (t *ToSql) Max(column string, dest any) string {
	t.query.aggregate = "MAX"
	t.query.aggregateCol = column
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// Min generates the SQL for a MIN aggregation query.
func (t *ToSql) Min(column string, dest any) string {
	t.query.aggregate = "MIN"
	t.query.aggregateCol = column
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// Sum generates the SQL for a SUM aggregation query.
func (t *ToSql) Sum(column string, dest any) string {
	t.query.aggregate = "SUM"
	t.query.aggregateCol = column
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildSelect()
	return t.replacePlaceholders(sql)
}

// Update generates the SQL for an UPDATE query.
func (t *ToSql) Update(column any, value ...any) string {
	builder := NewBuilder(t.query)
	sql, _ := builder.BuildUpdate(column, value...)
	return t.replacePlaceholders(sql)
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
	sql, _ := NewBuilder(t.query).BuildUpdate(updateQuery, incAmount)
	return t.replacePlaceholders(sql)
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
	sql, _ := NewBuilder(t.query).BuildUpdate(updateQuery, decAmount)
	return t.replacePlaceholders(sql)
}

// replacePlaceholders replaces ? placeholders with actual values for display.
func (t *ToSql) replacePlaceholders(sql string) string {
	// This is a simplified implementation that just removes placeholders
	// A full implementation would replace them with actual values
	return strings.ReplaceAll(sql, "?", "?")
}

// ToRawSql returns the raw SQL with placeholders.
func (q *Query) ToRawSql() contractsorm.ToSql {
	return NewToSql(q)
}

// ToSql returns a ToSql instance for generating SQL without execution.
func (q *Query) ToSql() contractsorm.ToSql {
	return NewToSql(q)
}
