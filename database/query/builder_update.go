package query

import (
	"fmt"
	"sort"
	"strings"
)

// updateSet accumulates the SET clause of an UPDATE statement together with
// its bind arguments, tracking the dialect-specific placeholder index.
type updateSet struct {
	b           *Builder
	placeholder func(int) string
	index       int
	exprs       []string
	args        []any
}

func (b *Builder) newUpdateSet() *updateSet {
	placeholder := func(n int) string { return "?" }
	if b.query.driver != nil {
		placeholder = b.query.driver.Placeholder
	}
	return &updateSet{b: b, placeholder: placeholder, index: 1, args: []any{}}
}

// nextPlaceholder returns the next placeholder and advances the index.
func (s *updateSet) nextPlaceholder() string {
	p := s.placeholder(s.index)
	s.index++
	return p
}

// replacePlaceholders replaces up to count "?" markers in sql with
// dialect-specific placeholders.
func (s *updateSet) replacePlaceholders(sql string, count int) string {
	for i := 0; i < count; i++ {
		sql = strings.Replace(sql, "?", s.nextPlaceholder(), 1)
	}
	return sql
}

// addValue adds "col = <value>", inlining the SQL of a RawExpression value.
func (s *updateSet) addValue(col string, val any) {
	quoted := s.b.quoteIdentifier(col)
	if raw, ok := val.(RawExpression); ok {
		s.exprs = append(s.exprs, fmt.Sprintf("%s = %s", quoted, s.replacePlaceholders(raw.SQL, len(raw.Args))))
		s.args = append(s.args, raw.Args...)
		return
	}
	s.exprs = append(s.exprs, fmt.Sprintf("%s = %s", quoted, s.nextPlaceholder()))
	s.args = append(s.args, val)
}

// addExpression adds a complete SET expression (e.g. "views = views + ?"),
// used by Increment/Decrement.
func (s *updateSet) addExpression(expr string, values ...any) {
	s.exprs = append(s.exprs, s.replacePlaceholders(expr, strings.Count(expr, "?")))
	s.args = append(s.args, values...)
}

// addJSONPath adds a JSON path update ("data->meta->active") using the
// dialect's JSON function, falling back to a plain column assignment.
func (s *updateSet) addJSONPath(colStr string, val any) {
	segments := strings.Split(colStr, "->")
	if len(segments) < 2 || !(s.b.query.isMySQL() || s.b.query.isSQLite()) {
		s.addValue(colStr, val)
		return
	}
	jsonColumn := s.b.quoteIdentifier(segments[0])
	jsonPath := "$." + strings.Join(segments[1:], ".")
	fn := "json_set"
	if s.b.query.isMySQL() {
		fn = "JSON_SET"
	}
	s.exprs = append(s.exprs, fmt.Sprintf("%s = %s(%s, '%s', %s)", jsonColumn, fn, jsonColumn, jsonPath, s.nextPlaceholder()))
	s.args = append(s.args, val)
}

// isOmitted reports whether the column is excluded from writes.
func (b *Builder) isOmitted(col string) bool {
	for _, omit := range b.query.omitColumns {
		if omit == col {
			return true
		}
	}
	return false
}

// buildUpdateSet builds the SET expressions for the supported input shapes:
// map[string]any, a single column (or expression) with a value, or a struct.
func (b *Builder) buildUpdateSet(column any, values []any) *updateSet {
	set := b.newUpdateSet()

	if m, ok := column.(map[string]any); ok {
		// Sort keys for deterministic SQL generation
		keys := make([]string, 0, len(m))
		for col := range m {
			keys = append(keys, col)
		}
		sort.Strings(keys)
		for _, col := range keys {
			if !b.isOmitted(col) {
				set.addValue(col, m[col])
			}
		}
		return set
	}

	if len(values) > 0 {
		colStr, ok := column.(string)
		if !ok {
			return set
		}
		switch {
		case strings.Contains(colStr, "="):
			set.addExpression(colStr, values...)
		case strings.Contains(colStr, "->"):
			set.addJSONPath(colStr, values[0])
		default:
			set.addValue(colStr, values[0])
		}
		return set
	}

	// Struct or pointer-to-struct: extract fields as col=? pairs
	cols, vals, err := b.extractColumnsAndValues(column)
	if err != nil || vals == nil {
		return set
	}
	for i, col := range cols {
		if !b.isOmitted(col) {
			set.addValue(col, vals[i])
		}
	}
	return set
}

