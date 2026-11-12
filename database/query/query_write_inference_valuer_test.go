package query_test

// Regression tests for docs/proposals/write-ops-model-inference-and-valuer.md
//
// 1. Write ops (and reads) should resolve the table from the value's
//    TableName() method or struct-name convention when .Model()/.Table()
//    was not called.
// 2. Fields whose type implements driver.Valuer must reach the driver even
//    when nil (or of a slice/map kind), so Value() decides the stored
//    representation.

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
	neatdriver "github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/query"
)

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

type inferRow struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func (inferRow) TableName() string { return "infer_items" }

// JSONMap is a map column type stored as TEXT via driver.Valuer.
// A nil map must still produce "{}" through Value().
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (m *JSONMap) Scan(src any) error {
	if src == nil {
		*m = nil
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		*m = nil
		return nil
	}
	return json.Unmarshal(b, m)
}

// JSONList is a slice column type stored as TEXT via driver.Valuer.
type JSONList []string

func (l JSONList) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (l *JSONList) Scan(src any) error {
	if src == nil {
		*l = nil
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		*l = nil
		return nil
	}
	return json.Unmarshal(b, l)
}

// scopedRow applies a tenant global scope — inferred models must apply it
// exactly like an explicit .Model() would.
type scopedRow struct {
	ID     string `db:"id"`
	Tenant string `db:"tenant"`
}

func (scopedRow) TableName() string { return "scoped_items" }

func (scopedRow) GlobalScopes() []func(orm.Query) orm.Query {
	return []func(orm.Query) orm.Query{
		func(q orm.Query) orm.Query { return q.Where("tenant = ?", "t1") },
	}
}

// EmbeddedVal is an exported embedded type implementing driver.Valuer. As an
// anonymous field it must be bound as a single column, not flattened.
type EmbeddedVal struct {
	Inner string
}

func (e EmbeddedVal) Value() (driver.Value, error) { return "EMB:" + e.Inner, nil }

type embedRow struct {
	ID string `db:"id"`
	EmbeddedVal
}

func (embedRow) TableName() string { return "embed_items" }

type valuerRow struct {
	ID   string   `db:"id"`
	Data JSONMap  `db:"data"`
	Tags JSONList `db:"tags"`
}

func (valuerRow) TableName() string { return "valuer_items" }

// softRow is a soft-deleting model: Delete(&softRow{}) must UPDATE
// deleted_at, not remove the row — even without an explicit .Model().
type softRow struct {
	ID        string  `db:"id"`
	Name      string  `db:"name"`
	DeletedAt *string `db:"deleted_at"`
}

func (softRow) TableName() string           { return "soft_items" }
func (softRow) SoftDeletedAtColumn() string { return "deleted_at" }

func createInferTable(t *testing.T, w *query.TestQuery) {
	t.Helper()
	execSQL(t, w, "CREATE TABLE infer_items (id TEXT PRIMARY KEY, name TEXT)")
}

func createSoftTable(t *testing.T, w *query.TestQuery) {
	t.Helper()
	execSQL(t, w, "CREATE TABLE soft_items (id TEXT PRIMARY KEY, name TEXT, deleted_at TEXT)")
}

func createValuerTable(t *testing.T, w *query.TestQuery) {
	t.Helper()
	execSQL(t, w, "CREATE TABLE valuer_items (id TEXT PRIMARY KEY, data TEXT NOT NULL, tags TEXT NOT NULL)")
}

func queryString(t *testing.T, db *sql.DB, stmt string, args ...any) string {
	t.Helper()
	var s string
	if err := db.QueryRow(stmt, args...).Scan(&s); err != nil {
		t.Fatalf("query %q: %v", stmt, err)
	}
	return s
}

// ---------------------------------------------------------------------------
// 1. Table inference without .Model()/.Table()
// ---------------------------------------------------------------------------

func TestCreateInfersTableFromTableNameMethod(t *testing.T) {
	w := openSQLiteQuery(t)
	createInferTable(t, w)

	err := w.Q.Create(&inferRow{ID: "i1", Name: "first"})
	if err != nil {
		t.Fatalf("Create without Model() failed: %v", err)
	}

	got := queryString(t, w.PrimaryDB(), "SELECT name FROM infer_items WHERE id = ?", "i1")
	if got != "first" {
		t.Fatalf("expected name 'first', got %q", got)
	}
}

