package query

import (
	"context"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestBuildInsert(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
	}

	user := User{Name: "Alice", Email: "alice@example.com"}
	sql, args := b.BuildInsert(user)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	_ = args // Just ensure it doesn't panic
}

func TestBuildInsertWithMap(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	data := map[string]any{"name": "Bob", "age": 30}
	sql, args := b.BuildInsert(data)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildInsertBulk(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
	}

	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	sql, args := b.BuildInsert(users)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
	// Bulk insert should have multiple row placeholders
	if !strings.Contains(sql, "), (") {
		t.Error("Expected multiple row placeholders for bulk insert")
	}
}

func TestBuildInsertEmptySlice(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	var users []struct {
		Name string
	}
	sql, args := b.BuildInsert(users)

	// Empty slice should return empty SQL (cannot insert nothing)
	if sql != "" {
		t.Errorf("Expected empty SQL for empty slice, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args for empty slice, got: %v", args)
	}
}

func TestBuildInsertWithPostgreSQLPlaceholders(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "postgres"}
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
	}

	user := User{Name: "Alice", Email: "alice@example.com"}
	sql, args := b.BuildInsert(user)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// PostgreSQL should use $1, $2 style placeholders
	if !strings.Contains(sql, "$1") {
		t.Error("Expected PostgreSQL-style placeholders ($1)")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildInsertBulkWithPostgreSQLPlaceholders(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "postgres"}
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
	}

	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	sql, args := b.BuildInsert(users)

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Should have incrementing placeholders: $1, $2, $3, $4
	if !strings.Contains(sql, "$1") || !strings.Contains(sql, "$4") {
		t.Error("Expected incrementing PostgreSQL-style placeholders")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildInsertWithRawExpression(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	builder := NewBuilder(q)

	// Test single insert with raw expression
	data := map[string]any{
		"name":       "John",
		"created_at": RawExpr("NOW()"),
	}

	sql, args := builder.BuildInsert(data)
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

func TestBuildInsertWithRawExpressionAndArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	builder := NewBuilder(q)

	// Test raw expression with arguments
	data := map[string]any{
		"name":  "John",
		"score": RawExpr("score + ?", 10),
	}

	sql, args := builder.BuildInsert(data)
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
	if args[1] != 10 {
		t.Errorf("Expected arg[1] to be 10, got %v", args[1])
	}
}

func TestBuildInsertWithMixedRawAndRegular(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	builder := NewBuilder(q)

	// Test mixing raw expressions with regular values
	// This tests the critical placeholder numbering fix
	// Note: map keys are sorted alphabetically, so order is: age, name, score
	data := map[string]any{
		"name":  "John",
		"score": RawExpr("score + ?", 10),
		"age":   25,
	}

	sql, args := builder.BuildInsert(data)
	if sql == "" {
		t.Fatal("Expected SQL to be generated")
	}

	// Check that raw expression is in the SQL
	if !strings.Contains(sql, "score + ?") {
		t.Errorf("Expected SQL to contain 'score + ?', got: %s", sql)
	}

	// Check that args are in correct order (alphabetical by key): age, name, raw expr arg
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != 25 {
		t.Errorf("Expected arg[0] to be 25 (age), got %v", args[0])
	}
	if args[1] != "John" {
		t.Errorf("Expected arg[1] to be 'John' (name), got %v", args[1])
	}
	if args[2] != 10 {
		t.Errorf("Expected arg[2] to be 10 (raw expr arg), got %v", args[2])
	}
}

func TestBuildBulkInsertWithRawExpression(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	builder := NewBuilder(q)

	// Test bulk insert with raw expressions
	data := []map[string]any{
		{"name": "John", "created_at": RawExpr("NOW()")},
		{"name": "Jane", "created_at": RawExpr("NOW()")},
	}

	sql, args := builder.BuildInsert(data)
	if sql == "" {
		t.Fatal("Expected SQL to be generated")
	}

	// Check that NOW() is in the SQL (not parameterized)
	if !strings.Contains(sql, "NOW()") {
		t.Errorf("Expected SQL to contain NOW(), got: %s", sql)
	}

	// Check that only non-raw values are in args (2 names)
	if len(args) != 2 {
		t.Errorf("Expected 2 args (names), got %d: %v", len(args), args)
	}
	if args[0] != "John" {
		t.Errorf("Expected arg[0] to be 'John', got %v", args[0])
	}
	if args[1] != "Jane" {
		t.Errorf("Expected arg[1] to be 'Jane', got %v", args[1])
	}
}

func TestBuildBulkInsertWithRawExpressionAndArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	builder := NewBuilder(q)

	// Test bulk insert with raw expressions that have arguments
	data := []map[string]any{
		{"name": "John", "score": RawExpr("score + ?", 10)},
		{"name": "Jane", "score": RawExpr("score + ?", 5)},
	}

	sql, args := builder.BuildInsert(data)
	if sql == "" {
		t.Fatal("Expected SQL to be generated")
	}

	// Check that raw expression is in the SQL
	if !strings.Contains(sql, "score + ?") {
		t.Errorf("Expected SQL to contain 'score + ?', got: %s", sql)
	}

	// Check that args are in correct order: John, 10, Jane, 5
	if len(args) != 4 {
		t.Errorf("Expected 4 args, got %d: %v", len(args), args)
	}
	if args[0] != "John" {
		t.Errorf("Expected arg[0] to be 'John', got %v", args[0])
	}
	if args[1] != 10 {
		t.Errorf("Expected arg[1] to be 10, got %v", args[1])
	}
	if args[2] != "Jane" {
		t.Errorf("Expected arg[2] to be 'Jane', got %v", args[2])
	}
	if args[3] != 5 {
		t.Errorf("Expected arg[3] to be 5, got %v", args[3])
	}
}

func TestBuildInsertWithSQLServerOutput(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlserver"}
	q.table = "users"
	builder := NewBuilder(q)

	// Test insert with OUTPUT clause for SQL Server
	data := map[string]any{
		"id":    1,
		"name":  "John",
		"email": "john@example.com",
	}

	sql, args := builder.BuildInsert(data)
	if sql == "" {
		t.Fatal("Expected SQL to be generated")
	}

	// Check that OUTPUT clause is present for SQL Server (column is quoted)
	if !strings.Contains(sql, "OUTPUT INSERTED") {
		t.Errorf("Expected OUTPUT INSERTED in SQL for SQL Server, got: %s", sql)
	}

	// Check that args are present
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(args), args)
	}
}

func TestBuildInsertWithSQLServerOutputUUID(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlserver"}
	q.table = "users"
	builder := NewBuilder(q)

	// Test insert with OUTPUT clause for UUID column
	data := map[string]any{
		"uuid":  "123e4567-e89b-12d3-a456-426614174000",
		"name":  "John",
		"email": "john@example.com",
	}

	sql, args := builder.BuildInsert(data)
	if sql == "" {
		t.Fatal("Expected SQL to be generated")
	}

	// Check that OUTPUT clause is present (column is quoted)
	if !strings.Contains(sql, "OUTPUT INSERTED") {
		t.Errorf("Expected OUTPUT INSERTED in SQL for SQL Server with UUID, got: %s", sql)
	}

	// Check that args are present
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(args), args)
	}
}
