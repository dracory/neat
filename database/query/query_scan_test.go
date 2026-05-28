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
