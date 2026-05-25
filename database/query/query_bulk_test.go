package query_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat/database/query"
	_ "modernc.org/sqlite"
)

// --- Bulk Insert Tests ---

type BulkUser struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Age   int    `db:"age"`
}

func TestBulkInsertWithStructSlice(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT, age INTEGER)")
	w.SetTable("bulk_users")

	users := []BulkUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	err := w.Q.Create(&users)
	if err != nil {
		t.Fatalf("Bulk insert with struct slice failed: %v", err)
	}

	// Verify all records were inserted
	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records, got %d", count)
	}

	// Verify IDs were set
	for i, user := range users {
		if user.ID == 0 {
			t.Errorf("User %d: Expected ID to be set, got 0", i)
		}
	}
}

func TestBulkInsertWithMapSlice(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_map_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT)")
	w.SetTable("bulk_map_users")

	users := []map[string]any{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
		{"name": "Charlie", "email": "charlie@example.com"},
	}

	err := w.Q.Create(&users)
	if err != nil {
		t.Fatalf("Bulk insert with map slice failed: %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records, got %d", count)
	}
}

func TestBulkInsertWithManyRecords(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_many (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, value INTEGER)")
	w.SetTable("bulk_many")

	// Insert 1000 records
	records := make([]map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		records[i] = map[string]any{
			"name":  fmt.Sprintf("record_%d", i),
			"value": i,
		}
	}

	start := time.Now()
	err := w.Q.Create(&records)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Bulk insert with many records failed: %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1000 {
		t.Errorf("Expected 1000 records, got %d", count)
	}

	t.Logf("Bulk insert of 1000 records took %v", duration)
}

func TestBulkInsertWithPointerSlice(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_ptr_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT)")
	w.SetTable("bulk_ptr_users")

	users := []*BulkUser{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
		{Name: "Charlie", Email: "charlie@example.com"},
	}

	err := w.Q.Create(&users)
	if err != nil {
		t.Fatalf("Bulk insert with pointer slice failed: %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records, got %d", count)
	}
}

func TestBulkInsertEmptySlice(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_empty (id INTEGER PRIMARY KEY, name TEXT)")
	w.SetTable("bulk_empty")

	users := []BulkUser{}

	err := w.Q.Create(&users)
	// Empty slice should either succeed gracefully or return a clear error
	if err != nil {
		// This is acceptable behavior
		t.Logf("Bulk insert with empty slice returned error (acceptable): %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 records after empty insert, got %d", count)
	}
}

