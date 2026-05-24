package query_test

import (
	"strings"
	"testing"
	"time"

	"github.com/dracory/neat/database/query"
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

func newSoftQuery(model any) *query.TestQuery {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))
	w.SetTable("soft_models")
	w.SetModel(model)
	return w
}

// TestBuildSelectInjectsSoftDeleteFilter verifies that `deleted_at IS NULL` is
// injected automatically for models with a *time.Time DeletedAt field.
func TestBuildSelectInjectsSoftDeleteFilter(t *testing.T) {
	w := newSoftQuery(&softModel{})
	sqlStr, _ := w.BuildSelectSQL()

	if !strings.Contains(sqlStr, "deleted_at IS NULL") {
		t.Errorf("expected 'deleted_at IS NULL' in SQL, got: %s", sqlStr)
	}
}

// TestBuildSelectWithTrashedSkipsFilter verifies WithTrashed() suppresses the filter.
func TestBuildSelectWithTrashedSkipsFilter(t *testing.T) {
	w := newSoftQuery(&softModel{})
	w.SetWithTrashed(true)
	sqlStr, _ := w.BuildSelectSQL()

	// Check if "deleted_at IS NULL" or "deleted_at IS NOT NULL" exists after "WHERE"
	if whereIdx := strings.Index(sqlStr, "WHERE"); whereIdx != -1 {
		whereClause := sqlStr[whereIdx:]
		if strings.Contains(whereClause, "deleted_at") {
			t.Errorf("expected no 'deleted_at' filter in WHERE clause with WithTrashed, got: %s", sqlStr)
		}
	}
}

// TestBuildSelectOnlyTrashedFilter verifies OnlyTrashed() uses IS NOT NULL.
func TestBuildSelectOnlyTrashedFilter(t *testing.T) {
	w := newSoftQuery(&softModel{})
	w.SetOnlyTrashed(true)
	sqlStr, _ := w.BuildSelectSQL()

	if !strings.Contains(sqlStr, "deleted_at IS NOT NULL") {
		t.Errorf("expected 'deleted_at IS NOT NULL' in SQL, got: %s", sqlStr)
	}
}

// TestBuildSelectNoFilterForNonSoftDeleteModel verifies plain models get no filter.
func TestBuildSelectNoFilterForNonSoftDeleteModel(t *testing.T) {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))
	w.SetTable("hard_models")
	w.SetModel(&hardModel{})
	sqlStr, _ := w.BuildSelectSQL()

	if strings.Contains(sqlStr, "deleted_at") {
		t.Errorf("expected no 'deleted_at' clause for non-soft-delete model, got: %s", sqlStr)
	}
}

// TestBuildSelectNoFilterWhenModelNil verifies nil model gets no soft-delete filter.
func TestBuildSelectNoFilterWhenModelNil(t *testing.T) {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))
	w.SetTable("users")
	sqlStr, _ := w.BuildSelectSQL()

	if strings.Contains(sqlStr, "deleted_at") {
		t.Errorf("expected no 'deleted_at' clause when model is nil, got: %s", sqlStr)
	}
}

type pointerModel struct {
	ID   int
	Name *string
}

func TestSelectPointerField(t *testing.T) {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))
	w.SetModel(&pointerModel{})
	sqlStr, _ := w.BuildSelectSQL()

	if !strings.Contains(sqlStr, "name") {
		t.Errorf("expected 'name' in SELECT, got: %s", sqlStr)
	}
}
