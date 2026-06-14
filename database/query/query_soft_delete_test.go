package query_test

import (
	"strings"
	"testing"
	"time"

	"github.com/dracory/neat/database/query"
)

// softModel has a *time.Time DeletedAt and implements SoftDeleteColumnNamer.
type softModel struct {
	ID        int
	Name      string
	DeletedAt *time.Time
}

// DeletedAtColumn implements SoftDeleteColumnNamer so the query builder applies
// soft-delete filtering using the "deleted_at" column.
func (m *softModel) SoftDeletedAtColumn() string { return "deleted_at" }

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

// TestSoftDeleteExecution tests actual DELETE execution with soft delete
func TestSoftDeleteExecution(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE soft_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, deleted_at DATETIME)")
	w.SetTable("soft_models")
	w.SetModel(&softModel{})

	// Create a record
	err := w.Q.Create(map[string]any{"name": "test_user"})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Get the created record
	var created softModel
	err = w.Q.Where("name = ?", "test_user").First(&created)
	if err != nil {
		t.Fatalf("Failed to get created record: %v", err)
	}

	// Soft delete the record
	res, err := w.Q.Where("name = ?", "test_user").Delete(&softModel{})
	if err != nil {
		t.Fatalf("Failed to soft delete: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify record is not found without WithTrashed
	var notFound softModel
	err = w.Q.Where("id = ?", created.ID).First(&notFound)
	if err == nil {
		t.Error("Expected error when finding soft deleted record without WithTrashed")
	}

	// Verify record is found with WithTrashed
	var found softModel
	err = w.Q.WithTrashed().Where("id = ?", created.ID).First(&found)
	if err != nil {
		t.Fatalf("Failed to find soft deleted record with WithTrashed: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, found.ID)
	}

	if found.DeletedAt == nil {
		t.Error("DeletedAt should be set for soft deleted record")
	}
}

// TestRestoreExecution tests actual Restore execution
func TestRestoreExecution(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE soft_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, deleted_at DATETIME)")
	w.SetTable("soft_models")
	w.SetModel(&softModel{})

	// Create a record
	err := w.Q.Create(map[string]any{"name": "user1"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	// Soft delete the record
	res, err := w.Q.Where("name = ?", "user1").Delete(&softModel{})
	if err != nil {
		t.Fatalf("Failed to soft delete: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Restore the record with WithTrashed and where condition
	res, err = w.Q.WithTrashed().Where("name = ?", "user1").Restore()
	if err != nil {
		t.Fatalf("Failed to restore user1: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify record is restored (can be found without WithTrashed)
	var restoredUser softModel
	err = w.Q.Where("name = ?", "user1").First(&restoredUser)
	if err != nil {
		t.Fatalf("Failed to find restored user: %v", err)
	}

	if restoredUser.DeletedAt != nil {
		t.Error("DeletedAt should be nil for restored record")
	}
}

// TestForceDeleteExecution tests actual ForceDelete execution
func TestForceDeleteExecution(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE soft_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, deleted_at DATETIME)")
	w.SetTable("soft_models")
	w.SetModel(&softModel{})

	// Create a record
	err := w.Q.Create(map[string]any{"name": "force_delete_user"})
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Get the created record
	var created softModel
	err = w.Q.Where("name = ?", "force_delete_user").First(&created)
	if err != nil {
		t.Fatalf("Failed to get created record: %v", err)
	}

	// Soft delete the record first
	res, err := w.Q.Where("name = ?", "force_delete_user").Delete(&softModel{})
	if err != nil {
		t.Fatalf("Failed to soft delete: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify record is soft deleted
	var softDeleted softModel
	err = w.Q.WithTrashed().Where("id = ?", created.ID).First(&softDeleted)
	if err != nil {
		t.Fatalf("Failed to find soft deleted record: %v", err)
	}

	if softDeleted.DeletedAt == nil {
		t.Error("Record should be soft deleted")
	}

	// Force delete the record (permanent deletion)
	res, err = w.Q.Where("name = ?", "force_delete_user").ForceDelete(&softModel{})
	if err != nil {
		t.Fatalf("Failed to force delete: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify record is permanently deleted (not found even with WithTrashed)
	var permanentlyDeleted softModel
	err = w.Q.WithTrashed().Where("id = ?", created.ID).First(&permanentlyDeleted)
	if err == nil {
		t.Error("Expected error when finding permanently deleted record")
	}

	if permanentlyDeleted.ID != 0 {
		t.Error("Record should be permanently deleted")
	}
}

// TestSoftDeleteWithRelations tests soft delete behavior with related models
func TestSoftDeleteWithRelations(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE soft_models (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, deleted_at DATETIME)")
	execSQL(t, w, "CREATE TABLE related_models (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, soft_model_id INTEGER, deleted_at DATETIME)")
	w.SetTable("soft_models")
	w.SetModel(&softModel{})

	// Create parent record
	err := w.Q.Create(map[string]any{"name": "parent"})
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	var parent softModel
	err = w.Q.Where("name = ?", "parent").First(&parent)
	if err != nil {
		t.Fatalf("Failed to get parent: %v", err)
	}

	// Create related records using raw SQL
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("Failed to get DB: %v", err)
	}
	_, err = db.Exec("INSERT INTO related_models (title, soft_model_id) VALUES (?, ?)", "related1", parent.ID)
	if err != nil {
		t.Fatalf("Failed to create related1: %v", err)
	}
	_, err = db.Exec("INSERT INTO related_models (title, soft_model_id) VALUES (?, ?)", "related2", parent.ID)
	if err != nil {
		t.Fatalf("Failed to create related2: %v", err)
	}

	// Soft delete parent record
	res, err := w.Q.Where("id = ?", parent.ID).Delete(&softModel{})
	if err != nil {
		t.Fatalf("Failed to soft delete parent: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify parent is soft deleted
	var deletedParent softModel
	err = w.Q.WithTrashed().Where("id = ?", parent.ID).First(&deletedParent)
	if err != nil {
		t.Fatalf("Failed to find soft deleted parent: %v", err)
	}

	if deletedParent.DeletedAt == nil {
		t.Error("Parent should be soft deleted")
	}

	// Verify related records still exist in database (soft delete doesn't cascade)
	var count int64
	err = db.QueryRow("SELECT COUNT(*) FROM related_models WHERE soft_model_id = ?", parent.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count related records: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 related records in database, got %d", count)
	}

	// Soft delete a related record using raw SQL
	_, err = db.Exec("UPDATE related_models SET deleted_at = datetime('now') WHERE title = ?", "related1")
	if err != nil {
		t.Fatalf("Failed to soft delete related1: %v", err)
	}

	// Verify one related record is soft deleted
	err = db.QueryRow("SELECT COUNT(*) FROM related_models WHERE soft_model_id = ? AND deleted_at IS NULL", parent.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count active related records: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 active related record, got %d", count)
	}

	// Verify all related records exist including deleted ones
	err = db.QueryRow("SELECT COUNT(*) FROM related_models WHERE soft_model_id = ?", parent.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count all related records: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 total related records, got %d", count)
	}
}
