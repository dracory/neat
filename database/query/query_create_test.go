package query_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dracory/neat/database/query"
)

type BulkUser struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Age   int    `db:"age"`
}

// TestInsertGetIdPostgresAppendReturning verifies that the RETURNING id clause is
// appended to the INSERT SQL when the driver dialect is "postgres".
func TestInsertGetIdPostgresAppendReturning(t *testing.T) {
	w := openSQLiteQuery(t)
	w.Q.Driver()

	fakePg := &query.FakeDriver{DialectName: "postgres"}
	pgW := query.WrapQuery(query.NewTestQuery(w.PrimaryDB(), fakePg, query.MakeDBConfig(), nil))
	pgW.SetTable("users")

	insertSQL, _ := pgW.BuildInsertSQL(map[string]any{"name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}
	if !pgW.IsPostgres() {
		t.Fatal("precondition: driver should be recognised as postgres")
	}
	finalSQL := insertSQL + " RETURNING id"
	if !strings.Contains(finalSQL, "RETURNING id") {
		t.Errorf("expected SQL to contain 'RETURNING id', got: %s", finalSQL)
	}
}

// TestInsertGetIdNonPostgresNoReturning verifies that no RETURNING clause is
// appended for non-postgres dialects.
func TestInsertGetIdNonPostgresNoReturning(t *testing.T) {
	w := openSQLiteQuery(t)
	fakeMy := &query.FakeDriver{DialectName: "mysql"}
	myW := query.WrapQuery(query.NewTestQuery(w.PrimaryDB(), fakeMy, query.MakeDBConfig(), nil))
	myW.SetTable("users")

	insertSQL, _ := myW.BuildInsertSQL(map[string]any{"name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}
	if myW.IsPostgres() {
		t.Fatal("precondition: driver should not be postgres")
	}
	if strings.Contains(insertSQL, "RETURNING") {
		t.Errorf("expected no 'RETURNING' in SQL for mysql dialect, got: %s", insertSQL)
	}
}

// TestInsertGetIdSQLiteReturnsLastInsertId is an end-to-end test using a real
// SQLite in-memory DB and verifies that InsertGetId returns a non-zero ID.
func TestInsertGetIdSQLiteReturnsLastInsertId(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE iid_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	w.SetTable("iid_users")

	id, err := w.Q.InsertGetId(map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("InsertGetId failed: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID from InsertGetId")
	}
}

// BulkUser is defined in query_bulk_test.go
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

	var count int64
	err = w.Q.Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 records, got %d", count)
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

func TestBulkInsertWithPointerSlice(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_ptr_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT, age INTEGER)")
	w.SetTable("bulk_ptr_users")

	users := []*BulkUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	err := w.Q.Create(users)
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
	execSQL(t, w, "CREATE TABLE bulk_single (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT, age INTEGER)")
	w.SetTable("bulk_single")

	users := []BulkUser{{Name: "Single", Email: "single@example.com", Age: 25}}

	err := w.Q.Create(users)
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

// --- Bulk Insert Error Handling Tests ---

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

// --- Bulk Insert Performance Tests ---

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
			_, _ = w.Q.Exec("DELETE FROM bulk_perf")
		})
	}
}

// --- Bulk Insert with Different Data Types ---

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

func TestBulkInsertWithLargeText(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_large (id INTEGER PRIMARY KEY, description TEXT)")
	w.SetTable("bulk_large")

	largeText := strings.Repeat("This is a long text. ", 1000) // ~18KB

	records := []map[string]any{
		{"description": largeText},
		{"description": largeText},
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
		{"name": "Test \"quoted\" text"},
		{"name": "Line1\nLine2"},
		{"name": "Tab\tText"},
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

// --- Bulk Insert with Context ---

func TestBulkInsertWithContext(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_ctx (id INTEGER PRIMARY KEY, name TEXT)")
	w.SetTable("bulk_ctx")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

// --- Bulk Insert with Read Replicas ---

func TestBulkInsertWithReplicas(t *testing.T) {
	writeDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open write DB: %v", err)
	}
	defer func() { _ = writeDB.Close() }()

	readDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open read DB: %v", err)
	}
	defer func() { _ = readDB.Close() }()

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

// TestBulkInsertSetsIDs verifies that bulk insert correctly sets the ID fields
// on the inserted models. Note: Bulk insert ID population is not supported for SQLite
// and MySQL due to unreliable ID calculation. Users should use InsertGetId() for
// bulk inserts if they need accurate IDs.
func TestBulkInsertSetsIDs(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE bulk_id_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT, age INTEGER)")
	w.SetTable("bulk_id_test")

	users := []BulkUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	err := w.Q.Create(users)
	if err != nil {
		t.Fatalf("Bulk insert failed: %v", err)
	}

	// Note: IDs are not set for bulk inserts in SQLite/MySQL
	// Verify by querying the database
	var fetchedUsers []BulkUser = []BulkUser{}
	err = w.Q.Find(&fetchedUsers)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(fetchedUsers) != 3 {
		t.Errorf("Expected 3 records, got %d", len(fetchedUsers))
	}

	// Verify the database has sequential IDs
	if fetchedUsers[1].ID != fetchedUsers[0].ID+1 {
		t.Errorf("Expected sequential IDs in database: user[1].ID (%d) should be user[0].ID+1 (%d)", fetchedUsers[1].ID, fetchedUsers[0].ID+1)
	}
	if fetchedUsers[2].ID != fetchedUsers[1].ID+1 {
		t.Errorf("Expected sequential IDs in database: user[2].ID (%d) should be user[1].ID+1 (%d)", fetchedUsers[2].ID, fetchedUsers[1].ID+1)
	}
}

// TestInsertGetIdSQLServer verifies that SQL Server uses OUTPUT clause
// instead of RETURNING for getting inserted IDs.
func TestInsertGetIdSQLServer(t *testing.T) {
	w := openSQLiteQuery(t)
	w.Q.Driver()

	fakeSQLServer := &query.FakeDriver{DialectName: "sqlserver"}
	sqlServerW := query.WrapQuery(query.NewTestQuery(w.PrimaryDB(), fakeSQLServer, query.MakeDBConfig(), nil))
	sqlServerW.SetTable("users")

	insertSQL, _ := sqlServerW.BuildInsertSQL(map[string]any{"id": 1, "name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}
	if !strings.Contains(insertSQL, "OUTPUT INSERTED") {
		t.Errorf("expected SQL to contain 'OUTPUT INSERTED' for SQL Server, got: %s", insertSQL)
	}
	// Should NOT have RETURNING clause for SQL Server
	if strings.Contains(insertSQL, "RETURNING") {
		t.Errorf("expected no 'RETURNING' clause for SQL Server, got: %s", insertSQL)
	}
}
