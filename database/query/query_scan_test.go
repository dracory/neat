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

func TestScanBytesToNormalizedString(t *testing.T) {
	w := openSQLiteQuery(t)
	// Create table with a BLOB column so SQLite returns []byte
	execSQL(t, w, "CREATE TABLE test_blob (val BLOB)")
	execSQL(t, w, "INSERT INTO test_blob VALUES (X'68656c6c6f')") // "hello" in hex
	execSQL(t, w, "INSERT INTO test_blob VALUES (NULL)")          // NULL row
	w.SetTable("test_blob")

	baseQ := w.Q.Clone().(*query.Query)

	// 1. Test []map[string]any via slice of interface{} path
	{
		q := baseQ.Clone().(*query.Query)
		var results []any
		if err := q.OrderBy("val", "desc").Get(&results); err != nil {
			t.Fatalf("Get with []any failed: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		// First row: hello
		m0, ok := results[0].(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", results[0])
		}
		val0, ok := m0["val"]
		if !ok {
			t.Fatal("expected key 'val' in map")
		}
		if s, ok := val0.(string); !ok || s != "hello" {
			t.Errorf("expected string 'hello', got %T (%v)", val0, val0)
		}
		// Second row: NULL
		m1, ok := results[1].(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", results[1])
		}
		val1 := m1["val"]
		if val1 != nil {
			t.Errorf("expected val to be nil for NULL column, got %T (%v)", val1, val1)
		}
	}

	// 2. Test slice of map destination (reflect.Slice elem.Kind() == reflect.Map)
	{
		q := baseQ.Clone().(*query.Query)
		var results []map[string]any
		if err := q.OrderBy("val", "desc").Get(&results); err != nil {
			t.Fatalf("Get with []map[string]any failed: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		// Row 1
		val0 := results[0]["val"]
		if s, ok := val0.(string); !ok || s != "hello" {
			t.Errorf("expected string 'hello' in []map, got %T (%v)", val0, val0)
		}
		// Row 2
		val1 := results[1]["val"]
		if val1 != nil {
			t.Errorf("expected nil for NULL column in []map, got %T (%v)", val1, val1)
		}
	}

	// 3. Test single map destination (*map[string]any)
	{
		q := baseQ.Clone().(*query.Query)
		q.EnableDebug()
		var result map[string]any
		if err := q.WhereNotNull("val").First(&result); err != nil {
			t.Fatalf("First with map[string]any failed: %v", err)
		}
		val := result["val"]
		if s, ok := val.(string); !ok || s != "hello" {
			t.Errorf("expected string 'hello' in single map, got %T (%v)", val, val)
		}

		// Test ErrNoRows for non-existent row on single map scan
		q2 := baseQ.Clone().(*query.Query)
		var noResult map[string]any
		err := q2.Where("val", "non-existent").First(&noResult)
		if err == nil {
			t.Fatal("expected First with map to return ErrNoRows, got nil")
		}
	}

	// 4. Test Chunk() callback with map elements
	{
		q := baseQ.Clone().(*query.Query)
		var chunkVals []any
		err := q.OrderBy("val", "desc").Chunk(1, func(chunk []map[string]any) error {
			for _, m := range chunk {
				chunkVals = append(chunkVals, m["val"])
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Chunk failed: %v", err)
		}
		if len(chunkVals) != 2 {
			t.Fatalf("expected 2 results, got %d", len(chunkVals))
		}
		if s, ok := chunkVals[0].(string); !ok || s != "hello" {
			t.Errorf("Chunk expected 'hello', got %T (%v)", chunkVals[0], chunkVals[0])
		}
		if chunkVals[1] != nil {
			t.Errorf("Chunk expected nil for NULL, got %T (%v)", chunkVals[1], chunkVals[1])
		}
	}

	// 5. Test Pluck() map destination
	{
		q := baseQ.Clone().(*query.Query)
		var results []map[string]any
		if err := q.OrderBy("val", "desc").Pluck("val", &results); err != nil {
			t.Fatalf("Pluck with []map[string]any failed: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results from Pluck, got %d", len(results))
		}
		// Row 1
		val0 := results[0]["val"]
		if s, ok := val0.(string); !ok || s != "hello" {
			t.Errorf("expected string 'hello' in Pluck map, got %T (%v)", val0, val0)
		}
		// Row 2
		val1 := results[1]["val"]
		if val1 != nil {
			t.Errorf("expected nil for NULL in Pluck map, got %T (%v)", val1, val1)
		}
	}

	// 6. Test Cursor() map destination
	{
		q := baseQ.Clone().(*query.Query)
		cursorChan, err := q.OrderBy("val", "desc").Cursor()
		if err != nil {
			t.Fatalf("Cursor failed: %v", err)
		}
		count := 0
		cursorVals := make([]any, 0)
		for cursor := range cursorChan {
			count++
			if cursor == nil {
				t.Error("Expected non-nil cursor")
				continue
			}
			result := make(map[string]any)
			if err := cursor.Scan(&result); err != nil {
				t.Fatalf("Cursor.Scan into map failed: %v", err)
			}
			cursorVals = append(cursorVals, result["val"])
		}
		if count != 2 {
			t.Fatalf("expected 2 cursor items, got %d", count)
		}
		if s, ok := cursorVals[0].(string); !ok || s != "hello" {
			t.Errorf("expected string 'hello' in Cursor map, got %T (%v)", cursorVals[0], cursorVals[0])
		}
		if cursorVals[1] != nil {
			t.Errorf("expected nil for NULL in Cursor map, got %T (%v)", cursorVals[1], cursorVals[1])
		}
	}

	// 7. Test FirstAsVar map destination
	{
		q := baseQ.Clone().(*query.Query)
		res, err := q.WhereNotNull("val").FirstAsVar()
		if err != nil {
			t.Fatalf("FirstAsVar failed: %v", err)
		}
		m, ok := res.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any from FirstAsVar, got %T", res)
		}
		val := m["val"]
		if s, ok := val.(string); !ok || s != "hello" {
			t.Errorf("expected string 'hello' from FirstAsVar map, got %T (%v)", val, val)
		}
	}
}
