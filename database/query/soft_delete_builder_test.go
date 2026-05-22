package query

import (
	"context"
	"strings"
	"testing"
	"time"
)

// softModel has a *time.Time DeletedAt — detected as soft-deletable.
type softModel struct {
	ID        int
	Name      string
	DeletedAt *time.Time
}

// hardModel has no DeletedAt — not soft-deletable.
type hardModel struct {
	ID   int
	Name string
}

func newSoftQuery(model any) *Query {
	q := NewQuery(context.Background(), nil, nil, "", testDBConfig(), nil)
	q.table = "soft_models"
	q.model = model
	return q
}

// TestBuildSelectInjectsSoftDeleteFilter verifies that `deleted_at IS NULL` is
// injected automatically for models with a *time.Time DeletedAt field.
func TestBuildSelectInjectsSoftDeleteFilter(t *testing.T) {
	q := newSoftQuery(&softModel{})
	sql, _ := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, "deleted_at IS NULL") {
		t.Errorf("expected 'deleted_at IS NULL' in SQL, got: %s", sql)
	}
}

// TestBuildSelectWithTrashedSkipsFilter verifies WithTrashed() suppresses the filter.
func TestBuildSelectWithTrashedSkipsFilter(t *testing.T) {
	q := newSoftQuery(&softModel{})
	q.withTrashed = true
	sql, _ := NewBuilder(q).BuildSelect()

	if strings.Contains(sql, "deleted_at") {
		t.Errorf("expected no 'deleted_at' clause with WithTrashed, got: %s", sql)
	}
}

// TestBuildSelectOnlyTrashedFilter verifies OnlyTrashed() uses IS NOT NULL.
func TestBuildSelectOnlyTrashedFilter(t *testing.T) {
	q := newSoftQuery(&softModel{})
	q.onlyTrashed = true
	sql, _ := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, "deleted_at IS NOT NULL") {
		t.Errorf("expected 'deleted_at IS NOT NULL' in SQL, got: %s", sql)
	}
}

// TestBuildSelectNoFilterForNonSoftDeleteModel verifies plain models get no filter.
func TestBuildSelectNoFilterForNonSoftDeleteModel(t *testing.T) {
	q := NewQuery(context.Background(), nil, nil, "", testDBConfig(), nil)
	q.table = "hard_models"
	q.model = &hardModel{}
	sql, _ := NewBuilder(q).BuildSelect()

	if strings.Contains(sql, "deleted_at") {
		t.Errorf("expected no 'deleted_at' clause for non-soft-delete model, got: %s", sql)
	}
}

// TestBuildSelectNoFilterWhenModelNil verifies nil model gets no soft-delete filter.
func TestBuildSelectNoFilterWhenModelNil(t *testing.T) {
	q := NewQuery(context.Background(), nil, nil, "", testDBConfig(), nil)
	q.table = "users"
	sql, _ := NewBuilder(q).BuildSelect()

	if strings.Contains(sql, "deleted_at") {
		t.Errorf("expected no 'deleted_at' clause when model is nil, got: %s", sql)
	}
}