func TestCreateInfersTableFromStructNameConvention(t *testing.T) {
	type conventionWidget struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE convention_widgets (id TEXT PRIMARY KEY, name TEXT)")

	err := w.Q.Create(&conventionWidget{ID: "w1", Name: "widget"})
	if err != nil {
		t.Fatalf("Create without Model() failed: %v", err)
	}

	got := queryString(t, w.PrimaryDB(), "SELECT name FROM convention_widgets WHERE id = ?", "w1")
	if got != "widget" {
		t.Fatalf("expected name 'widget', got %q", got)
	}
}

func TestSaveInfersTableFromValue(t *testing.T) {
	w := openSQLiteQuery(t)
	createInferTable(t, w)
	// Save() with a non-zero primary key performs an UPDATE, so the row must
	// already exist; this exercises table inference on the update path.
	execSQL(t, w, "INSERT INTO infer_items (id, name) VALUES ('s1', 'old')")

	err := w.Q.Save(&inferRow{ID: "s1", Name: "saved"})
	if err != nil {
		t.Fatalf("Save without Model() failed: %v", err)
	}

	got := queryString(t, w.PrimaryDB(), "SELECT name FROM infer_items WHERE id = ?", "s1")
	if got != "saved" {
		t.Fatalf("expected name 'saved', got %q", got)
	}
}

func TestDeleteInfersTableFromValue(t *testing.T) {
	w := openSQLiteQuery(t)
	createInferTable(t, w)
	execSQL(t, w, "INSERT INTO infer_items (id, name) VALUES ('d1', 'gone')")

	_, err := w.Q.Delete(&inferRow{ID: "d1"})
	if err != nil {
		t.Fatalf("Delete without Model() failed: %v", err)
	}

	var n int
	if err := w.PrimaryDB().QueryRow("SELECT COUNT(*) FROM infer_items WHERE id = 'd1'").Scan(&n); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected row d1 to be deleted, still present")
	}
}

func TestFirstInfersTableFromDestination(t *testing.T) {
	w := openSQLiteQuery(t)
	createInferTable(t, w)
	execSQL(t, w, "INSERT INTO infer_items (id, name) VALUES ('f1', 'found')")

	var dest inferRow
	err := w.Q.First(&dest)
	if err != nil {
		t.Fatalf("First without Model() failed: %v", err)
	}
	if dest.Name != "found" {
		t.Fatalf("expected name 'found', got %q", dest.Name)
	}
}

func TestFindInfersTableFromDestination(t *testing.T) {
	w := openSQLiteQuery(t)
	createInferTable(t, w)
	execSQL(t, w, "INSERT INTO infer_items (id, name) VALUES ('f2', 'a'),('f3', 'b')")

	var dest []inferRow
	err := w.Q.Find(&dest)
	if err != nil {
		t.Fatalf("Find without Model() failed: %v", err)
	}
	if len(dest) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(dest))
	}
}

func TestDeleteInferredModelUsesSoftDelete(t *testing.T) {
	w := openSQLiteQuery(t)
	createSoftTable(t, w)
	execSQL(t, w, "INSERT INTO soft_items (id, name) VALUES ('sd1', 'keep')")

	_, err := w.Q.Delete(&softRow{ID: "sd1"})
	if err != nil {
		t.Fatalf("Delete without Model() failed: %v", err)
	}

	var n int
	if err := w.PrimaryDB().QueryRow("SELECT COUNT(*) FROM soft_items WHERE id = 'sd1' AND deleted_at IS NOT NULL").Scan(&n); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected row sd1 to be soft-deleted (deleted_at set), got count %d", n)
	}
}

func TestFirstInferredModelExcludesSoftDeleted(t *testing.T) {
	w := openSQLiteQuery(t)
	createSoftTable(t, w)
	execSQL(t, w, "INSERT INTO soft_items (id, name, deleted_at) VALUES ('sd2', 'gone', '2026-01-01 00:00:00'),('sd3', 'alive', NULL)")

	var dest softRow
	if err := w.Q.Where("id = ?", "sd2").First(&dest); err == nil {
		t.Fatalf("expected soft-deleted row sd2 to be excluded, got %+v", dest)
	}

	// A fresh Query is required: WHERE clauses accumulate on the same builder.
	q2 := query.NewTestQuery(w.PrimaryDB(), neatdriver.NewSQLite(), query.MakeDBConfig(), nil)
	if err := q2.Where("id = ?", "sd3").First(&dest); err != nil {
		t.Fatalf("First for live row failed: %v", err)
	}
	if dest.Name != "alive" {
		t.Fatalf("expected name 'alive', got %q", dest.Name)
	}
}

