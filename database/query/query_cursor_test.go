package query_test

import (
	"testing"
)

func TestCursorBasic(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_cursor VALUES (1,'alice'),(2,'bob'),(3,'charlie')")

	w.SetTable("test_cursor")
	cursorChan, err := w.Q.Cursor()
	if err != nil {
		t.Fatalf("Cursor failed: %v", err)
	}

	count := 0
	for cursor := range cursorChan {
		count++
		if cursor == nil {
			t.Error("Expected non-nil cursor")
		}
	}

	if count != 3 {
		t.Errorf("Expected 3 cursor items, got %d", count)
	}
}

func TestCursorChannelCreationAndConsumption(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor_stream (id INTEGER, value TEXT)")
	execSQL(t, w, "INSERT INTO test_cursor_stream VALUES (1,'one'),(2,'two'),(3,'three'),(4,'four'),(5,'five')")

	w.SetTable("test_cursor_stream")
	cursorChan, err := w.Q.Cursor()
	if err != nil {
		t.Fatalf("Cursor failed: %v", err)
	}

	if cursorChan == nil {
		t.Fatal("Expected non-nil cursor channel")
	}

	results := make([]map[string]any, 0)
	for cursor := range cursorChan {
		if cursor == nil {
			t.Error("Expected non-nil cursor")
			continue
		}

		var result map[string]any
		if err := cursor.Scan(&result); err != nil {
			t.Errorf("Cursor.Scan failed: %v", err)
		}
		results = append(results, result)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

func TestCursorErrorHandling(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor_error (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_cursor_error VALUES (1,'test')")

	w.SetTable("test_cursor_error")
	cursorChan, err := w.Q.Cursor()
	if err != nil {
		t.Fatalf("Cursor failed: %v", err)
	}

	count := 0
	for cursor := range cursorChan {
		count++
		if cursor == nil {
			t.Error("Expected non-nil cursor")
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 cursor item, got %d", count)
	}
}

func TestCursorWithTransactions(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor_tx (id INTEGER, name TEXT)")

	execSQL(t, w, "INSERT INTO test_cursor_tx VALUES (1,'tx1'),(2,'tx2')")
	w.SetTable("test_cursor_tx")

	tx, err := w.Q.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	cursorChan, err := tx.Cursor()
	if err != nil {
		t.Fatalf("Cursor in transaction failed: %v", err)
	}

	count := 0
	for cursor := range cursorChan {
		count++
		if cursor == nil {
			t.Error("Expected non-nil cursor in transaction")
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 cursor items in transaction, got %d", count)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

func TestCursorWithWhereClauses(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_cursor_where (id INTEGER, name TEXT, status TEXT)")
	execSQL(t, w, "INSERT INTO test_cursor_where VALUES (1,'alice','active'),(2,'bob','inactive'),(3,'charlie','active')")

	w.SetTable("test_cursor_where")
	w.Q.Where("status = ?", "active")

	cursorChan, err := w.Q.Cursor()
	if err != nil {
		t.Fatalf("Cursor with WHERE failed: %v", err)
	}

	count := 0
	for cursor := range cursorChan {
		count++
		if cursor == nil {
			t.Error("Expected non-nil cursor with WHERE")
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 cursor items with WHERE clause, got %d", count)
	}
}
