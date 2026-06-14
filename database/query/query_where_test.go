package query

import (
	"context"
	"strings"
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/schema/constants"
)

func TestWhereIn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereIn("id", []any{1, 2, 3})

	if result == nil {
		t.Error("Expected non-nil Query from WhereIn")
	}
}

func TestOrWhereIn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereIn("id", []any{1, 2, 3})

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereIn")
	}
}

func TestWhereNotIn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNotIn("id", []any{1, 2, 3})

	if result == nil {
		t.Error("Expected non-nil Query from WhereNotIn")
	}
}

func TestOrWhereNotIn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereNotIn("id", []any{1, 2, 3})

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNotIn")
	}
}

func TestWhereBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereBetween("age", 18, 65)

	if result == nil {
		t.Error("Expected non-nil Query from WhereBetween")
	}
}

func TestOrWhereBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereBetween("age", 18, 65)

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereBetween")
	}
}

func TestWhereNotBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNotBetween("age", 18, 65)

	if result == nil {
		t.Error("Expected non-nil Query from WhereNotBetween")
	}
}

func TestOrWhereNotBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereNotBetween("age", 18, 65)

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNotBetween")
	}
}

func TestWhereNull(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNull(constants.DeletedAtColumnName)

	if result == nil {
		t.Error("Expected non-nil Query from WhereNull")
	}
}

func TestOrWhereNull(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereNull(constants.DeletedAtColumnName)

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNull")
	}
}

func TestWhereAnyWithEmptyColumns(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereAny([]string{}, "=", "test")

	// Should return the same query without adding a where clause
	if result == nil {
		t.Error("Expected non-nil Query from WhereAny with empty columns")
	}
	if len(result.(*Query).wheres) != 0 {
		t.Error("Expected no where clauses to be added for empty columns")
	}
}

func TestWhereAllWithEmptyColumns(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereAll([]string{}, "=", "test")

	// Should return the same query without adding a where clause
	if result == nil {
		t.Error("Expected non-nil Query from WhereAll with empty columns")
	}
	if len(result.(*Query).wheres) != 0 {
		t.Error("Expected no where clauses to be added for empty columns")
	}
}

func TestWhereNoneWithEmptyColumns(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNone([]string{}, "=", "test")

	// Should return the same query without adding a where clause
	if result == nil {
		t.Error("Expected non-nil Query from WhereNone with empty columns")
	}
	if len(result.(*Query).wheres) != 0 {
		t.Error("Expected no where clauses to be added for empty columns")
	}
}

func TestWhereNotWithClosure(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	result := q.WhereNot(func(q orm.Query) orm.Query {
		return q.Where("name = ?", "test")
	})

	if result == nil {
		t.Error("Expected non-nil Query from WhereNot with closure")
	}

	// Verify the where clause was wrapped in NOT
	qResult := result.(*Query)
	if len(qResult.wheres) != 1 {
		t.Errorf("Expected 1 where clause, got %d", len(qResult.wheres))
	}
	if !strings.Contains(qResult.wheres[0].query, "NOT") {
		t.Error("Expected where clause to contain NOT")
	}
}

func TestOrWhereNotWithClosure(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	result := q.OrWhereNot(func(q orm.Query) orm.Query {
		return q.Where("name = ?", "test")
	})

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNot with closure")
	}

	// Verify the where clause was wrapped in NOT and set to OR
	qResult := result.(*Query)
	if len(qResult.wheres) != 1 {
		t.Errorf("Expected 1 where clause, got %d", len(qResult.wheres))
	}
	if !strings.Contains(qResult.wheres[0].query, "NOT") {
		t.Error("Expected where clause to contain NOT")
	}
	if qResult.wheres[0]._type != "or" {
		t.Errorf("Expected where clause type to be 'or', got '%s'", qResult.wheres[0]._type)
	}
	// Check the actual query string
	t.Logf("OrWhereNot closure query: %s", qResult.wheres[0].query)
}

func TestWhereNotNull(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNotNull(constants.DeletedAtColumnName)

	if result == nil {
		t.Error("Expected non-nil Query from WhereNotNull")
	}
}

