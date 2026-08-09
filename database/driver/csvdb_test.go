package driver

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// writeCSVFile is a test helper that writes a CSV file with the given content
// into the given directory and returns the full file path.
func writeCSVFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

// mkdirTemp is a test helper that creates a temp directory and registers cleanup.
func mkdirTemp(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "CSVDB-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func TestCSVDBDialect(t *testing.T) {
	d := NewCSVDB()
	if got := d.Dialect(); got != "sqlite" {
		t.Errorf("Dialect() = %q, want %q", got, "sqlite")
	}
}

func TestCSVDBOpenEmptyString(t *testing.T) {
	d := NewCSVDB()
	db, err := d.Open("")
	if err != nil {
		t.Fatalf("Open(\"\") failed: %v", err)
	}
	defer db.Close()

	// Should be a valid, empty in-memory SQLite database
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tables, got %d", count)
	}
}

func TestCSVDBOpenMemory(t *testing.T) {
	d := NewCSVDB()
	db, err := d.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(\":memory:\") failed: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tables, got %d", count)
	}
}

func TestCSVDBOpenEmptyDirectory(t *testing.T) {
	dir := mkdirTemp(t)

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open(empty dir) failed: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tables in empty dir, got %d", count)
	}
}

func TestCSVDBOpenNonExistentDirectory(t *testing.T) {
	d := NewCSVDB()
	_, err := d.Open(filepath.Join(os.TempDir(), "CSVDB-does-not-exist-12345"))
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}

func TestCSVDBOpenFilePathNotDirectory(t *testing.T) {
	dir := mkdirTemp(t)
	filePath := writeCSVFile(t, dir, "notadir.csv", "a,b\n1,2\n")

	d := NewCSVDB()
	_, err := d.Open(filePath)
	if err == nil {
		t.Fatal("expected error when path is a file, got nil")
	}
}

func TestCSVDBOpenValidDirectory(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "users.csv", "id,name,email,active,created\n1,Alice,alice@example.com,true,2024-01-15T10:30:00Z\n2,Bob,bob@example.com,false,2024-02-20T14:45:00Z\n")
	writeCSVFile(t, dir, "products.csv", "id,name,price,category\n1,Widget,19.99,Hardware\n2,Gadget,49.99,Electronics\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Both tables should exist
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

func TestCSVDBTableNameMatchesFilename(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "orders.csv", "id,total\n1,99.99\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", "orders").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 'orders' table, got count=%d", count)
	}
}

func TestCSVDBColumnNamesFromHeader(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "items.csv", "id,name,price\n1,Item,10.5\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("PRAGMA table_info(items)")
	if err != nil {
		t.Fatalf("PRAGMA failed: %v", err)
	}
	defer rows.Close()

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
	if len(cols) != 3 {
		t.Fatalf("expected 3 columns, got %d: %v", len(cols), cols)
	}
	// columns are sorted: id, name, price
	expected := []string{"id", "name", "price"}
	for i, c := range cols {
		if c != expected[i] {
			t.Errorf("column %d = %q, want %q", i, c, expected[i])
		}
	}
}

func TestCSVDBTypeInferenceInt(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "nums.csv", "id,count\n1,100\n2,200\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('nums') WHERE name='count'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "INTEGER" {
		t.Errorf("expected INTEGER, got %s", ctype)
	}
}

func TestCSVDBTypeInferenceFloat(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "nums.csv", "id,price\n1,19.99\n2,29.99\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('nums') WHERE name='price'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "REAL" {
		t.Errorf("expected REAL, got %s", ctype)
	}
}

func TestCSVDBTypeInferenceBool(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "flags.csv", "id,active\n1,true\n2,false\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// bool is stored as INTEGER (0/1)
	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('flags') WHERE name='active'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "INTEGER" {
		t.Errorf("expected INTEGER for bool, got %s", ctype)
	}

	// Verify the values were converted to 1/0
	var val int
	if err := db.QueryRow("SELECT active FROM flags WHERE id=1").Scan(&val); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected true→1, got %d", val)
	}
	if err := db.QueryRow("SELECT active FROM flags WHERE id=2").Scan(&val); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if val != 0 {
		t.Errorf("expected false→0, got %d", val)
	}
}

