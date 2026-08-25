package driver

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"
)

// writeJSONFile is a test helper that writes a JSON/JSONL file with the given content
// into the given directory and returns the full file path.
func writeJSONFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

// mkdirTempJSONDB is a test helper that creates a temp directory and registers cleanup.
func mkdirTempJSONDB(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "JSONDB-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func TestJSONDBDialect(t *testing.T) {
	d := NewJSONDB()
	if got := d.Dialect(); got != "sqlite" {
		t.Errorf("Dialect() = %q, want %q", got, "sqlite")
	}
}

func TestJSONDBOpenEmptyString(t *testing.T) {
	d := NewJSONDB()
	db, err := d.Open("")
	if err != nil {
		t.Fatalf("Open(\"\") failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tables, got %d", count)
	}
}

func TestJSONDBOpenMemory(t *testing.T) {
	d := NewJSONDB()
	db, err := d.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(\":memory:\") failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tables, got %d", count)
	}
}

func TestJSONDBOpenEmptyDirectory(t *testing.T) {
	dir := mkdirTempJSONDB(t)

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open(empty dir) failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tables in empty dir, got %d", count)
	}
}

func TestJSONDBOpenNonExistentDirectory(t *testing.T) {
	d := NewJSONDB()
	_, err := d.Open(filepath.Join(os.TempDir(), "JSONDB-does-not-exist-12345"))
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}

func TestJSONDBOpenFilePathNotDirectory(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	filePath := writeJSONFile(t, dir, "notadir.json", "[{\"id\":1}]")

	d := NewJSONDB()
	_, err := d.Open(filePath)
	if err == nil {
		t.Fatal("expected error when path is a file, got nil")
	}
}

func TestJSONDBOpenValidDirectory(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "users.json", "[{\"id\": 1, \"name\": \"Alice\", \"email\": \"alice@example.com\", \"active\": true, \"created\": \"2024-01-15T10:30:00Z\"}]")
	writeJSONFile(t, dir, "products.jsonl", "{\"id\": 1, \"name\": \"Widget\", \"price\": 19.99, \"category\": \"Hardware\"}")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	for _, table := range []string{"users", "products"} {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("expected table %s to exist, got count=%d", table, count)
		}
	}
}

func TestJSONDBTableNameMatchesFilename(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "orders.json", "[{\"id\": 1, \"total\": 99.99}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", "orders").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 'orders' table, got count=%d", count)
	}
}

func TestJSONDBColumnNamesInferred(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "items.json", "[{\"id\": 1, \"name\": \"Item\", \"price\": 10.5}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query("PRAGMA table_info(items)")
	if err != nil {
		t.Fatalf("PRAGMA failed: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var cols []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		cols = append(cols, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration failed: %v", err)
	}
	if len(cols) != 3 {
		t.Fatalf("expected 3 columns, got %d: %v", len(cols), cols)
	}
	// Columns are sorted alphabetically: id, name, price
	expected := []string{"id", "name", "price"}
	for i, c := range cols {
		if c != expected[i] {
			t.Errorf("column %d = %q, want %q", i, c, expected[i])
		}
	}
}

func TestJSONDBTypeInferenceInt(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "nums.json", "[{\"id\": 1, \"count\": 100}, {\"id\": 2, \"count\": 200}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('nums') WHERE name='count'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "INTEGER" {
		t.Errorf("expected INTEGER, got %s", ctype)
	}
}

func TestJSONDBTypeInferenceFloat(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "nums.json", "[{\"id\": 1, \"price\": 19.99}, {\"id\": 2, \"price\": 29.99}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('nums') WHERE name='price'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "REAL" {
		t.Errorf("expected REAL, got %s", ctype)
	}
}

func TestJSONDBTypeInferenceBool(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "flags.json", "[{\"id\": 1, \"active\": true}, {\"id\": 2, \"active\": false}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('flags') WHERE name='active'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "INTEGER" {
		t.Errorf("expected INTEGER for bool, got %s", ctype)
	}

	var val int
	if err := db.QueryRow("SELECT active FROM flags WHERE id=1").Scan(&val); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected true→1, got %d", val)
	}
}

func TestJSONDBTypeInferenceTime(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "events.json", "[{\"id\": 1, \"created\": \"2024-01-15T10:30:00Z\"}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('events') WHERE name='created'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "DATETIME" {
		t.Errorf("expected DATETIME, got %s", ctype)
	}
}

