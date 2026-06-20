package query

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestConvertTimeArgs(t *testing.T) {
	ts := time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)
	args := []any{"string", 123, ts, &ts, nil, (*time.Time)(nil)}

	converted := convertTimeArgs(args)

	if len(converted) != 6 {
		t.Errorf("Expected 6 args, got %d", len(converted))
	}
	if converted[0] != "string" {
		t.Errorf("Expected string unchanged, got %v", converted[0])
	}
	if converted[1] != 123 {
		t.Errorf("Expected int unchanged, got %v", converted[1])
	}
	if s, ok := converted[2].(string); !ok || s != "2026-06-20 12:34:56" {
		t.Errorf("Expected time.Time converted to '2026-06-20 12:34:56', got %v (%T)", converted[2], converted[2])
	}
	if s, ok := converted[3].(string); !ok || s != "2026-06-20 12:34:56" {
		t.Errorf("Expected *time.Time converted to '2026-06-20 12:34:56', got %v (%T)", converted[3], converted[3])
	}
	if converted[4] != nil {
		t.Errorf("Expected nil unchanged, got %v", converted[4])
	}
	if _, ok := converted[5].(*time.Time); !ok {
		t.Errorf("Expected nil *time.Time to remain as (*time.Time)(nil), got %v (%T)", converted[5], converted[5])
	}
}

func TestBuildWheresWithTimeArg(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.wheres = []whereClause{
		{_type: "", query: "created_at = ?", args: []any{time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)}},
	}
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "created_at = ?") {
		t.Errorf("Expected 'created_at = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
	if s, ok := args[0].(string); !ok || s != "2026-06-20 12:34:56" {
		t.Errorf("Expected time arg converted to '2026-06-20 12:34:56', got %v (%T)", args[0], args[0])
	}
}

func TestBuildWheresWithPtrTimeArg(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	ts := time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)
	q.wheres = []whereClause{
		{_type: "", query: "deleted_at = ?", args: []any{&ts}},
	}
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "deleted_at = ?") {
		t.Errorf("Expected 'deleted_at = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
	if s, ok := args[0].(string); !ok || s != "2026-06-20 12:34:56" {
		t.Errorf("Expected *time.Time arg converted to '2026-06-20 12:34:56', got %v (%T)", args[0], args[0])
	}
}
