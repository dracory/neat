package common

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database"
)

// capturingLogger captures warning messages for testing
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

func (l *capturingLogger) getWarnings() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.warnings
}

func (l *capturingLogger) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warnings = []string{}
}

var _ log.Log = (*capturingLogger)(nil)

// TestSlowQueryWarningIntegration tests that slow query logging triggers correctly
func TestSlowQueryWarningIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := &capturingLogger{}

	// Configure database with low SlowThreshold (1ms)
	// Use file-based database to ensure tables persist across connections
	tempDir := t.TempDir()
	config := neat.DBConfig{
		Default: "default",
		Connections: map[string]neat.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: tempDir + "/slow_test.db",
			},
		},
		SlowThreshold: 1, // 1ms threshold
	}

	db, err := neat.New(config, database.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Get the underlying sql.DB for table creation
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Create a test table using raw SQL (this won't trigger slow query warning)
	_, err = sqlDB.Exec("CREATE TABLE slow_test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Clear any previous warnings
	logger.clear()

	// Execute a query that will exceed the threshold by adding a delay
	// We use a sleep to simulate a slow query
	start := time.Now()
	time.Sleep(2 * time.Millisecond) // Sleep for 2ms to exceed 1ms threshold

	// Use ORM query method (this will trigger slow query warning)
	var result map[string]any
	err = db.Query().Table("slow_test").Where("id = ?", 1).First(&result)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	elapsed := time.Since(start)

	// Verify slow query warning was logged
	if !logger.hasWarning() {
		t.Errorf("Expected slow query warning for query taking %v (threshold: 1ms)", elapsed)
	}

	// Verify warning contains "slow query"
	if !logger.warningContains("slow query") {
		t.Errorf("Expected warning message to contain 'slow query', got: %v", logger.getWarnings())
	}
}

// TestSlowQueryNoWarningBelowThreshold tests that no warning is logged for fast queries
func TestSlowQueryNoWarningBelowThreshold(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := &capturingLogger{}

	// Configure database with higher SlowThreshold (100ms)
	// Use file-based database to ensure tables persist across connections
	tempDir := t.TempDir()
	config := neat.DBConfig{
		Default: "default",
		Connections: map[string]neat.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: tempDir + "/fast_test.db",
			},
		},
		SlowThreshold: 100, // 100ms threshold
	}

	db, err := neat.New(config, database.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Get the underlying sql.DB for table creation
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Create a test table using raw SQL
	_, err = sqlDB.Exec("CREATE TABLE fast_test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Clear any previous warnings
	logger.clear()

	// Execute a fast query (should not exceed 100ms threshold)
	var result map[string]any
	err = db.Query().Table("fast_test").Where("id = ?", 1).First(&result)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// Verify no slow query warning was logged
	if logger.hasWarning() {
		t.Errorf("Expected no slow query warning for fast query, got: %v", logger.getWarnings())
	}
}

// TestSlowQueryDisabled tests that no warning is logged when SlowThreshold is disabled
func TestSlowQueryDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := &capturingLogger{}

	// Configure database with SlowThreshold disabled (0 or negative)
	// Use file-based database to ensure tables persist across connections
	tempDir := t.TempDir()
	config := neat.DBConfig{
		Default: "default",
		Connections: map[string]neat.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: tempDir + "/disabled_test.db",
			},
		},
		SlowThreshold: 0, // Disabled
	}

	db, err := neat.New(config, database.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Get the underlying sql.DB for table creation
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Create a test table using raw SQL
	_, err = sqlDB.Exec("CREATE TABLE disabled_test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Clear any previous warnings
	logger.clear()

	// Execute a query with artificial delay
	time.Sleep(2 * time.Millisecond)
	var result map[string]any
	err = db.Query().Table("disabled_test").Where("id = ?", 1).First(&result)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// Verify no slow query warning was logged when disabled
	if logger.hasWarning() {
		t.Errorf("Expected no slow query warning when SlowThreshold is disabled, got: %v", logger.getWarnings())
	}
}

// TestSlowQueryWithDifferentOperations tests slow query logging for different operation types
func TestSlowQueryWithDifferentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := &capturingLogger{}

	// Configure database with low SlowThreshold (1ms)
	// Use file-based database to ensure tables persist across connections
	tempDir := t.TempDir()
	config := neat.DBConfig{
		Default: "default",
		Connections: map[string]neat.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: tempDir + "/ops_test.db",
			},
		},
		SlowThreshold: 1, // 1ms threshold
	}

	db, err := neat.New(config, database.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Get the underlying sql.DB for table creation
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Create a test table using raw SQL
	_, err = sqlDB.Exec("CREATE TABLE ops_test (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	testCases := []struct {
		name    string
		operate func() error
	}{
		{
			name: "SELECT",
			operate: func() error {
				time.Sleep(2 * time.Millisecond)
				var result map[string]any
				err := db.Query().Table("ops_test").Where("id = ?", 1).First(&result)
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any previous warnings
			logger.clear()

			// Insert a fresh row for the test
			_, err = sqlDB.Exec("INSERT INTO ops_test (name, value) VALUES (?, ?)", "test", 42)
			if err != nil {
				t.Fatalf("Failed to insert test data: %v", err)
			}

			// Execute the operation
			err := tc.operate()
			if err != nil {
				t.Fatalf("Failed to execute %s operation: %v", tc.name, err)
			}

			// Verify slow query warning was logged
			if !logger.hasWarning() {
				t.Errorf("Expected slow query warning for %s operation", tc.name)
			}

			// Verify warning contains "slow query"
			if !logger.warningContains("slow query") {
				t.Errorf("Expected warning message to contain 'slow query' for %s operation", tc.name)
			}

			// Clean up the row for next test
			_, err = sqlDB.Exec("DELETE FROM ops_test WHERE id = ?", 1)
			if err != nil {
				t.Logf("Failed to clean up test data: %v", err)
			}
		})
	}
}
