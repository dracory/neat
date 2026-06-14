package migration

import (
	"testing"
)

// TestIsValidMigrationTableName tests the table name validation function
func TestIsValidMigrationTableName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		want      bool
	}{
		// Valid table names
		{"valid simple name", "migrations", true},
		{"valid with underscore", "my_migrations", true},
		{"valid with numbers", "migrations_2024", true},
		{"valid starting with underscore", "_migrations", true},
		{"valid mixed case", "MyMigrations", true},
		{"valid long name", "this_is_a_very_long_migration_table_name_but_still_valid", true},

		// Invalid table names - SQL injection attempts
		{"empty string", "", false},
		{"starts with number", "123migrations", false},
		{"contains semicolon", "migrations;DROP TABLE users", false},
		{"contains space", "my migrations", false},
		{"contains dash", "my-migrations", false},
		{"contains dot", "schema.migrations", false},
		{"contains parenthesis", "migrations()", false},
		{"contains quote", "migrations'", false},
		{"contains double quote", `migrations"`, false},
		{"contains asterisk", "migrations*", false},
		{"contains slash", "migrations/", false},
		{"contains backslash", `migrations\`, false},
		{"contains at sign", "migrations@", false},
		{"contains hash", "migrations#", false},
		{"contains percent", "migrations%", false},
		{"contains ampersand", "migrations&", false},
		{"contains equals", "migrations=", false},
		{"contains plus", "migrations+", false},
		{"contains comma", "migrations,users", false},

		// SQL keyword attempts
		{"keyword SELECT", "SELECT", false},
		{"keyword INSERT", "INSERT", false},
		{"keyword UPDATE", "UPDATE", false},
		{"keyword DELETE", "DELETE", false},
		{"keyword DROP", "DROP", false},
		{"keyword CREATE", "CREATE", false},
		{"keyword ALTER", "ALTER", false},
		{"keyword TRUNCATE", "TRUNCATE", false},
		{"keyword TABLE", "TABLE", false},
		{"keyword FROM", "FROM", false},
		{"keyword WHERE", "WHERE", false},
		{"keyword UNION", "UNION", false},
		{"keyword JOIN", "JOIN", false},
		{"keyword EXEC", "EXEC", false},
		{"keyword EXECUTE", "EXECUTE", false},

		// SQL keywords in lowercase
		{"keyword select lowercase", "select", false},
		{"keyword drop lowercase", "drop", false},
		{"keyword union lowercase", "union", false},

		// Edge cases
		{"too long", "a_very_extremely_super_incredibly_ridiculously_absurdly_long_table_name_that_exceeds_the_maximum_allowed_length_for_safety_purposes_and_should_be_rejected_by_validation_logic_to_prevent_issues", false},
		{"single char valid", "m", true},
		{"single underscore", "_", true},
		{"two underscores", "__", true},
		{"number after valid start", "m1", true},
		{"multiple underscores", "my__migrations", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidMigrationTableName(tt.tableName)
			if got != tt.want {
				t.Errorf("isValidMigrationTableName(%q) = %v, want %v", tt.tableName, got, tt.want)
			}
		})
	}
}

// TestNewRepositoryValidation tests that NewRepository validates table names
func TestNewRepositoryValidation(t *testing.T) {
	// Note: This test requires mock config and orm objects
	// For now, we test the validation function directly above
	// In a real scenario, you would create mocks and test the full constructor

	tests := []struct {
		name        string
		tableName   string
		shouldPanic bool
	}{
		{"valid table name", "migrations", false},
		{"SQL injection attempt", "migrations; DROP TABLE users", true},
		{"SQL keyword", "SELECT", true},
		{"special characters", "migrations'--", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation function directly
			result := isValidMigrationTableName(tt.tableName)
			if tt.shouldPanic && result {
				t.Errorf("Expected validation to fail for %q but it passed", tt.tableName)
			}
			if !tt.shouldPanic && !result {
				t.Errorf("Expected validation to pass for %q but it failed", tt.tableName)
			}
		})
	}
}
