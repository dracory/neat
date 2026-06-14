package sqlserver_test

import (
	"testing"
)

// TestIsValidDatabaseName tests database name validation for SQL Server
func TestIsValidDatabaseName(t *testing.T) {
	tests := []struct {
		name     string
		dbName   string
		expected bool
	}{
		// Valid names
		{"valid simple", "test", true},
		{"valid with underscore", "test_db", true},
		{"valid with numbers", "test123", true},
		{"valid mixed case", "TestDatabase", true},
		{"valid starting underscore", "_testdb", true},
		{"valid long", "my_integration_test_database_2024", true},

		// Invalid - SQL injection attempts
		{"empty string", "", false},
		{"starts with number", "123test", false},
		{"contains semicolon", "test;DROP DATABASE master", false},
		{"contains space", "test database", false},
		{"contains dash", "test-db", false},
		{"contains dot", "test.db", false},
		{"contains quote", "test'db", false},
		{"contains double quote", `test"db`, false},
		{"contains parenthesis", "test()", false},
		{"contains slash", "test/db", false},
		{"contains backslash", `test\db`, false},
		{"contains at", "test@db", false},
		{"contains hash", "test#db", false},

		// SQL keywords
		{"keyword SELECT", "SELECT", false},
		{"keyword INSERT", "INSERT", false},
		{"keyword DELETE", "DELETE", false},
		{"keyword DROP", "DROP", false},
		{"keyword MASTER", "MASTER", false},
		{"keyword DATABASE", "DATABASE", false},
		{"keyword EXEC", "EXEC", false},

		// SQL keywords lowercase
		{"keyword select lowercase", "select", false},
		{"keyword master lowercase", "master", false},
		{"keyword database lowercase", "database", false},

		// SQL Server system databases
		{"system db model", "model", false},
		{"system db msdb", "msdb", false},
		{"system db tempdb", "tempdb", false},

		// Too long
		{"too long", "a_very_extremely_long_database_name_that_exceeds_the_maximum_length_limit_of_128_characters_and_should_definitely_be_rejected_by_validation", false},

		// Edge cases
		{"single char", "a", true},
		{"two chars", "ab", true},
		{"numbers at end", "test123", true},
		{"multiple underscores", "test__db", true},
		{"underscore at end", "test_", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDatabaseName(tt.dbName)
			if result != tt.expected {
				t.Errorf("isValidDatabaseName(%q) = %v, want %v", tt.dbName, result, tt.expected)
			}
		})
	}
}

// TestSQLServerInjectionAttempts tests specific SQL Server injection patterns
func TestSQLServerInjectionAttempts(t *testing.T) {
	injectionAttempts := []string{
		"test; DROP DATABASE master",
		"test'; DROP DATABASE master; --",
		"test' OR '1'='1",
		"test-- comment",
		"test/* comment */",
		"'; EXEC sp_executesql N'DROP DATABASE master'",
		"xp_cmdshell",
		"test]; DROP DATABASE master; --",
		"test; SHUTDOWN; --",
	}

	for _, attempt := range injectionAttempts {
		t.Run("injection_"+attempt, func(t *testing.T) {
			if isValidDatabaseName(attempt) {
				t.Errorf("SQL injection attempt %q should be rejected", attempt)
			}
		})
	}
}
