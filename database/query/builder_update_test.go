package query

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestBuildUpdate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
	}

	user := User{Name: "Alice", Email: "alice@example.com"}
	sql, args := b.BuildUpdate(user)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "UPDATE") {
		t.Error("Expected UPDATE in SQL")
	}
	if !strings.Contains(sql, "SET") {
		t.Error("Expected SET in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithMap(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	data := map[string]any{"name": "Bob", "age": 30}
	sql, args := b.BuildUpdate(data)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "UPDATE") {
		t.Error("Expected UPDATE in SQL")
	}
	if !strings.Contains(sql, "SET") {
		t.Error("Expected SET in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithColumnAndValue(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("name", "Charlie")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "UPDATE") {
		t.Error("Expected UPDATE in SQL")
	}
	if !strings.Contains(sql, "SET") {
		t.Error("Expected SET in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithExpression(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	// Test with expression (for Increment/Decrement)
	sql, args := b.BuildUpdate("count = count + 1", 1)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "count = count + 1") {
		t.Error("Expected expression in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithSoftDelete(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	// Test soft delete operation (updating deleted_at column)
	data := map[string]any{"deleted_at": "2024-01-01"}
	sql, _ := b.BuildUpdate(data)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Should not include soft-delete filter when updating deleted_at
	if !strings.Contains(sql, "UPDATE") {
		t.Error("Expected UPDATE in SQL")
	}
}

func TestBuildUpdateWithOmittedColumns(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.omitColumns = []string{"password"}
	b := NewBuilder(q)

	data := map[string]any{"name": "Eve", "password": "secret"}
	sql, args := b.BuildUpdate(data)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Password should be omitted
	if strings.Contains(sql, "password") {
		t.Error("Expected password to be omitted from SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithJSONPathMySQL(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "mysql"}
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("data->name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "JSON_SET") {
		t.Error("Expected JSON_SET in SQL for MySQL")
	}
	if !strings.Contains(sql, "$.name") {
		t.Error("Expected JSON path $.name in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithJSONPathSQLite(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlite"}
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("data->name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "json_set") {
		t.Error("Expected json_set in SQL for SQLite")
	}
	if !strings.Contains(sql, "$.name") {
		t.Error("Expected JSON path $.name in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithJSONPathNested(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "mysql"}
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("data->meta->active", true)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "JSON_SET") {
		t.Error("Expected JSON_SET in SQL")
	}
	if !strings.Contains(sql, "$.meta.active") {
		t.Error("Expected JSON path $.meta.active in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithJSONPathOtherDialect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "postgres"}
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("data->name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Should fallback to normal behavior for non-MySQL/SQLite
	if strings.Contains(sql, "JSON_SET") || strings.Contains(sql, "json_set") {
		t.Error("Expected fallback to normal behavior for non-MySQL/SQLite")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithLimitMySQL(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "mysql"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT clause in SQL for MySQL")
	}
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithLimitSQLite(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlite"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// SQLite should use rowid subquery workaround
	if !strings.Contains(sql, "rowid IN") {
		t.Error("Expected rowid subquery workaround for SQLite")
	}
	if !strings.Contains(sql, "SELECT rowid FROM") {
		t.Error("Expected SELECT rowid subquery for SQLite")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithLimitSQLiteWithOrder(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlite"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	q.OrderBy("id", "desc")
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "rowid IN") {
		t.Error("Expected rowid subquery workaround for SQLite")
	}
	if !strings.Contains(sql, "ORDER BY") {
		t.Error("Expected ORDER BY in subquery for SQLite")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithLimitPostgreSQL(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "postgres"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT clause in SQL for PostgreSQL")
	}
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithLimitNoWhere(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "mysql"}
	limit := 10
	q.limit = &limit
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT clause in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateSoftDeleteOperation(t *testing.T) {
	type SoftDeleteModel struct {
		ID        int
		Name      string
		DeletedAt *string
	}

	model := &SoftDeleteModel{}
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.model = model
	q.Where("id = ?", 1)
	b := NewBuilder(q)

	// Soft delete operation: updating deleted_at column
	data := map[string]any{"deleted_at": "2024-01-01"}
	sql, args := b.BuildUpdate(data)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Should use regular WHERE without soft-delete filter
	if strings.Contains(sql, "deleted_at IS NULL") || strings.Contains(sql, "deleted_at IS NOT NULL") {
		t.Error("Expected no soft-delete filter when updating deleted_at column")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateNonSoftDeleteOperation(t *testing.T) {
	type SoftDeleteModel struct {
		ID        int
		Name      string
		DeletedAt *time.Time
	}

	model := &SoftDeleteModel{}
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.model = model
	q.Where("id = ?", 1)
	b := NewBuilder(q)

	// Normal operation: updating other columns
	data := map[string]any{"name": "Alice"}
	sql, args := b.BuildUpdate(data)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Should include soft-delete filter for normal operations
	if !strings.Contains(sql, "deleted_at IS NULL") {
		t.Error("Expected soft-delete filter when updating non-deleted_at columns")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildUpdateWithRawExpression(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	builder := NewBuilder(q)

	// Test update with raw expression
	data := map[string]any{
		"name":       "John",
		"updated_at": RawExpr("NOW()"),
	}

	sql, args := builder.BuildUpdate(data)
	if sql == "" {
		t.Fatal("Expected SQL to be generated")
	}

	// Check that NOW() is in the SQL (not parameterized)
	if !strings.Contains(sql, "NOW()") {
		t.Errorf("Expected SQL to contain NOW(), got: %s", sql)
	}

	// Check that only non-raw values are in args
	if len(args) != 1 {
		t.Errorf("Expected 1 arg (name), got %d: %v", len(args), args)
	}
	if args[0] != "John" {
		t.Errorf("Expected arg[0] to be 'John', got %v", args[0])
	}
}

func TestBuildUpdateWithRawExpressionAndArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	builder := NewBuilder(q)

	// Test raw expression with arguments
	data := map[string]any{
		"name":  "John",
		"score": RawExpr("score + ?", 5),
	}

	sql, args := builder.BuildUpdate(data)
	if sql == "" {
		t.Fatal("Expected SQL to be generated")
	}

	// Check that raw expression is in the SQL
	if !strings.Contains(sql, "score + ?") {
		t.Errorf("Expected SQL to contain 'score + ?', got: %s", sql)
	}

	// Check that both the name and the raw expression arg are in args
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "John" {
		t.Errorf("Expected arg[0] to be 'John', got %v", args[0])
	}
	if args[1] != 5 {
		t.Errorf("Expected arg[1] to be 5, got %v", args[1])
	}
}

func TestBuildUpdateWithMixedRawAndRegular(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	builder := NewBuilder(q)

	// Test mixing raw expressions with regular values in UPDATE
	// Note: map iteration order is not guaranteed, so we just check that all args are present
	data := map[string]any{
		"name":  "John",
		"score": RawExpr("score + ?", 5),
		"age":   30,
	}

	sql, args := builder.BuildUpdate(data)
	if sql == "" {
		t.Fatal("Expected SQL to be generated")
	}

	// Check that raw expression is in the SQL
	if !strings.Contains(sql, "score + ?") {
		t.Errorf("Expected SQL to contain 'score + ?', got: %s", sql)
	}

	// Check that we have 3 args total
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(args), args)
	}

	// Check that all expected values are in args (order may vary)
	argsMap := make(map[any]bool)
	for _, arg := range args {
		argsMap[arg] = true
	}
	if !argsMap["John"] {
		t.Errorf("Expected 'John' in args, got %v", args)
	}
	if !argsMap[30] {
		t.Errorf("Expected 30 in args, got %v", args)
	}
	if !argsMap[5] {
		t.Errorf("Expected 5 in args, got %v", args)
	}
}

func TestBuildUpdateWithLimitSQLServer(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlserver"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	b := NewBuilder(q)

	sql, args := b.BuildUpdate("name", "Alice")

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// SQL Server uses TOP instead of LIMIT
	if !strings.Contains(sql, "UPDATE TOP (10)") {
		t.Error("Expected TOP clause in SQL for SQL Server")
	}
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}
