package cursor

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec("CREATE TABLE t (id INTEGER, name TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec("INSERT INTO t VALUES (1,'alice'),(2,'bob'),(3,'carol')"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	return db
}

func queryRows(t *testing.T, db *sql.DB) *sql.Rows {
	t.Helper()
	rows, err := db.Query("SELECT id, name FROM t ORDER BY id")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	return rows
}

func TestNewCursor(t *testing.T) {
	db := openTestDB(t)
	rows := queryRows(t, db)
	c := NewCursor(rows)
	if c == nil {
		t.Fatal("NewCursor returned nil")
	}
	if c.closed {
		t.Error("new cursor should not be closed")
	}
	if c.err != nil {
		t.Error("new cursor should have nil err")
	}
	_ = c.Close()
}

func TestCursorColumns(t *testing.T) {
	db := openTestDB(t)
	rows := queryRows(t, db)
	c := NewCursor(rows)
	defer func() { _ = c.Close() }()

	cols, err := c.Columns()
	if err != nil {
		t.Fatalf("Columns: %v", err)
	}
	if len(cols) != 2 || cols[0] != "id" || cols[1] != "name" {
		t.Errorf("unexpected columns: %v", cols)
	}
}

func TestCursorIteration(t *testing.T) {
	db := openTestDB(t)
	// Single-column query: Cursor.Scan delegates to rows.Scan with one dest argument.
	rows, err := db.Query("SELECT id FROM t ORDER BY id")
	if err != nil {
		t.Fatal(err)
	}
	c := NewCursor(rows)
	defer func() { _ = c.Close() }()

	var ids []int
	for c.Next() {
		var id int
		if err := c.Scan(&id); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		ids = append(ids, id)
	}
	if err := c.Err(); err != nil {
		t.Errorf("Err after iteration: %v", err)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Errorf("unexpected ids: %v", ids)
	}
}

func TestCursorClose(t *testing.T) {
	db := openTestDB(t)
	rows := queryRows(t, db)
	c := NewCursor(rows)

	if err := c.Close(); err != nil {
		t.Errorf("first Close: %v", err)
	}
	if !c.closed {
		t.Error("cursor should be closed after Close")
	}
	if err := c.Close(); err != nil {
		t.Errorf("second Close (idempotent) should not error: %v", err)
	}
}

func TestCursorNextAfterClose(t *testing.T) {
	db := openTestDB(t)
	rows := queryRows(t, db)
	c := NewCursor(rows)
	_ = c.Close()

	if c.Next() {
		t.Error("Next on closed cursor should return false")
	}
}

func TestCursorScanAfterClose(t *testing.T) {
	db := openTestDB(t)
	rows := queryRows(t, db)
	c := NewCursor(rows)
	_ = c.Close()

	var id int
	err := c.Scan(&id)
	if err == nil {
		t.Error("Scan on closed cursor should return error")
	}
}

func TestCursorColumnsAfterClose(t *testing.T) {
	db := openTestDB(t)
	rows := queryRows(t, db)
	c := NewCursor(rows)
	_ = c.Close()

	_, err := c.Columns()
	if err == nil {
		t.Error("Columns on closed cursor should return error")
	}
}

func TestCursorErrWithPriorError(t *testing.T) {
	db := openTestDB(t)
	rows := queryRows(t, db)
	c := NewCursor(rows)
	defer func() { _ = c.Close() }()

	sentinel := sql.ErrNoRows
	c.err = sentinel

	if c.Err() != sentinel {
		t.Errorf("expected sentinel error, got %v", c.Err())
	}
	if c.Next() {
		t.Error("Next with prior error should return false")
	}
	if err := c.Scan(nil); err != sentinel {
		t.Errorf("Scan with prior error should return sentinel, got %v", err)
	}
}

func TestCursorEmptyTable(t *testing.T) {
	db := openTestDB(t)
	rows, err := db.Query("SELECT id, name FROM t WHERE 1=0")
	if err != nil {
		t.Fatal(err)
	}
	c := NewCursor(rows)
	defer func() { _ = c.Close() }()

	if c.Next() {
		t.Error("Next on empty result set should return false")
	}
	if err := c.Err(); err != nil {
		t.Errorf("Err on empty result: %v", err)
	}
}