func TestWhereColumn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereColumn("updated_at", "=", "created_at")

	if result == nil {
		t.Error("Expected non-nil Query from WhereColumn")
	}
}

func TestOrWhereColumn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereColumn("updated_at", "=", "created_at")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereColumn")
	}
}

func TestWhereExists(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereExists(func(q orm.Query) orm.Query {
		return q
	})

	if result == nil {
		t.Error("Expected non-nil Query from WhereExists")
	}
}

func TestWhereExistsSubquerySQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.WhereExists(func(sub orm.Query) orm.Query {
		return sub.Table("orders").Where("orders.user_id = users.id")
	})

	sql, _ := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, "EXISTS") {
		t.Errorf("Expected SQL to contain 'EXISTS', got: %s", sql)
	}
	if !strings.Contains(sql, "orders") {
		t.Errorf("Expected SQL to contain subquery table 'orders', got: %s", sql)
	}
	if !strings.Contains(sql, "SELECT") {
		t.Errorf("Expected SQL to contain SELECT from subquery, got: %s", sql)
	}
}

func TestWhereExistsWithArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.WhereExists(func(sub orm.Query) orm.Query {
		return sub.Table("orders").Where("orders.user_id = ?", 5)
	})

	sql, args := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, "EXISTS") {
		t.Errorf("Expected SQL to contain 'EXISTS', got: %s", sql)
	}
	if len(args) != 1 || args[0] != 5 {
		t.Errorf("Expected args [5] from subquery, got %v", args)
	}
}

func TestWhereNotExistsViaWhereNot(t *testing.T) {
	// WhereNot wraps an expression in NOT(…); used to negate an EXISTS subquery.
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.WhereNot(func(sub orm.Query) orm.Query {
		return sub.WhereExists(func(inner orm.Query) orm.Query {
			return inner.Table("bans").Where("bans.user_id = users.id")
		})
	})

	sql, _ := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, "NOT") {
		t.Errorf("Expected SQL to contain 'NOT', got: %s", sql)
	}
	if !strings.Contains(sql, "EXISTS") {
		t.Errorf("Expected SQL to contain 'EXISTS', got: %s", sql)
	}
}

func TestNestedWhereExists(t *testing.T) {
	// Outer WHERE EXISTS + inner WHERE EXISTS (nested)
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.WhereExists(func(sub orm.Query) orm.Query {
		return sub.Table("orders").WhereExists(func(inner orm.Query) orm.Query {
			return inner.Table("order_items").Where("order_items.order_id = orders.id")
		})
	})

	sql, _ := NewBuilder(q).BuildSelect()

	if strings.Count(sql, "EXISTS") < 2 {
		t.Errorf("Expected at least 2 'EXISTS' in nested SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "order_items") {
		t.Errorf("Expected SQL to contain inner subquery table 'order_items', got: %s", sql)
	}
}

func TestWhereNot(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNot("id = ?", 1)

	if result == nil {
		t.Error("Expected non-nil Query from WhereNot")
	}
}

func TestOrWhereNot(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereNot("id = ?", 1)

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNot")
	}
}

func TestWhereAny(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereAny([]string{"name", "email"}, "=", "John")

	if result == nil {
		t.Error("Expected non-nil Query from WhereAny")
	}
}

func TestWhereAll(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereAll([]string{"name", "email"}, "=", "John")

	if result == nil {
		t.Error("Expected non-nil Query from WhereAll")
	}
}

func TestWhereNone(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNone([]string{"name", "email"}, "=", "John")

	if result == nil {
		t.Error("Expected non-nil Query from WhereNone")
	}
}

func TestWhereJsonContains(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonContains("data", "value")

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonContains")
	}
}

func TestOrWhereJsonContains(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereJsonContains("data", "value")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereJsonContains")
	}
}

func TestWhereJsonDoesntContain(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonDoesntContain("data", "value")

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonDoesntContain")
	}
}

func TestOrWhereJsonDoesntContain(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereJsonDoesntContain("data", "value")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereJsonDoesntContain")
	}
}

func TestWhereJsonContainsKey(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonContainsKey("data->key")

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonContainsKey")
	}
}

func TestOrWhereJsonContainsKey(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereJsonContainsKey("data->key")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereJsonContainsKey")
	}
}

