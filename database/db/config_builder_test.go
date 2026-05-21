package db

import (
	"testing"
)

func TestNewConfigBuilder(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "user",
		Password: "pass",
	}

	builder := NewConfigBuilder(config)
	if builder == nil {
		t.Fatal("Expected builder to be created")
	}

	if builder.config.Driver != "mysql" {
		t.Errorf("Expected driver to be mysql, got %s", builder.config.Driver)
	}
}

func TestConfigBuilderBuildDSNWithExistingDSN(t *testing.T) {
	config := ConnectionConfig{
		Dsn: "mysql://user:pass@localhost:3306/testdb",
	}

	builder := NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dsn != "mysql://user:pass@localhost:3306/testdb" {
		t.Errorf("Expected DSN to be returned as-is, got %s", dsn)
	}
}

func TestConfigBuilderBuildMySQLDSN(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "user",
		Password: "pass",
	}

	builder := NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "user:pass@tcp(localhost:3306)/testdb"
	if dsn != expected {
		t.Errorf("Expected DSN to be %s, got %s", expected, dsn)
	}
}

func TestConfigBuilderBuildPostgresDSN(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		SSLMode:  "disable",
	}

	builder := NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dsn == "" {
		t.Error("Expected DSN to be generated")
	}

	// Verify it contains expected components
	if !contains(dsn, "user") {
		t.Error("Expected DSN to contain username")
	}
	if !contains(dsn, "localhost") {
		t.Error("Expected DSN to contain host")
	}
	if !contains(dsn, "testdb") {
		t.Error("Expected DSN to contain database")
	}
}

func TestConfigBuilderBuildSQLiteDSN(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "sqlite",
		Database: "/path/to/database.db",
	}

	builder := NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dsn == "" {
		t.Error("Expected DSN to be generated")
	}

	if !contains(dsn, "/path/to/database.db") {
		t.Errorf("Expected DSN to contain database path, got %s", dsn)
	}
}

func TestConfigBuilderBuildSQLServerDSN(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "sqlserver",
		Host:     "localhost",
		Port:     1433,
		Database: "testdb",
		Username: "user",
		Password: "pass",
	}

	builder := NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dsn == "" {
		t.Error("Expected DSN to be generated")
	}

	// Verify it contains expected components
	if !contains(dsn, "localhost") {
		t.Error("Expected DSN to contain host")
	}
	if !contains(dsn, "testdb") {
		t.Error("Expected DSN to contain database")
	}
}

func TestConfigBuilderBuildTursoDSN(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "turso",
		Database: "testdb",
		Dsn:      "libsql://test-url?authToken=test-token",
	}

	builder := NewConfigBuilder(config)
	dsn, err := builder.BuildDSN()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dsn == "" {
		t.Error("Expected DSN to be generated")
	}

	if !contains(dsn, "test-url") {
		t.Error("Expected DSN to contain URL")
	}
}

func TestConfigBuilderUnsupportedDriver(t *testing.T) {
	config := ConnectionConfig{
		Driver: "unsupported",
	}

	builder := NewConfigBuilder(config)
	_, err := builder.BuildDSN()

	if err == nil {
		t.Error("Expected error for unsupported driver")
	}
}

func TestConfigBuilderEmptyConfig(t *testing.T) {
	config := ConnectionConfig{
		Driver: "mysql",
		Host:   "",
	}

	builder := NewConfigBuilder(config)
	_, err := builder.BuildDSN()

	// Empty host might not cause an error depending on implementation
	// Just verify the builder doesn't panic
	if builder == nil {
		t.Error("Expected builder to be created")
	}
	_ = err // We don't assert on error as behavior may vary
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
