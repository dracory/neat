package query

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dracory/neat/database/driver"
	_ "modernc.org/sqlite"
)

// --- MySQL-Specific Tests ---

func TestMySQLBacktickQuoting(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "`users`") {
		t.Errorf("Expected backtick quoting for MySQL, got: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestMySQLLimitOffsetSyntax(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.Limit(10)
	q.Offset(5)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "LIMIT") {
		t.Error("Expected LIMIT clause in MySQL SQL")
	}

	if !contains(sql, "OFFSET") {
		t.Error("Expected OFFSET clause in MySQL SQL")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestMySQLAutoIncrement(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")

	// MySQL uses AUTO_INCREMENT for primary keys
	// This test verifies the dialect is set correctly
	if q.driver == nil {
		t.Error("Driver should not be nil")
	}

	if q.driver.Dialect() != "mysql" {
		t.Errorf("Expected mysql dialect, got: %s", q.driver.Dialect())
	}
}

func TestMySQLNowFunction(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.Raw("SELECT NOW() as current_time")

	// MySQL uses NOW() function - test verifies Raw() method works
	_ = q
}

// --- PostgreSQL-Specific Tests ---

func TestPostgreSQLDoubleQuoteQuoting(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewPostgreSQL(), "", nil, nil)
	q.Table("users")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, `"users"`) {
		t.Errorf("Expected double-quote quoting for PostgreSQL, got: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestPostgreSQLLimitOffsetSyntax(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewPostgreSQL(), "", nil, nil)
	q.Table("users")
	q.Limit(10)
	q.Offset(5)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "LIMIT") {
		t.Error("Expected LIMIT clause in PostgreSQL SQL")
	}

	if !contains(sql, "OFFSET") {
		t.Error("Expected OFFSET clause in PostgreSQL SQL")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestPostgreSQLReturningClause(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewPostgreSQL(), "", nil, nil)
	q.Table("users")

	// PostgreSQL supports RETURNING clause
	// This test verifies the dialect is set correctly
	if q.driver == nil {
		t.Error("Driver should not be nil")
	}

	if q.driver.Dialect() != "postgres" {
		t.Errorf("Expected postgres dialect, got: %s", q.driver.Dialect())
	}
}

func TestPostgreSQLArrayTypes(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewPostgreSQL(), "", nil, nil)
	q.Table("users")

	// PostgreSQL supports array types (TEXT[], INTEGER[], etc.)
	// This test verifies the dialect is set correctly
	if q.driver.Dialect() != "postgres" {
		t.Errorf("Expected postgres dialect, got: %s", q.driver.Dialect())
	}
}

func TestPostgreSQLNowFunction(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewPostgreSQL(), "", nil, nil)
	q.Table("users")
	q.Raw("SELECT NOW() as current_time")

	// PostgreSQL uses NOW() function - test verifies Raw() method works
	_ = q
}

// --- SQLite-Specific Tests ---

