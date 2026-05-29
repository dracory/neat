package query

import (
	"strings"
	"testing"
)

func TestBuildWheres(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.wheres = []whereClause{
		{_type: "", query: "name = ?", args: []any{"Alice"}},
		{_type: "AND", query: "age > ?", args: []any{18}},
	}
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if where == "" {
		t.Error("Expected non-empty WHERE clause")
	}
	if !strings.Contains(where, "name = ?") {
		t.Error("Expected 'name = ?' in WHERE clause")
	}
	if !strings.Contains(where, "age > ?") {
		t.Error("Expected 'age > ?' in WHERE clause")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildWheresWithIN(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.wheres = []whereClause{
		{_type: "", query: "id IN (?)", args: []any{[]any{1, 2, 3}}},
	}
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if where == "" {
		t.Error("Expected non-empty WHERE clause")
	}
	// IN (?) should be expanded to IN (?, ?, ?)
	if !strings.Contains(where, "IN (?, ?, ?)") {
		t.Error("Expected IN clause to be expanded")
	}
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

func TestBuildWheresEmpty(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if where != "" {
		t.Error("Expected empty WHERE clause")
	}
	if len(args) != 0 {
		t.Error("Expected empty args")
	}
}

func TestLaravelStyleWhere(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.Where("name", "Alice")
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 || args[0] != "Alice" {
		t.Errorf("Expected args [Alice], got %v", args)
	}
}

func TestLaravelStyleOrWhere(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.OrWhere("name", "Bob")
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 || args[0] != "Bob" {
		t.Errorf("Expected args [Bob], got %v", args)
	}
}

func TestExplicitOperatorWhere(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.Where("age > ?", 18)
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "age > ?") {
		t.Errorf("Expected 'age > ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 || args[0] != 18 {
		t.Errorf("Expected args [18], got %v", args)
	}
}

func TestMixedWhereStyles(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.Where("name", "Alice")
	q.Where("age > ?", 18)
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if !strings.Contains(where, "age > ?") {
		t.Errorf("Expected 'age > ?' in WHERE clause, got %s", where)
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}