func TestWhereJsonDoesntContainKey(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonDoesntContainKey("data->key")

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonDoesntContainKey")
	}
}

func TestOrWhereJsonDoesntContainKey(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereJsonDoesntContainKey("data->key")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereJsonDoesntContainKey")
	}
}

func TestWhereJsonLength(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonLength("data", ">", 0)

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonLength")
	}
}

// --- SQL Output Verification Tests ---

func TestWhereIn_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereIn("id", []any{1, 2, 3})

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}

	// Verify IN clause expansion (SQLite uses double quotes for identifiers)
	if !contains(sql, "IN (?, ?, ?)") && !contains(sql, "IN (?,?,?)") {
		t.Errorf("Expected IN clause with 3 placeholders, got: %s", sql)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 arguments, got %d", len(args))
	}

	if args[0] != 1 || args[1] != 2 || args[2] != 3 {
		t.Errorf("Expected args [1,2,3], got %v", args)
	}
}

func TestOrWhereIn_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.Where("status = ?", "active")
	q.OrWhereIn("id", []any{4, 5})

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "OR") {
		t.Error("Expected OR in WHERE clause")
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 arguments (status + 2 ids), got %d", len(args))
	}
}

func TestWhereBetween_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereBetween("age", 18, 65)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "BETWEEN") {
		t.Error("Expected BETWEEN in WHERE clause")
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 arguments for BETWEEN, got %d", len(args))
	}

	if args[0] != 18 || args[1] != 65 {
		t.Errorf("Expected args [18,65], got %v", args)
	}
}

func TestWhereNull_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereNull(constants.DeletedAtColumnName)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "IS NULL") {
		t.Error("Expected IS NULL in WHERE clause")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments for IS NULL, got %d", len(args))
	}
}

func TestWhereNotNull_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereNotNull(constants.DeletedAtColumnName)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "IS NOT NULL") {
		t.Error("Expected IS NOT NULL in WHERE clause")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments for IS NOT NULL, got %d", len(args))
	}
}

func TestWhereColumn_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereColumn("updated_at", "=", "created_at")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "updated_at") || !contains(sql, "created_at") {
		t.Error("Expected both columns in WHERE clause")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments for column comparison, got %d", len(args))
	}
}

func TestWhereNot_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereNot("id = ?", 1)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "NOT") {
		t.Error("Expected NOT in WHERE clause")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestWhereAny_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereAny([]string{"name", "email"}, "=", "John")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "name") || !contains(sql, "email") {
		t.Error("Expected both columns in WHERE clause")
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 arguments for WHERE ANY, got %d", len(args))
	}
}

func TestWhereAll_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereAll([]string{"name", "email"}, "=", "John")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "name") || !contains(sql, "email") {
		t.Error("Expected both columns in WHERE clause")
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 arguments for WHERE ALL, got %d", len(args))
	}
}

func TestWhereNone_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereNone([]string{"name", "email"}, "=", "John")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "name") || !contains(sql, "email") {
		t.Error("Expected both columns in WHERE clause")
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 arguments for WHERE NONE, got %d", len(args))
	}
}

// --- Dialect-Specific SQL Tests ---

func TestWhereIn_MySqlDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "users", nil, nil)
	q.WhereIn("id", []any{1, 2, 3})

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// MySQL uses backticks for identifiers
	if !contains(sql, "`id`") {
		t.Errorf("Expected MySQL backtick quoting, got: %s", sql)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 arguments, got %d", len(args))
	}
}

func TestWhereIn_PostgresDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewPostgreSQL(), "users", nil, nil)
	q.WhereIn("id", []any{1, 2, 3})

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// PostgreSQL uses double quotes for identifiers
	if !contains(sql, `"id"`) {
		t.Errorf("Expected PostgreSQL double quote quoting, got: %s", sql)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 arguments, got %d", len(args))
	}
}

