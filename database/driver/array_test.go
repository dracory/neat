package driver

import (
	"context"
	"database/sql"
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
	defer db.Close()

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
	defer func() {
		_ = db.Close()
	}()

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

	schema := driver.inferSchema(rows)

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
		_ = db.Close()
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
