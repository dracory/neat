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

// quoteIdentifier wraps an identifier in the appropriate quotes for the dialect.
func (b *Builder) quoteIdentifier(name string) string {
	if b.query.driver == nil || name == "*" || name == "" {
		return name
	}

	dialect := b.query.driver.Dialect()
	quoteChar := "\""
	if dialect == "mysql" {
		quoteChar = "`"
	}

	// If already quoted, return as is
	if (strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"")) ||
		(strings.HasPrefix(name, "`") && strings.HasSuffix(name, "`")) {
		return name
	}

	// Handle dotted names (e.g., table.column)
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		for i, part := range parts {
			parts[i] = b.quoteIdentifier(part)
		}
		return strings.Join(parts, ".")
	}

	// Handle " AS " alias (case insensitive)
	upperName := strings.ToUpper(name)
	if idx := strings.Index(upperName, " AS "); idx != -1 {
		identifier := strings.TrimSpace(name[:idx])
		alias := strings.TrimSpace(name[idx+4:])
		return fmt.Sprintf("%s AS %s", b.quoteIdentifier(identifier), b.quoteIdentifier(alias))
	}

	// Handle space alias (e.g., "users u")
	if idx := strings.Index(name, " "); idx != -1 {
		identifier := strings.TrimSpace(name[:idx])
		alias := strings.TrimSpace(name[idx+1:])
		return fmt.Sprintf("%s %s", b.quoteIdentifier(identifier), b.quoteIdentifier(alias))
	}

	return fmt.Sprintf("%s%s%s", quoteChar, name, quoteChar)
}

// BuildSelect builds a SELECT query from the query state.
func (b *Builder) BuildSelect() (string, []any) {
	if b.query.rawSQL != "" {
		return b.query.rawSQL, b.query.rawArgs
	}

	var parts []string
	var args []any

	// SELECT clause
	if b.query.aggregate != "" {
		// When aggregate is set, ignore SELECT list and use aggregate function
		// Handle COUNT with DISTINCT
		if b.query.aggregate == "COUNT" && b.query.distinct && len(b.query.distinctCols) > 0 {
			parts = append(parts, fmt.Sprintf("SELECT COUNT(DISTINCT %s)", strings.Join(b.query.distinctCols, ", ")))
		} else {
			parts = append(parts, fmt.Sprintf("SELECT %s(%s)", b.query.aggregate, b.query.aggregateCol))
		}
	} else if len(b.query.selects) > 0 {
		var selectParts []string
		for _, s := range b.query.selects {
			selectParts = append(selectParts, s.expr)
			args = append(args, s.args...)
		}
		// Prepend DISTINCT if set
		if b.query.distinct {
			parts = append(parts, fmt.Sprintf("SELECT DISTINCT %s", strings.Join(selectParts, ", ")))
		} else {
			parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(selectParts, ", ")))
		}
	} else {
		// No explicit SELECT, derive from model
		if b.query.model != nil {
			cols := b.extractColumnNames(b.query.model)
			if len(cols) > 0 {
				// Filter out omitted columns
				var filteredCols []string
				for _, col := range cols {
					omitted := false
					for _, omit := range b.query.omitColumns {
						if omit == col {
							omitted = true
							break
						}
					}
					if !omitted {
						filteredCols = append(filteredCols, col)
					}
				}
				if len(filteredCols) > 0 {
					parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(filteredCols, ", ")))
				} else {
					parts = append(parts, "SELECT *")
				}
			} else {
				parts = append(parts, "SELECT *")
			}
		} else {
			parts = append(parts, "SELECT *")
		}
	}

	// FROM clause
	if b.query.table != "" {
		if strings.Contains(b.query.table, "(") && strings.Contains(b.query.table, ")") {
			// Subquery in FROM, don't quote
			parts = append(parts, fmt.Sprintf("FROM %s", b.query.table))
		} else {
			parts = append(parts, fmt.Sprintf("FROM %s", b.quoteIdentifier(b.query.table)))
		}
		args = append(args, b.query.tableArgs...)
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
		var havingParts []string
		for _, having := range b.query.havings {
			havingParts = append(havingParts, having.query)
			args = append(args, having.args...)
		}
		parts = append(parts, fmt.Sprintf("HAVING %s", strings.Join(havingParts, " AND ")))
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
		// SQLite requires LIMIT when using OFFSET
		if b.query.limit == nil && b.query.driver != nil && b.query.driver.Dialect() == "sqlite" {
			parts = append(parts, "LIMIT -1")
		}
		parts = append(parts, fmt.Sprintf("OFFSET %d", *b.query.offset))
	}

	// Locking clauses
	// Skip lock clauses for SQLite as it doesn't support them
	if b.query.driver == nil || b.query.driver.Dialect() != "sqlite" {
		if b.query.lockForUpdate {
			parts = append(parts, "FOR UPDATE")
		} else if b.query.sharedLock {
			parts = append(parts, "LOCK IN SHARE MODE")
		}
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
		parts = append(parts, fmt.Sprintf("INTO %s", b.quoteIdentifier(b.query.table)))
		args = append(args, b.query.tableArgs...)
	}

	// Extract columns and values from the value
	columns, values, err := b.extractColumnsAndValues(value)
	if err != nil {
		return "", nil
	}

	if len(columns) > 0 {
		// Quote column names
		quotedColumns := make([]string, len(columns))
		for i, col := range columns {
			quotedColumns[i] = b.quoteIdentifier(col)
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(quotedColumns, ", ")))
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
			args = append(args, values...)
		} else {
			// Single insert
			placeholders := make([]string, len(columns))
			for i := range placeholders {
				placeholders[i] = "?"
			}
			parts = append(parts, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
			args = append(args, values...)
		}
	}

	return strings.Join(parts, " "), args
}