func TestWhereBetween_DialectComparison(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
		quote  string
	}{
		{"MySQL", driver.NewMySQL(), "`"},
		{"PostgreSQL", driver.NewPostgreSQL(), `"`},
		{"SQLite", driver.NewSQLite(), `"`},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.TODO(), nil, tc.driver, "users", nil, nil)
			q.WhereBetween("age", 18, 65)

			builder := NewBuilder(q)
			sql, args := builder.BuildSelect()

			if !contains(sql, tc.quote+"age"+tc.quote) {
				t.Errorf("Expected %s quoting for age, got: %s", tc.name, sql)
			}

			if len(args) != 2 {
				t.Errorf("Expected 2 arguments, got %d", len(args))
			}
		})
	}
}

// --- Complex WHERE Clause Combinations ---

func TestWhereMultipleConditions(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.Where("status = ?", "active")
	q.Where("age > ?", 18)
	q.Where("country = ?", "US")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Should have AND between conditions
	andCount := countOccurrences(sql, "AND")
	if andCount < 2 {
		t.Errorf("Expected at least 2 ANDs for 3 conditions, got %d in: %s", andCount, sql)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 arguments, got %d", len(args))
	}
}

func TestWhereAndOrCombination(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.Where("status = ?", "active")
	q.OrWhere("status = ?", "pending")
	q.Where("age > ?", 18)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "OR") {
		t.Error("Expected OR in WHERE clause")
	}

	if !contains(sql, "AND") {
		t.Error("Expected AND in WHERE clause")
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 arguments, got %d", len(args))
	}
}

func TestWhereInAndBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereIn("id", []any{1, 2, 3})
	q.WhereBetween("age", 18, 65)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "IN") {
		t.Error("Expected IN in WHERE clause")
	}

	if !contains(sql, "BETWEEN") {
		t.Error("Expected BETWEEN in WHERE clause")
	}

	if len(args) != 5 { // 3 for IN + 2 for BETWEEN
		t.Errorf("Expected 5 arguments, got %d", len(args))
	}
}

func TestWhereNullAndNotNull(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereNull(constants.DeletedAtColumnName)
	q.WhereNotNull("email")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "IS NULL") {
		t.Error("Expected IS NULL in WHERE clause")
	}

	if !contains(sql, "IS NOT NULL") {
		t.Error("Expected IS NOT NULL in WHERE clause")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments for NULL checks, got %d", len(args))
	}
}

func TestWhereNestedConditions(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.Where("(status = ? OR status = ?)", "active", "pending")
	q.Where("age > ?", 18)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "(") || !contains(sql, ")") {
		t.Error("Expected parentheses for grouping")
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 arguments, got %d", len(args))
	}
}

func TestWhereColumnComparison(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereColumn("created_at", "<", "updated_at")
	q.Where("status = ?", "active")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "created_at") || !contains(sql, "updated_at") {
		t.Error("Expected both columns in comparison")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument for status check, got %d", len(args))
	}
}

// --- JSON WHERE Clause Tests ---

func TestWhereJsonContains_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereJsonContains("data", "value")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "data") {
		t.Error("Expected data column in WHERE clause")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestWhereJsonContainsKey_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereJsonContainsKey("data->key")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// The JSON path should be present in the SQL (may be quoted)
	if !contains(sql, "data") && !contains(sql, "key") {
		t.Error("Expected JSON path components in WHERE clause")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments for key check, got %d", len(args))
	}
}

func TestWhereJsonLength_SqlOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereJsonLength("data", ">", 0)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "data") {
		t.Error("Expected data column in WHERE clause")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

// --- Dialect-Specific JSON WHERE Clause Tests ---

func TestWhereJsonContains_SqliteDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "users", nil, nil)
	q.WhereJsonContains("data", "value")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// SQLite uses json_extract() for JSON operations
	if !contains(sql, "json_extract") {
		t.Errorf("Expected SQLite json_extract() function, got: %s", sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}

	if args[0] != "value" {
		t.Errorf("Expected 'value' argument, got %v", args[0])
	}
}

