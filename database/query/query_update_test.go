package query_test

import (
	"testing"
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
