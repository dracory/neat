package query_test

import (
	"testing"
)

func TestFind(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_find (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_find VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_find VALUES (2, 'Bob')")

	w.SetTable("test_find")

	type User struct {
		ID   int
		Name string
	}

	results := make([]User, 0)
	if err := w.Q.Find(&results); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestFindWithConditions(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_find_cond (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_find_cond VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_find_cond VALUES (2, 'Bob')")

	w.SetTable("test_find_cond")
	w.Q.Where("id = ?", 1)

	type User struct {
		ID   int
		Name string
	}

	results := make([]User, 0)
	if err := w.Q.Find(&results); err != nil {
		t.Fatalf("Find with conditions failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Alice" {
		t.Errorf("expected name 'Alice', got %s", results[0].Name)
	}
}

func TestFindOrFail(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_find_or_fail (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_find_or_fail VALUES (1, 'Alice')")

	w.SetTable("test_find_or_fail")

	type User struct {
		ID   int
		Name string
	}

	results := make([]User, 0)
	if err := w.Q.FindOrFail(&results); err != nil {
		t.Fatalf("FindOrFail failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestFindOrFailEmpty(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_find_or_fail_empty (id INTEGER PRIMARY KEY, name TEXT)")

	w.SetTable("test_find_or_fail_empty")

	type User struct {
		ID   int
		Name string
	}

	results := make([]User, 0)
	err := w.Q.FindOrFail(&results)
	if err == nil {
		t.Error("expected error for empty results")
	}
}
