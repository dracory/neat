package query_test

import (
	"testing"
)

func TestSaveInsert(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_save (id INTEGER PRIMARY KEY, name TEXT)")
	w.SetTable("test_save")

	type User struct {
		ID   int
		Name string
	}

	user := &User{Name: "Alice"}
	if err := w.Q.Save(user); err != nil {
		t.Fatalf("Save (insert) failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected ID to be set after insert")
	}
}

func TestSaveUpdate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_save_update (id INTEGER PRIMARY KEY, name TEXT)")
	w.SetTable("test_save_update")

	type User struct {
		ID   int
		Name string
	}

	// First insert
	user := &User{Name: "Alice"}
	if err := w.Q.Save(user); err != nil {
		t.Fatalf("Save (insert) failed: %v", err)
	}

	// Then update
	user.Name = "Bob"
	if err := w.Q.Save(user); err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	// Verify update
	var result User
	if err := w.Q.Where("id = ?", user.ID).First(&result); err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}
	if result.Name != "Bob" {
		t.Errorf("expected name 'Bob', got %s", result.Name)
	}
}

func TestSaveQuietly(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_save_quietly (id INTEGER PRIMARY KEY, name TEXT)")
	w.SetTable("test_save_quietly")

	type User struct {
		ID   int
		Name string
	}

	user := &User{Name: "Alice"}
	if err := w.Q.SaveQuietly(user); err != nil {
		t.Fatalf("SaveQuietly failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected ID to be set after insert")
	}
}
