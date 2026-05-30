package query

import (
	"context"
	"strings"
	"testing"
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