// BuildUpdate builds an UPDATE query from the query state.
func (b *Builder) BuildUpdate(column any, values ...any) (string, []any) {
	var parts []string
	var args []any
	var setArgs []any // Store SET args separately to add them after WHERE args

	// UPDATE clause
	parts = append(parts, "UPDATE")

	// Table name
	if b.query.table != "" {
		parts = append(parts, b.quoteIdentifier(b.query.table))
		args = append(args, b.query.tableArgs...)
	}

	// SET clause
	var setParts []string

	// Handle map[string]any for column/value pairs
	if m, ok := column.(map[string]any); ok {
		for col, val := range m {
			// Skip omitted columns
			omitted := false
			for _, omit := range b.query.omitColumns {
				if omit == col {
					omitted = true
					break
				}
			}
			if omitted {
				continue
			}
			setParts = append(setParts, fmt.Sprintf("%s = ?", b.quoteIdentifier(col)))
			setArgs = append(setArgs, val)
		}
	} else if len(values) > 0 {
		// Handle single column with value
		if colStr, ok := column.(string); ok {
			// Check if the column string is already a complete SET expression (contains =)
			if strings.Contains(colStr, "=") {
				// Use the expression as-is (for Increment/Decrement)
				setParts = append(setParts, colStr)
				setArgs = append(setArgs, values...)
			} else {
				setParts = append(setParts, fmt.Sprintf("%s = ?", b.quoteIdentifier(colStr)))
				setArgs = append(setArgs, values[0])
			}
		}
	} else {
		// Handle struct or pointer-to-struct: extract fields as col=? pairs
		cols, vals, err := b.extractColumnsAndValues(column)
		if err == nil {
			for i, col := range cols {
				// Skip omitted columns
				omitted := false
				for _, omit := range b.query.omitColumns {
					if omit == col {
						omitted = true
						break
					}
				}
				if omitted {
					continue
				}
				setParts = append(setParts, fmt.Sprintf("%s = ?", b.quoteIdentifier(col)))
				setArgs = append(setArgs, vals[i])
			}
		}
	}

	if len(setParts) > 0 {
		parts = append(parts, fmt.Sprintf("SET %s", strings.Join(setParts, ", ")))
	}

	// WHERE clauses (with automatic soft-delete filter)
	// Skip soft-delete filter if we're updating deleted_at column (for soft delete operations)
	isSoftDeleteOperation := false
	if m, ok := column.(map[string]any); ok {
		if _, hasDeletedAt := m["deleted_at"]; hasDeletedAt {
			isSoftDeleteOperation = true
		}
	}

	// Add SET args first (they appear first in the SQL: SET ... WHERE ...)
	args = append(args, setArgs...)

	// Build WHERE clause (will be used for both normal WHERE and LIMIT workaround)
	var whereParts string
	var whereArgs []any
	if !isSoftDeleteOperation {
		whereParts, whereArgs = b.buildWheresWithSoftDelete()
	} else {
		// For soft delete operations, use regular WHERE without soft-delete filter
		whereParts, whereArgs = b.buildWheres()
	}

	// LIMIT clause
	// MySQL supports LIMIT directly in UPDATE
	// SQLite requires a subquery workaround: UPDATE ... WHERE rowid IN (SELECT rowid FROM ... ORDER BY ... LIMIT N)
	// PostgreSQL supports LIMIT directly in UPDATE
	if b.query.limit != nil {
		if b.query.driver != nil && b.query.driver.Dialect() == "mysql" {
			// Add WHERE clause if it exists
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
			parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
		} else if b.query.driver != nil && b.query.driver.Dialect() == "sqlite" {
			// SQLite workaround: wrap in subquery with rowid
			if whereParts == "" {
				whereParts = "1=1"
			}
			// Build ORDER BY clause for deterministic row selection
			var orderClause string
			if len(b.query.orders) > 0 {
				var orderParts []string
				for _, order := range b.query.orders {
					orderParts = append(orderParts, fmt.Sprintf("%s %s", order.column, order.direction))
				}
				orderClause = fmt.Sprintf(" ORDER BY %s", strings.Join(orderParts, ", "))
			}
			// Add WHERE clause with rowid subquery including ORDER BY
			parts = append(parts, fmt.Sprintf("WHERE rowid IN (SELECT rowid FROM %s WHERE %s%s LIMIT %d)", b.quoteIdentifier(b.query.table), whereParts, orderClause, *b.query.limit))
			args = append(args, whereArgs...)
		} else if b.query.driver != nil && b.query.driver.Dialect() == "postgres" {
			// PostgreSQL supports LIMIT directly in UPDATE
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
			parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
		} else {
			// Other databases: add WHERE clause normally (LIMIT may or may not be supported)
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
		}
	} else {
		// No LIMIT: add WHERE clause normally
		if whereParts != "" {
			parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
			args = append(args, whereArgs...)
		}
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
		parts = append(parts, fmt.Sprintf("FROM %s", b.quoteIdentifier(b.query.table)))
		args = append(args, b.query.tableArgs...)
	}

	// Build WHERE clause (will be used for both normal WHERE and LIMIT workaround)
	whereParts, whereArgs := b.buildWheresWithSoftDelete()

	// LIMIT clause
	// MySQL supports LIMIT directly in DELETE
	// SQLite requires a subquery workaround: DELETE FROM ... WHERE rowid IN (SELECT rowid FROM ... ORDER BY ... LIMIT N)
	// PostgreSQL supports LIMIT directly in DELETE
	if b.query.limit != nil {
		if b.query.driver != nil && b.query.driver.Dialect() == "mysql" {
			// Add WHERE clause if it exists
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
			parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
		} else if b.query.driver != nil && b.query.driver.Dialect() == "sqlite" {
			// SQLite workaround: wrap in subquery with rowid
			if whereParts == "" {
				whereParts = "1=1"
			}
			// Build ORDER BY clause for deterministic row selection
			var orderClause string
			if len(b.query.orders) > 0 {
				var orderParts []string
				for _, order := range b.query.orders {
					orderParts = append(orderParts, fmt.Sprintf("%s %s", order.column, order.direction))
				}
				orderClause = fmt.Sprintf(" ORDER BY %s", strings.Join(orderParts, ", "))
			}
			// Add WHERE clause with rowid subquery including ORDER BY
			parts = append(parts, fmt.Sprintf("WHERE rowid IN (SELECT rowid FROM %s WHERE %s%s LIMIT %d)", b.quoteIdentifier(b.query.table), whereParts, orderClause, *b.query.limit))
			args = append(args, whereArgs...)
		} else if b.query.driver != nil && b.query.driver.Dialect() == "postgres" {
			// PostgreSQL supports LIMIT directly in DELETE
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
			parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
		} else {
			// Other databases: add WHERE clause normally (LIMIT may or may not be supported)
			if whereParts != "" {
				parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
				args = append(args, whereArgs...)
			}
		}
	} else {
		// No LIMIT: add WHERE clause normally
		if whereParts != "" {
			parts = append(parts, fmt.Sprintf("WHERE %s", whereParts))
			args = append(args, whereArgs...)
		}
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
			prefix = fmt.Sprintf("%s IS NOT NULL", b.quoteIdentifier("deleted_at"))
		case b.query.withTrashed:
			// include all rows — no filter
		default:
			prefix = fmt.Sprintf("%s IS NULL", b.quoteIdentifier("deleted_at"))
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

		// Quote identifiers in the WHERE clause
		clauseQuery = b.quoteWhereIdentifiers(clauseQuery)

		parts = append(parts, clauseQuery)
		args = append(args, clauseArgs...)
	}

	return strings.Join(parts, " "), args
}

