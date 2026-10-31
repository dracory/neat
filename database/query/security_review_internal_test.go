package query

// Security-review proof tests for docs/security-review.md Findings #4, #5, #7.
// These tests FAIL on the current implementation.

import (
	"context"
	"database/sql"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	_ "modernc.org/sqlite"
)

// Finding #4: sanitizeError wraps with %w, so err.Error() still contains the
// full SQL detail the function claims to hide in production mode.
func TestSecurityReviewSanitizeErrorStillLeaksDetail(t *testing.T) {
	q := NewQuery(context.Background(), nil, nil, "default", nil, nil)

	raw := errors.New(`pq: syntax error at or near "pasword" at character 15 in table users`)
	err := q.sanitizeError(raw)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if strings.Contains(err.Error(), "pasword") || strings.Contains(err.Error(), "syntax error") {
		t.Errorf("sanitized error still contains SQL detail: %q", err.Error())
	}
}

// Finding #7: Paginate accepts negative limit without validation. On SQLite,
// LIMIT -5 means "no limit" — silently returning the wrong page instead of an
// error.
func TestSecurityReviewPaginateNegativeLimit(t *testing.T) {
	drv := driver.NewSQLite()
	sqlDB, err := drv.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if _, err := sqlDB.Exec("CREATE TABLE items (id INTEGER, name TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	for i := 0; i < 20; i++ {
		if _, err := sqlDB.Exec("INSERT INTO items (id, name) VALUES (?, ?)", i, "item"); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	q := NewQuery(context.Background(), sqlDB, drv, "default", nil, nil)
	q.Table("items")

	var results []map[string]any
	var total int64
	err = q.Paginate(1, -5, &results, &total)

	// A negative limit is invalid input; Paginate should reject it, not
	// silently run LIMIT -5 (unbounded in SQLite).
	if err == nil {
		t.Errorf("Paginate accepted limit=-5 and returned %d rows; expected a validation error", len(results))
	}
}

// Finding #5: abandoning the cursor channel must not leave the producer
// goroutine blocked forever — the query timeout bounds its lifetime.
func TestSecurityReviewCursorGoroutineLeak(t *testing.T) {
	drv := driver.NewSQLite()
	sqlDB, err := drv.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if _, err := sqlDB.Exec("CREATE TABLE rows_tbl (id INTEGER)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	// More rows than the channel buffer (10) so the goroutine blocks on send.
	for i := 0; i < 50; i++ {
		if _, err := sqlDB.Exec("INSERT INTO rows_tbl (id) VALUES (?)", i); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	cfg := &db.DBConfig{}
	cfg.Pool.QueryTimeout = 1 // seconds

	before := runtime.NumGoroutine()

	q := NewQuery(context.Background(), sqlDB, drv, "default", cfg, nil)
	q.Table("rows_tbl")
	ch, err := q.Cursor()
	if err != nil {
		t.Fatalf("Cursor: %v", err)
	}
	// Read a few items, then abandon the channel — the goroutine must exit
	// once the query timeout fires.
	for i := 0; i < 3; i++ {
		<-ch
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before {
			return // goroutine exited — no leak
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("cursor producer goroutine still alive %s after abandonment (before=%d after=%d)",
		"5s", before, runtime.NumGoroutine())
}

// Finding #5 (companion): with a cancelable query context the consumer can
// always release the pool slot an abandoned cursor pins.
func TestSecurityReviewCursorRowsStayOpen(t *testing.T) {
	drv := driver.NewSQLite()
	// Private-cache in-memory DB so a leaked cursor from another test on the
	// shared cache cannot interfere.
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if _, err := sqlDB.Exec("CREATE TABLE t (id INTEGER)"); err != nil {
		t.Fatalf("create: %v", err)
	}
	for i := 0; i < 50; i++ {
		if _, err := sqlDB.Exec("INSERT INTO t (id) VALUES (?)", i); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	// Constrain the pool AFTER seeding: one abandoned cursor then pins the
	// only connection.
	sqlDB.SetMaxOpenConns(1)

	ctx, cancel := context.WithCancel(context.Background())
	q := NewQuery(ctx, sqlDB, drv, "default", nil, nil)
	q.Table("t")
	ch, err := q.Cursor()
	if err != nil {
		t.Fatalf("Cursor: %v", err)
	}
	<-ch // consume one, then abandon and cancel

	cancel()

	// After cancellation the cursor goroutine must exit and release its
	// connection, letting a subsequent query complete.
	done := make(chan error, 1)
	go func() {
		var n int
		done <- sqlDB.QueryRow("SELECT COUNT(*) FROM t").Scan(&n)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("count query failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Errorf("pool exhausted: subsequent query blocked >3s behind abandoned cursor")
	}
}
