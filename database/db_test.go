package database

import (
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
				Database: "testdb",
			},
		},
	}

	db, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db.DatabaseName() != "testdb" {
		t.Errorf("Expected database name 'testdb', got '%s'", db.DatabaseName())
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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
				defer db.Close()
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