// quoteWhereIdentifiers quotes column names in WHERE clauses.
// It uses a conservative approach that only quotes simple identifiers
// to avoid breaking complex expressions, function calls, or subqueries.
func (b *Builder) quoteWhereIdentifiers(query string) string {
	// SQL keywords that should never be quoted
	sqlKeywords := map[string]bool{
		"AND": true, "OR": true, "NOT": true, "NULL": true,
		"TRUE": true, "FALSE": true, "IS": true, "IN": true,
		"LIKE": true, "BETWEEN": true, "SELECT": true, "FROM": true,
		"WHERE": true, "JOIN": true, "ON": true, "AS": true,
		"GROUP": true, "ORDER": true, "BY": true, "HAVING": true,
		"LIMIT": true, "OFFSET": true, "CASE": true, "WHEN": true,
		"THEN": true, "ELSE": true, "END": true, "EXISTS": true,
	}

	// Collect all replacements first, then apply them in reverse order
	type replacement struct {
		start int
		end   int
		value string
	}
	var replacements []replacement

	// Tokenize the query to identify potential column names
	// We look for simple identifiers that appear before operators
	// Only use operators with leading space to avoid matching inside quoted identifiers
	operators := []string{" != ", " <> ", " >= ", " <= ", " LIKE ", " NOT LIKE ", " = ", " > ", " < ", " IS ", " IS NOT ", " IN ", " NOT IN ", " BETWEEN ", " NOT BETWEEN "}

	for _, op := range operators {
		start := 0
		for {
			idx := strings.Index(strings.ToUpper(query[start:]), strings.ToUpper(op))
			if idx == -1 {
				break
			}
			idx += start

			// Get text before operator
			beforeOp := query[:idx]
			trimmed := strings.TrimSpace(beforeOp)

			// Get last word (potential column name)
			lastSpace := strings.LastIndex(trimmed, " ")
			var colName string
			if lastSpace == -1 {
				colName = trimmed
			} else {
				colName = trimmed[lastSpace+1:]
			}

			// Only quote if it's a simple identifier:
			// - Not already quoted
			// - Not a SQL keyword
			// - Contains only alphanumeric characters and underscores
			// - Doesn't contain dots (table.column handled separately)
			// - Doesn't contain parentheses (function calls)
			// - Doesn't start with a number
			if colName != "" &&
				!strings.HasPrefix(colName, "\"") && !strings.HasPrefix(colName, "`") &&
				!sqlKeywords[strings.ToUpper(colName)] &&
				isSimpleIdentifier(colName) {
				quotedCol := b.quoteIdentifier(colName)
				// Find the last occurrence of colName in beforeOp
				colIdx := strings.LastIndex(beforeOp, colName)
				if colIdx != -1 {
					replacements = append(replacements, replacement{
						start: colIdx,
						end:   colIdx + len(colName),
						value: quotedCol,
					})
				}
			}
			start = idx + len(op)
		}
	}

	// Sort replacements by start position descending
	for i := 0; i < len(replacements); i++ {
		for j := i + 1; j < len(replacements); j++ {
			if replacements[i].start < replacements[j].start {
				replacements[i], replacements[j] = replacements[j], replacements[i]
			}
		}
	}

	// Apply replacements from end to start
	result := query
	for _, r := range replacements {
		result = result[:r.start] + r.value + result[r.end:]
	}

	return result
}

