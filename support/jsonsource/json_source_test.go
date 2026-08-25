package jsonsource

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// writeTempJSON writes content to a temporary file and returns its path.
func writeTempJSON(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp JSON: %v", err)
	}
	return path
}

// --- NewJsonSource (string) tests ---

func TestNewJsonSource_String_Basic(t *testing.T) {
	content := `[{"id": 1, "name": "Alice", "active": true},` +
		`{"id": 2, "name": "Bob", "active": false},` +
		`{"id": 3, "name": "Charlie", "active": true}]`

	model := NewJsonSource(content, "users", false)

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
	if rows[0]["active"] != true {
		t.Errorf("expected active=true, got %v", rows[0]["active"])
	}
}

func TestNewJsonSource_String_JSONL(t *testing.T) {
	content := `{"id": 1, "type": "login", "user_id": 42}` + "\n" +
		`{"id": 2, "type": "logout", "user_id": 42}` + "\n" +
		`{"id": 3, "type": "purchase", "user_id": 17, "amount": 49.99}`

	model := NewJsonSource(content, "events", true)

	if model.TableName() != "events" {
		t.Errorf("expected table name 'events', got '%s'", model.TableName())
	}

	rows, err := model.Rows()
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if rows[0]["type"] != "login" {
		t.Errorf("expected type='login', got %v", rows[0]["type"])
	}
	if rows[2]["amount"] != 49.99 {
		t.Errorf("expected amount=49.99, got %v", rows[2]["amount"])
	}
}

func TestNewJsonSource_String_JSONLWithEmptyLines(t *testing.T) {
	content := `{"id": 1, "name": "Alice"}` + "\n\n" +
		`{"id": 2, "name": "Bob"}` + "\n"

	model := NewJsonSource(content, "users", true)

	rows, _ := model.Rows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (empty lines skipped), got %d", len(rows))
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected first row 'Alice', got %v", rows[0]["name"])
	}
}

func TestNewJsonSource_String_TimestampDetection(t *testing.T) {
	content := `[{"id": 1, "created": "2024-01-15T10:30:00Z"}]`

	model := NewJsonSource(content, "events", false)

	rows, _ := model.Rows()
	row := rows[0]

	if _, ok := row["created"].(time.Time); !ok {
		t.Errorf("expected created to be time.Time, got %T (%v)", row["created"], row["created"])
	}
}

func TestNewJsonSource_String_NestedObjectAsJSONString(t *testing.T) {
	content := `[{"id": 1, "address": {"city": "NYC", "zip": "10001"}}]`

	model := NewJsonSource(content, "users", false)

	rows, _ := model.Rows()
	row := rows[0]

	addrStr, ok := row["address"].(string)
	if !ok {
		t.Fatalf("expected address to be string (JSON), got %T (%v)", row["address"], row["address"])
	}

	var addr map[string]any
	if err := json.Unmarshal([]byte(addrStr), &addr); err != nil {
		t.Fatalf("expected address to be valid JSON, got parse error: %v", err)
	}
	if addr["city"] != "NYC" {
		t.Errorf("expected city='NYC', got %v", addr["city"])
	}
}

func TestNewJsonSource_String_NestedArrayAsJSONString(t *testing.T) {
	content := `[{"id": 1, "tags": ["red", "blue", "green"]}]`

	model := NewJsonSource(content, "items", false)

	rows, _ := model.Rows()
	row := rows[0]

	tagsStr, ok := row["tags"].(string)
	if !ok {
		t.Fatalf("expected tags to be string (JSON), got %T (%v)", row["tags"], row["tags"])
	}

	var tags []any
	if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
		t.Fatalf("expected tags to be valid JSON array, got parse error: %v", err)
	}
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
}

func TestNewJsonSource_String_EmptyArray(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on empty JSON array")
		}
	}()
	NewJsonSource(`[]`, "empty", false)
}

func TestNewJsonSource_String_PanicsOnInvalidJSON(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on invalid JSON")
		}
	}()
	NewJsonSource(`{not valid json}`, "bad", false)
}

func TestNewJsonSource_String_PanicsOnNonArrayJSON(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on non-array JSON")
		}
	}()
	NewJsonSource(`{"id": 1, "name": "Alice"}`, "obj", false)
}

func TestNewJsonSource_String_PanicsOnTrailingData(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on trailing data after JSON array")
		}
	}()
	NewJsonSource(`[{"id": 1}][1,2]`, "bad", false)
}

// --- NewJsonFileSource (file path) tests ---

func TestNewJsonFileSource_Basic(t *testing.T) {
	content := `[{"id": 1, "name": "Alice", "active": true},` +
		`{"id": 2, "name": "Bob", "active": false},` +
		`{"id": 3, "name": "Charlie", "active": true}]`

	path := writeTempJSON(t, "users.json", content)
	model := NewJsonFileSource(path)

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

	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
}

func TestNewJsonFileSource_JSONL(t *testing.T) {
	content := `{"id": 1, "type": "login", "user_id": 42}` + "\n" +
		`{"id": 2, "type": "logout", "user_id": 42}` + "\n" +
		`{"id": 3, "type": "purchase", "user_id": 17, "amount": 49.99}`

	path := writeTempJSON(t, "events.jsonl", content)
	model := NewJsonFileSource(path)

	if model.TableName() != "events" {
		t.Errorf("expected table name 'events', got '%s'", model.TableName())
	}

	rows, _ := model.Rows()
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[0]["type"] != "login" {
		t.Errorf("expected type='login', got %v", rows[0]["type"])
	}
}

