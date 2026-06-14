package query_test

import (
	"strings"
	"testing"
	"time"

	"github.com/dracory/neat/database/query"
	"github.com/dracory/neat/database/soft_delete"
)

// altSoftModel embeds SoftDeletedAt, using "soft_deleted_at" as the column name.
type altSoftModel struct {
	soft_delete.SoftDeletedAt
	ID   int
	Name string
}

// customColumnModel overrides the soft delete column to "removed_at" via a method.
type customColumnModel struct {
	ID        int
	Name      string
	RemovedAt *time.Time
}

func (m *customColumnModel) DeletedAtColumn() string { return "removed_at" }

func newAltSoftQuery(model any) *query.TestQuery {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))
	w.SetTable("alt_soft_models")
	w.SetModel(model)
	return w
}

// ── SQL generation tests ──────────────────────────────────────────────────────

// TestSoftDeletedAtInjectsSoftDeleteFilter verifies that "soft_deleted_at IS NULL"
// is injected automatically for models embedding SoftDeletedAt.
func TestSoftDeletedAtInjectsSoftDeleteFilter(t *testing.T) {
	w := newAltSoftQuery(&altSoftModel{})
	sqlStr, _ := w.BuildSelectSQL()

	if !strings.Contains(sqlStr, "soft_deleted_at IS NULL") {
		t.Errorf("expected 'soft_deleted_at IS NULL' in SQL, got: %s", sqlStr)
	}
}

// TestSoftDeletedAtWithTrashedSkipsFilter verifies WithTrashed() suppresses the filter.
func TestSoftDeletedAtWithTrashedSkipsFilter(t *testing.T) {
	w := newAltSoftQuery(&altSoftModel{})
	w.SetWithTrashed(true)
	sqlStr, _ := w.BuildSelectSQL()

	if whereIdx := strings.Index(sqlStr, "WHERE"); whereIdx != -1 {
		whereClause := sqlStr[whereIdx:]
		if strings.Contains(whereClause, "soft_deleted_at") {
			t.Errorf("expected no 'soft_deleted_at' filter with WithTrashed, got: %s", sqlStr)
		}
	}
}

// TestSoftDeletedAtOnlyTrashedFilter verifies OnlyTrashed() uses IS NOT NULL.
func TestSoftDeletedAtOnlyTrashedFilter(t *testing.T) {
	w := newAltSoftQuery(&altSoftModel{})
	w.SetOnlyTrashed(true)
	sqlStr, _ := w.BuildSelectSQL()

	if !strings.Contains(sqlStr, "soft_deleted_at IS NOT NULL") {
		t.Errorf("expected 'soft_deleted_at IS NOT NULL' in SQL, got: %s", sqlStr)
	}
}

// TestCustomColumnModelInjectsFilter verifies a model with a custom DeletedAtColumn()
// gets the correct column name injected into the WHERE clause.
func TestCustomColumnModelInjectsFilter(t *testing.T) {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))
	w.SetTable("custom_models")
	w.SetModel(&customColumnModel{})
	sqlStr, _ := w.BuildSelectSQL()

	if !strings.Contains(sqlStr, "removed_at IS NULL") {
		t.Errorf("expected 'removed_at IS NULL' in SQL, got: %s", sqlStr)
	}
	if strings.Contains(sqlStr, "deleted_at") {
		t.Errorf("expected no 'deleted_at' reference in SQL, got: %s", sqlStr)
	}
}

// ── Execution tests (real SQLite) ─────────────────────────────────────────────

// TestSoftDeletedAtExecution tests full soft delete lifecycle with SoftDeletedAt.
func TestSoftDeletedAtExecution(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE alt_soft_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, soft_deleted_at DATETIME)")
	w.SetTable("alt_soft_models")
	w.SetModel(&altSoftModel{})

	// Create a record
	if err := w.Q.Create(map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Fetch before deletion
	var before altSoftModel
	if err := w.Q.Where("name = ?", "alice").First(&before); err != nil {
		t.Fatalf("First before delete: %v", err)
	}

	// Soft delete
	res, err := w.Q.Where("name = ?", "alice").Delete(&altSoftModel{})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", res.RowsAffected)
	}

	// Record should not be visible without WithTrashed
	var notFound altSoftModel
	if err := w.Q.Where("id = ?", before.ID).First(&notFound); err == nil {
		t.Error("expected record to be hidden after soft delete")
	}

	// Record should be visible with WithTrashed
	var found altSoftModel
	if err := w.Q.WithTrashed().Where("id = ?", before.ID).First(&found); err != nil {
		t.Fatalf("WithTrashed First: %v", err)
	}
	if found.ID != before.ID {
		t.Errorf("expected ID %d, got %d", before.ID, found.ID)
	}
	if !found.IsDeleted() {
		t.Error("SoftDeletedAt should be non-nil after soft delete")
	}
}

