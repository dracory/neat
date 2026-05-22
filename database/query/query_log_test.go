package query

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
)

// capturingLogger records all Warningf calls for assertion.
type capturingLogger struct {
	mu       sync.Mutex
	warnings []string
}

func (l *capturingLogger) Debugf(format string, args ...any) {}
func (l *capturingLogger) Infof(format string, args ...any)  {}
func (l *capturingLogger) Errorf(format string, args ...any) {}
func (l *capturingLogger) Warning(args ...any)               {}
func (l *capturingLogger) Warningf(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warnings = append(l.warnings, format)
}
func (l *capturingLogger) hasWarning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.warnings) > 0
}
func (l *capturingLogger) warningContains(s string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, w := range l.warnings {
		if strings.Contains(w, s) {
			return true
		}
	}
	return false
}

var _ log.Log = (*capturingLogger)(nil)

// openSQLiteQueryWithConfig returns a Query with a real SQLite DB and a custom DBConfig.
func openSQLiteQueryWithConfig(t *testing.T, cfg *db.DBConfig, lg log.Log) *Query {
	t.Helper()
	drv := driver.NewSQLite()
	sqlDB, err := drv.Open(":memory:")
	if err != nil {
		t.Fatalf("openSQLiteQueryWithConfig: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return NewQuery(context.Background(), sqlDB, drv, "default", cfg, lg)
}

func logTestDBConfig(slowThresholdMs int) *db.DBConfig {
	return &db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {Driver: "sqlite", Database: ":memory:"},
		},
		SlowThreshold: slowThresholdMs,
	}
}

// --- EnableQueryLog / DisableQueryLog / FlushQueryLog ---

func TestEnableQueryLogCapturesEntries(t *testing.T) {
	q := openSQLiteQueryWithConfig(t, logTestDBConfig(0), nil)
	execSQL(t, q, "CREATE TABLE ql_test (id INTEGER)")
	execSQL(t, q, "INSERT INTO ql_test VALUES (1)")

	q.EnableQueryLog()
	q.table = "ql_test"
	var result []map[string]any
	_ = q.Get(&result)

	logs := q.GetQueryLog()
	if len(logs) == 0 {
		t.Error("expected at least one query log entry after EnableQueryLog")
	}
}

func TestDisableQueryLogSuppressesEntries(t *testing.T) {
	q := openSQLiteQueryWithConfig(t, logTestDBConfig(0), nil)
	execSQL(t, q, "CREATE TABLE ql_off (id INTEGER)")
	execSQL(t, q, "INSERT INTO ql_off VALUES (1)")

	q.EnableQueryLog()
	q.DisableQueryLog()
	q.table = "ql_off"
	var result []map[string]any
	_ = q.Get(&result)

	if len(q.GetQueryLog()) != 0 {
		t.Error("expected no log entries after DisableQueryLog")
	}
}

func TestFlushQueryLogClearsEntries(t *testing.T) {
	q := openSQLiteQueryWithConfig(t, logTestDBConfig(0), nil)
	execSQL(t, q, "CREATE TABLE ql_flush (id INTEGER)")
	execSQL(t, q, "INSERT INTO ql_flush VALUES (1)")

	q.EnableQueryLog()
	q.table = "ql_flush"
	var result []map[string]any
	_ = q.Get(&result)

	if len(q.GetQueryLog()) == 0 {
		t.Fatal("precondition: expected log entries before flush")
	}
	q.FlushQueryLog()
	if len(q.GetQueryLog()) != 0 {
		t.Error("expected empty log after FlushQueryLog")
	}
}

// --- QueryLog entry has Time >= 0 ---

func TestLogQueryRecordsDuration(t *testing.T) {
	q := openSQLiteQueryWithConfig(t, logTestDBConfig(0), nil)
	execSQL(t, q, "CREATE TABLE ql_dur (id INTEGER)")
	execSQL(t, q, "INSERT INTO ql_dur VALUES (1)")

	q.EnableQueryLog()
	q.table = "ql_dur"
	var result []map[string]any
	_ = q.Get(&result)

	logs := q.GetQueryLog()
	if len(logs) == 0 {
		t.Fatal("expected query log entry")
	}
	if logs[0].Time < 0 {
		t.Errorf("expected Time >= 0, got %v", logs[0].Time)
	}
}

// --- SlowThreshold triggers Warningf ---

func TestLogQuerySlowThresholdEmitsWarning(t *testing.T) {
	lg := &capturingLogger{}
	// threshold of 1ms — call logQuery directly with elapsed=1ms to guarantee trigger
	q := openSQLiteQueryWithConfig(t, logTestDBConfig(1), lg)

	// Call logQuery directly so we control the elapsed time precisely.
	startTime := time.Now().Add(-time.Millisecond * 2) // 2ms ago
	q.logQuery("SELECT 1", nil, startTime)

	if !lg.hasWarning() {
		t.Error("expected slow-query warning when elapsed >= SlowThreshold")
	}
}

func TestLogQuerySlowThresholdContainsSlow(t *testing.T) {
	lg := &capturingLogger{}
	q := openSQLiteQueryWithConfig(t, logTestDBConfig(1), lg)

	startTime := time.Now().Add(-time.Millisecond * 2) // 2ms elapsed
	q.logQuery("SELECT 1", nil, startTime)

	if !lg.warningContains("slow query") {
		t.Error("expected warning message to contain 'slow query'")
	}
}

func TestLogQueryNoWarnWhenThresholdNotSet(t *testing.T) {
	lg := &capturingLogger{}
	// SlowThreshold=0 in DBConfig but we'll leave it unset (use nil dbConfig check)
	cfg := &db.DBConfig{
		Default:       "default",
		Connections:   map[string]db.ConnectionConfig{"default": {Driver: "sqlite", Database: ":memory:"}},
		SlowThreshold: -1, // negative → disabled
	}
	q := openSQLiteQueryWithConfig(t, cfg, lg)
	execSQL(t, q, "CREATE TABLE ql_nowarn (id INTEGER)")
	execSQL(t, q, "INSERT INTO ql_nowarn VALUES (1)")

	q.EnableQueryLog()
	q.table = "ql_nowarn"
	var result []map[string]any
	_ = q.Get(&result)

	if lg.hasWarning() {
		t.Error("expected no warning when SlowThreshold is negative")
	}
}