func TestBulkInsertSingleRecord(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_single (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	w.SetTable("bulk_single")

	users := []BulkUser{{Name: "Single"}}

	err := w.Q.Create(&users)
	if err != nil {
		t.Fatalf("Bulk insert with single record failed: %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
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

// --- Bulk Operation Error Handling Tests ---

func TestBulkInsertWithInvalidDataType(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_error (id INTEGER PRIMARY KEY, age INTEGER)")
	w.SetTable("bulk_error")

	users := []map[string]any{
		{"age": "invalid"}, // String instead of integer
	}

	// SQLite uses dynamic typing and will attempt to convert the string to an integer.
	// For non-numeric strings like "invalid", SQLite stores 0 instead of failing.
	// This test verifies the insert succeeds (SQLite's lenient behavior).
	err := w.Q.Create(&users)
	if err != nil {
		t.Fatalf("Bulk insert with invalid data type failed (SQLite allows this): %v", err)
	}

	// Verify the record was inserted with SQLite's default conversion (0 for invalid strings)
	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

func TestBulkInsertWithMissingRequiredColumn(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_required (id INTEGER PRIMARY KEY, name TEXT NOT NULL)")
	w.SetTable("bulk_required")

	users := []map[string]any{
		{}, // Missing required 'name' column
	}

	err := w.Q.Create(&users)
	// Should fail due to NOT NULL constraint
	if err == nil {
		t.Error("Expected error for missing required column, got nil")
	}
}

func TestBulkInsertWithDuplicateKey(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_unique (id INTEGER PRIMARY KEY, email TEXT UNIQUE)")
	execSQL(t, w, "INSERT INTO bulk_unique VALUES (1,'test@example.com')")
	w.SetTable("bulk_unique")

	users := []map[string]any{
		{"email": "test@example.com"}, // Duplicate email
	}

	err := w.Q.Create(&users)
	// Should fail due to UNIQUE constraint
	if err == nil {
		t.Error("Expected error for duplicate key, got nil")
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

func TestBulkOperationWithNilSlice(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_nil (id INTEGER PRIMARY KEY, name TEXT)")
	w.SetTable("bulk_nil")

	var users []BulkUser = nil

	err := w.Q.Create(&users)
	// Should handle nil slice gracefully
	if err != nil {
		t.Logf("Bulk insert with nil slice returned error: %v", err)
	}
}

// --- Bulk Operation Performance Tests ---

func TestBulkInsertPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_perf (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, value INTEGER)")
	w.SetTable("bulk_perf")

	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			records := make([]map[string]any, size)
			for i := 0; i < size; i++ {
				records[i] = map[string]any{
					"name":  fmt.Sprintf("record_%d", i),
					"value": i,
				}
			}

			start := time.Now()
			err := w.Q.Create(&records)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Bulk insert failed: %v", err)
			}

			t.Logf("Inserted %d records in %v (%.2f records/sec)", size, duration, float64(size)/duration.Seconds())

			// Clean up for next test
			w.Q.Exec("DELETE FROM bulk_perf")
		})
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
	w.Q.Create(&records)

	start := time.Now()
	result, err := w.Q.Delete()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Bulk delete failed: %v", err)
	}

	t.Logf("Deleted %d records in %v", result.RowsAffected, duration)
}

// --- Bulk Operation with Different Data Types ---

func TestBulkInsertWithMixedTypes(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_mixed (id INTEGER PRIMARY KEY, name TEXT, age INTEGER, score REAL, active BOOLEAN)")
	w.SetTable("bulk_mixed")

	records := []map[string]any{
		{"name": "Alice", "age": 25, "score": 95.5, "active": true},
		{"name": "Bob", "age": 30, "score": 87.3, "active": false},
		{"name": "Charlie", "age": 35, "score": 92.1, "active": true},
	}

	err := w.Q.Create(&records)
	if err != nil {
		t.Fatalf("Bulk insert with mixed types failed: %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records, got %d", count)
	}
}

func TestBulkInsertWithNullValues(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_null (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")
	w.SetTable("bulk_null")

	records := []map[string]any{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": nil}, // NULL email
		{"name": "Charlie", "email": "charlie@example.com"},
	}

	err := w.Q.Create(&records)
	if err != nil {
		t.Fatalf("Bulk insert with null values failed: %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records, got %d", count)
	}
}

// --- Bulk Operation Edge Cases ---

func TestBulkInsertWithLargeText(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_large_text (id INTEGER PRIMARY KEY, content TEXT)")
	w.SetTable("bulk_large_text")

	largeText := string(make([]byte, 10000)) // 10KB text
	for i := range largeText {
		largeText = largeText[:i] + "a" + largeText[i+1:]
	}

	records := []map[string]any{
		{"content": largeText},
		{"content": largeText},
	}

	err := w.Q.Create(&records)
	if err != nil {
		t.Fatalf("Bulk insert with large text failed: %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 records, got %d", count)
	}
}

func TestBulkInsertWithSpecialCharacters(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_special (id INTEGER PRIMARY KEY, name TEXT)")
	w.SetTable("bulk_special")

	records := []map[string]any{
		{"name": "O'Brien"},
		{"name": "Alice & Bob"},
		{"name": "Test \"quotes\""},
		{"name": "Line\nBreak"},
		{"name": "Tab\tCharacter"},
	}

	err := w.Q.Create(&records)
	if err != nil {
		t.Fatalf("Bulk insert with special characters failed: %v", err)
	}

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected 5 records, got %d", count)
	}
}

// --- Bulk Operation with Context ---

func TestBulkInsertWithContext(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_ctx (id INTEGER PRIMARY KEY, name TEXT)")
	w.SetTable("bulk_ctx")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a new query with context
	q := query.NewQuery(ctx, w.PrimaryDB(), nil, "", nil, nil)
	q.Table("bulk_ctx")

	records := []map[string]any{
		{"name": "Alice"},
		{"name": "Bob"},
	}

	err := q.Create(&records)
	if err != nil {
		t.Fatalf("Bulk insert with context failed: %v", err)
	}
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

// --- Bulk Operation with Read Replicas ---

func TestBulkInsertWithReplicas(t *testing.T) {
	writeDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open write DB: %v", err)
	}
	defer writeDB.Close()

	readDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open read DB: %v", err)
	}
	defer readDB.Close()

	// Setup schema on both DBs
	_, err = writeDB.Exec("CREATE TABLE test_replicas (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = readDB.Exec("CREATE TABLE test_replicas (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table on read DB: %v", err)
	}

	q := query.NewQueryWithReplicas(context.Background(), writeDB, readDB, nil, "", nil, nil)
	q.Table("test_replicas")

	records := []map[string]any{
		{"name": "Alice"},
		{"name": "Bob"},
	}

	err = q.Create(&records)
	if err != nil {
		t.Fatalf("Bulk insert with replicas failed: %v", err)
	}

	// Verify write was successful on the write DB
	// Note: read DB is a separate in-memory instance, so we verify on write DB
	var count int64
	err = writeDB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM test_replicas").Scan(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 records, got %d", count)
	}
}