func TestNewJsonFileSource_NDJSONExtension(t *testing.T) {
	content := `{"id": 1, "name": "Alice"}`
	path := writeTempJSON(t, "users.ndjson", content)
	model := NewJsonFileSource(path)

	rows, _ := model.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
}

func TestNewJsonFileSource_TableNameFromPath(t *testing.T) {
	content := `[{"id": 1}]`
	path := writeTempJSON(t, "my_table.json", content)
	model := NewJsonFileSource(path)

	if model.TableName() != "my_table" {
		t.Errorf("expected table name 'my_table', got '%s'", model.TableName())
	}
}

func TestNewJsonFileSource_EmptyArray(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on empty JSON array file")
		}
	}()
	path := writeTempJSON(t, "empty.json", `[]`)
	NewJsonFileSource(path)
}

func TestNewJsonFileSource_PanicsOnNonExistentFile(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on non-existent file")
		}
	}()
	NewJsonFileSource("/nonexistent/path/to/file.json")
}

func TestNewJsonFileSource_PanicsOnInvalidJSON(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on invalid JSON")
		}
	}()
	path := writeTempJSON(t, "bad.json", `{not valid json}`)
	NewJsonFileSource(path)
}

// --- NewJsonFSSource (fstest.MapFS) tests ---

func TestNewJsonFSSource_Basic(t *testing.T) {
	sys := fstest.MapFS{
		"data/users.json": &fstest.MapFile{
			Data: []byte(`[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`),
		},
	}

	model := NewJsonFSSource(sys, "data/users.json")
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

// --- Internal function tests ---

func TestNormalizeValue(t *testing.T) {
	if v := NormalizeValue("hello"); v != "hello" {
		t.Errorf("NormalizeValue(string) = %v, want 'hello'", v)
	}
	if v, ok := NormalizeValue("2024-01-15T10:30:00Z").(time.Time); !ok {
		t.Errorf("NormalizeValue(RFC3339) expected time.Time, got %T", v)
	}
	if v := NormalizeValue("just text"); v != "just text" {
		t.Errorf("NormalizeValue(plain string) = %v, want 'just text'", v)
	}
	if v := NormalizeValue(float64(42)); v != int64(42) {
		t.Errorf("NormalizeValue(float64(42)) = %v (%T), want int64(42)", v, v)
	}
	if v := NormalizeValue(float64(19.99)); v != float64(19.99) {
		t.Errorf("NormalizeValue(float64(19.99)) = %v (%T), want float64(19.99)", v, v)
	}
	if v := NormalizeValue(true); v != true {
		t.Errorf("NormalizeValue(bool) = %v, want true", v)
	}
	if v := NormalizeValue(nil); v != nil {
		t.Errorf("NormalizeValue(nil) = %v, want nil", v)
	}

	m := map[string]any{"city": "NYC"}
	v := NormalizeValue(m)
	s, ok := v.(string)
	if !ok {
		t.Fatalf("NormalizeValue(map) expected string, got %T", v)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		t.Errorf("NormalizeValue(map) produced invalid JSON: %v", err)
	}

	arr := []any{"red", "blue"}
	v = NormalizeValue(arr)
	s, ok = v.(string)
	if !ok {
		t.Fatalf("NormalizeValue(slice) expected string, got %T", v)
	}
	var parsedArr []any
	if err := json.Unmarshal([]byte(s), &parsedArr); err != nil {
		t.Errorf("NormalizeValue(slice) produced invalid JSON: %v", err)
	}
}

func TestDeriveTableName_JSON(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"data/users.json", "users"},
		{"users.json", "users"},
		{"/home/user/data/products.json", "products"},
		{"events.jsonl", "events"},
		{"events.ndjson", "events"},
	}
	for _, tt := range tests {
		got := DeriveTableName(tt.path)
		if got != tt.want {
			t.Errorf("DeriveTableName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestIsJSONLFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"data.json", false},
		{"data.jsonl", true},
		{"data.ndjson", true},
		{"data.JSONL", true},
		{"data.NDJSON", true},
		{"data.csv", false},
	}
	for _, tt := range tests {
		got := IsJSONLFile(tt.path)
		if got != tt.want {
			t.Errorf("IsJSONLFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// Test MaxJSONRows limit enforcement in ParseJSONArrayReader and ParseJSONLReader
func TestParseJSON_MaxJSONRowsLimit(t *testing.T) {
	// Create a JSON array with MaxJSONRows + 1 rows
	var buf bytes.Buffer
	buf.WriteString("[")
	for i := 0; i <= MaxJSONRows; i++ {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(`{"id": 1}`)
	}
	buf.WriteString("]")

	_, err := ParseJSONArrayReader(&buf)
	if err == nil {
		t.Fatal("Expected error for exceeding MaxJSONRows in ParseJSONArrayReader, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds the limit") {
		t.Errorf("Expected limit error message, got: %v", err)
	}

	// Create a JSONL with MaxJSONRows + 1 rows
	var buf2 bytes.Buffer
	for i := 0; i <= MaxJSONRows; i++ {
		buf2.WriteString(`{"id": 1}` + "\n")
	}

	_, err = ParseJSONLReader(&buf2)
	if err == nil {
		t.Fatal("Expected error for exceeding MaxJSONRows in ParseJSONLReader, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds the limit") {
		t.Errorf("Expected limit error message, got: %v", err)
	}
}
