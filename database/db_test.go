package database

import (
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/database/seeder"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
)

func TestDatabaseName(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	if db.Name() != "default" {
		t.Errorf("Expected connection name 'default', got '%s'", db.Name())
	}
}

func TestDatabaseName_MultipleConnections(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
			"secondary": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	conn, err := db.Connection("secondary")
	if err != nil {
		t.Fatalf("Failed to get secondary connection: %v", err)
	}

	if conn.Name() != "secondary" {
		t.Errorf("Expected connection name 'secondary', got '%s'", conn.Name())
	}
}

func TestDatabaseName_EmptyStringUsesDefault(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	conn, err := db.Connection("")
	if err != nil {
		t.Fatalf("Failed to get default connection with empty string: %v", err)
	}

	if conn.Name() != "default" {
		t.Errorf("Expected connection name 'default', got '%s'", conn.Name())
	}
}

func TestDatabaseName_DatabaseName(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	if db.DatabaseName() != ":memory:" {
		t.Errorf("Expected database name ':memory:', got '%s'", db.DatabaseName())
	}
}

func TestDatabase_Connection(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
			"secondary": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	t.Run("Get existing connection", func(t *testing.T) {
		conn, err := db.Connection("secondary")
		if err != nil {
			t.Fatalf("Failed to get secondary connection: %v", err)
		}
		if conn == nil {
			t.Error("Expected non-nil connection")
		}
	})

	t.Run("Get default connection with empty string", func(t *testing.T) {
		conn, err := db.Connection("")
		if err != nil {
			t.Fatalf("Failed to get default connection: %v", err)
		}
		if conn.Name() != "default" {
			t.Errorf("Expected connection name 'default', got '%s'", conn.Name())
		}
	})

	t.Run("Non-existent connection returns error", func(t *testing.T) {
		_, err := db.Connection("nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent connection")
		}
	})
}

func TestDatabase_Query(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	query := db.Query()
	if query == nil {
		t.Error("Expected non-nil query")
	}
}

func TestDatabase_Schema(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	if schema == nil {
		t.Error("Expected non-nil schema")
	}
}

func TestDatabase_DB(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	if sqlDB == nil {
		t.Error("Expected non-nil sql.DB")
	}
}

func TestDatabase_Close(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}

func TestDatabase_QueryLog(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Test EnableQueryLog
	db.EnableQueryLog()

	// Test GetQueryLog
	logs := db.GetQueryLog()
	if logs == nil {
		t.Error("Expected non-nil query log")
	}

	// Test FlushQueryLog
	db.FlushQueryLog()
	logs = db.GetQueryLog()
	if len(logs) != 0 {
		t.Errorf("Expected empty query log after flush, got %d logs", len(logs))
	}

	// Test DisableQueryLog
	db.DisableQueryLog()
}

func TestDatabase_Transaction(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Test successful transaction
	err = db.Transaction(func(tx orm.Query) error {
		// This is a simplified test - in real usage, tx would be used for queries
		return nil
	})
	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}
}

func TestNew_AutoDefaultConnection(t *testing.T) {
	config := db.DBConfig{
		Connections: map[string]db.ConnectionConfig{
			"custom": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	if db.Name() != "custom" {
		t.Errorf("Expected auto-selected connection name 'custom', got '%s'", db.Name())
	}
}

func TestNewFromDSN(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "SQLite DSN",
			dsn:     "sqlite://:memory:",
			wantErr: false,
		},
		{
			name:    "Empty DSN",
			dsn:     "",
			wantErr: true,
		},
		{
			name:    "Invalid DSN",
			dsn:     "invalid://dsn",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewFromDSN(tt.dsn, WithLogger(log.NewNoopLogger()))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromDSN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() { _ = db.Close() }()
				if db == nil {
					t.Error("Expected non-nil database")
				}
			}
		})
	}
}

func TestRedactDSN(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{
			name: "PostgreSQL DSN with credentials",
			dsn:  "postgres://user:pass@localhost:5432/mydb",
			want: "postgres://[REDACTED]@localhost:5432/mydb",
		},
		{
			name: "MySQL DSN with credentials",
			dsn:  "mysql://user:pass@tcp(localhost:3306)/mydb",
			want: "mysql://[REDACTED]@tcp(localhost:3306)/mydb",
		},
		{
			name: "DSN without credentials",
			dsn:  "sqlite://:memory:",
			want: "sqlite://:memory:",
		},
		{
			name: "Empty DSN",
			dsn:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactDSN(tt.dsn)
			if got != tt.want {
				t.Errorf("redactDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Mock seeder for testing
type mockSeeder struct {
	signature string
	runCalled bool
}

func (m *mockSeeder) Signature() string {
	return m.signature
}

func (m *mockSeeder) Run() error {
	m.runCalled = true
	return nil
}

func TestDatabase_Seed(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	seeder1 := &mockSeeder{signature: "seeder_1"}
	seeder2 := &mockSeeder{signature: "seeder_2"}

	seeders := []seeder.Seeder{seeder1, seeder2}

	err = db.Seed(seeders)
	if err != nil {
		t.Errorf("Seed() failed: %v", err)
	}

	if !seeder1.runCalled {
		t.Error("Expected seeder1.Run() to be called")
	}
	if !seeder2.runCalled {
		t.Error("Expected seeder2.Run() to be called")
	}
}

func TestDatabase_SeedOnce(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	seeder1 := &mockSeeder{signature: "seeder_1"}
	seeders := []seeder.Seeder{seeder1}

	// First call
	err = db.SeedOnce(seeders)
	if err != nil {
		t.Errorf("SeedOnce() failed on first call: %v", err)
	}

	if !seeder1.runCalled {
		t.Error("Expected seeder1.Run() to be called on first SeedOnce")
	}

	// Reset runCalled flag
	seeder1.runCalled = false

	// Second call - should skip
	err = db.SeedOnce(seeders)
	if err != nil {
		t.Errorf("SeedOnce() failed on second call: %v", err)
	}

	if seeder1.runCalled {
		t.Error("Expected seeder1.Run() to NOT be called on second SeedOnce")
	}
}

func TestDatabase_Seeder(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	facade := db.Seeder()
	if facade == nil {
		t.Fatal("Expected non-nil seeder facade")
	}

	seeder1 := &mockSeeder{signature: "seeder_1"}
	seeders := []seeder.Seeder{seeder1}

	// Register seeders
	facade.Register(seeders)

	// Get seeder
	retrieved := facade.GetSeeder("seeder_1")
	if retrieved == nil {
		t.Error("Expected seeder to be found")
		return
	}
	if retrieved.Signature() != "seeder_1" {
		t.Errorf("Expected signature 'seeder_1', got '%s'", retrieved.Signature())
	}

	// Get all seeders
	allSeeders := facade.GetSeeders()
	if len(allSeeders) != 1 {
		t.Errorf("Expected 1 seeder, got %d", len(allSeeders))
	}
}

func TestDatabase_Migrate(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Test Migrate with empty paths (should use default)
	err = db.Migrate()
	if err != nil {
		t.Errorf("Migrate() failed: %v", err)
	}
}

func TestDatabase_MigrationStatus(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Test MigrationStatus - should return empty slice when no migrations
	status, err := db.MigrationStatus()
	if err != nil {
		t.Errorf("MigrationStatus() failed: %v", err)
	}
	// When no migrations are registered, status should be empty
	if len(status) != 0 {
		t.Errorf("Expected empty status when no migrations, got %d items", len(status))
	}
}
