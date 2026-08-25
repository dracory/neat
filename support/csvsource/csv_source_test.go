package csvsource

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
	"time"
)

// writeTempCSV writes content to a temporary file and returns its path.
// The caller is responsible for removing the file via t.Cleanup.
func writeTempCSV(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp CSV: %v", err)
	}
	return path
}

// --- NewCsvSource (string) tests ---

func TestNewCsvSource_String_Basic(t *testing.T) {
	csvString := "id,name,email,active\n" +
		"1,Alice,alice@example.com,true\n" +
		"2,Bob,bob@example.com,false\n" +
		"3,Charlie,charlie@example.com,true\n"

	model := NewCsvSource(csvString, "users")

	if model.TableName() != "users" {
		t.Errorf("expected table name 'users', got '%s'", model.TableName())
	}

	rows, err := model.Rows()
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if rows[0]["id"] != int64(1) {
		t.Errorf("expected id=1 (int64), got %v (%T)", rows[0]["id"], rows[0]["id"])
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
	if rows[0]["active"] != int64(1) {
		t.Errorf("expected active=1 (bool→int64), got %v (%T)", rows[0]["active"], rows[0]["active"])
	}
}

func TestNewCsvSource_String_TypeInference(t *testing.T) {
	csvString := "id,price,name,created,active\n" +
		"1,19.99,Widget,2024-01-15T10:30:00Z,true\n" +
		"2,29.99,Gadget,2024-02-20T14:45:00Z,false\n"

	model := NewCsvSource(csvString, "products")

	rows, _ := model.Rows()
	row := rows[0]

	if _, ok := row["id"].(int64); !ok {
		t.Errorf("expected id to be int64, got %T (%v)", row["id"], row["id"])
	}
	if f, ok := row["price"].(float64); !ok || f != 19.99 {
		t.Errorf("expected price=19.99 (float64), got %T (%v)", row["price"], row["price"])
	}
	if s, ok := row["name"].(string); !ok || s != "Widget" {
		t.Errorf("expected name='Widget' (string), got %T (%v)", row["name"], row["name"])
	}
	if _, ok := row["created"].(time.Time); !ok {
		t.Errorf("expected created to be time.Time, got %T (%v)", row["created"], row["created"])
	}
	if v, ok := row["active"].(int64); !ok || v != 1 {
		t.Errorf("expected active=1 (int64), got %T (%v)", row["active"], row["active"])
	}
}

func TestNewCsvSource_String_TypeWidening(t *testing.T) {
	csvString := "value\n1\n2.5\n3\n"
	model := NewCsvSource(csvString, "mixed")

	rows, _ := model.Rows()
	if f, ok := rows[0]["value"].(float64); !ok || f != 1.0 {
		t.Errorf("expected value=1.0 (float64 after widening), got %T (%v)", rows[0]["value"], rows[0]["value"])
	}
	if f, ok := rows[1]["value"].(float64); !ok || f != 2.5 {
		t.Errorf("expected value=2.5 (float64), got %T (%v)", rows[1]["value"], rows[1]["value"])
	}
}

func TestNewCsvSource_String_EmptyCells(t *testing.T) {
	csvString := "id,name,email\n1,Alice,\n2,,bob@example.com\n"
	model := NewCsvSource(csvString, "users")

	rows, _ := model.Rows()
	if rows[0]["email"] != nil {
		t.Errorf("expected nil for empty email, got %v", rows[0]["email"])
	}
	if rows[1]["name"] != nil {
		t.Errorf("expected nil for empty name, got %v", rows[1]["name"])
	}
}

func TestNewCsvSource_String_HeaderOnly(t *testing.T) {
	csvString := "id,name,email\n"
	model := NewCsvSource(csvString, "users")

	if model.TableName() != "users" {
		t.Errorf("expected table name 'users', got '%s'", model.TableName())
	}

	rows, _ := model.Rows()
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for header-only CSV, got %d", len(rows))
	}

	schema := model.Schema()
	if schema == nil {
		t.Fatal("expected schema to be set for header-only CSV")
	}
	if schema["id"] != "string" {
		t.Errorf("expected schema['id']='string' (no data to infer from), got '%s'", schema["id"])
	}
}

