package query

import (
	"context"
	"strings"
	"testing"
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

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	_ = args // Just ensure it doesn't panic
}