func TestWhereJsonContains_MySqlDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.WhereJsonContains("data", "value")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// MySQL uses JSON_CONTAINS() function
	if !contains(sql, "JSON_CONTAINS") {
		t.Errorf("Expected MySQL JSON_CONTAINS() function, got: %s", sql)
	}

	// MySQL uses backticks for table name
	if !contains(sql, "`users`") {
		t.Errorf("Expected MySQL backtick quoting for table, got: %s", sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestWhereJsonContains_PostgresDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewPostgreSQL(), "", nil, nil)
	q.Table("users")
	q.WhereJsonContains("data", "value")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// PostgreSQL uses @> operator for JSONB containment
	if !contains(sql, "@>") {
		t.Errorf("Expected PostgreSQL @> operator for JSONB containment, got: %s", sql)
	}

	// PostgreSQL uses double quotes for table name
	if !contains(sql, `"users"`) {
		t.Errorf("Expected PostgreSQL double quote quoting for table, got: %s", sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestWhereJsonContainsKey_SqliteDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "", nil, nil)
	q.Table("users")
	q.WhereJsonContainsKey("data->key")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// SQLite uses json_type() with path for key existence
	if !contains(sql, "json_type") {
		t.Errorf("Expected SQLite json_type() function, got: %s", sql)
	}

	// JSON path should be present
	if !contains(sql, "key") {
		t.Error("Expected JSON key in WHERE clause")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments for key check, got %d", len(args))
	}
}

func TestWhereJsonContainsKey_MySqlDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.WhereJsonContainsKey("data->key")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// MySQL uses JSON_CONTAINS_PATH() or JSON_EXTRACT()
	if !contains(sql, "JSON_CONTAINS_PATH") && !contains(sql, "JSON_EXTRACT") {
		t.Errorf("Expected MySQL JSON function (JSON_CONTAINS_PATH or JSON_EXTRACT), got: %s", sql)
	}

	// MySQL uses backticks for table name
	if !contains(sql, "`users`") {
		t.Errorf("Expected MySQL backtick quoting for table, got: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments for key check, got %d", len(args))
	}
}

func TestWhereJsonContainsKey_PostgresDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewPostgreSQL(), "", nil, nil)
	q.Table("users")
	q.WhereJsonContainsKey("data->key")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// PostgreSQL uses -> operator with IS NOT NULL for key existence
	if !contains(sql, "->") {
		t.Errorf("Expected PostgreSQL -> operator for JSONB key existence, got: %s", sql)
	}
	if !contains(sql, "IS NOT NULL") {
		t.Errorf("Expected IS NOT NULL for key existence check, got: %s", sql)
	}

	// PostgreSQL uses double quotes for table name
	if !contains(sql, `"users"`) {
		t.Errorf("Expected PostgreSQL double quote quoting for table, got: %s", sql)
	}

	// PostgreSQL key existence with path should have 0 arguments (path is embedded in SQL)
	if len(args) != 0 {
		t.Errorf("Expected 0 arguments for key path, got %d", len(args))
	}
}

