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
		err := q.validateAggregate(col, nil)
		if err != nil {
			t.Errorf("Expected no error for valid column '%s', got: %v", col, err)
		}
	}

	// Test invalid column names
	invalidColumns := []string{"id;", "user'id", "count--", "table`column"}
	for _, col := range invalidColumns {
		err := q.validateAggregate(col, nil)
		if err == nil {
			t.Errorf("Expected error for invalid column '%s'", col)
		}
	}

	// Test nil destination
	err := q.validateAggregate("id", nil)
	if err == nil {
		t.Error("Expected error for nil destination")
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

// UpdateOrInsert tests are in query_upsert_test.go
// This file contains helper-specific tests
