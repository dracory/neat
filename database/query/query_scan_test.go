package query_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dracory/neat/database/query"
)

// --- db tag ---

type dbTagModel struct {
	MyCol string `db:"my_col"`
}

func TestScanRowsByDbTag(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_db_tag (my_col TEXT)")
	execSQL(t, w, "INSERT INTO test_db_tag VALUES ('hello')")

	w.SetTable("test_db_tag")
	var result dbTagModel
	if err := w.Q.Find(&result); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if result.MyCol != "hello" {
		t.Errorf("expected MyCol='hello', got %q", result.MyCol)
	}
}

// --- neat tag ---

type neatTagModel struct {
	MyCol string `neat:"my_col"`
}

func TestScanRowsByNeatTag(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_neat_tag (my_col TEXT)")
	execSQL(t, w, "INSERT INTO test_neat_tag VALUES ('world')")

	w.SetTable("test_neat_tag")
	var result neatTagModel
	if err := w.Q.Find(&result); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if result.MyCol != "world" {
		t.Errorf("expected MyCol='world', got %q", result.MyCol)
	}
}

// --- snake_case fallback (no tag) ---

type snakeCaseModel struct {
	UserName string
}

func TestScanRowsBySnakeCase(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_snake (user_name TEXT)")
	execSQL(t, w, "INSERT INTO test_snake VALUES ('snake')")

	w.SetTable("test_snake")
	var result snakeCaseModel
	if err := w.Q.Find(&result); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if result.UserName != "snake" {
		t.Errorf("expected UserName='snake', got %q", result.UserName)
	}
}

// --- extra (unmatched) columns don't panic ---

type narrowModel struct {
	Name string `db:"name"`
}

func TestScanRowsUnmatchedColumnIgnored(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_wide (name TEXT, extra TEXT, another INTEGER)")
	execSQL(t, w, "INSERT INTO test_wide VALUES ('alice', 'ignored', 42)")

	w.SetTable("test_wide")
	var result narrowModel
	if err := w.Q.Find(&result); err != nil {
		t.Fatalf("Find should not error on extra columns: %v", err)
	}
	if result.Name != "alice" {
		t.Errorf("expected Name='alice', got %q", result.Name)
	}
}

// --- slice scan ---

type rowModel struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func TestScanRowsIntoSlice(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE test_slice (id INTEGER, name TEXT)")
	execSQL(t, w, "INSERT INTO test_slice VALUES (1,'a'),(2,'b'),(3,'c')")

	w.SetTable("test_slice")
	var results []rowModel
	if err := w.Q.Find(&results); err != nil {
		t.Fatalf("Find into slice failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}
	if results[0].ID != 1 || results[0].Name != "a" {
		t.Errorf("unexpected first row: %+v", results[0])
	}
	if results[2].Name != "c" {
		t.Errorf("unexpected last row: %+v", results[2])
	}
}

// --- structFieldColumnName resolution order: db > neat > gorm > snake_case ---

func TestStructFieldColumnNameDbTagPriority(t *testing.T) {
	type m struct {
		F string `db:"db_col" neat:"neat_col" gorm:"column:gorm_col"`
	}
	col := query.StructFieldColumnName(reflect.TypeOf(m{}).Field(0))
	if col != "db_col" {
		t.Errorf("expected db tag to win, got %q", col)
	}
}

func TestStructFieldColumnNameNeatTagFallback(t *testing.T) {
	type m struct {
		F string `neat:"neat_col" gorm:"column:gorm_col"`
	}
	col := query.StructFieldColumnName(reflect.TypeOf(m{}).Field(0))
	if col != "neat_col" {
		t.Errorf("expected neat tag fallback, got %q", col)
	}
}

func TestStructFieldColumnNameGormTagFallback(t *testing.T) {
	type m struct {
		F string `gorm:"column:gorm_col"`
	}
	col := query.StructFieldColumnName(reflect.TypeOf(m{}).Field(0))
	if col != "gorm_col" {
		t.Errorf("expected gorm tag fallback, got %q", col)
	}
}

func TestStructFieldColumnNameSnakeCaseFallback(t *testing.T) {
	type m struct {
		MyFieldName string
	}
	col := query.StructFieldColumnName(reflect.TypeOf(m{}).Field(0))
	if !strings.Contains(col, "my_field_name") {
		t.Errorf("expected snake_case fallback, got %q", col)
	}
}

// --- cursor tests ---

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
