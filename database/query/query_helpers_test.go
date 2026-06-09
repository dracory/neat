package query

import (
	"context"
	"testing"
	"time"

	"github.com/dracory/neat/contracts/log"
)

func TestValidateAggregate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	// Test valid column names
	validColumns := []string{"id", "user_id", "count", "*", "table.column"}
	for _, col := range validColumns {
		err := q.validateAggregate(col)
		if err != nil {
			t.Errorf("Expected no error for valid column '%s', got: %v", col, err)
		}
	}

	// Test invalid column names
	invalidColumns := []string{"id;", "user'id", "count--", "table`column"}
	for _, col := range invalidColumns {
		err := q.validateAggregate(col)
		if err == nil {
			t.Errorf("Expected error for invalid column '%s'", col)
		}
	}
}

func TestLogQuery(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.enableLog = true

	start := time.Now()
	q.logQuery("SELECT * FROM users", []any{1, 2}, start)

	if len(*q.queryLog) != 1 {
		t.Errorf("Expected 1 query log entry, got %d", len(*q.queryLog))
	}

	if (*q.queryLog)[0].Query != "SELECT * FROM users" {
		t.Errorf("Query log SQL mismatch")
	}

	if len((*q.queryLog)[0].Bindings) != 2 {
		t.Errorf("Query log bindings count mismatch")
	}
}

func TestLogQueryDisabled(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.enableLog = false

	start := time.Now()
	q.logQuery("SELECT * FROM users", []any{1, 2}, start)

	if len(*q.queryLog) != 0 {
		t.Errorf("Expected 0 query log entries when disabled, got %d", len(*q.queryLog))
	}
}

func TestLogQuerySlowThreshold(t *testing.T) {
	dbConfig := MakeDBConfig()
	dbConfig.SlowThreshold = 100 // 100ms

	logger := log.NewNoopLogger()
	q := NewQuery(context.TODO(), nil, nil, "", dbConfig, logger)
	q.enableLog = true

	// Test slow query
	start := time.Now().Add(-200 * time.Millisecond)
	q.logQuery("SELECT * FROM users", []any{}, start)

	// The warning should be emitted (but we can't easily test the log output)
	// Just ensure it doesn't panic
}

func TestLogQueryBelowThreshold(t *testing.T) {
	dbConfig := MakeDBConfig()
	dbConfig.SlowThreshold = 100 // 100ms

	logger := log.NewNoopLogger()
	q := NewQuery(context.TODO(), nil, nil, "", dbConfig, logger)
	q.enableLog = true

	// Test fast query
	start := time.Now().Add(-10 * time.Millisecond)
	q.logQuery("SELECT * FROM users", []any{}, start)

	// Should not emit warning
	if len(*q.queryLog) != 1 {
		t.Errorf("Expected 1 query log entry, got %d", len(*q.queryLog))
	}
}

func TestTimeoutContextNilContext(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.ctx = nil

	ctx, cancel := q.timeoutContext()
	defer cancel()

	if ctx == nil {
		t.Error("Expected non-nil context even when q.ctx is nil")
	}
}

func TestTimeoutContextNoTimeout(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	ctx, cancel := q.timeoutContext()
	defer cancel()

	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		t.Error("Expected no deadline when QueryTimeout is not configured")
	}
}

func TestTimeoutContextZeroTimeout(t *testing.T) {
	dbConfig := MakeDBConfig()
	dbConfig.Pool.QueryTimeout = 0

	q := NewQuery(context.TODO(), nil, nil, "", dbConfig, nil)

	ctx, cancel := q.timeoutContext()
	defer cancel()

	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		t.Error("Expected no deadline when QueryTimeout is 0")
	}
}

func TestTimeoutContextWithTimeout(t *testing.T) {
	dbConfig := MakeDBConfig()
	dbConfig.Pool.QueryTimeout = 30

	q := NewQuery(context.TODO(), nil, nil, "", dbConfig, nil)

	ctx, cancel := q.timeoutContext()
	defer cancel()

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		t.Fatal("Expected deadline when QueryTimeout is 30s")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > 35*time.Second {
		t.Errorf("Expected deadline ~30s from now, got %v remaining", remaining)
	}
}

// UpdateOrInsert tests are in query_upsert_test.go
// This file contains helper-specific tests
