package database

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/database/seeder"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	_ "modernc.org/sqlite"
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

func TestNewFromSQLDB_NilDB(t *testing.T) {
	_, err := NewFromSQLDB(nil, WithDriver("sqlite"), WithLogger(log.NewNoopLogger()))
	if err == nil {
		t.Error("Expected error for nil *sql.DB")
	}
}

func TestNewFromSQLDB_AutoDetect(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	neatDB, err := NewFromSQLDB(sqlDB, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("NewFromSQLDB auto-detect failed: %v", err)
	}
	defer func() { _ = neatDB.Close() }()

	if neatDB.Name() != "default" {
		t.Errorf("Expected connection name 'default', got %q", neatDB.Name())
	}
}

func TestNewFromSQLDB_ExplicitDriver(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	neatDB, err := NewFromSQLDB(sqlDB, WithDriver("sqlite"), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("NewFromSQLDB explicit driver failed: %v", err)
	}
	defer func() { _ = neatDB.Close() }()

	// Verify a basic ORM query works through the provided connection
	_, err = sqlDB.Exec("CREATE TABLE test_from_sqldb (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = sqlDB.Exec("INSERT INTO test_from_sqldb (name) VALUES ('hello')")
	if err != nil {
		t.Fatalf("Failed to insert row: %v", err)
	}

	var rows []map[string]any
	err = neatDB.Query().Table("test_from_sqldb").Find(&rows)
	if err != nil {
		t.Fatalf("ORM query failed: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}
}

func ExampleNew() {
	// Create a database instance with configuration
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
		panic(err)
	}
	defer db.Close()

	// Use the database
	query := db.Query()
	_ = query
}

func ExampleNew_multipleConnections() {
	// Create a database instance with multiple connections
	config := db.DBConfig{
		Default: "primary",
		Connections: map[string]db.ConnectionConfig{
			"primary": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
			"replica": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Get a specific connection
	replica, err := db.Connection("replica")
	if err != nil {
		panic(err)
	}
	_ = replica
}

func ExampleNew_withPoolConfig() {
	// Create a database instance with custom pool configuration
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := New(
		config,
		WithPool(db.PoolConfig{
			MaxIdleConns:    5,
			MaxOpenConns:    25,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 3600,
			QueryTimeout:    30,
		}),
		WithLogger(log.NewNoopLogger()),
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db
}

func ExampleNewFromDSN() {
	// Create a database instance from a DSN string
	db, err := NewFromDSN("sqlite://:memory:", WithLogger(log.NewNoopLogger()))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db
}

func ExampleNewFromDSN_postgres() {
	// PostgreSQL DSN with SSL mode
	db, err := NewFromDSN("postgres://user:pass@localhost:5432/mydb?sslmode=require", WithLogger(log.NewNoopLogger()))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db
}

func ExampleNewFromDSN_mysql() {
	// MySQL DSN with charset
	db, err := NewFromDSN("mysql://user:pass@tcp(localhost:3306)/mydb?charset=utf8mb4", WithLogger(log.NewNoopLogger()))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db
}

func ExampleNewFromDSN_sqlite() {
	// SQLite DSN with file path
	db, err := NewFromDSN("sqlite://./database.db", WithLogger(log.NewNoopLogger()))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db
}

func ExampleNewFromDSN_turso() {
	// Turso (SQLite edge) DSN
	db, err := NewFromDSN("turso://lib-name.turso.io", WithLogger(log.NewNoopLogger()))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db
}

func ExampleNewFromSQLDB() {
	// Create a database instance from an existing *sql.DB
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	db, err := NewFromSQLDB(sqlDB, WithLogger(log.NewNoopLogger()))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db
}

func ExampleNewFromSQLDB_withDriver() {
	// Create a database instance from *sql.DB with explicit driver
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	db, err := NewFromSQLDB(sqlDB, WithDriver("sqlite"), WithLogger(log.NewNoopLogger()))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db
}

func ExampleDatabase_Query() {
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
		panic(err)
	}
	defer db.Close()

	// Get the ORM query builder
	query := db.Query()
	_ = query
}

func ExampleDatabase_Schema() {
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
		panic(err)
	}
	defer db.Close()

	// Get the schema builder
	schema := db.Schema()
	_ = schema
}

func ExampleDatabase_Transaction() {
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
		panic(err)
	}
	defer db.Close()

	// Execute a transaction with automatic rollback on error
	err = db.Transaction(func(tx orm.Query) error {
		// Perform database operations within the transaction
		// If an error is returned, the transaction is rolled back
		// If nil is returned, the transaction is committed
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func ExampleDatabase_Transaction_withError() {
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
		panic(err)
	}
	defer db.Close()

	// Transaction with error handling - automatically rolls back on error
	err = db.Transaction(func(tx orm.Query) error {
		// Simulate an error
		return fmt.Errorf("operation failed")
	})
	if err != nil {
		// Transaction was rolled back
		fmt.Println("Transaction rolled back:", err)
	}
}

func ExampleDatabase_Transaction_withIsolation() {
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
		panic(err)
	}
	defer db.Close()

	// Transaction with isolation level
	err = db.Transaction(func(tx orm.Query) error {
		// Perform operations with specific isolation level
		return nil
	}, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		panic(err)
	}
}