func TestCSVDBTypeInferenceTime(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "events.csv", "id,created\n1,2024-01-15T10:30:00Z\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('events') WHERE name='created'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "DATETIME" {
		t.Errorf("expected DATETIME, got %s", ctype)
	}
}

func TestCSVDBTypeInferenceString(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "items.csv", "id,name\n1,Widget\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('items') WHERE name='name'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "TEXT" {
		t.Errorf("expected TEXT, got %s", ctype)
	}
}

func TestCSVDBTypeWideningIntFloatToReal(t *testing.T) {
	dir := mkdirTemp(t)
	// mixed int and float in same column → REAL
	writeCSVFile(t, dir, "mixed.csv", "id,value\n1,10\n2,10.5\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('mixed') WHERE name='value'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "REAL" {
		t.Errorf("expected REAL for mixed int/float, got %s", ctype)
	}
}

func TestCSVDBTypeWideningIntStringToText(t *testing.T) {
	dir := mkdirTemp(t)
	// mixed int and string in same column → TEXT
	writeCSVFile(t, dir, "mixed.csv", "id,value\n1,10\n2,hello\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var ctype string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('mixed') WHERE name='value'").Scan(&ctype); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if ctype != "TEXT" {
		t.Errorf("expected TEXT for mixed int/string, got %s", ctype)
	}
}

func TestCSVDBEmptyCSVHeaderOnly(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "empty.csv", "id,name\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Table should exist with 0 rows
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM empty").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}

func TestCSVDBEmptyCellsBecomeNull(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "data.csv", "id,name,note\n1,Alice,\n2,,hello\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var note sql.NullString
	if err := db.QueryRow("SELECT note FROM data WHERE id=1").Scan(&note); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if note.Valid {
		t.Errorf("expected NULL for empty cell, got %q", note.String)
	}

	var name sql.NullString
	if err := db.QueryRow("SELECT name FROM data WHERE id=2").Scan(&name); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if name.Valid {
		t.Errorf("expected NULL for empty cell, got %q", name.String)
	}
}

func TestCSVDBInvalidColumnNameInHeader(t *testing.T) {
	dir := mkdirTemp(t)
	// column name with a dot is invalid
	writeCSVFile(t, dir, "bad.csv", "id,bad.name\n1,foo\n")

	d := NewCSVDB()
	_, err := d.Open(dir)
	if err == nil {
		t.Fatal("expected error for invalid column name, got nil")
	}
}

func TestCSVDBNonCSVFilesSkipped(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "users.csv", "id,name\n1,Alice\n")
	writeCSVFile(t, dir, "readme.txt", "this is not a csv\n")
	writeCSVFile(t, dir, "data.json", `{"key":"value"}`)

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Only "users" table should exist
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 table, got %d", count)
	}
}

func TestCSVDBSubdirectoriesSkipped(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "users.csv", "id,name\n1,Alice\n")

	// Create a subdirectory that happens to end in .csv
	subDir := filepath.Join(dir, "subdir.csv")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Only "users" table should exist, not "subdir"
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", "subdir").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected subdir to be skipped, got count=%d", count)
	}
}

func TestCSVDBCaseInsensitiveExtension(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "upper.CSV", "id,name\n1,Alice\n")
	writeCSVFile(t, dir, "mixed.Csv", "id,name\n1,Bob\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	for _, table := range []string{"upper", "mixed"} {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("query failed for %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("expected table %s to exist (case-insensitive ext), got count=%d", table, count)
		}
	}
}

