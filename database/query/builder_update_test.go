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
