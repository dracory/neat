package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dracory/neat/support/str"
)

// BuildSelect builds a SELECT query from the query state.
func (b *Builder) BuildSelect() (string, []any) {
	if b.query.rawSQL != "" {
		return b.query.rawSQL, b.query.rawArgs
	}

	var parts []string
	args := make([]any, 0)

	// SELECT clause
	if b.query.aggregate != "" {
		// When aggregate is set, ignore SELECT list and use aggregate function
		// Handle COUNT with DISTINCT
		if b.query.aggregate == "COUNT" && b.query.distinct {
			if len(b.query.distinctCols) > 0 {
				parts = append(parts, fmt.Sprintf("SELECT COUNT(DISTINCT %s)", strings.Join(b.query.distinctCols, ", ")))
			} else if len(b.query.selects) > 0 {
				var selectParts []string
				for _, s := range b.query.selects {
					selectParts = append(selectParts, s.expr)
					args = append(args, s.args...)
				}
				parts = append(parts, fmt.Sprintf("SELECT COUNT(DISTINCT %s)", strings.Join(selectParts, ", ")))
			} else {
				parts = append(parts, fmt.Sprintf("SELECT %s(%s)", b.query.aggregate, b.query.aggregateCol))
			}
		} else {
			parts = append(parts, fmt.Sprintf("SELECT %s(%s)", b.query.aggregate, b.query.aggregateCol))
		}
	} else if len(b.query.selects) > 0 {
		// Get placeholder function for the dialect
		placeholderFunc := func(n int) string { return "?" }
		if b.query.driver != nil {
			placeholderFunc = b.query.driver.Placeholder
		}

		// Start placeholder index for SELECT clause
		placeholderIndex := 1
		var selectParts []string
		for _, s := range b.query.selects {
			// Replace ? with dialect-specific placeholder
			// Count placeholders first to avoid infinite loop if placeholderFunc returns "?"
			placeholderCount := strings.Count(s.expr, "?")
			replacedQuery := s.expr
			for i := 0; i < placeholderCount; i++ {
				replacedQuery = strings.Replace(replacedQuery, "?", placeholderFunc(placeholderIndex), 1)
				placeholderIndex++
			}
			selectParts = append(selectParts, replacedQuery)
			args = append(args, s.args...)
		}

		// Add count subqueries
		for _, cq := range b.query.withCountQueries {
			countSubquery := b.buildCountSubquery(cq, placeholderFunc, placeholderIndex)
			if countSubquery.sql == "" {
				continue // Skip this subquery if validation failed
			}
			selectParts = append(selectParts, fmt.Sprintf("(%s) AS %s", countSubquery.sql, b.quoteIdentifier(cq.column)))
			args = append(args, countSubquery.args...)
			placeholderIndex += len(countSubquery.args)
		}

		// Add exists subqueries
		for _, eq := range b.query.withExistsQueries {
			existsSubquery := b.buildExistsSubquery(eq, placeholderFunc, placeholderIndex)
			if existsSubquery.sql == "" {
				continue // Skip this subquery if validation failed
			}
			existsAlias := str.Of(eq.relation).Snake().String() + "_exists"
			selectParts = append(selectParts, fmt.Sprintf("(%s) AS %s", existsSubquery.sql, b.quoteIdentifier(existsAlias)))
			args = append(args, existsSubquery.args...)
			placeholderIndex += len(existsSubquery.args)
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

		// Add count and exists subqueries to default SELECT
		placeholderIndex := 1
		if len(b.query.withCountQueries) > 0 || len(b.query.withExistsQueries) > 0 {
			placeholderFunc := func(n int) string { return "?" }
			if b.query.driver != nil {
				placeholderFunc = b.query.driver.Placeholder
			}

			for _, cq := range b.query.withCountQueries {
				countSubquery := b.buildCountSubquery(cq, placeholderFunc, placeholderIndex)
				if countSubquery.sql == "" {
					continue // Skip this subquery if validation failed
				}
				parts[0] = parts[0] + fmt.Sprintf(", (%s) AS %s", countSubquery.sql, b.quoteIdentifier(cq.column))
				args = append(args, countSubquery.args...)
				placeholderIndex += len(countSubquery.args)
			}

			for _, eq := range b.query.withExistsQueries {
				existsSubquery := b.buildExistsSubquery(eq, placeholderFunc, placeholderIndex)
				if existsSubquery.sql == "" {
					continue // Skip this subquery if validation failed
				}
				existsAlias := str.Of(eq.relation).Snake().String() + "_exists"
				parts[0] = parts[0] + fmt.Sprintf(", (%s) AS %s", existsSubquery.sql, b.quoteIdentifier(existsAlias))
				args = append(args, existsSubquery.args...)
				placeholderIndex += len(existsSubquery.args)
			}
		}
	}

	// FROM clause
	if b.query.table != "" {
		if strings.Contains(b.query.table, "(") && strings.Contains(b.query.table, ")") {
			// Subquery in FROM, don't quote
			// Get placeholder function for the dialect
			placeholderFunc := func(n int) string { return "?" }
			if b.query.driver != nil {
				placeholderFunc = b.query.driver.Placeholder
			}
			// Start placeholder index after SELECT clause parameters
			placeholderIndex := len(args) + 1
			// Replace ? with dialect-specific placeholder
			// Count placeholders first to avoid infinite loop if placeholderFunc returns "?"
			placeholderCount := strings.Count(b.query.table, "?")
			replacedQuery := b.query.table
			for i := 0; i < placeholderCount; i++ {
				replacedQuery = strings.Replace(replacedQuery, "?", placeholderFunc(placeholderIndex), 1)
				placeholderIndex++
			}
			parts = append(parts, fmt.Sprintf("FROM %s", replacedQuery))
		} else {
			parts = append(parts, fmt.Sprintf("FROM %s", b.quoteIdentifier(b.query.table)))
		}
		args = append(args, b.query.tableArgs...)
	}

	// JOIN clauses
	if len(b.query.joins) > 0 {
		// Get placeholder function for the dialect
		placeholderFunc := func(n int) string { return "?" }
		if b.query.driver != nil {
			placeholderFunc = b.query.driver.Placeholder
		}

		// Start placeholder index after FROM clause parameters
		placeholderIndex := len(args) + 1
		for _, join := range b.query.joins {
			// Replace ? with dialect-specific placeholder
			// Count placeholders first to avoid infinite loop if placeholderFunc returns "?"
			placeholderCount := strings.Count(join.query, "?")
			replacedQuery := join.query
			for i := 0; i < placeholderCount; i++ {
				replacedQuery = strings.Replace(replacedQuery, "?", placeholderFunc(placeholderIndex), 1)
				placeholderIndex++
			}
			parts = append(parts, fmt.Sprintf("%s %s", join._type, replacedQuery))
			args = append(args, join.args...)
		}
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
		// Get placeholder function for the dialect
		placeholderFunc := func(n int) string { return "?" }
		if b.query.driver != nil {
			placeholderFunc = b.query.driver.Placeholder
		}

		// Start placeholder index after WHERE clause parameters
		placeholderIndex := len(args) + 1
		var havingParts []string
		for _, having := range b.query.havings {
			// Replace ? with dialect-specific placeholder
			// Count placeholders first to avoid infinite loop if placeholderFunc returns "?"
			placeholderCount := strings.Count(having.query, "?")
			replacedQuery := having.query
			for i := 0; i < placeholderCount; i++ {
				replacedQuery = strings.Replace(replacedQuery, "?", placeholderFunc(placeholderIndex), 1)
				placeholderIndex++
			}
			havingParts = append(havingParts, replacedQuery)
			args = append(args, having.args...)
		}
		parts = append(parts, fmt.Sprintf("HAVING %s", strings.Join(havingParts, " AND ")))
	}

	// ORDER BY clauses - skip for aggregate queries
	if b.query.aggregate == "" && len(b.query.orders) > 0 {
		var orderParts []string
		for _, order := range b.query.orders {
			orderParts = append(orderParts, fmt.Sprintf("%s %s", order.column, order.direction))
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(orderParts, ", ")))
	}

	// LIMIT clause - skip for aggregate queries
	if b.query.aggregate == "" && b.query.limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *b.query.limit))
	}

	// OFFSET clause - skip for aggregate queries
	if b.query.aggregate == "" && b.query.offset != nil {
		// SQLite requires LIMIT when using OFFSET
		if b.query.limit == nil && b.query.driver != nil && b.query.driver.Dialect() == "sqlite" {
			parts = append(parts, "LIMIT -1")
		}
		parts = append(parts, fmt.Sprintf("OFFSET %d", *b.query.offset))
	}

	// Locking clauses - skip for aggregate queries
	// Skip lock clauses for SQLite as it doesn't support them
	if b.query.aggregate == "" && (b.query.driver == nil || b.query.driver.Dialect() != "sqlite") {
		if b.query.lockForUpdate {
			parts = append(parts, "FOR UPDATE")
		} else if b.query.sharedLock {
			if b.query.driver != nil && b.query.driver.Dialect() == "postgres" {
				parts = append(parts, "FOR SHARE")
			} else {
				parts = append(parts, "LOCK IN SHARE MODE")
			}
		}
	}

	return strings.Join(parts, " "), args
}