func TestWhereJsonLength_SqliteDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "", nil, nil)
	q.Table("users")
	q.WhereJsonLength("data", ">", 0)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// SQLite uses json_array_length() function
	if !contains(sql, "json_array_length") {
		t.Errorf("Expected SQLite json_array_length() function, got: %s", sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestWhereJsonLength_MySqlDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.WhereJsonLength("data", ">", 0)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// MySQL uses JSON_LENGTH() function
	if !contains(sql, "JSON_LENGTH") {
		t.Errorf("Expected MySQL JSON_LENGTH() function, got: %s", sql)
	}

	// MySQL uses backticks for table name
	if !contains(sql, "`users`") {
		t.Errorf("Expected MySQL backtick quoting for table, got: %s", sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestWhereJsonLength_PostgresDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewPostgreSQL(), "", nil, nil)
	q.Table("users")
	q.WhereJsonLength("data", ">", 0)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// PostgreSQL uses jsonb_array_length() function
	if !contains(sql, "jsonb_array_length") {
		t.Errorf("Expected PostgreSQL jsonb_array_length() function, got: %s", sql)
	}

	// PostgreSQL uses double quotes for table name
	if !contains(sql, `"users"`) {
		t.Errorf("Expected PostgreSQL double quote quoting for table, got: %s", sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestSplitJsonColumn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	tests := []struct {
		name           string
		input          string
		expectedColumn string
		expectedPath   string
	}{
		{
			name:           "no path",
			input:          "data",
			expectedColumn: "data",
			expectedPath:   "",
		},
		{
			name:           "single level path",
			input:          "data->key",
			expectedColumn: "data",
			expectedPath:   "$.key",
		},
		{
			name:           "nested path",
			input:          "data->meta->active",
			expectedColumn: "data",
			expectedPath:   "$.meta.active",
		},
		{
			name:           "deeply nested path",
			input:          "data->level1->level2->level3",
			expectedColumn: "data",
			expectedPath:   "$.level1.level2.level3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			col, path := q.splitJsonColumn(tc.input)
			if col != tc.expectedColumn {
				t.Errorf("Expected column %q, got %q", tc.expectedColumn, col)
			}
			if path != tc.expectedPath {
				t.Errorf("Expected path %q, got %q", tc.expectedPath, path)
			}
		})
	}
}

func TestJsonPathHandling_NestedPath(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"SQLite", driver.NewSQLite()},
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.TODO(), nil, tc.driver, "users", nil, nil)
			q.WhereJsonContainsKey("data->nested->key")

			builder := NewBuilder(q)
			sql, args := builder.BuildSelect()

			// Should handle nested JSON paths
			if !contains(sql, "data") {
				t.Errorf("Expected data column in %s, got: %s", tc.name, sql)
			}

			if !contains(sql, "key") {
				t.Errorf("Expected key in %s, got: %s", tc.name, sql)
			}

			// For SQLite, verify that -> is replaced with . in the JSON path
			if tc.name == "SQLite" {
				if contains(sql, "->") {
					t.Errorf("SQLite should not contain -> in JSON path, got: %s", sql)
				}
				// Should contain dot notation for nested paths
				if !contains(sql, ".nested.") {
					t.Errorf("SQLite should use dot notation for nested paths, got: %s", sql)
				}
			}

			if len(args) != 0 {
				t.Errorf("Expected 0 arguments for nested key check in %s, got %d", tc.name, len(args))
			}
		})
	}
}

func TestJsonPathHandling_ArrayIndex(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"SQLite", driver.NewSQLite()},
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.TODO(), nil, tc.driver, "users", nil, nil)
			q.WhereJsonContainsKey("data->items[0]")

			builder := NewBuilder(q)
			sql, args := builder.BuildSelect()

			// Should handle array index in JSON paths
			if !contains(sql, "data") {
				t.Errorf("Expected data column in %s, got: %s", tc.name, sql)
			}

			if len(args) != 0 {
				t.Errorf("Expected 0 arguments for array index check in %s, got %d", tc.name, len(args))
			}
		})
	}
}

func TestJsonOperatorDifferences_Comparison(t *testing.T) {
	testCases := []struct {
		name         string
		driver       driver.Driver
		expectedFunc string
		expectedOp   string
	}{
		{"SQLite", driver.NewSQLite(), "json_extract", ""},
		{"MySQL", driver.NewMySQL(), "JSON_CONTAINS", ""},
		{"PostgreSQL", driver.NewPostgreSQL(), "", "@>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.TODO(), nil, tc.driver, "users", nil, nil)
			q.WhereJsonContains("data", "value")

			builder := NewBuilder(q)
			sql, _ := builder.BuildSelect()

			if tc.expectedFunc != "" && !contains(sql, tc.expectedFunc) {
				t.Errorf("Expected %s function in %s, got: %s", tc.expectedFunc, tc.name, sql)
			}

			if tc.expectedOp != "" && !contains(sql, tc.expectedOp) {
				t.Errorf("Expected %s operator in %s, got: %s", tc.expectedOp, tc.name, sql)
			}
		})
	}
}

func TestJsonOperatorDifferences_KeyExistence(t *testing.T) {
	testCases := []struct {
		name         string
		driver       driver.Driver
		expectedFunc string
		expectedOp   string
	}{
		{"SQLite", driver.NewSQLite(), "json_type", ""},
		{"MySQL", driver.NewMySQL(), "JSON_CONTAINS_PATH", ""},
		{"PostgreSQL", driver.NewPostgreSQL(), "", "->"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.TODO(), nil, tc.driver, "users", nil, nil)
			q.WhereJsonContainsKey("data->key")

			builder := NewBuilder(q)
			sql, _ := builder.BuildSelect()

			if tc.expectedFunc != "" && !contains(sql, tc.expectedFunc) {
				t.Errorf("Expected %s function in %s, got: %s", tc.expectedFunc, tc.name, sql)
			}

			if tc.expectedOp != "" && !contains(sql, tc.expectedOp) {
				t.Errorf("Expected %s operator in %s, got: %s", tc.expectedOp, tc.name, sql)
			}
		})
	}
}

