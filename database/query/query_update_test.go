package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/dracory/neat/database/query"
)

func TestUpdate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_update (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_update VALUES (1, 'Alice')")
	w.SetTable("test_update")

	result, err := w.Q.Update("name", "Bob")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if result.RowsAffected == 0 {
		t.Error("expected rows affected to be > 0")
	}

	// Verify update
	var row map[string]any
	if err := w.Q.Where("id = ?", 1).First(&row); err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}
	if row["name"] != "Bob" {
		t.Errorf("expected name 'Bob', got %v", row["name"])
	}
}

func TestUpdateWithMap(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_update_map (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	execSQL(t, w, "INSERT INTO test_update_map VALUES (1, 'Alice', 25)")
	w.SetTable("test_update_map")

	result, err := w.Q.Update(map[string]any{"name": "Bob", "age": 30})
	if err != nil {
		t.Fatalf("Update with map failed: %v", err)
	}
	if result.RowsAffected == 0 {
		t.Error("expected rows affected to be > 0")
	}

	// Verify update
	var row map[string]any
	if err := w.Q.Where("id = ?", 1).First(&row); err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}
	if row["name"] != "Bob" {
		t.Errorf("expected name 'Bob', got %v", row["name"])
	}
	if int(row["age"].(int64)) != 30 {
		t.Errorf("expected age 30, got %v", row["age"])
	}
}

// --- Bulk Update Tests ---

func TestBulkUpdateWithWhereClause(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update (id INTEGER PRIMARY KEY, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO bulk_update VALUES (1,'user1','active'),(2,'user2','active'),(3,'user3','inactive')")
	w.SetTable("bulk_update")

	result, err := w.Q.Where("status = ?", "active").Update(map[string]any{"status": "updated"})
	if err != nil {
		t.Fatalf("Bulk update failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	// Verify the update - use raw SQL to avoid query state pollution
	var updatedCount int64
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM bulk_update WHERE status = ?", "updated").Scan(&updatedCount)
	if err != nil {
		t.Fatalf("Raw count failed: %v", err)
	}
	if updatedCount != 2 {
		t.Errorf("Expected 2 records with status 'updated', got %d", updatedCount)
	}
}

func TestBulkUpdateAllRecords(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_all (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_all VALUES (1,10),(2,20),(3,30),(4,40)")
	w.SetTable("bulk_update_all")

	updateResult, err := w.Q.Update(map[string]any{"value": 100})
	if err != nil {
		t.Fatalf("Bulk update all failed: %v", err)
	}

	if updateResult.RowsAffected != 4 {
		t.Errorf("Expected 4 rows affected, got %d", updateResult.RowsAffected)
	}

	// Verify all records were updated
	var results []map[string]any
	err = w.Q.Get(&results)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	for _, r := range results {
		if r["value"] != int64(100) {
			t.Errorf("Expected value 100, got %v", r["value"])
		}
	}
}

func TestBulkUpdateWithNoMatches(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_none (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO bulk_update_none VALUES (1,'test')")
	w.SetTable("bulk_update_none")

	result, err := w.Q.Where("name = ?", "nonexistent").Update(map[string]any{"name": "updated"})
	if err != nil {
		t.Fatalf("Bulk update with no matches failed: %v", err)
	}

	if result.RowsAffected != 0 {
		t.Errorf("Expected 0 rows affected, got %d", result.RowsAffected)
	}
}

func TestBulkUpdateWithLimit(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_limit (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_limit VALUES (1,10),(2,20),(3,30),(4,40),(5,50)")
	w.SetTable("bulk_update_limit")

	result, err := w.Q.Limit(2).Update(map[string]any{"value": 999})
	if err != nil {
		t.Fatalf("Bulk update with limit failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}
}

func TestBulkUpdateMultipleColumns(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_multi (id INTEGER PRIMARY KEY, name TEXT, status TEXT, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_multi VALUES (1,'test','active',100)")
	w.SetTable("bulk_update_multi")

	updateResult, err := w.Q.Where("id = ?", 1).Update(map[string]any{
		"name":   "updated",
		"status": "inactive",
		"value":  200,
	})
	if err != nil {
		t.Fatalf("Bulk update multiple columns failed: %v", err)
	}

	if updateResult.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", updateResult.RowsAffected)
	}

	// Verify all columns were updated
	var record map[string]any
	err = w.Q.Where("id = ?", 1).First(&record)
	if err != nil {
		t.Fatalf("First failed: %v", err)
	}
	if record["name"] != "updated" || record["status"] != "inactive" || record["value"] != int64(200) {
		t.Errorf("Columns not updated correctly: %+v", record)
	}
}

func TestBulkUpdateInTransaction(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_tx (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_tx VALUES (1,10),(2,20),(3,30)")
	w.SetTable("bulk_update_tx")

	tx, err := w.Q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	result, err := tx.Where("value > ?", 15).Update(map[string]any{"value": 999})
	if err != nil {
		t.Fatalf("Bulk update in transaction failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify rollback worked
	var count int64
	err = w.Q.Where("value = ?", 999).Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 records with value 999 after rollback, got %d", count)
	}
}

func TestBulkUpdateWithInvalidColumn(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_error (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO bulk_update_error VALUES (1,'test')")
	w.SetTable("bulk_update_error")

	_, err := w.Q.Update(map[string]any{"nonexistent_column": "value"})
	// Should fail due to invalid column
	if err == nil {
		t.Error("Expected error for invalid column, got nil")
	}
}

func TestBulkUpdatePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_perf (id INTEGER PRIMARY KEY, value INTEGER)")
	w.SetTable("bulk_update_perf")

	// Insert test data
	records := make([]map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		records[i] = map[string]any{"value": i}
	}
	w.Q.Create(&records)

	start := time.Now()
	result, err := w.Q.Update(map[string]any{"value": 999})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Bulk update failed: %v", err)
	}

	t.Logf("Updated %d records in %v", result.RowsAffected, duration)
}

func TestBulkUpdateWithContext(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_ctx (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO bulk_update_ctx VALUES (1,'test')")
	w.SetTable("bulk_update_ctx")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q := query.NewQuery(ctx, w.PrimaryDB(), nil, "", nil, nil)
	q.Table("bulk_update_ctx")

	_, err := q.Update(map[string]any{"name": "updated"})
	if err != nil {
		t.Fatalf("Bulk update with context failed: %v", err)
	}
}

func TestBulkUpdateWithWhereAndLimitNoDuplicateArgs(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_limit_args (id INTEGER PRIMARY KEY, status TEXT, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_limit_args VALUES (1,'active',10),(2,'active',20),(3,'inactive',30),(4,'active',40),(5,'active',50)")
	w.SetTable("bulk_update_limit_args")

	result, err := w.Q.Where("status = ?", "active").Limit(2).Update(map[string]any{"value": 999})
	if err != nil {
		t.Fatalf("Bulk update with where and limit failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	// Verify only 2 records were updated
	var count int64
	err = w.Q.Where("value = ?", 999).Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 records with value 999, got %d", count)
	}
}

func TestBulkUpdateWithLimitNoWhere(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_limit_no_where (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_limit_no_where VALUES (1,10),(2,20),(3,30),(4,40),(5,50)")
	w.SetTable("bulk_update_limit_no_where")

	result, err := w.Q.Limit(2).Update(map[string]any{"value": 999})
	if err != nil {
		t.Fatalf("Bulk update with limit no where failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}
}
