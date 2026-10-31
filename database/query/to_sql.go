package query

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/dracory/neat/contracts/database/orm"
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
	query := t.query.Clone().(*Query).applyScopes()
	query.aggregate = "COUNT"
	query.aggregateCol = "*"
	builder := NewBuilder(query)
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
	query := t.query.Clone().(*Query).applyScopes()
	if len(value) > 0 {
		applyConditions(query, value)
	}
	builder := NewBuilder(query)
	sql, args := builder.BuildDelete()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Find generates the SQL for a SELECT query.
func (t *ToSql) Find(dest any, conds ...any) string {
	// Apply scopes and conditions on a clone to avoid mutating the original query.
	query := t.query.Clone().(*Query)
	query = query.applyScopes()
	if len(conds) > 0 {
		applyConditions(query, conds)
	}
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// First generates the SQL for a SELECT query with LIMIT 1.
func (t *ToSql) First(dest any) string {
	query := t.query.Clone().(*Query).applyScopes()
	limit := 1
	query.limit = &limit
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// ForceDelete generates the SQL for a permanent DELETE query (bypasses soft delete).
//
// Deprecated: Use HardDelete() instead.
func (t *ToSql) ForceDelete(value ...any) string {
	query := t.query.Clone().(*Query).applyScopes()
	query.includeSoftDeleted = true
	if len(value) > 0 {
		applyConditions(query, value)
	}
	builder := NewBuilder(query)
	sql, args := builder.BuildDelete()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// HardDelete generates the SQL for a permanent DELETE query (bypasses soft delete).
// This is an intuitive alias for ForceDelete().
func (t *ToSql) HardDelete(value ...any) string {
	return t.ForceDelete(value...)
}

// SoftDelete generates the SQL for a soft-delete UPDATE query.
// Returns an empty string if the model does not implement SoftDeleteColumnNamer
// (i.e., does not support soft deletes), mirroring the fail-fast behavior of
// Query.SoftDelete().
func (t *ToSql) SoftDelete(value ...any) string {
	query := t.query.Clone().(*Query).applyScopes()
	if len(value) > 0 {
		applyConditions(query, value)
	}

	if !hasSoftDeleteCapability(query.model) {
		return ""
	}

	query.includeSoftDeleted = true
	builder := NewBuilder(query)
	col := getSoftDeleteColumn(query.model)
	var deleteValue any = time.Now()
	if strat, ok := query.model.(orm.SoftDeleteStrategy); ok {
		deleteValue = strat.SoftDeleteValue()
	}
	sql, args := builder.BuildUpdate(map[string]any{col: deleteValue})
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Get generates the SQL for a SELECT query.
func (t *ToSql) Get(dest any) string {
	query := t.query.Clone().(*Query).applyScopes()
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Pluck generates the SQL for a SELECT query with a single column.
func (t *ToSql) Pluck(column string, dest any) string {
	query := t.query.Clone().(*Query).applyScopes()
	query.selects = []selectClause{{expr: column}}
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Value generates the SQL for a SELECT query with a single column and LIMIT 1.
func (t *ToSql) Value(column string, dest any) string {
	query := t.query.Clone().(*Query).applyScopes()
	query.selects = []selectClause{{expr: column}}
	limit := 1
	query.limit = &limit
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Save generates the SQL for an INSERT or UPDATE query based on primary key.
func (t *ToSql) Save(value any) string {
	var sql string
	var args []any

	if isPrimaryKeyZero(value) {
		// INSERT: generate INSERT query
		builder := NewBuilder(t.query)
		sql, args = builder.BuildInsert(value)
	} else {
		// UPDATE: set WHERE id = <id> on a clone, then generate UPDATE query
		idVal, _ := getPrimaryKeyValueAny(value)
		clone := t.query.Clone().(*Query).applyScopes()
		clone.wheres = append(clone.wheres, whereClause{_type: "and", query: "id = ?", args: []any{idVal}})
		builder := NewBuilder(clone)
		sql, args = builder.BuildUpdate(value)
	}

	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Avg generates the SQL for an AVG aggregation query.
func (t *ToSql) Avg(column string, dest any) string {
	query := t.query.Clone().(*Query).applyScopes()
	query.aggregate = "AVG"
	query.aggregateCol = column
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Max generates the SQL for a MAX aggregation query.
func (t *ToSql) Max(column string, dest any) string {
	query := t.query.Clone().(*Query).applyScopes()
	query.aggregate = "MAX"
	query.aggregateCol = column
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Min generates the SQL for a MIN aggregation query.
func (t *ToSql) Min(column string, dest any) string {
	query := t.query.Clone().(*Query).applyScopes()
	query.aggregate = "MIN"
	query.aggregateCol = column
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Sum generates the SQL for a SUM aggregation query.
func (t *ToSql) Sum(column string, dest any) string {
	query := t.query.Clone().(*Query).applyScopes()
	query.aggregate = "SUM"
	query.aggregateCol = column
	builder := NewBuilder(query)
	sql, args := builder.BuildSelect()
	if t.useValues {
		return t.replacePlaceholdersWithValues(sql, args)
	}
	return t.replacePlaceholders(sql, args)
}

// Update generates the SQL for an UPDATE query.
func (t *ToSql) Update(column any, value ...any) string {
	query := t.query.Clone().(*Query).applyScopes()
	builder := NewBuilder(query)
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
	query := t.query.Clone().(*Query).applyScopes()
	builder := NewBuilder(query)
	quoted := builder.quoteIdentifier(column)
	updateQuery := fmt.Sprintf("%s = %s + ?", quoted, quoted)
	sql, args := builder.BuildUpdate(updateQuery, incAmount)
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
	query := t.query.Clone().(*Query).applyScopes()
	builder := NewBuilder(query)
	quoted := builder.quoteIdentifier(column)
	updateQuery := fmt.Sprintf("%s = %s - ?", quoted, quoted)
	sql, args := builder.BuildUpdate(updateQuery, decAmount)
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

// replacePlaceholdersWithValues replaces placeholders with their literal
// values. String literals are escaped (single quotes doubled, backslashes
// doubled on MySQL-family dialects) so a bound value cannot break out of the
// literal. Placeholders are located by scanning forward from the end of the
// previously injected value, so '?'-like sequences inside an injected value
// are never mistaken for remaining placeholders.
func (t *ToSql) replacePlaceholdersWithValues(sql string, args []any) string {
	// Determine the placeholder pattern based on the dialect
	dialect := ""
	placeholderPattern := "?"
	if t.query.driver != nil {
		dialect = t.query.driver.Dialect()
		switch dialect {
		case "oracle":
			placeholderPattern = ":%d"
		case "postgres":
			placeholderPattern = "$%d"
		case "sqlserver":
			placeholderPattern = "@p%d"
		}
	}

	var b strings.Builder
	b.Grow(len(sql) + len(args)*8)
	pos := 0
	for i, arg := range args {
		placeholder := placeholderPattern
		if placeholderPattern != "?" {
			placeholder = fmt.Sprintf(placeholderPattern, i+1)
		}
		idx := strings.Index(sql[pos:], placeholder)
		if idx < 0 {
			break
		}
		idx += pos
		b.WriteString(sql[pos:idx])
		b.WriteString(formatSQLLiteral(arg, dialect))
		pos = idx + len(placeholder)
	}
	b.WriteString(sql[pos:])
	return b.String()
}

// formatSQLLiteral renders a bound argument as an escaped SQL literal.
func formatSQLLiteral(arg any, dialect string) string {
	switch v := arg.(type) {
	case string:
		return "'" + escapeStringLiteral(v, dialect) + "'"
	case []byte:
		return "X'" + hex.EncodeToString(v) + "'"
	case time.Time:
		return "'" + v.Format("2006-01-02 15:04:05.999999") + "'"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		if dialect == "sqlserver" {
			if v {
				return "1"
			}
			return "0"
		}
		if v {
			return "TRUE"
		}
		return "FALSE"
	case nil:
		return "NULL"
	default:
		return "'" + escapeStringLiteral(fmt.Sprintf("%v", v), dialect) + "'"
	}
}

// escapeStringLiteral escapes a value placed inside a single-quoted SQL
// string literal: single quotes are doubled for every dialect, and
// backslashes are additionally doubled on MySQL-family dialects where
// backslash is an escape character inside string literals.
func escapeStringLiteral(s string, dialect string) string {
	s = strings.ReplaceAll(s, "'", "''")
	switch dialect {
	case "mysql", "tidb":
		s = strings.ReplaceAll(s, "\\", "\\\\")
	}
	return s
}

// ToRawSql returns the raw SQL with placeholders replaced by values.
func (q *Query) ToRawSql() orm.ToSql {
	return &ToSql{query: q, useValues: true}
}

// ToSql returns a ToSql instance for generating SQL without execution.
func (q *Query) ToSql() orm.ToSql {
	return NewToSql(q)
}