func TestNewCsvSource_String_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on empty string")
		}
	}()
	NewCsvSource("", "users")
}

// --- NewCsvFileSource (file path) tests ---

func TestNewCsvFileSource_Basic(t *testing.T) {
	csvContent := "id,name,email,active\n" +
		"1,Alice,alice@example.com,true\n" +
		"2,Bob,bob@example.com,false\n" +
		"3,Charlie,charlie@example.com,true\n"

	path := writeTempCSV(t, "users.csv", csvContent)
	model := NewCsvFileSource(path)

	if model.TableName() != "users" {
		t.Errorf("expected table name 'users', got '%s'", model.TableName())
	}

	rows, err := model.Rows()
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if rows[0]["id"] != int64(1) {
		t.Errorf("expected id=1 (int64), got %v (%T)", rows[0]["id"], rows[0]["id"])
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
}

func TestNewCsvFileSource_TableNameFromPath(t *testing.T) {
	csvContent := "id,name\n1,Alice\n"
	path := writeTempCSV(t, "my_table.csv", csvContent)
	model := NewCsvFileSource(path)

	if model.TableName() != "my_table" {
		t.Errorf("expected table name 'my_table', got '%s'", model.TableName())
	}
}

func TestNewCsvFileSource_DateFormats(t *testing.T) {
	csvContent := "date_only,datetime,us_date\n" +
		"2024-01-15,2024-01-15 10:30:00,01/15/2024\n"
	path := writeTempCSV(t, "dates.csv", csvContent)
	model := NewCsvFileSource(path)

	rows, _ := model.Rows()
	row := rows[0]

	if _, ok := row["date_only"].(time.Time); !ok {
		t.Errorf("expected date_only to be time.Time, got %T (%v)", row["date_only"], row["date_only"])
	}
	if _, ok := row["datetime"].(time.Time); !ok {
		t.Errorf("expected datetime to be time.Time, got %T (%v)", row["datetime"], row["datetime"])
	}
	if _, ok := row["us_date"].(time.Time); !ok {
		t.Errorf("expected us_date to be time.Time, got %T (%v)", row["us_date"], row["us_date"])
	}
}

func TestNewCsvFileSource_WithDelimiter(t *testing.T) {
	tsvContent := "id\tname\tprice\n1\tWidget\t19.99\n2\tGadget\t29.99\n"
	path := writeTempCSV(t, "products.tsv", tsvContent)
	model := NewCsvFileSourceWithDelimiter(path, '\t')

	if model.TableName() != "products" {
		t.Errorf("expected table name 'products', got '%s'", model.TableName())
	}

	rows, _ := model.Rows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "Widget" {
		t.Errorf("expected name='Widget', got %v", rows[0]["name"])
	}
	if f, ok := rows[0]["price"].(float64); !ok || f != 19.99 {
		t.Errorf("expected price=19.99, got %v", rows[0]["price"])
	}
}

func TestNewCsvFileSource_PanicsOnNonExistentFile(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on non-existent file")
		}
	}()
	NewCsvFileSource("/nonexistent/path/to/file.csv")
}

func TestNewCsvFileSource_PanicsOnEmptyFile(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on empty file")
		}
	}()
	path := writeTempCSV(t, "empty.csv", "")
	NewCsvFileSource(path)
}

// --- NewCsvFSSource (fstest.MapFS) tests ---

func TestNewCsvFSSource_Basic(t *testing.T) {
	sys := fstest.MapFS{
		"data/users.csv": &fstest.MapFile{
			Data: []byte("id,name,active\n1,Alice,true\n2,Bob,false\n"),
		},
	}

	model := NewCsvFSSource(sys, "data/users.csv")
	if model.TableName() != "users" {
		t.Errorf("expected table name 'users', got '%s'", model.TableName())
	}

	rows, err := model.Rows()
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
}

func TestNewCsvFSSource_WithDelimiter(t *testing.T) {
	sys := fstest.MapFS{
		"events.tsv": &fstest.MapFile{
			Data: []byte("id\tevent\n100\tlogin\n"),
		},
	}

	model := NewCsvFSSourceWithDelimiter(sys, "events.tsv", '\t')
	if model.TableName() != "events" {
		t.Errorf("expected table name 'events', got '%s'", model.TableName())
	}

	rows, _ := model.Rows()
	if len(rows) != 1 || rows[0]["event"] != "login" {
		t.Errorf("unexpected rows: %v", rows)
	}
}