func TestJsonOperatorDifferences_ArrayLength(t *testing.T) {
	testCases := []struct {
		name         string
		driver       driver.Driver
		expectedFunc string
	}{
		{"SQLite", driver.NewSQLite(), "json_array_length"},
		{"MySQL", driver.NewMySQL(), "JSON_LENGTH"},
		{"PostgreSQL", driver.NewPostgreSQL(), "jsonb_array_length"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.TODO(), nil, tc.driver, "users", nil, nil)
			q.WhereJsonLength("data", ">", 0)

			builder := NewBuilder(q)
			sql, _ := builder.BuildSelect()

			if !contains(sql, tc.expectedFunc) {
				t.Errorf("Expected %s function in %s, got: %s", tc.expectedFunc, tc.name, sql)
			}
		})
	}
}

// countOccurrences counts how many times substr appears in s
func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
		}
	}
	return count
}

func TestJsonPathSyntax_MySQL(t *testing.T) {
	mysqlDriver := driver.NewMySQL()

	t.Run("WhereJsonContains with path", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, mysqlDriver, "json_data", nil, nil)
		q.WhereJsonContains("data->name", "json1")

		builder := NewBuilder(q)
		sql, args := builder.BuildSelect()

		// Should generate: JSON_CONTAINS(data, ?, '$.name')
		if !contains(sql, "JSON_CONTAINS(data, ?, '$.name')") {
			t.Errorf("Expected JSON_CONTAINS with quoted path, got: %s", sql)
		}

		// MySQL JSON_CONTAINS requires valid JSON, so strings are wrapped in quotes
		if len(args) != 1 || args[0] != "\"json1\"" {
			t.Errorf("Expected args [\"json1\"], got: %v", args)
		}
	})

	t.Run("WhereJsonContainsKey with nested path", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, mysqlDriver, "json_data", nil, nil)
		q.WhereJsonContainsKey("data->meta->active")

		builder := NewBuilder(q)
		sql, args := builder.BuildSelect()

		// Should generate: JSON_CONTAINS_PATH(data, 'one', '$.meta.active')
		if !contains(sql, "JSON_CONTAINS_PATH(data, 'one', '$.meta.active')") {
			t.Errorf("Expected JSON_CONTAINS_PATH with quoted nested path, got: %s", sql)
		}

		if len(args) != 0 {
			t.Errorf("Expected no args, got: %v", args)
		}
	})

	t.Run("WhereJsonLength with path", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, mysqlDriver, "json_data", nil, nil)
		q.WhereJsonLength("data->tags", "=", 2)

		builder := NewBuilder(q)
		sql, args := builder.BuildSelect()

		// Should generate: JSON_LENGTH(data, '$.tags') = ?
		if !contains(sql, "JSON_LENGTH(data, '$.tags') = ?") {
			t.Errorf("Expected JSON_LENGTH with quoted path, got: %s", sql)
		}

		if len(args) != 1 || args[0] != 2 {
			t.Errorf("Expected args [2], got: %v", args)
		}
	})

	t.Run("WhereJsonContains without path", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, mysqlDriver, "json_data", nil, nil)
		q.WhereJsonContains("data", "value")

		builder := NewBuilder(q)
		sql, args := builder.BuildSelect()

		// Should generate: JSON_CONTAINS(data, ?) - no path parameter
		if !contains(sql, "JSON_CONTAINS(data, ?)") {
			t.Errorf("Expected JSON_CONTAINS without path, got: %s", sql)
		}

		// MySQL JSON_CONTAINS requires valid JSON, so strings are wrapped in quotes
		if len(args) != 1 || args[0] != "\"value\"" {
			t.Errorf("Expected args [\"value\"], got: %v", args)
		}
	})
}

