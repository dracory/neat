package query_test

import (
	"testing"
)

func TestFirst(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_first VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_first VALUES (2, 'Bob')")

	w.SetTable("test_first")

	type User struct {
		ID   int
		Name string
	}

	var result User
	if err := w.Q.First(&result); err != nil {
		t.Fatalf("First failed: %v", err)
	}
	if result.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %s", result.Name)
	}
}

func TestFirstWithWhere(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_where (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_first_where VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_first_where VALUES (2, 'Bob')")

	w.SetTable("test_first_where")
	w.Q.Where("name = ?", "Bob")

	type User struct {
		ID   int
		Name string
	}

	var result User
	if err := w.Q.First(&result); err != nil {
		t.Fatalf("First with Where failed: %v", err)
	}
	if result.Name != "Bob" {
		t.Errorf("expected name 'Bob', got %s", result.Name)
	}
}

func TestFirstOrFail(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_fail (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_first_or_fail VALUES (1, 'Alice')")

	w.SetTable("test_first_or_fail")

	type User struct {
		ID   int
		Name string
	}

	var result User
	if err := w.Q.FirstOrFail(&result); err != nil {
		t.Fatalf("FirstOrFail failed: %v", err)
	}
	if result.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %s", result.Name)
	}
}

func TestFirstOrFailNotFound(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_first_or_fail_not_found (id INTEGER PRIMARY KEY, name TEXT)")

	w.SetTable("test_first_or_fail_not_found")

	type User struct {
		ID   int
		Name string
	}

	var result User
	err := w.Q.FirstOrFail(&result)
	if err == nil {
		t.Error("expected error for not found")
	}
}