type subqueryResult struct {
	sql  string
	args []any
}

// buildCountSubquery builds a COUNT subquery for a relationship.
func (b *Builder) buildCountSubquery(cq countQuery, placeholderFunc func(int) string, startIndex int) subqueryResult {
	// Infer the related table name from the relation name
	relationName := str.Of(cq.relation).Snake().String()
	relationTable := relationName
	// Simple pluralization: add 's' if not already ending with 's'
	if !strings.HasSuffix(relationName, "s") {
		relationTable = relationName + "s"
	}

	// Get the parent type name to build the foreign key
	parentTypeName := ""
	if b.query.model != nil {
		v := reflect.ValueOf(b.query.model)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			parentTypeName = v.Type().Name()
		}
	}

	// Build foreign key column name
	// If model is not available, use the table name (singularized)
	foreignKeyColumn := ""
	if parentTypeName != "" {
		foreignKeyColumn = str.Of(parentTypeName).Snake().String() + "_id"
	} else if b.query.table != "" {
		// Use table name and remove trailing 's' if present
		tableName := strings.TrimSuffix(b.query.table, "s")
		foreignKeyColumn = tableName + "_id"
	}

	// Validate that foreign key column was determined
	if foreignKeyColumn == "" {
		return subqueryResult{sql: "", args: nil}
	}

	// Build the count subquery
	// SELECT COUNT(*) FROM related_table WHERE foreign_key = main_table.id
	subquerySQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = %s.id",
		b.quoteIdentifier(relationTable),
		b.quoteIdentifier(foreignKeyColumn),
		b.quoteIdentifier(b.query.table))

	// If a constraint is provided, apply it to the subquery
	if cq.constraint != nil {
		// Create a temporary query to build the constraint
		tempQuery := b.query.newQuery()
		tempQuery.table = relationTable
		tempQuery.driver = b.query.driver // Set driver for dialect support
		tempQuery = cq.constraint(tempQuery).(*Query)

		// Build the WHERE clause from the constraint
		if len(tempQuery.wheres) > 0 {
			whereParts, whereArgs := b.buildWheresFromSlice(tempQuery.wheres, placeholderFunc, startIndex)
			if whereParts != "" {
				subquerySQL += " AND " + whereParts
				return subqueryResult{sql: subquerySQL, args: whereArgs}
			}
		}
	}

	return subqueryResult{sql: subquerySQL, args: []any{}}
}