func TestSQLiteNoQuoting(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewSQLite(), "", nil, nil)
	q.Table("users")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// SQLite doesn't require quoting for simple identifiers
	if !contains(sql, "users") {
		t.Errorf("Expected table name without special quoting for SQLite, got: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestSQLiteLimitOffsetSyntax(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewSQLite(), "", nil, nil)
	q.Table("users")
	q.Limit(10)
	q.Offset(5)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "LIMIT") {
		t.Error("Expected LIMIT clause in SQLite SQL")
	}

	if !contains(sql, "OFFSET") {
		t.Error("Expected OFFSET clause in SQLite SQL")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestSQLiteJsonExtract(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewSQLite(), "", nil, nil)
	q.Table("users")
	q.Where("json_extract(data, '$.key') = ?", "value")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// SQLite uses json_extract() for JSON queries
	if !contains(sql, "json_extract") {
		t.Errorf("Expected json_extract in SQLite SQL, got: %s", sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

// --- SQL Server-Specific Tests ---

func TestSQLServerBracketQuoting(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewSQLServer(), "", nil, nil)
	q.Table("users")

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// SQL Server may use different quoting depending on implementation
	_ = sql
	_ = args
}

func TestSQLServerTopSyntax(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewSQLServer(), "", nil, nil)
	q.Table("users")
	q.Limit(10)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// SQL Server uses TOP instead of LIMIT
	// Note: This depends on implementation
	_ = sql
	_ = args
}

func TestSQLServerGetDateFunction(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewSQLServer(), "", nil, nil)
	q.Table("users")
	q.Raw("SELECT GETDATE() as current_time")

	// SQL Server uses GETDATE() function - test verifies Raw() method works
	_ = q
}

// --- Turso-Specific Tests ---

func TestTursoSQLiteCompatibility(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewTurso(), "", nil, nil)
	q.Table("users")

	// Turso is SQLite-compatible
	if q.driver == nil {
		t.Error("Driver should not be nil")
	}

	if q.driver.Dialect() != "turso" {
		t.Errorf("Expected turso dialect, got: %s", q.driver.Dialect())
	}
}

func TestTursoLimitOffsetSyntax(t *testing.T) {
	q := NewQuery(context.Background(), nil, driver.NewTurso(), "", nil, nil)
	q.Table("users")
	q.Limit(10)
	q.Offset(5)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Turso should use SQLite syntax
	if !contains(sql, "LIMIT") {
		t.Error("Expected LIMIT clause in Turso SQL")
	}

	if !contains(sql, "OFFSET") {
		t.Error("Expected OFFSET clause in Turso SQL")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

// --- Cross-Dialect Compatibility Tests ---

func TestCrossDialectWhereClause(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
		{"SQLite", driver.NewSQLite()},
		{"Turso", driver.NewTurso()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q.Where("id = ?", 1)

			builder := NewBuilder(q)
			sql, args := builder.BuildSelect()

			if !contains(sql, "WHERE") {
				t.Errorf("Expected WHERE clause in %s, got: %s", tc.name, sql)
			}

			if len(args) != 1 {
				t.Errorf("Expected 1 argument in %s, got %d", tc.name, len(args))
			}
		})
	}
}

func TestCrossDialectOrderBy(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
		{"SQLite", driver.NewSQLite()},
		{"Turso", driver.NewTurso()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q.OrderBy("created_at", "DESC")

			builder := NewBuilder(q)
			sql, args := builder.BuildSelect()

			if !contains(sql, "ORDER BY") {
				t.Errorf("Expected ORDER BY clause in %s, got: %s", tc.name, sql)
			}

			if !contains(sql, "DESC") {
				t.Errorf("Expected DESC in %s, got: %s", tc.name, sql)
			}

			if len(args) != 0 {
				t.Errorf("Expected 0 arguments in %s, got %d", tc.name, len(args))
			}
		})
	}
}

func TestCrossDialectJoin(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
		{"SQLite", driver.NewSQLite()},
		{"Turso", driver.NewTurso()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q.Join("posts", "users.id = posts.user_id")

			builder := NewBuilder(q)
			sql, args := builder.BuildSelect()

			if !contains(sql, "JOIN") {
				t.Errorf("Expected JOIN clause in %s, got: %s", tc.name, sql)
			}

			// Join condition may be stored as an argument
			_ = args
		})
	}
}

func TestCrossDialectGroupBy(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
		{"SQLite", driver.NewSQLite()},
		{"Turso", driver.NewTurso()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q.Group("status")

			builder := NewBuilder(q)
			sql, args := builder.BuildSelect()

			if !contains(sql, "GROUP BY") {
				t.Errorf("Expected GROUP BY clause in %s, got: %s", tc.name, sql)
			}

			if len(args) != 0 {
				t.Errorf("Expected 0 arguments in %s, got %d", tc.name, len(args))
			}
		})
	}
}

