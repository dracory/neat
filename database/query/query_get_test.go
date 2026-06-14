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

	results := make([]User, 0)
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

	results := make([]User, 0)
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

func TestAll(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_all (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_all VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_all VALUES (2, 'Bob')")

	w.SetTable("test_all")

	type User struct {
		ID   int
		Name string
	}

	results := make([]User, 0)
	if err := w.Q.All(&results); err != nil {
		t.Fatalf("All failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestAllWithFilter(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_all_filter (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_all_filter VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_all_filter VALUES (2, 'Bob')")

	w.SetTable("test_all_filter")
	w.Q.Filter("name = ?", "Bob")

	type User struct {
		ID   int
		Name string
	}

	results := make([]User, 0)
	if err := w.Q.All(&results); err != nil {
		t.Fatalf("All with Filter failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Bob" {
		t.Errorf("expected name 'Bob', got %s", results[0].Name)
	}
}

func TestFindAll(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_find_all (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_find_all VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_find_all VALUES (2, 'Bob')")

	w.SetTable("test_find_all")

	type User struct {
		ID   int
		Name string
	}

	results := make([]User, 0)
	if err := w.Q.FindAll(&results); err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestFindAllAsAliasForAll(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_find_all_alias (id INTEGER PRIMARY KEY, name TEXT)")
	execSQL(t, w, "INSERT INTO test_find_all_alias VALUES (1, 'Alice')")
	execSQL(t, w, "INSERT INTO test_find_all_alias VALUES (2, 'Bob')")

	w.SetTable("test_find_all_alias")

	type User struct {
		ID   int
		Name string
	}

	results1 := make([]User, 0)
	results2 := make([]User, 0)

	if err := w.Q.All(&results1); err != nil {
		t.Fatalf("All failed: %v", err)
	}

	w.SetTable("test_find_all_alias")
	if err := w.Q.FindAll(&results2); err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(results1) != len(results2) {
		t.Errorf("FindAll and All should return same number of results. Got %d vs %d", len(results1), len(results2))
	}
}
