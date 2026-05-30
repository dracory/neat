package db

import (
	"strings"
	"testing"
)

func TestParseDSNMySQL(t *testing.T) {
	dsn := "mysql://user:pass@localhost:3306/testdb"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Driver != "mysql" {
		t.Errorf("Expected driver to be mysql, got %s", config.Driver)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host to be localhost, got %s", config.Host)
	}

	if config.Database != "testdb" {
		t.Errorf("Expected database to be testdb, got %s", config.Database)
	}

	if config.Username != "user" {
		t.Errorf("Expected username to be user, got %s", config.Username)
	}

	if config.Password != "pass" {
		t.Errorf("Expected password to be pass, got %s", config.Password)
	}

	if config.Port != 3306 {
		t.Errorf("Expected port to be 3306, got %d", config.Port)
	}
}

func TestParseDSNPostgres(t *testing.T) {
	dsn := "postgres://user:pass@localhost:5432/testdb"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Driver != "postgres" {
		t.Errorf("Expected driver to be postgres, got %s", config.Driver)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host to be localhost, got %s", config.Host)
	}

	if config.Database != "testdb" {
		t.Errorf("Expected database to be testdb, got %s", config.Database)
	}

	if config.Username != "user" {
		t.Errorf("Expected username to be user, got %s", config.Username)
	}

	if config.Password != "pass" {
		t.Errorf("Expected password to be pass, got %s", config.Password)
	}

	if config.Port != 5432 {
		t.Errorf("Expected port to be 5432, got %d", config.Port)
	}
}

func TestParseDSNPostgreSQL(t *testing.T) {
	dsn := "postgresql://user:pass@localhost:5432/testdb"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Driver != "postgres" {
		t.Errorf("Expected driver to be postgres, got %s", config.Driver)
	}
}

func TestParseDSNSQLite(t *testing.T) {
	dsn := "sqlite:///path/to/database.db"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Driver != "sqlite" {
		t.Errorf("Expected driver to be sqlite, got %s", config.Driver)
	}

	if config.Database != "/path/to/database.db" {
		t.Errorf("Expected database to be /path/to/database.db, got %s", config.Database)
	}
}

func TestParseDSNSQLiteFilePrefix(t *testing.T) {
	dsn := "file:/path/to/database.db"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Driver != "sqlite" {
		t.Errorf("Expected driver to be sqlite, got %s", config.Driver)
	}

	if config.Database != "/path/to/database.db" {
		t.Errorf("Expected database to be /path/to/database.db, got %s", config.Database)
	}
}

func TestParseDSNSQLServer(t *testing.T) {
	dsn := "sqlserver://user:pass@localhost:1433?database=testdb"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Driver != "sqlserver" {
		t.Errorf("Expected driver to be sqlserver, got %s", config.Driver)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host to be localhost, got %s", config.Host)
	}

	if config.Database != "testdb" {
		t.Errorf("Expected database to be testdb, got %s", config.Database)
	}

	if config.Username != "user" {
		t.Errorf("Expected username to be user, got %s", config.Username)
	}

	if config.Password != "pass" {
		t.Errorf("Expected password to be pass, got %s", config.Password)
	}

	if config.Port != 1433 {
		t.Errorf("Expected port to be 1433, got %d", config.Port)
	}
}

func TestParseDSNTurso(t *testing.T) {
	dsn := "libsql://test-db.turso.io?authToken=test-token"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Driver != "turso" {
		t.Errorf("Expected driver to be turso, got %s", config.Driver)
	}

	if config.Database != "test-db.turso.io" {
		t.Errorf("Expected database to be test-db.turso.io, got %s", config.Database)
	}

	if config.Password != "test-token" {
		t.Errorf("Expected password to be test-token, got %s", config.Password)
	}
}

func TestParseDSNWithQueryParameters(t *testing.T) {
	// PostgreSQL with query parameters
	dsn := "postgres://user:pass@localhost:5432/testdb?sslmode=disable&TimeZone=UTC&search_path=public"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.SSLMode != "disable" {
		t.Errorf("Expected SSLMode to be disable, got %s", config.SSLMode)
	}

	if config.Timezone != "UTC" {
		t.Errorf("Expected Timezone to be UTC, got %s", config.Timezone)
	}

	if config.Schema != "public" {
		t.Errorf("Expected Schema to be public, got %s", config.Schema)
	}
}

func TestParseDSNMySQLWithQueryParameters(t *testing.T) {
	dsn := "mysql://user:pass@localhost:3306/testdb?charset=utf8mb4&loc=Local"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Charset != "utf8mb4" {
		t.Errorf("Expected Charset to be utf8mb4, got %s", config.Charset)
	}

	if config.Loc != "Local" {
		t.Errorf("Expected Loc to be Local, got %s", config.Loc)
	}
}