// TestSoftDeletedAtRestoreExecution tests restoring a soft-deleted record with SoftDeletedAt.
func TestSoftDeletedAtRestoreExecution(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE alt_soft_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, soft_deleted_at DATETIME)")
	w.SetTable("alt_soft_models")
	w.SetModel(&altSoftModel{})

	if err := w.Q.Create(map[string]any{"name": "bob"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Soft delete
	if _, err := w.Q.Where("name = ?", "bob").Delete(&altSoftModel{}); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Restore
	res, err := w.Q.WithTrashed().Where("name = ?", "bob").Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected on restore, got %d", res.RowsAffected)
	}

	// Record should now be visible without WithTrashed
	var restored altSoftModel
	if err := w.Q.Where("name = ?", "bob").First(&restored); err != nil {
		t.Fatalf("First after restore: %v", err)
	}
	if restored.IsDeleted() {
		t.Error("SoftDeletedAt should be nil after restore")
	}
}

// TestCustomColumnExecution tests full soft delete lifecycle with a custom column name.
func TestCustomColumnExecution(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE custom_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, removed_at DATETIME)")
	w.SetTable("custom_models")
	w.SetModel(&customColumnModel{})

	if err := w.Q.Create(map[string]any{"name": "carol"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var created customColumnModel
	if err := w.Q.Where("name = ?", "carol").First(&created); err != nil {
		t.Fatalf("First: %v", err)
	}

	// Soft delete sets "removed_at"
	res, err := w.Q.Where("name = ?", "carol").Delete(&customColumnModel{})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", res.RowsAffected)
	}

	// Hidden without WithTrashed
	var notFound customColumnModel
	if err := w.Q.Where("id = ?", created.ID).First(&notFound); err == nil {
		t.Error("expected record to be hidden after soft delete")
	}

	// Visible with OnlyTrashed
	var trashed customColumnModel
	if err := w.Q.OnlyTrashed().Where("id = ?", created.ID).First(&trashed); err != nil {
		t.Fatalf("OnlyTrashed First: %v", err)
	}
	if trashed.RemovedAt == nil {
		t.Error("RemovedAt should be non-nil after soft delete")
	}
}

// ── Unit tests for SoftDeletedAt struct methods ──────────────────────────────

func TestSoftDeletedAtMethods(t *testing.T) {
	sd := &soft_delete.SoftDeletedAt{}

	if sd.DeletedAtColumn() != "soft_deleted_at" {
		t.Errorf("expected 'soft_deleted_at', got %q", sd.DeletedAtColumn())
	}
	if sd.IsDeleted() {
		t.Error("expected IsDeleted() to be false initially")
	}

	sd.Delete()
	if !sd.IsDeleted() {
		t.Error("expected IsDeleted() to be true after Delete()")
	}
	if sd.GetDeletedAt() == nil {
		t.Error("expected GetDeletedAt() to be non-nil after Delete()")
	}

	sd.Restore()
	if sd.IsDeleted() {
		t.Error("expected IsDeleted() to be false after Restore()")
	}
	if sd.GetDeletedAt() != nil {
		t.Error("expected GetDeletedAt() to be nil after Restore()")
	}
}

// TestSoftDeletesDefaultColumnMethod verifies SoftDeletes.DeletedAtColumn() returns "deleted_at".
func TestSoftDeletesDefaultColumnMethod(t *testing.T) {
	sd := &soft_delete.SoftDeletes{}
	if sd.DeletedAtColumn() != "deleted_at" {
		t.Errorf("expected 'deleted_at', got %q", sd.DeletedAtColumn())
	}
}
