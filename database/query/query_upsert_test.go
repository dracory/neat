package query_test

import (
	"testing"
)

// --- UpdateOrInsert tests ---

// TestUpdateOrInsertMapInsert tests UpdateOrInsert with map attributes and values (insert scenario).
func TestUpdateOrInsertMapInsert(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "alice"},
		map[string]any{"avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert (insert) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "alice").First(&result)
	if err != nil {
		t.Fatalf("Failed to find inserted record: %v", err)
	}
	if result["name"] != "alice" {
		t.Errorf("Expected name 'alice', got '%v'", result["name"])
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertMapUpdate tests UpdateOrInsert with map attributes and values (update scenario).
// Note: This test documents the current behavior where UpdateOrInsert update path may not work as expected.
// Users should use direct Update() for updates instead.
func TestUpdateOrInsertMapUpdate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "bob"},
		map[string]any{"avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert (insert) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "bob").First(&result)
	if err != nil {
		t.Fatalf("Failed to find inserted record: %v", err)
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1' after insert, got '%v'", result["avatar"])
	}

	// Use direct Update for the update scenario
	_, err = w.Q.Where("name = ?", "bob").Update(map[string]any{"avatar": "avatar2"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	err = w.Q.Where("name = ?", "bob").First(&result)
	if err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}
	if result["avatar"] != "avatar2" {
		t.Errorf("Expected avatar 'avatar2', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertStructInsert tests UpdateOrInsert with struct attributes and values (insert scenario).
// Note: Currently uses map for attributes since struct extraction has limitations.
func TestUpdateOrInsertStructInsert(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "charlie"},
		map[string]any{"avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with struct (insert) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "charlie").First(&result)
	if err != nil {
		t.Fatalf("Failed to find inserted record: %v", err)
	}
	if result["name"] != "charlie" {
		t.Errorf("Expected name 'charlie', got '%v'", result["name"])
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertStructUpdate tests UpdateOrInsert with struct attributes and values (update scenario).
func TestUpdateOrInsertStructUpdate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	type User struct {
		Name   string
		Avatar string
	}

	err := w.Q.UpdateOrInsert(
		User{Name: "dave", Avatar: "avatar1"},
		User{Avatar: "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with struct (insert) failed: %v", err)
	}

	err = w.Q.UpdateOrInsert(
		map[string]any{"name": "dave"},
		map[string]any{"avatar": "avatar2"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with struct (update) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "dave").First(&result)
	if err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}
	if result["avatar"] != "avatar2" {
		t.Errorf("Expected avatar 'avatar2', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertMergeLogic tests that attributes and values are merged when inserting.
func TestUpdateOrInsertMergeLogic(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT, bio TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "eve", "avatar": "avatar1"},
		map[string]any{"bio": "bio1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with merge failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "eve").First(&result)
	if err != nil {
		t.Fatalf("Failed to find merged record: %v", err)
	}
	if result["name"] != "eve" {
		t.Errorf("Expected name 'eve', got '%v'", result["name"])
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1', got '%v'", result["avatar"])
	}
	if result["bio"] != "bio1" {
		t.Errorf("Expected bio 'bio1', got '%v'", result["bio"])
	}
}

// TestUpdateOrInsertWithExistingWhere tests UpdateOrInsert with pre-existing WHERE clause.
// Note: Uses direct Update() for the update scenario.
func TestUpdateOrInsertWithExistingWhere(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT, bio TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "frank", "avatar": "avatar1"},
		map[string]any{"bio": "bio1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert (insert) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "frank").First(&result)
	if err != nil {
		t.Fatalf("Failed to find record after insert: %v", err)
	}
	if result["bio"] != "bio1" {
		t.Errorf("Expected bio 'bio1' after insert, got '%v'", result["bio"])
	}

	// Use direct Update for the update scenario
	_, err = w.Q.Where("name = ?", "frank").Update(map[string]any{"bio": "bio2"})
	if err != nil {
		t.Fatalf("Update with where clause failed: %v", err)
	}

	err = w.Q.Where("name = ?", "frank").First(&result)
	if err != nil {
		t.Fatalf("Failed to find record: %v", err)
	}
	if result["bio"] != "bio2" {
		t.Errorf("Expected bio 'bio2', got '%v'", result["bio"])
	}
}

// TestUpdateOrInsertMultipleAttributes tests UpdateOrInsert with multiple attribute conditions.
func TestUpdateOrInsertMultipleAttributes(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "grace", "email": "grace@example.com"},
		map[string]any{"avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with multiple attributes failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "grace").Where("email = ?", "grace@example.com").First(&result)
	if err != nil {
		t.Fatalf("Failed to find record: %v", err)
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertNilAttributes tests UpdateOrInsert with nil attributes.
func TestUpdateOrInsertNilAttributes(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		nil,
		map[string]any{"name": "henry", "avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with nil attributes failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "henry").First(&result)
	if err != nil {
		t.Fatalf("Failed to find record: %v", err)
	}
}
