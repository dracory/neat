package query

import (
	"context"
	"testing"
)

func TestCloneQueryStateIsolation(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Where("status = ?", "active")
	q.Select("id", "name")
	q.OrderBy("name")
	limit := 10
	q.limit = &limit

	clone := q.Clone().(*Query)

	// Mutate the clone — original must be unaffected
	clone.Table("posts")
	clone.Where("id = ?", 99)

	if q.table != "users" {
		t.Errorf("Original table mutated by clone modification, got: %s", q.table)
	}
	if len(q.wheres) != 1 {
		t.Errorf("Original wheres mutated by clone modification, got %d wheres", len(q.wheres))
	}
	if clone.table != "posts" {
		t.Errorf("Clone table not updated, got: %s", clone.table)
	}
	if len(clone.wheres) != 2 {
		t.Errorf("Clone wheres not updated, got %d wheres", len(clone.wheres))
	}
}

func TestClonePreservesSelects(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Select("id", "name")

	clone := q.Clone().(*Query)

	if len(clone.selects) != len(q.selects) {
		t.Errorf("Clone selects count mismatch: original %d, clone %d", len(q.selects), len(clone.selects))
	}
}

func TestClonePreservesLimit(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Limit(5)

	clone := q.Clone().(*Query)

	if clone.limit == nil {
		t.Fatal("Clone limit should not be nil")
	}
	if *clone.limit != 5 {
		t.Errorf("Clone limit mismatch: expected 5, got %d", *clone.limit)
	}

	// Mutating clone's limit must not affect original
	newLimit := 99
	clone.limit = &newLimit
	if *q.limit != 5 {
		t.Errorf("Original limit mutated by clone modification")
	}
}

func TestClonePropagatesReplicas(t *testing.T) {
	// This test requires the testing infrastructure from query_accessors_test.go
	// The actual test is in query_accessors_test.go
	// This is a placeholder to ensure the file structure is correct
}

func TestClonePreservesAllFields(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Where("status = ?", "active")
	q.Select("id", "name")
	q.OrderBy("name")
	q.Limit(10)
	q.Offset(5)
	q.distinct = true
	q.lockForUpdate = true
	q.sharedLock = false
	q.withTrashed = true
	q.onlyTrashed = false
	q.withoutTrashed = false

	clone := q.Clone().(*Query)

	if clone.table != q.table {
		t.Errorf("Clone table mismatch")
	}
	if len(clone.wheres) != len(q.wheres) {
		t.Errorf("Clone wheres count mismatch")
	}
	if len(clone.selects) != len(q.selects) {
		t.Errorf("Clone selects count mismatch")
	}
	if len(clone.orders) != len(q.orders) {
		t.Errorf("Clone orders count mismatch")
	}
	if clone.limit == nil || *clone.limit != *q.limit {
		t.Errorf("Clone limit mismatch")
	}
	if clone.offset == nil || *clone.offset != *q.offset {
		t.Errorf("Clone offset mismatch")
	}
	if clone.distinct != q.distinct {
		t.Errorf("Clone distinct mismatch")
	}
	if clone.lockForUpdate != q.lockForUpdate {
		t.Errorf("Clone lockForUpdate mismatch")
	}
	if clone.sharedLock != q.sharedLock {
		t.Errorf("Clone sharedLock mismatch")
	}
	if clone.withTrashed != q.withTrashed {
		t.Errorf("Clone withTrashed mismatch")
	}
	if clone.onlyTrashed != q.onlyTrashed {
		t.Errorf("Clone onlyTrashed mismatch")
	}
	if clone.withoutTrashed != q.withoutTrashed {
		t.Errorf("Clone withoutTrashed mismatch")
	}
}