func TestJsonPathSyntax_SQLite(t *testing.T) {
	sqliteDriver := driver.NewSQLite()

	t.Run("WhereJsonContains with path", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, sqliteDriver, "json_data", nil, nil)
		q.WhereJsonContains("data->name", "json1")

		builder := NewBuilder(q)
		sql, args := builder.BuildSelect()

		// Should generate: json_extract(data, '$.name') = ?
		if !contains(sql, "json_extract(data, '$.name') = ?") {
			t.Errorf("Expected json_extract with path, got: %s", sql)
		}

		if len(args) != 1 || args[0] != "json1" {
			t.Errorf("Expected args [json1], got: %v", args)
		}
	})

	t.Run("WhereJsonContainsKey with nested path", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, sqliteDriver, "json_data", nil, nil)
		q.WhereJsonContainsKey("data->meta->active")

		builder := NewBuilder(q)
		sql, args := builder.BuildSelect()

		// Should generate: json_type(data, '$.meta.active') IS NOT NULL
		if !contains(sql, "json_type(data, '$.meta.active') IS NOT NULL") {
			t.Errorf("Expected json_type with nested path, got: %s", sql)
		}

		if len(args) != 0 {
			t.Errorf("Expected no args, got: %v", args)
		}
	})

	t.Run("WhereJsonLength with path", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, sqliteDriver, "json_data", nil, nil)
		q.WhereJsonLength("data->tags", "=", 2)

		builder := NewBuilder(q)
		sql, args := builder.BuildSelect()

		// Should generate: json_array_length(data, '$.tags') = ?
		if !contains(sql, "json_array_length(data, '$.tags') = ?") {
			t.Errorf("Expected json_array_length with path, got: %s", sql)
		}

		if len(args) != 1 || args[0] != 2 {
			t.Errorf("Expected args [2], got: %v", args)
		}
	})
}

func TestJsonUpdatePathSyntax(t *testing.T) {
	t.Run("MySQL JSON_SET with path", func(t *testing.T) {
		mysqlDriver := driver.NewMySQL()
		q := NewQuery(context.TODO(), nil, mysqlDriver, "json_data", nil, nil)
		q.Where("id = ?", 1)

		builder := NewBuilder(q)
		sql, args := builder.BuildUpdate("data->name", "updated_name")

		// Should generate: SET `data` = JSON_SET(`data`, '$.name', ?)
		if !contains(sql, "JSON_SET(`data`, '$.name', ?)") {
			t.Errorf("Expected JSON_SET with quoted path, got: %s", sql)
		}

		if len(args) != 2 {
			t.Errorf("Expected 2 args, got: %v", args)
		}
	})

	t.Run("SQLite json_set with path", func(t *testing.T) {
		sqliteDriver := driver.NewSQLite()
		q := NewQuery(context.TODO(), nil, sqliteDriver, "json_data", nil, nil)
		q.Where("id = ?", 1)

		builder := NewBuilder(q)
		sql, args := builder.BuildUpdate("data->name", "updated_name")

		// Should generate: SET "data" = json_set("data", '$.name', ?)
		if !contains(sql, "json_set") && !contains(sql, "'$.name'") {
			t.Errorf("Expected json_set with quoted path, got: %s", sql)
		}

		if len(args) != 2 {
			t.Errorf("Expected 2 args, got: %v", args)
		}
	})
}

func TestExclude(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Exclude("status", "inactive")

	if result == nil {
		t.Error("Expected non-nil Query from Exclude")
	}
}

func TestExcludeAsAliasForWhereNot(t *testing.T) {
	q1 := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q1.Table("users")
	q1.WhereNot("is_active = ?", false)

	q2 := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q2.Table("users")
	q2.Exclude("is_active = ?", false)

	wrapped1 := WrapQuery(q1)
	wrapped2 := WrapQuery(q2)

	sql1, args1 := wrapped1.BuildSelectSQL()
	sql2, args2 := wrapped2.BuildSelectSQL()

	if sql1 != sql2 {
		t.Errorf("Exclude and WhereNot should generate identical SQL. Got:\nWhereNot: %s\nExclude: %s", sql1, sql2)
	}
	if len(args1) != len(args2) {
		t.Errorf("Exclude and WhereNot should generate identical args. Got %v vs %v", args1, args2)
	}
}
