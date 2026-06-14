package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/dracory/neat/database/query"
)

func TestDelete(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_delete (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_delete VALUES (1, 'Alice')")
	w.SetTable("test_delete")

	result, err := w.Q.Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if result.RowsAffected == 0 {
		t.Error("expected rows affected to be > 0")
	}

	// Verify deletion
	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 records after delete, got %d", count)
	}
}

// --- Bulk Delete Tests ---

func TestBulkDeleteWithWhereClause(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete (id INTEGER PRIMARY KEY, status TEXT)")
	execSQL(t, w, "INSERT INTO bulk_delete VALUES (1,'active'),(2,'active'),(3,'inactive'),(4,'active')")
	w.SetTable("bulk_delete")

	result, err := w.Q.Where("status = ?", "active").Delete()
	if err != nil {
		t.Fatalf("Bulk delete failed: %v", err)
	}

	if result.RowsAffected != 3 {
		t.Errorf("Expected 3 rows affected, got %d", result.RowsAffected)
	}

	// Verify with raw SQL to avoid query state pollution
	var count int64
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM bulk_delete").Scan(&count)
	if err != nil {
		t.Fatalf("Raw count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 record remaining, got %d", count)
	}
}

func TestBulkDeleteAllRecords(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_all (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO bulk_delete_all VALUES (1,'a'),(2,'b'),(3,'c')")
	w.SetTable("bulk_delete_all")

	result, err := w.Q.Delete()
	if err != nil {
		t.Fatalf("Bulk delete all failed: %v", err)
	}

	if result.RowsAffected != 3 {
		t.Errorf("Expected 3 rows affected, got %d", result.RowsAffected)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 records remaining, got %d", count)
	}
}

func TestBulkDeleteWithNoMatches(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_none (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO bulk_delete_none VALUES (1,'test')")
	w.SetTable("bulk_delete_none")

	result, err := w.Q.Where("name = ?", "nonexistent").Delete()
	if err != nil {
		t.Fatalf("Bulk delete with no matches failed: %v", err)
	}

	if result.RowsAffected != 0 {
		t.Errorf("Expected 0 rows affected, got %d", result.RowsAffected)
	}
}

func TestBulkDeleteWithLimit(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_limit (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_delete_limit VALUES (1,10),(2,20),(3,30),(4,40),(5,50)")
	w.SetTable("bulk_delete_limit")

	result, err := w.Q.Limit(2).Delete()
	if err != nil {
		t.Fatalf("Bulk delete with limit failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	// Verify with raw SQL to avoid query state pollution
	var count int64
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM bulk_delete_limit").Scan(&count)
	if err != nil {
		t.Fatalf("Raw count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records remaining, got %d", count)
	}
}

func TestBulkDeleteInTransaction(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_tx (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_delete_tx VALUES (1,10),(2,20),(3,30)")
	w.SetTable("bulk_delete_tx")

	tx, err := w.Q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	result, err := tx.Where("value > ?", 15).Delete()
	if err != nil {
		t.Fatalf("Bulk delete in transaction failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify rollback worked
	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records after rollback, got %d", count)
	}
}

func TestBulkDeleteOnEmptyTable(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_empty (id INTEGER PRIMARY KEY)")
	w.SetTable("bulk_delete_empty")

	result, err := w.Q.Delete()
	if err != nil {
		t.Fatalf("Delete on empty table failed: %v", err)
	}

	if result.RowsAffected != 0 {
		t.Errorf("Expected 0 rows affected on empty table, got %d", result.RowsAffected)
	}
}

func TestBulkDeletePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_perf (id INTEGER PRIMARY KEY, value INTEGER)")
	w.SetTable("bulk_delete_perf")

	// Insert test data
	records := make([]map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		records[i] = map[string]any{"value": i}
	}
	_ = w.Q.Create(&records)

	start := time.Now()
	result, err := w.Q.Delete()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Bulk delete failed: %v", err)
	}

	t.Logf("Deleted %d records in %v", result.RowsAffected, duration)
}

func TestBulkDeleteWithContext(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_ctx (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO bulk_delete_ctx VALUES (1,'test')")
	w.SetTable("bulk_delete_ctx")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q := query.NewQuery(ctx, w.PrimaryDB(), nil, "", nil, nil)
	q.Table("bulk_delete_ctx")

	_, err := q.Delete()
	if err != nil {
		t.Fatalf("Bulk delete with context failed: %v", err)
	}
}

func TestBulkDeleteWithWhereAndLimitNoDuplicateArgs(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_limit_args (id INTEGER PRIMARY KEY, status TEXT)")
	execSQL(t, w, "INSERT INTO bulk_delete_limit_args VALUES (1,'active'),(2,'active'),(3,'inactive'),(4,'active'),(5,'active')")
	w.SetTable("bulk_delete_limit_args")

	// This test verifies that WHERE arguments are not duplicated when using LIMIT
	result, err := w.Q.Where("status = ?", "active").Limit(2).Delete()
	if err != nil {
		t.Fatalf("Bulk delete with WHERE and LIMIT failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	// Verify with raw SQL to avoid query state pollution
	var count int64
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM bulk_delete_limit_args").Scan(&count)
	if err != nil {
		t.Fatalf("Raw count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records remaining, got %d", count)
	}
}

func TestBulkDeleteWithLimitNoWhere(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_limit_no_where (id INTEGER PRIMARY KEY)")
	execSQL(t, w, "INSERT INTO bulk_delete_limit_no_where VALUES (1),(2),(3),(4),(5)")
	w.SetTable("bulk_delete_limit_no_where")

	// Test LIMIT without WHERE clause
	result, err := w.Q.Limit(2).Delete()
	if err != nil {
		t.Fatalf("Bulk delete with LIMIT and no WHERE failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	// Verify with raw SQL to avoid query state pollution
	var count int64
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM bulk_delete_limit_no_where").Scan(&count)
	if err != nil {
		t.Fatalf("Raw count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records remaining, got %d", count)
	}
}

func TestBulkDeleteWithLimitAndOrderBy(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_delete_order (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_delete_order VALUES (1,10),(2,20),(3,30),(4,40),(5,50)")
	w.SetTable("bulk_delete_order")

	// Delete the first 2 rows by ID (ORDER BY id)
	result, err := w.Q.OrderBy("id").Limit(2).Delete()
	if err != nil {
		t.Fatalf("Bulk delete with LIMIT and ORDER BY failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	// Verify with raw SQL to avoid query state pollution
	var count int64
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM bulk_delete_order").Scan(&count)
	if err != nil {
		t.Fatalf("Raw count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records remaining, got %d", count)
	}
}

func TestDestroy(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_destroy (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_destroy VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_destroy VALUES (2, 'Bob')")

	w.SetTable("test_destroy")
	w.Q.Where("name = ?", "Alice")

	result, err := w.Q.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", result.RowsAffected)
	}

	// Verify deletion using raw SQL to avoid query state pollution
	var count int64
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM test_destroy").Scan(&count)
	if err != nil {
		t.Fatalf("Raw count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row remaining, got %d", count)
	}
}

func TestDestroyAsAliasForDelete(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_destroy_alias (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_destroy_alias VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_destroy_alias VALUES (2, 'Bob')")

	// Test Delete
	w.SetTable("test_destroy_alias")
	w.Q.Where("name = ?", "Alice")
	result1, err := w.Q.Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Reset and test Destroy
	execSQL(t, w, "DELETE FROM test_destroy_alias")
	execSQL(t, w, "INSERT INTO test_destroy_alias VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_destroy_alias VALUES (2, 'Bob')")

	w.SetTable("test_destroy_alias")
	w.Q.Where("name = ?", "Alice")
	result2, err := w.Q.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	if result1.RowsAffected != result2.RowsAffected {
		t.Errorf("Destroy and Delete should affect same number of rows. Got %d vs %d", result1.RowsAffected, result2.RowsAffected)
	}
}
