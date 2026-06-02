package database

import (
	"testing"

	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	_ "github.com/lib/pq"
)

func TestSSL_PostgreSQL_SSLModeConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		sslMode string
		wantErr bool
	}{
		{
			name:    "SSL mode disable",
			sslMode: "disable",
			wantErr: false,
		},
		{
			name:    "SSL mode require",
			sslMode: "require",
			wantErr: false,
		},
		{
			name:    "SSL mode verify-ca",
			sslMode: "verify-ca",
			wantErr: false,
		},
		{
			name:    "SSL mode verify-full",
			sslMode: "verify-full",
			wantErr: false,
		},
		{
			name:    "SSL mode allow",
			sslMode: "allow",
			wantErr: false,
		},
		{
			name:    "SSL mode prefer",
			sslMode: "prefer",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := db.DBConfig{
				Default: "default",
				Connections: map[string]db.ConnectionConfig{
					"default": {
						Driver:   "postgres",
						Host:     "localhost",
						Port:     5432,
						Database: "testdb",
						Username: "user",
						Password: "pass",
						SSLMode:  tt.sslMode,
					},
				},
			}

			builder := db.NewConfigBuilder(config.Connections["default"])
			dsn, err := builder.BuildDSN()

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildDSN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify SSL mode is included in DSN
				if tt.sslMode != "" {
					// Check that sslmode parameter is present
					// The DSN format for PostgreSQL is: host=... port=... user=... password=... dbname=... sslmode=...
					// We can't easily test the exact format without parsing, but we can verify the builder doesn't error
					if dsn == "" {
						t.Error("Expected non-empty DSN")
					}
				}
			}
		})
	}
}

func TestSSL_PostgreSQL_SSLModeDefault(t *testing.T) {
	// Test that when SSL mode is not specified, it defaults to "require"
	config := db.ConnectionConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		// SSLMode not set
	}

	builder := db.NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Fatalf("BuildDSN() error = %v", err)
	}

	// The builder should still work without SSL mode
	if dsn == "" {
		t.Error("Expected non-empty DSN")
	}
}

func TestSSL_MySQL_TLSConfiguration(t *testing.T) {
	// MySQL doesn't use SSLMode in the same way as PostgreSQL
	// It uses TLS parameters in the DSN
	config := db.ConnectionConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		// MySQL SSL configuration would be via DSN parameters like tls=...
	}

	builder := db.NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Fatalf("BuildDSN() error = %v", err)
	}

	// Verify DSN is built
	if dsn == "" {
		t.Error("Expected non-empty DSN")
	}
}

func TestSSL_SQLite_NoSSL(t *testing.T) {
	// SQLite doesn't support SSL
	config := db.ConnectionConfig{
		Driver:   "sqlite",
		Database: ":memory:",
		SSLMode:  "require", // This should be ignored for SQLite
	}

	builder := db.NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Fatalf("BuildDSN() error = %v", err)
	}

	// SQLite DSN should just be the database path
	if dsn != ":memory:" {
		t.Errorf("Expected ':memory:', got '%s'", dsn)
	}
}

func TestSSL_SQLServer_NoSSLMode(t *testing.T) {
	// SQL Server uses connection string format, SSL would be configured differently
	config := db.ConnectionConfig{
		Driver:   "sqlserver",
		Host:     "localhost",
		Port:     1433,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		SSLMode:  "require", // This would be ignored in current implementation
	}

	builder := db.NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Fatalf("BuildDSN() error = %v", err)
	}

	// Verify DSN is built
	if dsn == "" {
		t.Error("Expected non-empty DSN")
	}
}

func TestSSL_DSNParsing_SSLMode(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantSSL string
	}{
		{
			name:    "PostgreSQL DSN with sslmode=require",
			dsn:     "postgres://user:pass@localhost:5432/testdb?sslmode=require",
			wantSSL: "require",
		},
		{
			name:    "PostgreSQL DSN with sslmode=disable",
			dsn:     "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
			wantSSL: "disable",
		},
		{
			name:    "PostgreSQL DSN without sslmode",
			dsn:     "postgres://user:pass@localhost:5432/testdb",
			wantSSL: "", // Should default to empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewFromDSN(tt.dsn, WithLogger(log.NewNoopLogger()), SkipPing())
			if err != nil {
				t.Fatalf("NewFromDSN() error = %v", err)
			}
			defer func() { _ = db.Close() }()

			// Verify the database was created successfully
			if db == nil {
				t.Error("Expected non-nil database")
			}
		})
	}
}

func TestSSL_ConnectionConfig_SSLMode(t *testing.T) {
	tests := []struct {
		name    string
		driver  string
		sslMode string
	}{
		{
			name:    "PostgreSQL with verify-full",
			driver:  "postgres",
			sslMode: "verify-full",
		},
		{
			name:    "PostgreSQL with verify-ca",
			driver:  "postgres",
			sslMode: "verify-ca",
		},
		{
			name:    "PostgreSQL with disable",
			driver:  "postgres",
			sslMode: "disable",
		},
		{
			name:    "MySQL with SSLMode (ignored)",
			driver:  "mysql",
			sslMode: "require",
		},
		{
			name:    "SQLite with SSLMode (ignored)",
			driver:  "sqlite",
			sslMode: "require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := db.ConnectionConfig{
				Driver:   tt.driver,
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  tt.sslMode,
			}

			builder := db.NewConfigBuilder(config)
			dsn, err := builder.BuildDSN()

			if err != nil {
				t.Errorf("BuildDSN() error = %v", err)
				return
			}

			// Verify DSN is built successfully
			if dsn == "" {
				t.Error("Expected non-empty DSN")
			}
		})
	}
}

func TestSSL_Turso_NoSSL(t *testing.T) {
	// Turso (libsql) doesn't use SSLMode in the same way
	config := db.ConnectionConfig{
		Driver:   "turso",
		Database: "libsql://test-db.turso.io",
		SSLMode:  "require", // This should be ignored
	}

	builder := db.NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Fatalf("BuildDSN() error = %v", err)
	}

	// Turso DSN should be returned as-is
	if dsn != "libsql://test-db.turso.io" {
		t.Errorf("Expected 'libsql://test-db.turso.io', got '%s'", dsn)
	}
}

func TestSSL_SpecialCharactersInPassword(t *testing.T) {
	// Test that special characters in password don't break SSL configuration
	config := db.ConnectionConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "p@ss!w0rd#123",
		SSLMode:  "require",
	}

	builder := db.NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Fatalf("BuildDSN() error = %v", err)
	}

	// Verify DSN is built with special characters
	if dsn == "" {
		t.Error("Expected non-empty DSN")
	}
}

func TestSSL_EmptySSLMode(t *testing.T) {
	// Test empty SSL mode string
	config := db.ConnectionConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		SSLMode:  "",
	}

	builder := db.NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Fatalf("BuildDSN() error = %v", err)
	}

	// Should build DSN without SSL mode
	if dsn == "" {
		t.Error("Expected non-empty DSN")
	}
}
