package database

import (
	"testing"

	"github.com/dracory/neat/database/db"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// TestParseDSNLibsqlValid verifies that valid libsql:// DSNs are parsed correctly,
// extracting only the hostname as Database (not the full URL with authToken).
func TestParseDSNLibsqlValid(t *testing.T) {
	dsn := "libsql://my-db.turso.io?authToken=secret-token"
	driverName, config, err := parseDSN(dsn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if driverName != "turso" {
		t.Errorf("Expected driver 'turso', got '%s'", driverName)
	}
	if config.Driver != "turso" {
		t.Errorf("Expected config.Driver 'turso', got '%s'", config.Driver)
	}
	if config.Database != "my-db.turso.io" {
		t.Errorf("Expected Database 'my-db.turso.io', got '%s' (should not contain authToken)", config.Database)
	}
	if config.Password != "secret-token" {
		t.Errorf("Expected Password 'secret-token', got '%s'", config.Password)
	}
	if config.Dsn != dsn {
		t.Errorf("Expected Dsn to preserve original DSN, got '%s'", config.Dsn)
	}
}

// TestParseDSNLibsqlInvalid verifies that malformed libsql:// DSNs return an error
// rather than silently accepting an invalid connection string.
func TestParseDSNLibsqlInvalid(t *testing.T) {
	// url.Parse will fail on this due to the space
	invalidDSN := "libsql://my-db .turso.io?authToken=secret"
	_, _, err := parseDSN(invalidDSN)
	if err == nil {
		t.Fatal("Expected error for malformed libsql DSN, got nil")
	}
}

// TestRedactDSNLibsqlAuthToken verifies that authToken query parameters are redacted
// in libsql:// DSNs.
func TestRedactDSNLibsqlAuthToken(t *testing.T) {
	dsn := "libsql://my-db.turso.io?authToken=secret-token"
	got := redactDSN(dsn)
	if got == dsn {
		t.Errorf("Expected authToken to be redacted, but DSN was returned unchanged: %s", got)
	}
	if !contains(got, "[REDACTED]") {
		t.Errorf("Expected '[REDACTED]' in redacted DSN, got: %s", got)
	}
	if contains(got, "secret-token") {
		t.Errorf("Expected 'secret-token' to be redacted, but it appears in: %s", got)
	}
}

// TestRedactDSNLibsqlEncodedAuthToken verifies that percent-encoded authToken parameter names
// (e.g. auth%54oken) are properly decoded and redacted.
func TestRedactDSNLibsqlEncodedAuthToken(t *testing.T) {
	dsn := "libsql://my-db.turso.io?auth%54oken=secret-token"
	got := redactDSN(dsn)
	if contains(got, "secret-token") {
		t.Errorf("Expected encoded authToken value to be redacted, but it appears in: %s", got)
	}
}

// TestRedactDSNLibsqlAuthUnderscoreToken verifies that auth_token parameter is redacted.
func TestRedactDSNLibsqlAuthUnderscoreToken(t *testing.T) {
	dsn := "libsql://my-db.turso.io?auth_token=secret-token"
	got := redactDSN(dsn)
	if contains(got, "secret-token") {
		t.Errorf("Expected auth_token value to be redacted, but it appears in: %s", got)
	}
}

// TestRedactDSNLibsqlJWT verifies that jwt parameter is redacted.
func TestRedactDSNLibsqlJWT(t *testing.T) {
	dsn := "libsql://my-db.turso.io?jwt=secret-jwt"
	got := redactDSN(dsn)
	if contains(got, "secret-jwt") {
		t.Errorf("Expected jwt value to be redacted, but it appears in: %s", got)
	}
}

// TestRedactDSNLibsqlPreservesNonSensitiveParams verifies that non-sensitive query
// parameters are preserved alongside redacted ones.
func TestRedactDSNLibsqlPreservesNonSensitiveParams(t *testing.T) {
	dsn := "libsql://my-db.turso.io?authToken=secret&foo=bar"
	got := redactDSN(dsn)
	if !contains(got, "foo=bar") {
		t.Errorf("Expected non-sensitive param 'foo=bar' to be preserved, got: %s", got)
	}
	if contains(got, "secret") {
		t.Errorf("Expected authToken 'secret' to be redacted, got: %s", got)
	}
}

// TestRedactDSNPostgresAuthToken verifies authToken redaction works for postgres DSNs too.
func TestRedactDSNPostgresAuthToken(t *testing.T) {
	dsn := "postgres://user:pass@localhost:5432/mydb?authToken=secret"
	got := redactDSN(dsn)
	if contains(got, "secret") {
		t.Errorf("Expected authToken 'secret' to be redacted in postgres DSN, got: %s", got)
	}
	if !contains(got, "[REDACTED]") {
		t.Errorf("Expected '[REDACTED]' in redacted DSN, got: %s", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestParseDSNLibsqlNoAuthToken verifies libsql:// DSNs without authToken parse correctly.
func TestParseDSNLibsqlNoAuthToken(t *testing.T) {
	dsn := "libsql://my-db.turso.io"
	driverName, config, err := parseDSN(dsn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if driverName != "turso" {
		t.Errorf("Expected driver 'turso', got '%s'", driverName)
	}
	if config.Database != "my-db.turso.io" {
		t.Errorf("Expected Database 'my-db.turso.io', got '%s'", config.Database)
	}
	if config.Password != "" {
		t.Errorf("Expected empty Password, got '%s'", config.Password)
	}
}

// TestParseDSNTursoScheme verifies turso:// DSNs still parse correctly.
func TestParseDSNTursoScheme(t *testing.T) {
	dsn := "turso://my-db.turso.io"
	driverName, config, err := parseDSN(dsn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if driverName != "turso" {
		t.Errorf("Expected driver 'turso', got '%s'", driverName)
	}
	if config.Database != "my-db.turso.io" {
		t.Errorf("Expected Database 'my-db.turso.io', got '%s'", config.Database)
	}
}

// TestNewFromDSNLibsql verifies the full NewFromDSN path with a libsql:// DSN.
// Since we can't actually connect to a remote Turso instance, we use SkipPing
// and expect the driver to be correctly identified as "turso".
func TestNewFromDSNLibsql(t *testing.T) {
	dsn := "libsql://my-db.turso.io?authToken=secret"
	database, err := NewFromDSN(dsn, SkipPing())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = database.Close() }()

	// Verify the connection config was set correctly
	connConfig := database.config.Connections["default"]
	if connConfig.Driver != "turso" {
		t.Errorf("Expected driver 'turso', got '%s'", connConfig.Driver)
	}
	if connConfig.Database != "my-db.turso.io" {
		t.Errorf("Expected Database 'my-db.turso.io', got '%s'", connConfig.Database)
	}
}

// TestNewFromDSNLibsqlInvalid verifies that malformed libsql DSNs cause NewFromDSN to fail.
func TestNewFromDSNLibsqlInvalid(t *testing.T) {
	invalidDSN := "libsql://my-db .turso.io?authToken=secret"
	_, err := NewFromDSN(invalidDSN, SkipPing())
	if err == nil {
		t.Fatal("Expected error for malformed libsql DSN, got nil")
	}
}

// TestRedactDSNLibsqlUserPassAndAuthToken verifies both user:pass@ and authToken are redacted.
func TestRedactDSNLibsqlUserPassAndAuthToken(t *testing.T) {
	dsn := "libsql://user:pass@my-db.turso.io?authToken=secret"
	got := redactDSN(dsn)
	if contains(got, "pass") && !contains(got, "[REDACTED]@") {
		t.Errorf("Expected user:pass to be redacted, got: %s", got)
	}
	if contains(got, "secret") {
		t.Errorf("Expected authToken 'secret' to be redacted, got: %s", got)
	}
}

// Ensure db package is referenced to avoid unused import
var _ = db.ConnectionConfig{}