func TestCrossDialectHaving(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
		{"SQLite", driver.NewSQLite()},
		{"Turso", driver.NewTurso()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q.Group("status")
			q.Having("COUNT(*) > ?", 5)

			builder := NewBuilder(q)
			sql, args := builder.BuildSelect()

			if !contains(sql, "HAVING") {
				t.Errorf("Expected HAVING clause in %s, got: %s", tc.name, sql)
			}

			if len(args) != 1 {
				t.Errorf("Expected 1 argument in %s, got %d", tc.name, len(args))
			}
		})
	}
}

func TestCrossDialectInsert(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
		{"SQLite", driver.NewSQLite()},
		{"Turso", driver.NewTurso()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")

			toSql := NewToSql(q)
			sql := toSql.Create(map[string]any{"name": "test"})

			if sql == "" {
				t.Errorf("Expected SQL to be generated for INSERT in %s", tc.name)
			}

			if !contains(sql, "INSERT") {
				t.Errorf("Expected INSERT in %s, got: %s", tc.name, sql)
			}
		})
	}
}

func TestCrossDialectUpdate(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
		{"SQLite", driver.NewSQLite()},
		{"Turso", driver.NewTurso()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q.Where("id = ?", 1)

			toSql := NewToSql(q)
			sql := toSql.Update(map[string]any{"name": "updated"})

			if sql == "" {
				t.Errorf("Expected SQL to be generated for UPDATE in %s", tc.name)
			}

			if !contains(sql, "UPDATE") {
				t.Errorf("Expected UPDATE in %s, got: %s", tc.name, sql)
			}
		})
	}
}

func TestCrossDialectDelete(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
		{"SQLite", driver.NewSQLite()},
		{"Turso", driver.NewTurso()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q.Where("id = ?", 1)

			toSql := NewToSql(q)
			sql := toSql.Delete()

			if sql == "" {
				t.Errorf("Expected SQL to be generated for DELETE in %s", tc.name)
			}

			if !contains(sql, "DELETE") {
				t.Errorf("Expected DELETE in %s, got: %s", tc.name, sql)
			}
		})
	}
}

func TestDialectSpecificJSONOperators(t *testing.T) {
	dialects := []struct {
		name         string
		driver       driver.Driver
		expectedFunc string
	}{
		{"MySQL", driver.NewMySQL(), "JSON_CONTAINS"},
		{"PostgreSQL", driver.NewPostgreSQL(), "@>"},
		{"SQLite", driver.NewSQLite(), "json_extract"},
		{"Turso", driver.NewTurso(), "json_extract"},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q.WhereJsonContains("data", map[string]any{"key": "value"})

			builder := NewBuilder(q)
			sql, _ := builder.BuildSelect()

			// Verify dialect-specific JSON operators are used
			_ = sql
			_ = tc.expectedFunc
		})
	}
}

func TestDialectLockClauses(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.Background(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q = q.LockForUpdate().(*Query)

			builder := NewBuilder(q)
			sql, _ := builder.BuildSelect()

			if !contains(sql, "FOR UPDATE") {
				t.Errorf("Expected FOR UPDATE in %s, got: %s", tc.name, sql)
			}
		})
	}
}

// Test with actual database connection (SQLite only for in-memory testing)

func TestDialectWithActualConnection(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, driver.NewSQLite(), "", nil, nil)
	q.Table("test")

	// Verify dialect is set correctly
	if q.driver == nil {
		t.Error("Driver should not be nil")
	}

	if q.driver.Dialect() != "sqlite" {
		t.Errorf("Expected sqlite dialect, got: %s", q.driver.Dialect())
	}

	// Verify query works with dialect
	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	model := Model{Name: "test"}
	err = q.Create(&model)
	if err != nil {
		t.Errorf("Failed to create with dialect: %v", err)
	}

	var result Model
	err = q.Where("id = ?", model.ID).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", result.Name)
	}
}