func TestJSONDBTypeInferenceString(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "items.json", "[{\"id\": 1, \"name\": \"Widget\"}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('items') WHERE name='name'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "TEXT" {
		t.Errorf("expected TEXT, got %s", ctype)
	}
}

func TestJSONDBTypeWideningIntFloatToReal(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "mixed.json", "[{\"id\": 1, \"value\": 10}, {\"id\": 2, \"value\": 10.5}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('mixed') WHERE name='value'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "REAL" {
		t.Errorf("expected REAL for mixed int/float, got %s", ctype)
	}
}

func TestJSONDBTypeWideningIntStringToText(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "mixed.json", "[{\"id\": 1, \"value\": 10}, {\"id\": 2, \"value\": \"hello\"}]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('mixed') WHERE name='value'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "TEXT" {
		t.Errorf("expected TEXT for mixed int/string, got %s", ctype)
	}
}

func TestJSONDBEmptyJSONFileSkipsTable(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "empty.json", "[]")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='empty'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected empty.json to be skipped (no table created), got count=%d", count)
	}
}

func TestJSONDBInvalidColumnNameInKeys(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "bad.json", "[{\"id\": 1, \"bad.name\": \"foo\"}]")

	d := NewJSONDB()
	_, err := d.Open(dir)
	if err == nil {
		t.Fatal("expected error for invalid column name, got nil")
	}
}

func TestJSONDBNonJSONFilesSkipped(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "users.json", "[{\"id\": 1, \"name\": \"Alice\"}]")
	writeJSONFile(t, dir, "readme.txt", "this is not a json\n")
	writeJSONFile(t, dir, "data.csv", "id,name\n1,Alice\n")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 table, got %d", count)
	}
}

func TestJSONDBSubdirectoriesSkipped(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "users.json", "[{\"id\": 1, \"name\": \"Alice\"}]")

	subDir := filepath.Join(dir, "subdir.json")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='subdir'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected subdir to be skipped, got count=%d", count)
	}
}

func TestJSONDBCaseInsensitiveExtension(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "upper.JSON", "[{\"id\": 1, \"name\": \"Alice\"}]")
	writeJSONFile(t, dir, "mixed.Jsonl", "{\"id\": 1, \"name\": \"Bob\"}")

	d := NewJSONDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	for _, table := range []string{"upper", "mixed"} {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("query failed for %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("expected table %s to exist, got count=%d", table, count)
		}
	}
}

func TestJSONDBTableNameCollisionCaseInsensitive(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "Users.json", "[{\"id\": 1, \"name\": \"Alice\"}]")
	users2Path := filepath.Join(dir, "users.json")
	if err := os.WriteFile(users2Path, []byte("[{\"id\": 2, \"name\": \"Bob\"}]"), 0644); err != nil {
		t.Skipf("cannot create both Users.json and users.json (case-insensitive filesystem): %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	jsonCount := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			jsonCount++
		}
	}
	if jsonCount < 2 {
		t.Skipf("filesystem is case-insensitive, skipping collision test")
	}

	d := NewJSONDB()
	_, err = d.Open(dir)
	if err == nil {
		t.Fatal("expected error for table name collision, got nil")
	}
}

func TestJSONDBDuplicateColumnNames(t *testing.T) {
	dir := mkdirTempJSONDB(t)
	writeJSONFile(t, dir, "dup.json", "[{\"id\": 1, \"name\": \"Alice\", \"Id\": 2}]")

	d := NewJSONDB()
	_, err := d.Open(dir)
	if err == nil {
		t.Fatal("expected error for duplicate column names, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate column name") {
		t.Errorf("expected 'duplicate column name' in error, got: %v", err)
	}
}

func TestJSONDBOpenWithSetFS(t *testing.T) {
	sys := fstest.MapFS{
		"data/users.json": &fstest.MapFile{
			Data: []byte(`[{"id": 1, "name": "Alice"}]`),
		},
		"data/events.jsonl": &fstest.MapFile{
			Data: []byte(`{"id": 100, "type": "click"}`),
		},
	}

	d := NewJSONDB()
	d.SetFS(sys)

	db, err := d.Open("data")
	if err != nil {
		t.Fatalf("Open with SetFS failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var userCount, eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		t.Fatalf("query users failed: %v", err)
	}
	if userCount != 1 {
		t.Errorf("expected 1 user, got %d", userCount)
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM events").Scan(&eventCount); err != nil {
		t.Fatalf("query events failed: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("expected 1 event, got %d", eventCount)
	}
}
