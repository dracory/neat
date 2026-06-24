package driver

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

type mockArraySource struct {
	tableName string
	rows      []map[string]any
}

func (m *mockArraySource) TableName() string {
	return m.tableName
}

func (m *mockArraySource) Rows() ([]map[string]any, error) {
	return m.rows, nil
}

type mockArraySourceWithSchema struct {
	mockArraySource
	schema map[string]string
}

func (m *mockArraySourceWithSchema) Schema() map[string]string {
	return m.schema
}

func TestArrayPopulate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	driver := NewArray()
	now := time.Now().Round(time.Second)

	source := &mockArraySource{
		tableName: "users",
		rows: []map[string]any{
			{"id": 1, "name": "John", "active": true, "created_at": now},
			{"id": 2, "name": "Jane", "active": false, "created_at": now},
		},
	}

	ctx := context.Background()
	err = driver.Populate(ctx, db, source)
	if err != nil {
		t.Fatalf("Populate failed: %v", err)
	}

	// Verify data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}

	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if name != "John" {
		t.Errorf("Expected John, got %s", name)
	}

	// Verify all columns and types (round-trip)
	rows, err := db.Query("SELECT id, name, active, created_at FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			t.Errorf("failed to close rows: %v", err)
		}
	}()
	for rows.Next() {
		var id int
		var name string
		var active bool
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &active, &createdAt); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		if id == 1 {
			if name != "John" || active != true || !createdAt.Equal(now) {
				t.Errorf("Round-trip failed for ID 1: got name=%s, active=%v, created_at=%v", name, active, createdAt)
			}
		}
	}

	// Test idempotency
	err = driver.Populate(ctx, db, source)
	if err != nil {
		t.Fatalf("Second Populate failed: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 rows after second populate, got %d", count)
	}
}

func TestArraySchemaInference(t *testing.T) {
	driver := NewArray()

	rows := []map[string]any{
		{"id": nil, "name": nil, "score": nil},
		{"id": 1, "name": "John", "score": 1.5, "active": true},
	}

	schema, _ := driver.inferSchema(rows)

	if schema["id"] != "INTEGER" {
		t.Errorf("Expected id to be INTEGER, got %s", schema["id"])
	}
	if schema["name"] != "TEXT" {
		t.Errorf("Expected name to be TEXT, got %s", schema["name"])
	}
	if schema["score"] != "REAL" {
		t.Errorf("Expected score to be REAL, got %s", schema["score"])
	}
	if schema["active"] != "INTEGER" {
		t.Errorf("Expected active to be INTEGER, got %s", schema["active"])
	}
}

func TestArrayEmptyWithSchema(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	driver := NewArray()
	source := &mockArraySourceWithSchema{
		mockArraySource: mockArraySource{
			tableName: "empty_table",
			rows:      nil,
		},
		schema: map[string]string{
			"id":   "int",
			"name": "string",
		},
	}

	ctx := context.Background()
	err = driver.Populate(ctx, db, source)
	if err != nil {
		t.Fatalf("Populate failed: %v", err)
	}

	// Verify table exists
	_, err = db.Exec("SELECT * FROM empty_table")
	if err != nil {
		t.Errorf("Expected table to exist, but got error: %v", err)
	}
}

func TestArrayInvalidIdentifiers(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	driver := NewArray()
	ctx := context.Background()

	// Invalid table name
	source1 := &mockArraySource{
		tableName: "users; DROP TABLE users",
		rows:      []map[string]any{{"id": 1}},
	}
	if err := driver.Populate(ctx, db, source1); err == nil {
		t.Error("Expected error for invalid table name, got nil")
	}

	// Invalid column name
	source2 := &mockArraySource{
		tableName: "valid_table",
		rows:      []map[string]any{{"id\"; --": 1}},
	}
	if err := driver.Populate(ctx, db, source2); err == nil {
		t.Error("Expected error for invalid column name, got nil")
	}
}

func TestArrayTypeWidening(t *testing.T) {
	driver := NewArray()

	rows := []map[string]any{
		{"val": 1},
		{"val": 1.5},
	}

	schema, err := driver.inferSchema(rows)
	if err != nil {
		t.Fatalf("inferSchema failed: %v", err)
	}
	if schema["val"] != "REAL" {
		t.Errorf("Expected val to be REAL (widened from INTEGER), got %s", schema["val"])
	}

	rows2 := []map[string]any{
		{"val": "string"},
		{"val": 1},
	}
	schema2, err := driver.inferSchema(rows2)
	if err != nil {
		t.Fatalf("inferSchema failed: %v", err)
	}
	if schema2["val"] != "TEXT" {
		t.Errorf("Expected val to be TEXT (widened from incompatible types), got %s", schema2["val"])
	}
}