// buildExistsSubquery builds an EXISTS subquery for a relationship.
func (b *Builder) buildExistsSubquery(eq existsQuery, placeholderFunc func(int) string, startIndex int) subqueryResult {
	// Infer the related table name from the relation name
	relationName := str.Of(eq.relation).Snake().String()
	relationTable := relationName
	// Simple pluralization: add 's' if not already ending with 's'
	if !strings.HasSuffix(relationName, "s") {
		relationTable = relationName + "s"
	}

	// Get the parent type name to build the foreign key
	parentTypeName := ""
	if b.query.model != nil {
		v := reflect.ValueOf(b.query.model)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			parentTypeName = v.Type().Name()
		}
	}

	// Build foreign key column name
	// If model is not available, use the table name (singularized)
	foreignKeyColumn := ""
	if parentTypeName != "" {
		foreignKeyColumn = str.Of(parentTypeName).Snake().String() + "_id"
	} else if b.query.table != "" {
		// Use table name and remove trailing 's' if present
		tableName := strings.TrimSuffix(b.query.table, "s")
		foreignKeyColumn = tableName + "_id"
	}

	// Validate that foreign key column was determined
	if foreignKeyColumn == "" {
		return subqueryResult{sql: "", args: nil}
	}

	// Build the exists subquery
	// SELECT EXISTS(SELECT 1 FROM related_table WHERE foreign_key = main_table.id)
	subquerySQL := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = %s.id",
		b.quoteIdentifier(relationTable),
		b.quoteIdentifier(foreignKeyColumn),
		b.quoteIdentifier(b.query.table))

	// If a constraint is provided, apply it to the subquery
	if eq.constraint != nil {
		// Create a temporary query to build the constraint
		tempQuery := b.query.newQuery()
		tempQuery.table = relationTable
		tempQuery.driver = b.query.driver // Set driver for dialect support
		tempQuery = eq.constraint(tempQuery).(*Query)

		// Build the WHERE clause from the constraint
		if len(tempQuery.wheres) > 0 {
			whereParts, whereArgs := b.buildWheresFromSlice(tempQuery.wheres, placeholderFunc, startIndex)
			if whereParts != "" {
				subquerySQL += " AND " + whereParts
				return subqueryResult{sql: subquerySQL + ")", args: whereArgs}
			}
		}
	}

	return subqueryResult{sql: subquerySQL + ")", args: []any{}}
}

// buildWheresFromSlice builds WHERE clauses from a whereClause slice with custom placeholder index.
func (b *Builder) buildWheresFromSlice(wheres []whereClause, placeholderFunc func(int) string, startIndex int) (string, []any) {
	if len(wheres) == 0 {
		return "", nil
	}

	var parts []string
	args := []any{}
	placeholderIndex := startIndex

	for i, w := range wheres {
		// Replace ? with dialect-specific placeholder
		placeholderCount := strings.Count(w.query, "?")
		replacedQuery := w.query
		for j := 0; j < placeholderCount; j++ {
			replacedQuery = strings.Replace(replacedQuery, "?", placeholderFunc(placeholderIndex), 1)
			placeholderIndex++
		}

		if i == 0 {
			parts = append(parts, replacedQuery)
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", w._type, replacedQuery))
		}
		args = append(args, w.args...)
	}

	return strings.Join(parts, " "), args
}