// --- Internal function tests ---

func TestInferValueType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123", "INTEGER"},
		{"-456", "INTEGER"},
		{"19.99", "REAL"},
		{"-0.5", "REAL"},
		{"true", "INTEGER"},
		{"false", "INTEGER"},
		{"True", "INTEGER"},
		{"False", "INTEGER"},
		{"2024-01-15T10:30:00Z", "DATETIME"},
		{"2024-01-15", "DATETIME"},
		{"2024-01-15 10:30:00", "DATETIME"},
		{"01/15/2024", "DATETIME"},
		{"hello", "TEXT"},
		{"", "TEXT"},
		{"123abc", "TEXT"},
	}

	for _, tt := range tests {
		got := inferValueType(tt.input)
		if got != tt.want {
			t.Errorf("inferValueType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWidenType(t *testing.T) {
	tests := []struct {
		current, new, want string
	}{
		{"INTEGER", "INTEGER", "INTEGER"},
		{"INTEGER", "REAL", "REAL"},
		{"REAL", "INTEGER", "REAL"},
		{"REAL", "REAL", "REAL"},
		{"INTEGER", "TEXT", "TEXT"},
		{"TEXT", "INTEGER", "TEXT"},
		{"DATETIME", "DATETIME", "DATETIME"},
		{"DATETIME", "INTEGER", "TEXT"},
		{"INTEGER", "DATETIME", "TEXT"},
		{"", "INTEGER", "INTEGER"},
		{"", "DATETIME", "DATETIME"},
		{"", "TEXT", "TEXT"},
	}

	for _, tt := range tests {
		got := widenType(tt.current, tt.new)
		if got != tt.want {
			t.Errorf("widenType(%q, %q) = %q, want %q", tt.current, tt.new, got, tt.want)
		}
	}
}

func TestConvertValue(t *testing.T) {
	if v := convertValue("42", "INTEGER"); v != int64(42) {
		t.Errorf("convertValue(\"42\", INTEGER) = %v, want 42", v)
	}
	if v := convertValue("true", "INTEGER"); v != int64(1) {
		t.Errorf("convertValue(\"true\", INTEGER) = %v, want 1", v)
	}
	if v := convertValue("false", "INTEGER"); v != int64(0) {
		t.Errorf("convertValue(\"false\", INTEGER) = %v, want 0", v)
	}
	if v := convertValue("19.99", "REAL"); v != 19.99 {
		t.Errorf("convertValue(\"19.99\", REAL) = %v, want 19.99", v)
	}
	if v, ok := convertValue("2024-01-15T10:30:00Z", "DATETIME").(time.Time); !ok {
		t.Errorf("convertValue(\"2024-01-15T10:30:00Z\", DATETIME) expected time.Time, got %T", v)
	}
	if v := convertValue("hello", "TEXT"); v != "hello" {
		t.Errorf("convertValue(\"hello\", TEXT) = %v, want 'hello'", v)
	}
	if v := convertValue("", "INTEGER"); v != nil {
		t.Errorf("convertValue(\"\", INTEGER) = %v, want nil", v)
	}
	if v := convertValue("", "TEXT"); v != nil {
		t.Errorf("convertValue(\"\", TEXT) = %v, want nil", v)
	}
}

func TestDeriveTableName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"data/users.csv", "users"},
		{"users.csv", "users"},
		{"/home/user/data/products.csv", "products"},
		{"events.tsv", "events"},
	}
	for _, tt := range tests {
		got := deriveTableName(tt.path)
		if got != tt.want {
			t.Errorf("deriveTableName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}

	// Windows-style paths only work on Windows (filepath.Base uses OS-specific separators)
	if runtime.GOOS == "windows" {
		got := deriveTableName("C:\\data\\orders.csv")
		if got != "orders" {
			t.Errorf("deriveTableName(\"C:\\\\data\\\\orders.csv\") = %q, want \"orders\"", got)
		}
	}
}