func TestArraySchemaMismatch(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	driver := NewArray()
	ctx := context.Background()

	source := &mockArraySourceWithSchema{
		mockArraySource: mockArraySource{
			tableName: "mismatch_table",
			rows: []map[string]any{
				{"id": 1, "extra": "data"},
			},
		},
		schema: map[string]string{
			"id": "int",
		},
	}

	err = driver.Populate(ctx, db, source)
	if err == nil {
		t.Error("Expected error for schema/row mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "contains key \"extra\" which is not in the explicit schema") {
		t.Errorf("Expected mismatch error message, got: %v", err)
	}
}

func TestArrayUnsupportedType(t *testing.T) {
	driver := NewArray()
	rows := []map[string]any{
		{"val": []string{"unsupported"}},
	}
	_, err := driver.inferSchema(rows)
	if err == nil {
		t.Error("Expected error for unsupported type, got nil")
	}
}

func TestArrayConcurrentPopulation(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	driver := NewArray()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tableName := "concurrent_table"
			source := &mockArraySource{
				tableName: tableName,
				rows:      []map[string]any{{"id": 1}},
			}
			if err := driver.Populate(ctx, db, source); err != nil {
				t.Errorf("Populate failed for %s: %v", tableName, err)
			}
		}(i)
	}
	wg.Wait()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM concurrent_table").Scan(&count); err != nil {
		t.Errorf("table check failed: %v", err)
	}
}

func TestArrayCleanup(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()

	drv := NewArray()
	ctx := context.Background()

	source := &mockArraySource{
		tableName: "cleanup_test",
		rows:      []map[string]any{{"id": 1}},
	}

	if err := drv.Populate(ctx, db, source); err != nil {
		t.Fatalf("Populate failed: %v", err)
	}

	// Verify the populated entry exists
	if !drv.isPopulated(db, "cleanup_test") {
		t.Fatal("Expected table to be marked as populated before Cleanup")
	}

	// Call Cleanup and verify the entry is removed
	drv.Cleanup(db)
	if drv.isPopulated(db, "cleanup_test") {
		t.Error("Expected table to NOT be marked as populated after Cleanup")
	}

	// After Cleanup, Populate should re-populate successfully
	if err := drv.Populate(ctx, db, source); err != nil {
		t.Fatalf("Populate after Cleanup failed: %v", err)
	}
	if !drv.isPopulated(db, "cleanup_test") {
		t.Error("Expected table to be marked as populated after re-Populate")
	}
}

func TestArrayMaxRowsLimit(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()

	drv := NewArray()
	ctx := context.Background()

	// Create a source with MaxArrayRows + 1 rows
	rows := make([]map[string]any, MaxArrayRows+1)
	for i := range rows {
		rows[i] = map[string]any{"id": i}
	}

	source := &mockArraySource{
		tableName: "too_many_rows",
		rows:      rows,
	}

	err = drv.Populate(ctx, db, source)
	if err == nil {
		t.Fatal("Expected error for exceeding MaxArrayRows, got nil")
	}
	if !strings.Contains(err.Error(), "exceeding the limit") {
		t.Errorf("Expected 'exceeding the limit' error, got: %v", err)
	}
}

func TestArraySQLKeywordColumnNames(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()

	drv := NewArray()
	ctx := context.Background()

	// Column names that are SQL keywords — these must be accepted because
	// all identifiers are double-quoted in the generated SQL, making them safe.
	source := &mockArraySource{
		tableName: "keyword_cols",
		rows: []map[string]any{
			{"id": 1, "order": "first", "group": "A", "index": 10},
			{"id": 2, "order": "second", "group": "B", "index": 20},
		},
	}

	if err := drv.Populate(ctx, db, source); err != nil {
		t.Fatalf("Populate with SQL keyword column names failed: %v", err)
	}

	var orderVal, groupVal string
	var indexVal int
	err = db.QueryRow(`SELECT "order", "group", "index" FROM keyword_cols WHERE id = 1`).Scan(&orderVal, &groupVal, &indexVal)
	if err != nil {
		t.Fatalf("Query with quoted keyword columns failed: %v", err)
	}
	if orderVal != "first" || groupVal != "A" || indexVal != 10 {
		t.Errorf("Unexpected values: order=%s, group=%s, index=%d", orderVal, groupVal, indexVal)
	}
}

func TestArrayBlobType(t *testing.T) {
	drv := NewArray()

	rows := []map[string]any{
		{"id": 1, "data": []byte("binary data")},
	}

	schema, err := drv.inferSchema(rows)
	if err != nil {
		t.Fatalf("inferSchema failed: %v", err)
	}
	if schema["data"] != "BLOB" {
		t.Errorf("Expected data to be BLOB, got %s", schema["data"])
	}

	// Full Populate round-trip — verifies createTable accepts BLOB type
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	source := &mockArraySource{
		tableName: "blob_test",
		rows:      rows,
	}
	if err := drv.Populate(ctx, db, source); err != nil {
		t.Fatalf("Populate with []byte column failed: %v", err)
	}

	var id int
	var data []byte
	err = db.QueryRow("SELECT id, data FROM blob_test WHERE id = 1").Scan(&id, &data)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if string(data) != "binary data" {
		t.Errorf("Expected 'binary data', got %q", string(data))
	}
}
