package query

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/contracts/log"
)

func TestDebugToggleIntegration(t *testing.T) {
	// Create a mock database connection
	mockDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer mockDB.Close()

	// Create a Query instance
	q := &Query{
		ctx:        context.Background(),
		db:         mockDB,
		dbConfig:   &db.DBConfig{Debug: false},
		log:        log.NewStdLogger(),
		debugState: false,
	}

	// Test initial state
	if q.IsDebug() {
		t.Error("Expected debug to be false initially")
	}

	// Enable debug via Query
	q.EnableDebug()
	if !q.IsDebug() {
		t.Error("Expected debug to be true after EnableDebug")
	}

	// Disable debug via Query
	q.DisableDebug()
	if q.IsDebug() {
		t.Error("Expected debug to be false after DisableDebug")
	}
}

func TestDebugToggleWithDbConfig(t *testing.T) {
	// Create a mock database connection
	mockDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer mockDB.Close()

	// Test with dbConfig.Debug = true
	cfg := &db.DBConfig{Debug: true}
	q := &Query{
		ctx:      context.Background(),
		db:       mockDB,
		dbConfig: cfg,
		log:      log.NewStdLogger(),
	}

	if !q.IsDebug() {
		t.Error("Expected debug to be true when dbConfig.Debug is true")
	}

	// Disable debug should override dbConfig
	q.DisableDebug()
	if q.IsDebug() {
		t.Error("Expected debug to be false after DisableDebug even if dbConfig.Debug is true")
	}
	if cfg.Debug {
		t.Error("Expected dbConfig.Debug to be false after DisableDebug")
	}
}

func TestDebugToggleErrorSanitization(t *testing.T) {
	// Create a mock database connection
	mockDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer mockDB.Close()

	// Create a Query instance with debug disabled
	q := &Query{
		ctx:        context.Background(),
		db:         mockDB,
		dbConfig:   &db.DBConfig{Debug: false},
		log:        log.NewStdLogger(),
		debugState: false,
	}

	// Test error sanitization with debug disabled
	testErr := q.sanitizeError(testErrorWithSQLDetails("SQL syntax error near 'SELECT'"))
	if testErr.Error() == "SQL syntax error near 'SELECT'" {
		t.Error("Expected error to be sanitized when debug is disabled")
	}

	// Enable debug
	q.EnableDebug()

	// Test error sanitization with debug enabled
	testErr = q.sanitizeError(testErrorWithSQLDetails("SQL syntax error near 'SELECT'"))
	if testErr.Error() != "SQL syntax error near 'SELECT'" {
		t.Error("Expected error to not be sanitized when debug is enabled")
	}
}

func testErrorWithSQLDetails(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