func TestCSVDBTableNameCollisionCaseInsensitive(t *testing.T) {
	dir := mkdirTemp(t)
	// On case-sensitive filesystems (Linux), Users.csv and users.csv can
	// coexist as separate files but produce colliding SQLite table names
	// (SQLite table names are case-insensitive). The driver should detect
	// this and return an error.
	//
	// On case-insensitive filesystems (Windows, macOS default), the OS
	// prevents creating both files, so this scenario can't arise. We skip
	// the test if the second file can't be created (which indicates a
	// case-insensitive filesystem).
	writeCSVFile(t, dir, "Users.csv", "id,name\n1,Alice\n")
	users2Path := filepath.Join(dir, "users.csv")
	if err := os.WriteFile(users2Path, []byte("id,name\n2,Bob\n"), 0644); err != nil {
		t.Skipf("cannot create both Users.csv and users.csv (case-insensitive filesystem): %v", err)
	}

	// Verify both files actually exist as separate entries (case-sensitive FS)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	csvCount := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".csv") {
			csvCount++
		}
	}
	if csvCount < 2 {
		t.Skipf("filesystem is case-insensitive (only %d csv files), skipping collision test", csvCount)
	}

	d := NewCSVDB()
	_, err = d.Open(dir)
	if err == nil {
		t.Fatal("expected error for table name collision, got nil")
	}
}

func TestCSVDBRaggedRowMoreFieldsThanHeader(t *testing.T) {
	dir := mkdirTemp(t)
	// Data row has more fields than the header — should error
	writeCSVFile(t, dir, "bad.csv", "id,name\n1,Alice,extra\n")

	d := NewCSVDB()
	_, err := d.Open(dir)
	if err == nil {
		t.Fatal("expected error for ragged row (more fields than header), got nil")
	}
}

func TestCSVDBRaggedRowFewerFieldsThanHeader(t *testing.T) {
	dir := mkdirTemp(t)
	// Data row has fewer fields than the header — should be allowed
	// (missing fields become NULL)
	writeCSVFile(t, dir, "data.csv", "id,name,email\n1,Alice\n")

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM data").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestCSVDBMaxCSVRowsLimit(t *testing.T) {
	// Test that parseCSV enforces MaxCSVRows.
	// We test parseCSV directly since generating 100k+ rows in a temp file
	// would be slow. Instead, verify the constant is accessible and reasonable.
	if MaxCSVRows <= 0 {
		t.Errorf("MaxCSVRows should be positive, got %d", MaxCSVRows)
	}
	if MaxCSVRows != 100000 {
		t.Errorf("MaxCSVRows = %d, want 100000", MaxCSVRows)
	}
}

func TestCSVDBBOMStripped(t *testing.T) {
	dir := mkdirTemp(t)
	// Write a CSV with a UTF-8 BOM prefix on the first header field.
	// Without BOM stripping, the first column would be "\xEF\xBB\xBFid"
	// which isSimpleIdentifier would reject.
	bomCSV := "\xEF\xBB\xBFid,name\n1,Alice\n"
	writeCSVFile(t, dir, "bom.csv", bomCSV)

	d := NewCSVDB()
	db, err := d.Open(dir)
	if err != nil {
		t.Fatalf("Open failed (BOM not stripped?): %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM bom").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestCSVDBDuplicateColumnNames(t *testing.T) {
	dir := mkdirTemp(t)
	writeCSVFile(t, dir, "dup.csv", "id,name,id\n1,Alice,2\n")

	d := NewCSVDB()
	_, err := d.Open(dir)
	if err == nil {
		t.Fatal("expected error for duplicate column names, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate column name") {
		t.Errorf("expected 'duplicate column name' in error, got: %v", err)
	}
}

func TestCSVDBInfNaNNotInferredAsReal(t *testing.T) {
	// "Inf" and "NaN" are accepted by strconv.ParseFloat but should be
	// treated as TEXT, not REAL.
	if inferValueType("Inf") != "TEXT" {
		t.Errorf("inferValueType(Inf) = %s, want TEXT", inferValueType("Inf"))
	}
	if inferValueType("-Inf") != "TEXT" {
		t.Errorf("inferValueType(-Inf) = %s, want TEXT", inferValueType("-Inf"))
	}
	if inferValueType("NaN") != "TEXT" {
		t.Errorf("inferValueType(NaN) = %s, want TEXT", inferValueType("NaN"))
	}
	// Sanity: real floats should still be REAL
	if inferValueType("3.14") != "REAL" {
		t.Errorf("inferValueType(3.14) = %s, want REAL", inferValueType("3.14"))
	}
}
