package common

import (
	"testing"

	"github.com/dracory/neat"
)

// TestReplicaRoutingIntegration tests that replica configuration is properly accepted
// Note: Actual routing behavior (readConn/writeConn) is tested in unit tests (query_accessors_test.go)
// This integration test verifies the configuration is accepted and database can be created
func TestReplicaRoutingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Configure database with read and write replicas
	config := neat.DBConfig{
		Default: "default",
		Connections: map[string]neat.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:?multi_stmts=true",
				Write: []neat.ReplicaConfig{
					{
						Host:     "primary",
						Port:     0,
						Database: ":memory:?multi_stmts=true",
						Username: "",
						Password: "",
					},
				},
				Read: []neat.ReplicaConfig{
					{
						Host:     "replica",
						Port:     0,
						Database: ":memory:?multi_stmts=true",
						Username: "",
						Password: "",
					},
				},
			},
		},
	}

	// Test that database can be created with replica configuration
	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create database with replica config: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Get the underlying sql.DB to verify connection works
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Verify we can execute basic operations
	_, err = sqlDB.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test passed - replica configuration is accepted and database works
}

// TestReplicaRoutingFallback tests that read operations fall back to primary when no replica is configured
func TestReplicaRoutingFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"

	// Configure database without read replicas
	config := neat.DBConfig{
		Default: "default",
		Connections: map[string]neat.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: dbPath,
				// No Read replicas configured
			},
		},
	}

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Get the underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Create test table
	_, err = sqlDB.Exec("CREATE TABLE fallback_test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert data
	_, err = sqlDB.Exec("INSERT INTO fallback_test (name) VALUES (?)", "test")
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	// SELECT should work (falling back to primary)
	var result map[string]any
	err = db.Query().Table("fallback_test").Where("name = ?", "test").First(&result)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}

	// Verify data was retrieved
	if result["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", result["name"])
	}
}