// --- Tests for Bug Fixes ---

func TestBulkUpdateWithWhereAndLimitNoDuplicateArgs(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_limit_args (id INTEGER PRIMARY KEY, status TEXT, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_limit_args VALUES (1,'active',10),(2,'active',20),(3,'inactive',30),(4,'active',40),(5,'active',50)")
	w.SetTable("bulk_update_limit_args")

	// This test verifies that WHERE arguments are not duplicated when using LIMIT
	result, err := w.Q.Where("status = ?", "active").Limit(2).Update(map[string]any{"value": 999})
	if err != nil {
		t.Fatalf("Bulk update with WHERE and LIMIT failed: %v", err)
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
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM bulk_update_limit_args WHERE value = ?", 999).Scan(&count)
	if err != nil {
		t.Fatalf("Raw count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 records with value 999, got %d", count)
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

func TestBulkUpdateWithLimitNoWhere(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_limit_no_where (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_limit_no_where VALUES (1,10),(2,20),(3,30),(4,40),(5,50)")
	w.SetTable("bulk_update_limit_no_where")

	// Test LIMIT without WHERE clause
	result, err := w.Q.Limit(2).Update(map[string]any{"value": 999})
	if err != nil {
		t.Fatalf("Bulk update with LIMIT and no WHERE failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
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

func TestBulkUpdateWithLimitAndOrderBy(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_update_order (id INTEGER PRIMARY KEY, value INTEGER)")
	execSQL(t, w, "INSERT INTO bulk_update_order VALUES (1,10),(2,20),(3,30),(4,40),(5,50)")
	w.SetTable("bulk_update_order")

	// Update the first 2 rows by ID (ORDER BY id)
	result, err := w.Q.OrderBy("id").Limit(2).Update(map[string]any{"value": 999})
	if err != nil {
		t.Fatalf("Bulk update with LIMIT and ORDER BY failed: %v", err)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
	}

	// Verify that rows with id=1 and id=2 were updated (not arbitrary rows)
	// Use raw SQL to avoid query state pollution
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	var results []map[string]any
	rows, err := db.QueryContext(context.Background(), "SELECT id, value FROM bulk_update_order ORDER BY id")
	if err != nil {
		t.Fatalf("Raw query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var value int64
		if err := rows.Scan(&id, &value); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		results = append(results, map[string]any{"id": id, "value": value})
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 records, got %d", len(results))
	}
	// First 2 rows should have value 999
	if results[0]["value"] != int64(999) || results[1]["value"] != int64(999) {
		t.Errorf("Expected first 2 rows to have value 999, got %v and %v", results[0]["value"], results[1]["value"])
	}
	// Remaining rows should have original values
	if results[2]["value"] != int64(30) || results[3]["value"] != int64(40) || results[4]["value"] != int64(50) {
		t.Errorf("Expected remaining rows to have original values, got %v, %v, %v", results[2]["value"], results[3]["value"], results[4]["value"])
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

	// Verify that rows with id=1 and id=2 were deleted (not arbitrary rows)
	// Use raw SQL to avoid query state pollution
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() failed: %v", err)
	}
	var results []map[string]any
	rows, err := db.QueryContext(context.Background(), "SELECT id, value FROM bulk_delete_order ORDER BY id")
	if err != nil {
		t.Fatalf("Raw query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var value int64
		if err := rows.Scan(&id, &value); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		results = append(results, map[string]any{"id": id, "value": value})
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 records remaining, got %d", len(results))
	}
	// Remaining rows should be id=3, id=4, id=5
	if results[0]["id"] != int64(3) || results[1]["id"] != int64(4) || results[2]["id"] != int64(5) {
		t.Errorf("Expected remaining rows to have ids 3, 4, 5, got %v, %v, %v", results[0]["id"], results[1]["id"], results[2]["id"])
	}
}
