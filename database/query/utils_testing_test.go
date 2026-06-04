package query_test

import (
	"context"
	"testing"

	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/query"
	_ "modernc.org/sqlite"
)

// openSQLiteQuery returns a TestQuery wrapper backed by a real in-memory SQLite DB.
func openSQLiteQuery(t *testing.T) *query.TestQuery {
	t.Helper()
	drv := driver.NewSQLite()
	sqlDB, err := drv.Open("file::memory:?cache=shared&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("openSQLiteQuery: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	q := query.NewTestQuery(sqlDB, drv, query.MakeDBConfig(), nil)
	return query.WrapQuery(q)
}

// execSQL runs raw DDL/DML on the TestQuery's write connection.
func execSQL(t *testing.T, w *query.TestQuery, stmt string) {
	t.Helper()
	db, err := w.Q.DB()
	if err != nil {
		t.Fatalf("DB() error: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), stmt); err != nil {
		t.Fatalf("execSQL %q: %v", stmt, err)
	}
}
