package query_test

import (
	"testing"
)

func TestGet(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_get (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_get VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_get VALUES (2, 'Bob')")

	w.SetTable("test_get")

	type User struct {
		ID   int
		Name string
	}

	var results []User
	if err := w.Q.Get(&results); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestGetWithWhere(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_get_where (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_get_where VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_get_where VALUES (2, 'Bob')")

	w.SetTable("test_get_where")
	w.Q.Where("name = ?", "Bob")

	type User struct {
		ID   int
		Name string
	}

	var results []User
	if err := w.Q.Get(&results); err != nil {
		t.Fatalf("Get with Where failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Bob" {
		t.Errorf("expected name 'Bob', got %s", results[0].Name)
	}
}
