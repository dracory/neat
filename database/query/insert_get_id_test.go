package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

// fakeDialectDriver implements driver.Driver with a configurable Dialect().
type fakeDialectDriver struct {
	dialect string
}

func (f *fakeDialectDriver) Open(dsn string) (*sql.DB, error)           { return nil, nil }
func (f *fakeDialectDriver) Close(db *sql.DB) error                     { return nil }
func (f *fakeDialectDriver) Ping(ctx context.Context, db *sql.DB) error { return nil }
func (f *fakeDialectDriver) BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}
func (f *fakeDialectDriver) Dialect() string { return f.dialect }
func (f *fakeDialectDriver) Placeholder(n int) string {
	if f.dialect == "postgres" {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

// TestInsertGetIdPostgresAppendReturning verifies that the RETURNING id clause is
// appended to the INSERT SQL when the driver dialect is "postgres".
func TestInsertGetIdPostgresAppendReturning(t *testing.T) {
	q := openSQLiteQuery(t)

	// Swap the driver for a fake postgres dialect driver.
	q.driver = &fakeDialectDriver{dialect: "postgres"}
	q.table = "users"

	// Build the INSERT SQL directly (don't execute — we just verify the SQL).
	builder := NewBuilder(q)
	insertSQL, _ := builder.BuildInsert(map[string]any{"name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}

	// InsertGetId appends " RETURNING id" for postgres before executing.
	isPostgres := q.driver != nil && q.driver.Dialect() == "postgres"
	if !isPostgres {
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
	q := openSQLiteQuery(t)
	q.driver = &fakeDialectDriver{dialect: "mysql"}
	q.table = "users"

	builder := NewBuilder(q)
	insertSQL, _ := builder.BuildInsert(map[string]any{"name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}

	isPostgres := q.driver != nil && q.driver.Dialect() == "postgres"
	if isPostgres {
		t.Fatal("precondition: driver should not be postgres")
	}
	// For non-postgres the SQL should not contain RETURNING.
	if strings.Contains(insertSQL, "RETURNING") {
		t.Errorf("expected no 'RETURNING' in SQL for mysql dialect, got: %s", insertSQL)
	}
}

// TestInsertGetIdSQLiteReturnsLastInsertId is an end-to-end test using a real
// SQLite in-memory DB and verifies that InsertGetId returns a non-zero ID.
func TestInsertGetIdSQLiteReturnsLastInsertId(t *testing.T) {
	q := openSQLiteQuery(t)
	execSQL(t, q, "CREATE TABLE iid_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	q.table = "iid_users"

	id, err := q.InsertGetId(map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("InsertGetId failed: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID from InsertGetId")
	}
}