func TestParseDSNSpecialCharactersInPasswordURL(t *testing.T) {
	// Test with URL-encoded special characters
	dsn := "mysql://user:p%40ssw0rd%21%40%23%24@localhost:3306/testdb"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// URL-encoded password should be decoded
	if config.Password != "p@ssw0rd!@#$" {
		t.Errorf("Expected password to be p@ssw0rd!@#$, got %s", config.Password)
	}
}

func TestParseDSNMaxLength(t *testing.T) {
	// Test DSN at max length (4096 characters)
	longDSN := "mysql://user:pass@localhost:3306/testdb"
	for len(longDSN) < maxDSNLength {
		longDSN += "a"
	}

	// Trim to exactly maxDSNLength
	longDSN = longDSN[:maxDSNLength]

	config, err := ParseDSN(longDSN)

	if err != nil {
		t.Errorf("Expected no error for DSN at max length, got %v", err)
	}

	if config.Dsn != longDSN {
		t.Errorf("Expected DSN to be preserved, got truncated version")
	}
}

func TestParseDSNExceedsMaxLength(t *testing.T) {
	// Test DSN exceeding max length
	longDSN := "mysql://user:pass@localhost:3306/testdb"
	for len(longDSN) <= maxDSNLength {
		longDSN += "a"
	}

	_, err := ParseDSN(longDSN)

	if err == nil {
		t.Error("Expected error for DSN exceeding max length")
	}

	expectedError := "DSN exceeds maximum length"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got %v", expectedError, err)
	}
}

func TestParseDSNEmpty(t *testing.T) {
	_, err := ParseDSN("")

	if err == nil {
		t.Error("Expected error for empty DSN")
	}

	expectedError := "DSN cannot be empty"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got %v", expectedError, err)
	}
}

func TestParseDSNInvalidURL(t *testing.T) {
	// Test with invalid URL format
	invalidDSN := "postgres://invalid url with spaces"
	_, err := ParseDSN(invalidDSN)

	if err == nil {
		t.Error("Expected error for invalid URL format")
	}
}

func TestParseDSNUnrecognizedPrefix(t *testing.T) {
	// Test with unrecognized prefix - should return as-is
	dsn := "unknown://user:pass@host/db"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error for unrecognized prefix, got %v", err)
	}

	if config.Dsn != dsn {
		t.Errorf("Expected DSN to be returned as-is, got %s", config.Dsn)
	}
}

func TestBuildDSN(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "user",
		Password: "pass",
	}

	dsn, err := BuildDSN(config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dsn == "" {
		t.Error("Expected DSN to be generated")
	}

	// Verify it contains expected components
	if !strings.Contains(dsn, "user") {
		t.Error("Expected DSN to contain username")
	}
	if !strings.Contains(dsn, "localhost") {
		t.Error("Expected DSN to contain host")
	}
	if !strings.Contains(dsn, "testdb") {
		t.Error("Expected DSN to contain database")
	}
}

func TestBuildDSNWithExistingDSN(t *testing.T) {
	config := ConnectionConfig{
		Dsn: "mysql://user:pass@localhost:3306/testdb",
	}

	dsn, err := BuildDSN(config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dsn != "mysql://user:pass@localhost:3306/testdb" {
		t.Errorf("Expected DSN to be returned as-is, got %s", dsn)
	}
}

func TestParseDSNWithoutPort(t *testing.T) {
	// Test DSN without explicit port - should parse successfully with port = 0
	dsn := "mysql://user:pass@localhost/testdb"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host to be localhost, got %s", config.Host)
	}

	if config.Port != 0 {
		t.Errorf("Expected port to be 0 (default), got %d", config.Port)
	}
}

func TestParseDSNWithoutPassword(t *testing.T) {
	// Test DSN without password
	dsn := "mysql://user@localhost:3306/testdb"
	config, err := ParseDSN(dsn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Username != "user" {
		t.Errorf("Expected username to be user, got %s", config.Username)
	}

	if config.Password != "" {
		t.Errorf("Expected password to be empty, got %s", config.Password)
	}
}

func TestParseDSNInvalidPort(t *testing.T) {
	// Test DSN with invalid (non-numeric) port - URL parser will reject this
	dsn := "mysql://user:pass@localhost:invalid/testdb"
	_, err := ParseDSN(dsn)

	if err == nil {
		t.Error("Expected error for invalid port")
	}
}