func TestFirstInferredModelAppliesGlobalScope(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE scoped_items (id TEXT PRIMARY KEY, tenant TEXT)")
	// t2 first so an unscoped First() returns the wrong row
	execSQL(t, w, "INSERT INTO scoped_items (id, tenant) VALUES ('b','t2'),('a','t1')")

	var dest scopedRow
	err := w.Q.First(&dest)
	if err != nil {
		t.Fatalf("First without Model() failed: %v", err)
	}
	if dest.Tenant != "t1" {
		t.Fatalf("expected global scope to restrict to tenant t1, got %+v", dest)
	}
}

func TestFindInferredModelAppliesGlobalScope(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE scoped_items (id TEXT PRIMARY KEY, tenant TEXT)")
	execSQL(t, w, "INSERT INTO scoped_items (id, tenant) VALUES ('a','t1'),('b','t2')")

	var dest []scopedRow
	err := w.Q.Find(&dest)
	if err != nil {
		t.Fatalf("Find without Model() failed: %v", err)
	}
	if len(dest) != 1 || dest[0].Tenant != "t1" {
		t.Fatalf("expected only the t1 row, got %+v", dest)
	}
}

func TestUpdateOrCreateInfersTableFromDest(t *testing.T) {
	w := openSQLiteQuery(t)
	createInferTable(t, w)

	var dest inferRow
	err := w.Q.UpdateOrCreate(&dest, map[string]any{"id": "u1"}, map[string]any{"name": "created"})
	if err != nil {
		t.Fatalf("UpdateOrCreate without Model() failed: %v", err)
	}

	got := queryString(t, w.PrimaryDB(), "SELECT name FROM infer_items WHERE id = ?", "u1")
	if got != "created" {
		t.Fatalf("expected name 'created', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// 2. driver.Valuer honoured for nil and slice-kind fields
// ---------------------------------------------------------------------------

func TestCreateBindsEmbeddedValuerField(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE embed_items (id TEXT PRIMARY KEY, embedded_val TEXT)")

	row := &embedRow{ID: "e1", EmbeddedVal: EmbeddedVal{Inner: "hello"}}
	if err := w.Q.Model(&embedRow{}).Create(row); err != nil {
		t.Fatalf("Create with embedded Valuer failed: %v", err)
	}

	got := queryString(t, w.PrimaryDB(), "SELECT embedded_val FROM embed_items WHERE id = ?", "e1")
	if got != "EMB:hello" {
		t.Fatalf("expected 'EMB:hello', got %q", got)
	}
}

func TestCreateCallsValuerOnNilMapField(t *testing.T) {
	w := openSQLiteQuery(t)
	createValuerTable(t, w)

	row := &valuerRow{ID: "v1", Data: nil, Tags: JSONList{"a"}}
	if err := w.Q.Model(&valuerRow{}).Create(row); err != nil {
		t.Fatalf("Create with nil JSONMap failed: %v", err)
	}

	got := queryString(t, w.PrimaryDB(), "SELECT data FROM valuer_items WHERE id = ?", "v1")
	if got != "{}" {
		t.Fatalf("expected data '{}', got %q", got)
	}
}

func TestCreateBindsNonNilSliceValuerField(t *testing.T) {
	w := openSQLiteQuery(t)
	createValuerTable(t, w)

	row := &valuerRow{ID: "v2", Data: JSONMap{"k": "v"}, Tags: JSONList{"x", "y"}}
	if err := w.Q.Model(&valuerRow{}).Create(row); err != nil {
		t.Fatalf("Create with JSONList failed: %v", err)
	}

	got := queryString(t, w.PrimaryDB(), "SELECT tags FROM valuer_items WHERE id = ?", "v2")
	if got != `["x","y"]` {
		t.Fatalf("expected tags '[\"x\",\"y\"]', got %q", got)
	}
}

func TestCreateCallsValuerOnNilSliceField(t *testing.T) {
	w := openSQLiteQuery(t)
	createValuerTable(t, w)

	row := &valuerRow{ID: "v3", Data: JSONMap{"k": "v"}, Tags: nil}
	if err := w.Q.Model(&valuerRow{}).Create(row); err != nil {
		t.Fatalf("Create with nil JSONList failed: %v", err)
	}

	got := queryString(t, w.PrimaryDB(), "SELECT tags FROM valuer_items WHERE id = ?", "v3")
	if got != "[]" {
		t.Fatalf("expected tags '[]', got %q", got)
	}
}