// isSoftDeleteUpdate reports whether the update targets the soft-delete
// column (soft delete / restore), in which case the automatic soft-delete
// filter must not be applied.
func (b *Builder) isSoftDeleteUpdate(column any) bool {
	m, ok := column.(map[string]any)
	if !ok {
		return false
	}
	_, has := m[getSoftDeleteColumn(b.query.model)]
	return has
}

// BuildUpdate builds an UPDATE query from the query state.
func (b *Builder) BuildUpdate(column any, values ...any) (string, []any) {
	parts := []string{"UPDATE"}
	if b.query.table != "" {
		parts = append(parts, b.quoteIdentifier(b.query.table))
	}

	set := b.buildUpdateSet(column, values)
	if len(set.exprs) > 0 {
		parts = append(parts, "SET "+strings.Join(set.exprs, ", "))
	}

	// SET args come first in the SQL (SET ... WHERE ...), with times normalized to UTC
	args := b.convertTimeArgs(set.args)

	// Soft delete / restore operations skip the automatic soft-delete filter
	var whereParts string
	var whereArgs []any
	if b.isSoftDeleteUpdate(column) {
		whereParts, whereArgs = b.buildWheresWithIndex(set.index)
	} else {
		whereParts, whereArgs = b.buildWheresWithSoftDeleteIndex(set.index)
	}

	parts, args = b.appendUpdateWhereAndLimit(parts, args, whereParts, whereArgs)
	return strings.Join(parts, " "), args
}

// appendUpdateWhereAndLimit appends the WHERE and LIMIT handling of an UPDATE,
// which differs per dialect:
//   - MySQL and PostgreSQL: WHERE ... LIMIT n
//   - SQLite: WHERE rowid IN (SELECT rowid FROM ... ORDER BY ... LIMIT n)
//   - SQL Server: UPDATE TOP (n) ... WHERE ...
//   - others: LIMIT is ignored
func (b *Builder) appendUpdateWhereAndLimit(parts []string, args []any, whereParts string, whereArgs []any) ([]string, []any) {
	limit := b.query.limit

	if limit != nil && b.query.isSQLite() {
		if whereParts == "" {
			whereParts = "1=1"
		}
		parts = append(parts, fmt.Sprintf("WHERE rowid IN (SELECT rowid FROM %s WHERE %s%s LIMIT %d)",
			b.quoteIdentifier(b.query.table), whereParts, b.buildUpdateOrderClause(), *limit))
		return parts, append(args, whereArgs...)
	}

	if limit != nil && b.query.isSQLServer() {
		parts[0] = fmt.Sprintf("UPDATE TOP (%d)", *limit)
	}

	if whereParts != "" {
		parts = append(parts, "WHERE "+whereParts)
		args = append(args, whereArgs...)
	}

	if limit != nil && (b.query.isMySQL() || b.query.isPostgres()) {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *limit))
	}
	return parts, args
}

// buildUpdateOrderClause builds the ORDER BY used for deterministic row
// selection in the SQLite LIMIT workaround.
func (b *Builder) buildUpdateOrderClause() string {
	if len(b.query.orders) == 0 {
		return ""
	}
	orderParts := make([]string, 0, len(b.query.orders))
	for _, order := range b.query.orders {
		orderParts = append(orderParts, fmt.Sprintf("%s %s", b.quoteIdentifier(order.column), order.direction))
	}
	return " ORDER BY " + strings.Join(orderParts, ", ")
}
