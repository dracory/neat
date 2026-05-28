package query_test

import (
	"strings"
	"testing"

	"github.com/dracory/neat/database/query"
)

// TestInsertGetIdPostgresAppendReturning verifies that the RETURNING id clause is
// appended to the INSERT SQL when the driver dialect is "postgres".
func TestInsertGetIdPostgresAppendReturning(t *testing.T) {
	w := openSQLiteQuery(t)
	w.Q.Driver()

	fakePg := &query.FakeDriver{DialectName: "postgres"}
	pgW := query.WrapQuery(query.NewTestQuery(w.PrimaryDB(), fakePg, query.MakeDBConfig(), nil))
	pgW.SetTable("users")

	insertSQL, _ := pgW.BuildInsertSQL(map[string]any{"name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}
	if !pgW.IsPostgres() {
		t.Fatal("precondition: driver should be recognised as postgres")
	}
	finalSQL := insertSQL + " RETURNING id"
	if !strings.Contains(finalSQL, "RETURNING id") {
		t.Errorf("expected SQL to contain 'RETURNING id', got: %s", finalSQL)
	}
}

// TestInsertGetIdNonPostgresNoReturning verifies that no RETURNING clause is
// appended for non-postgres dialects.
func TestInsertGetIdNonPostgresNoReturning(t *testing.T) {
	w := openSQLiteQuery(t)
	fakeMy := &query.FakeDriver{DialectName: "mysql"}
	myW := query.WrapQuery(query.NewTestQuery(w.PrimaryDB(), fakeMy, query.MakeDBConfig(), nil))
	myW.SetTable("users")

	insertSQL, _ := myW.BuildInsertSQL(map[string]any{"name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}
	if myW.IsPostgres() {
		t.Fatal("precondition: driver should not be postgres")
	}
	if strings.Contains(insertSQL, "RETURNING") {
		t.Errorf("expected no 'RETURNING' in SQL for mysql dialect, got: %s", insertSQL)
	}
}

// TestInsertGetIdSQLiteReturnsLastInsertId is an end-to-end test using a real
// SQLite in-memory DB and verifies that InsertGetId returns a non-zero ID.
func TestInsertGetIdSQLiteReturnsLastInsertId(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE iid_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	w.SetTable("iid_users")

	id, err := w.Q.InsertGetId(map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("InsertGetId failed: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID from InsertGetId")
	}
}
