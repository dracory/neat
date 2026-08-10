package driver

import (
	"testing"
	"time"
)

func TestParseCSVBasic(t *testing.T) {
	dir := mkdirTemp(t)
	path := writeCSVFile(t, dir, "basic.csv", "id,name\n1,Alice\n2,Bob\n")

	columns, rows, err := parseCSV(path)
	if err != nil {
		t.Fatalf("parseCSV failed: %v", err)
	}
	if len(columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(columns))
	}
	if columns[0] != "id" || columns[1] != "name" {
		t.Errorf("columns = %v, want [id name]", columns)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "1" || rows[0][1] != "Alice" {
		t.Errorf("row 0 = %v, want [1 Alice]", rows[0])
	}
	if rows[1][0] != "2" || rows[1][1] != "Bob" {
		t.Errorf("row 1 = %v, want [2 Bob]", rows[1])
	}
}

func TestParseCSVEmptyFile(t *testing.T) {
	dir := mkdirTemp(t)
	path := writeCSVFile(t, dir, "empty.csv", "")

	_, _, err := parseCSV(path)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

func TestParseCSVHeaderOnly(t *testing.T) {
	dir := mkdirTemp(t)
	path := writeCSVFile(t, dir, "header.csv", "id,name\n")

	columns, rows, err := parseCSV(path)
	if err != nil {
		t.Fatalf("parseCSV failed: %v", err)
	}
	if len(columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(columns))
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestInferValueTypeInt(t *testing.T) {
	if got := inferValueType("123"); got != "INTEGER" {
		t.Errorf("inferValueType(\"123\") = %q, want INTEGER", got)
	}
	if got := inferValueType("-456"); got != "INTEGER" {
		t.Errorf("inferValueType(\"-456\") = %q, want INTEGER", got)
	}
}

func TestInferValueTypeFloat(t *testing.T) {
	if got := inferValueType("19.99"); got != "REAL" {
		t.Errorf("inferValueType(\"19.99\") = %q, want REAL", got)
	}
	if got := inferValueType("-0.5"); got != "REAL" {
		t.Errorf("inferValueType(\"-0.5\") = %q, want REAL", got)
	}
}

func TestInferValueTypeBool(t *testing.T) {
	for _, v := range []string{"true", "false", "True", "False"} {
		if got := inferValueType(v); got != "INTEGER" {
			t.Errorf("inferValueType(%q) = %q, want INTEGER", v, got)
		}
	}
}

func TestInferValueTypeTimeRFC3339(t *testing.T) {
	if got := inferValueType("2024-01-15T10:30:00Z"); got != "DATETIME" {
		t.Errorf("inferValueType(RFC3339) = %q, want DATETIME", got)
	}
}

func TestInferValueTypeTimeDateOnly(t *testing.T) {
	if got := inferValueType("2024-01-15"); got != "DATETIME" {
		t.Errorf("inferValueType(date only) = %q, want DATETIME", got)
	}
}

func TestInferValueTypeString(t *testing.T) {
	if got := inferValueType("hello"); got != "TEXT" {
		t.Errorf("inferValueType(\"hello\") = %q, want TEXT", got)
	}
	if got := inferValueType("abc123"); got != "TEXT" {
		t.Errorf("inferValueType(\"abc123\") = %q, want TEXT", got)
	}
}

func TestInferColumnTypesAllSameType(t *testing.T) {
	columns := []string{"id", "name"}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	types := inferColumnTypes(columns, rows)
	if types[0] != "INTEGER" {
		t.Errorf("col 0 type = %q, want INTEGER", types[0])
	}
	if types[1] != "TEXT" {
		t.Errorf("col 1 type = %q, want TEXT", types[1])
	}
}

func TestInferColumnTypesMixedTypes(t *testing.T) {
	columns := []string{"value"}
	rows := [][]string{
		{"10"},
		{"10.5"},
	}
	types := inferColumnTypes(columns, rows)
	if types[0] != "REAL" {
		t.Errorf("mixed int/float type = %q, want REAL", types[0])
	}
}

func TestInferColumnTypesEmptyCells(t *testing.T) {
	columns := []string{"id", "name"}
	rows := [][]string{
		{"1", ""},
		{"2", "Alice"},
	}
	types := inferColumnTypes(columns, rows)
	// Empty cells should not affect type inference
	if types[0] != "INTEGER" {
		t.Errorf("col 0 type = %q, want INTEGER", types[0])
	}
	if types[1] != "TEXT" {
		t.Errorf("col 1 type = %q, want TEXT", types[1])
	}
}

func TestWidenType(t *testing.T) {
	tests := []struct {
		current, new, want string
	}{
		{"INTEGER", "INTEGER", "INTEGER"},
		{"INTEGER", "REAL", "REAL"},
		{"REAL", "INTEGER", "REAL"},
		{"INTEGER", "TEXT", "TEXT"},
		{"REAL", "TEXT", "TEXT"},
		{"TEXT", "INTEGER", "TEXT"},
		{"DATETIME", "DATETIME", "DATETIME"},
		{"DATETIME", "TEXT", "TEXT"},
		{"TEXT", "DATETIME", "TEXT"},
		// DATETIME mixed with numeric types → TEXT (incompatible)
		{"DATETIME", "INTEGER", "TEXT"},
		{"INTEGER", "DATETIME", "TEXT"},
		{"DATETIME", "REAL", "TEXT"},
		{"REAL", "DATETIME", "TEXT"},
		// BLOB mixed with any other type → BLOB (preserves binary data integrity)
		{"BLOB", "BLOB", "BLOB"},
		{"BLOB", "TEXT", "BLOB"},
		{"TEXT", "BLOB", "BLOB"},
		{"BLOB", "INTEGER", "BLOB"},
		{"INTEGER", "BLOB", "BLOB"},
		{"BLOB", "REAL", "BLOB"},
		{"REAL", "BLOB", "BLOB"},
		{"BLOB", "DATETIME", "BLOB"},
		{"DATETIME", "BLOB", "BLOB"},
	}
	for _, tt := range tests {
		if got := widenType(tt.current, tt.new); got != tt.want {
			t.Errorf("widenType(%q, %q) = %q, want %q", tt.current, tt.new, got, tt.want)
		}
	}
}

func TestConvertValueInt(t *testing.T) {
	v := convertValue("42", "INTEGER")
	n, ok := v.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", v)
	}
	if n != 42 {
		t.Errorf("got %d, want 42", n)
	}
}

func TestConvertValueFloat(t *testing.T) {
	v := convertValue("19.99", "REAL")
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", v)
	}
	if f != 19.99 {
		t.Errorf("got %f, want 19.99", f)
	}
}