// isSimpleIdentifier checks if a string is a simple column identifier
// that can be safely quoted. Returns false for:
// - Identifiers with dots (table.column)
// - Identifiers with parentheses (function calls)
// - Identifiers starting with numbers
// - Empty strings
func isSimpleIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Check for dots (table.column) or parentheses (function calls)
	if strings.Contains(s, ".") || strings.Contains(s, "(") || strings.Contains(s, ")") {
		return false
	}

	// Check if starts with a number
	if s[0] >= '0' && s[0] <= '9' {
		return false
	}

	// Check if contains only valid identifier characters
	for _, r := range s {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isDigit := r >= '0' && r <= '9'
		isUnderscore := r == '_'
		if !isLetter && !isDigit && !isUnderscore {
			return false
		}
	}

	return true
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
		// Sort keys for deterministic SQL generation
		keys := v.MapKeys()
		sortedKeys := make([]reflect.Value, len(keys))
		copy(sortedKeys, keys)
		// Sort by string key name
		for i := 0; i < len(sortedKeys); i++ {
			for j := i + 1; j < len(sortedKeys); j++ {
				if sortedKeys[i].String() > sortedKeys[j].String() {
					sortedKeys[i], sortedKeys[j] = sortedKeys[j], sortedKeys[i]
				}
			}
		}
		for _, key := range sortedKeys {
			columns = append(columns, key.String())
			values = append(values, v.MapIndex(key).Interface())
		}
		return columns, values, nil
	}

	// Handle struct using reflection
	if v.Kind() == reflect.Struct {
		cols, vals := b.extractStructColumnsAndValues(v)
		return cols, vals, nil
	}

	return nil, nil, fmt.Errorf("unsupported value type for INSERT: %T", value)
}

