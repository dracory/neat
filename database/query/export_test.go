package query

// This file exposes internal types and helpers exclusively for use in
// package query_test. It is compiled only when running tests.

import (
	"context"
	"database/sql"
	"reflect"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
)

// NewTestQuery constructs a Query with all internal fields accessible via
// the TestQuery wrapper below.
func NewTestQuery(sqlDB *sql.DB, drv driver.Driver, cfg *db.DBConfig, lg log.Log) *Query {
	return NewQuery(context.Background(), sqlDB, drv, "default", cfg, lg)
}

// NewTestQueryWithReplicas constructs a Query with explicit read/write DBs.
func NewTestQueryWithReplicas(writeDB, readDB *sql.DB, drv driver.Driver, cfg *db.DBConfig) *Query {
	return NewQueryWithReplicas(context.Background(), writeDB, readDB, drv, "default", cfg, nil)
}

// NewTestQueryWithConfig opens a real SQLite in-memory DB and constructs a Query
// with the given DBConfig and logger. The caller must arrange cleanup via t.Cleanup.
func NewTestQueryWithConfig(t interface {
	Helper()
	Fatalf(string, ...any)
	Cleanup(func())
}, cfg *db.DBConfig, lg log.Log) *Query {
	drv := driver.NewSQLite()
	sqlDB, err := drv.Open(":memory:")
	if err != nil {
		t.Fatalf("NewTestQueryWithConfig: open sqlite: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return NewQuery(context.Background(), sqlDB, drv, "default", cfg, lg)
}

// TestQuery wraps *Query to expose internals to query_test.
type TestQuery struct {
	Q *Query
}

func WrapQuery(q *Query) *TestQuery { return &TestQuery{Q: q} }

func (w *TestQuery) ReadDB() *sql.DB    { return w.Q.readDB }
func (w *TestQuery) WriteDB() *sql.DB   { return w.Q.writeDB }
func (w *TestQuery) PrimaryDB() *sql.DB { return w.Q.db }

func (w *TestQuery) SetReadDB(d *sql.DB)            { w.Q.readDB = d }
func (w *TestQuery) SetWriteDB(d *sql.DB)           { w.Q.writeDB = d }
func (w *TestQuery) SetTx(tx *sql.Tx)               { w.Q.tx = tx }
func (w *TestQuery) SetTable(t string)              { w.Q.table = t }
func (w *TestQuery) SetModel(m any)                 { w.Q.model = m }
func (w *TestQuery) SetWithTrashed(v bool)          { w.Q.withTrashed = v }
func (w *TestQuery) SetOnlyTrashed(v bool)          { w.Q.onlyTrashed = v }
func (w *TestQuery) SetDBConfig(cfg *db.DBConfig)   { w.Q.dbConfig = cfg }
func (w *TestQuery) SetContext(ctx context.Context) { w.Q.ctx = ctx }
func (w *TestQuery) Context() context.Context       { return w.Q.ctx }

func (w *TestQuery) GetTable() string { return w.Q.table }

func (w *TestQuery) GetModelToObserver() []contractsorm.ModelToObserver { return w.Q.modelToObserver }
func (w *TestQuery) GetWithoutEvents() bool                             { return w.Q.withoutEvents }
func (w *TestQuery) GetDistinct() bool                                  { return w.Q.distinct }
func (w *TestQuery) GetDistinctCols() []string                          { return w.Q.distinctCols }

func (w *TestQuery) ReadConn() *sql.DB  { return w.Q.readConn() }
func (w *TestQuery) WriteConn() *sql.DB { return w.Q.writeConn() }

// BuildSelectSQL builds a SELECT statement from the current query state.
func (w *TestQuery) BuildSelectSQL() (string, []any) {
	return NewBuilder(w.Q).BuildSelect()
}

// BuildInsertSQL builds an INSERT statement from the provided values.
func (w *TestQuery) BuildInsertSQL(values any) (string, []any) {
	return NewBuilder(w.Q).BuildInsert(values)
}

// StructFieldColumnName exposes the unexported structFieldColumnName for tests.
func StructFieldColumnName(f reflect.StructField) string {
	return structFieldColumnName(f)
}

// IsPostgres returns true when the query's driver dialect is "postgres".
func (w *TestQuery) IsPostgres() bool {
	return w.Q.driver != nil && w.Q.driver.Dialect() == "postgres"
}

// LogQuery calls the unexported logQuery method.
func (w *TestQuery) LogQuery(sqlStr string, bindings []any, start time.Time) {
	w.Q.logQuery(sqlStr, bindings, start)
}

// MakeDBConfig builds a minimal DBConfig suitable for unit tests.
func MakeDBConfig() *db.DBConfig {
	return &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {Driver: "sqlite", Database: ":memory:"},
		},
	}
}

// MakeDBConfigSlowThreshold builds a DBConfig with the given slow threshold.
func MakeDBConfigSlowThreshold(ms int) *db.DBConfig {
	cfg := MakeDBConfig()
	cfg.SlowThreshold = ms
	return cfg
}

// FakeDriver is a driver.Driver stub with a configurable Dialect for tests.
type FakeDriver struct {
	DialectName string
}

func (f *FakeDriver) Open(dsn string) (*sql.DB, error)          { return nil, nil }
func (f *FakeDriver) Close(d *sql.DB) error                     { return nil }
func (f *FakeDriver) Ping(ctx context.Context, d *sql.DB) error { return nil }
func (f *FakeDriver) BeginTx(ctx context.Context, d *sql.DB, o *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}
func (f *FakeDriver) Dialect() string { return f.DialectName }
func (f *FakeDriver) Placeholder(n int) string {
	if f.DialectName == "postgres" {
		return "$1"
	}
	return "?"
}