func TestConvertValueBoolTrue(t *testing.T) {
	v := convertValue("true", "INTEGER")
	n, ok := v.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", v)
	}
	if n != 1 {
		t.Errorf("got %d, want 1", n)
	}
}

func TestConvertValueBoolFalse(t *testing.T) {
	v := convertValue("false", "INTEGER")
	n, ok := v.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", v)
	}
	if n != 0 {
		t.Errorf("got %d, want 0", n)
	}
}

func TestConvertValueTime(t *testing.T) {
	v := convertValue("2024-01-15T10:30:00Z", "DATETIME")
	_, ok := v.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", v)
	}
}

func TestConvertValueEmptyStringReturnsNil(t *testing.T) {
	v := convertValue("", "INTEGER")
	if v != nil {
		t.Errorf("expected nil for empty string, got %v", v)
	}
}

func TestIsSimpleIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"users", true},
		{"id", true},
		{"user_id", true},
		{"order", true}, // SQL keyword but valid identifier
		{"", false},
		{"1abc", false}, // starts with digit
		{"table.column", false},
		{"func()", false},
		{"name with space", false},
		{"name-dash", false},
		{"naïve", false}, // non-ASCII
	}
	for _, tt := range tests {
		if got := isSimpleIdentifier(tt.input); got != tt.want {
			t.Errorf("isSimpleIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