func (b *Builder) extractStructColumnsAndValues(v reflect.Value) ([]string, []any) {
	var columns []string
	var values []any

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			cols, vals := b.extractStructColumnsAndValues(fieldValue)
			columns = append(columns, cols...)
			values = append(values, vals...)
			continue
		}

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get column name from tag or field name
		columnName := structFieldColumnName(field)
		if columnName == "" {
			continue
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
			} else if fieldValue.Kind() == reflect.Ptr {
				// Skip nil pointers except for deleted_at (soft delete)
				if columnName != "deleted_at" {
					continue
				}
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

		// Skip zero values except for boolean, time.Time, and deleted_at (soft delete)
		// For deleted_at (nil pointer), we want to include it as NULL in INSERT
		if fieldValue.IsZero() && fieldValue.Kind() != reflect.Bool && fieldValue.Type() != reflect.TypeOf(time.Time{}) && !(columnName == "deleted_at" && fieldValue.Kind() == reflect.Ptr) {
			continue
		}

		columns = append(columns, columnName)
		values = append(values, fieldValue.Interface())
	}
	return columns, values
}

// extractColumnNames extracts column names from a struct without checking values.
// This is used for SELECT clause generation where we want all columns regardless of their values.
func (b *Builder) extractColumnNames(value any) []string {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	return b.extractStructColumnNames(v)
}

func (b *Builder) extractStructColumnNames(v reflect.Value) []string {
	var columns []string

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			columns = append(columns, b.extractStructColumnNames(fieldValue)...)
			continue
		}

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get column name from tag or field name
		columnName := structFieldColumnName(field)
		if columnName == "" {
			continue
		}

		// Skip slice/struct fields that are not handled as basic types
		// But allow pointers to basic types or time.Time
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if (fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Struct) &&
			fieldType != reflect.TypeOf(time.Time{}) {
			continue
		}

		columns = append(columns, columnName)
	}

	return columns
}
