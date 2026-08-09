package driver

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// writeXMLFile is a test helper that writes an XML file with the given content
// into the given directory and returns the full file path.
func writeXMLFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

// mkdirTempXMLDB is a test helper that creates a temp directory and registers cleanup.
func mkdirTempXMLDB(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "XMLDB-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func TestXMLDBDialect(t *testing.T) {
	d := NewXMLDB()
	if got := d.Dialect(); got != "sqlite" {
		t.Errorf("Dialect() = %q, want %q", got, "sqlite")
	}
}

func TestXMLDBOpenEmptyString(t *testing.T) {
	d := NewXMLDB()
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

func TestXMLDBOpenMemory(t *testing.T) {
	d := NewXMLDB()
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

func TestXMLDBOpenEmptyDirectory(t *testing.T) {
	dir := mkdirTempXMLDB(t)

	d := NewXMLDB()
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

func TestXMLDBOpenNonExistentDirectory(t *testing.T) {
	d := NewXMLDB()
	_, err := d.Open(filepath.Join(os.TempDir(), "XMLDB-does-not-exist-12345"))
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}

func TestXMLDBOpenFilePathNotDirectory(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	filePath := writeXMLFile(t, dir, "notadir.xml", "<users><user id=\"1\"/></users>")

	d := NewXMLDB()
	_, err := d.Open(filePath)
	if err == nil {
		t.Fatal("expected error when path is a file, got nil")
	}
}

func TestXMLDBOpenValidDirectory(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "users.xml", `<users><user id="1"><name>Alice</name><email>alice@example.com</email><active>true</active><created>2024-01-15T10:30:00Z</created></user></users>`)
	writeXMLFile(t, dir, "products.xml", `<products><product id="1"><name>Widget</name><price>19.99</price><category>Hardware</category></product></products>`)

	d := NewXMLDB()
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

func TestXMLDBTableNameMatchesFilename(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "orders.xml", `<orders><order id="1"><total>99.99</total></order></orders>`)

	d := NewXMLDB()
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

func TestXMLDBColumnNamesInferred(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "items.xml", `<items><item id="1"><name>Item</name><price>10.5</price></item></items>`)

	d := NewXMLDB()
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

func TestXMLDBTypeInferenceInt(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "nums.xml", `<nums><num id="1"><count>100</count></num><num id="2"><count>200</count></num></nums>`)

	d := NewXMLDB()
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

func TestXMLDBTypeInferenceFloat(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "nums.xml", `<nums><num id="1"><price>19.99</price></num><num id="2"><price>29.99</price></num></nums>`)

	d := NewXMLDB()
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

func TestXMLDBTypeInferenceBool(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "flags.xml", `<flags><flag id="1"><active>true</active></flag><flag id="2"><active>false</active></flag></flags>`)

	d := NewXMLDB()
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

func TestXMLDBTypeInferenceTime(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "events.xml", `<events><event id="1"><created>2024-01-15T10:30:00Z</created></event></events>`)

	d := NewXMLDB()
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

func TestXMLDBTypeInferenceString(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "items.xml", `<items><item id="1"><name>Widget</name></item></items>`)

	d := NewXMLDB()
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

func TestXMLDBTypeWideningIntFloatToReal(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "mixed.xml", `<mixed><entry id="1"><value>10</value></entry><entry id="2"><value>10.5</value></entry></mixed>`)

	d := NewXMLDB()
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

func TestXMLDBTypeWideningIntStringToText(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "mixed.xml", `<mixed><entry id="1"><value>10</value></entry><entry id="2"><value>hello</value></entry></mixed>`)

	d := NewXMLDB()
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

func TestXMLDBEmptyXMLFileSkipsTable(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "empty.xml", "<users></users>")

	d := NewXMLDB()
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
		t.Errorf("expected empty.xml to be skipped (no table created), got count=%d", count)
	}
}

func TestXMLDBInvalidColumnNameInKeys(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "bad.xml", `<bad><entry id="1"><bad.name>foo</bad.name></entry></bad>`)

	d := NewXMLDB()
	_, err := d.Open(dir)
	if err == nil {
		t.Fatal("expected error for invalid column name, got nil")
	}
}

func TestXMLDBNonXMLFilesSkipped(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "users.xml", `<users><user id="1"><name>Alice</name></user></users>`)
	writeXMLFile(t, dir, "readme.txt", "this is not an xml\n")
	writeXMLFile(t, dir, "data.csv", "id,name\n1,Alice\n")

	d := NewXMLDB()
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

func TestXMLDBSubdirectoriesSkipped(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "users.xml", `<users><user id="1"><name>Alice</name></user></users>`)

	subDir := filepath.Join(dir, "subdir.xml")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	d := NewXMLDB()
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

func TestXMLDBCaseInsensitiveExtension(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "upper.XML", `<upper><entry id="1"><name>Alice</name></entry></upper>`)
	writeXMLFile(t, dir, "mixed.Xml", `<mixed><entry id="1"><name>Bob</name></entry></mixed>`)

	d := NewXMLDB()
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

func TestXMLDBTableNameCollisionCaseInsensitive(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "Users.xml", `<users><user id="1"><name>Alice</name></user></users>`)
	users2Path := filepath.Join(dir, "users.xml")
	if err := os.WriteFile(users2Path, []byte(`<users><user id="2"><name>Bob</name></user></users>`), 0644); err != nil {
		t.Skipf("cannot create both Users.xml and users.xml (case-insensitive filesystem): %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	xmlCount := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".xml") {
			xmlCount++
		}
	}
	if xmlCount < 2 {
		t.Skipf("filesystem is case-insensitive, skipping collision test")
	}

	d := NewXMLDB()
	_, err = d.Open(dir)
	if err == nil {
		t.Fatal("expected error for table name collision, got nil")
	}
}

func TestXMLDBDuplicateColumnNames(t *testing.T) {
	dir := mkdirTempXMLDB(t)
	writeXMLFile(t, dir, "dup.xml", `<dup><entry id="1" name="Alice" Id="2"/></dup>`)

	d := NewXMLDB()
	_, err := d.Open(dir)
	if err == nil {
		t.Fatal("expected error for duplicate column names, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate column name") {
		t.Errorf("expected 'duplicate column name' in error, got: %v", err)
	}
}
