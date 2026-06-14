package query

import (
	"strings"
	"testing"
)

// TestIsSimpleIdentifier_SecurityEnhanced tests the enhanced isSimpleIdentifier
// function with SQL keyword blocking for Oracle ID retrieval security
func TestIsSimpleIdentifier_SecurityEnhanced(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       bool
	}{
		// Valid identifiers
		{"valid simple name", "users", true},
		{"valid with underscore", "user_accounts", true},
		{"valid with numbers", "users_2024", true},
		{"valid mixed case", "UserAccounts", true},
		{"valid starting with underscore", "_temp", true},
		{"valid long name", "very_long_table_name_with_underscores", true},
		{"valid single char", "u", true},
		{"valid two chars", "id", true},

		// Invalid - SQL keywords (critical for Oracle ID retrieval)
		{"keyword SELECT", "SELECT", false},
		{"keyword INSERT", "INSERT", false},
		{"keyword UPDATE", "UPDATE", false},
		{"keyword DELETE", "DELETE", false},
		{"keyword DROP", "DROP", false},
		{"keyword CREATE", "CREATE", false},
		{"keyword ALTER", "ALTER", false},
		{"keyword TRUNCATE", "TRUNCATE", false},
		{"keyword UNION", "UNION", false},
		{"keyword WHERE", "WHERE", false},
		{"keyword FROM", "FROM", false},
		{"keyword TABLE", "TABLE", false},
		{"keyword DUAL", "DUAL", false},
		{"keyword EXEC", "EXEC", false},
		{"keyword EXECUTE", "EXECUTE", false},

		// SQL keywords in lowercase
		{"keyword select lowercase", "select", false},
		{"keyword drop lowercase", "drop", false},
		{"keyword union lowercase", "union", false},
		{"keyword dual lowercase", "dual", false},

		// SQL keywords in mixed case
		{"keyword Select mixed", "Select", false},
		{"keyword DroP mixed", "DroP", false},

		// Invalid - special characters (SQL injection vectors)
		{"empty string", "", false},
		{"starts with number", "123table", false},
		{"contains semicolon", "users;DROP", false},
		{"contains space", "user accounts", false},
		{"contains dash", "user-accounts", false},
		{"contains dot", "schema.users", false},
		{"contains parenthesis", "COUNT()", false},
		{"contains single quote", "users'", false},
		{"contains double quote", `users"`, false},
		{"contains asterisk", "users*", false},
		{"contains slash", "users/", false},
		{"contains backslash", `users\`, false},
		{"contains at sign", "users@", false},
		{"contains hash", "users#", false},
		{"contains percent", "users%", false},
		{"contains ampersand", "users&", false},
		{"contains equals", "users=", false},
		{"contains plus", "users+", false},
		{"contains comma", "users,accounts", false},

		// Oracle-specific injection attempts
		{"oracle comment", "users--", false},
		{"oracle block comment", "users/*", false},
		{"oracle concatenation", "users||'x'", false},

		// Sequence name patterns (should be valid if table name is valid)
		{"valid sequence pattern", "USERS_SEQ", true},
		{"valid sequence with underscore", "USER_ACCOUNTS_SEQ", true},
		{"valid sequence ID pattern", "USERS_ID_SEQ", true},

		// Edge cases
		{"single underscore", "_", true},
		{"double underscore", "__", true},
		{"triple underscore", "___", true},
		{"underscore at end", "users_", true},
		{"multiple underscores", "user__accounts", true},
		{"number at end", "users1", true},
		{"mixed numbers", "user1account2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSimpleIdentifier(tt.identifier)
			if got != tt.want {
				t.Errorf("isSimpleIdentifier(%q) = %v, want %v", tt.identifier, got, tt.want)
			}
		})
	}
}

// TestOracleSequenceNaming tests that common Oracle sequence naming patterns
// are correctly validated
func TestOracleSequenceNaming(t *testing.T) {
	tableNames := []string{"users", "user_accounts", "orders"}

	for _, table := range tableNames {
		t.Run("table_"+table, func(t *testing.T) {
			// Test that table name itself is valid
			if !isSimpleIdentifier(table) {
				t.Errorf("table name %q should be valid", table)
			}

			// Test sequence naming patterns
			patterns := []string{
				table + "_SEQ",
				"SEQ_" + table,
				table + "_ID_SEQ",
			}

			for _, pattern := range patterns {
				upperPattern := strings.ToUpper(pattern)
				if !isSimpleIdentifier(upperPattern) {
					t.Errorf("sequence pattern %q should be valid", upperPattern)
				}
			}
		})
	}
}

// TestSQLKeywordBlocking specifically tests that all SQL keywords are blocked
func TestSQLKeywordBlocking(t *testing.T) {
	keywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"ALTER", "TRUNCATE", "REPLACE", "MERGE", "UNION", "EXCEPT",
		"INTERSECT", "WHERE", "FROM", "JOIN", "INNER", "OUTER",
		"LEFT", "RIGHT", "FULL", "CROSS", "ON", "USING", "AND",
		"OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS",
		"NULL", "TRUE", "FALSE", "CASE", "WHEN", "THEN", "ELSE",
		"END", "GROUP", "HAVING", "ORDER", "BY", "LIMIT", "OFFSET",
		"DISTINCT", "ALL", "AS", "TABLE", "VIEW", "INDEX", "TRIGGER",
		"PROCEDURE", "FUNCTION", "DATABASE", "SCHEMA", "GRANT", "REVOKE",
		"EXEC", "EXECUTE", "DUAL", "SYSDATE", "SYSTIMESTAMP",
	}

	for _, keyword := range keywords {
		t.Run("keyword_"+keyword, func(t *testing.T) {
			// Test uppercase
			if isSimpleIdentifier(keyword) {
				t.Errorf("SQL keyword %q should be rejected (uppercase)", keyword)
			}

			// Test lowercase
			if isSimpleIdentifier(strings.ToLower(keyword)) {
				t.Errorf("SQL keyword %q should be rejected (lowercase)", strings.ToLower(keyword))
			}

			// Test mixed case
			if len(keyword) > 1 {
				mixed := strings.ToLower(keyword)
				mixed = strings.ToUpper(string(mixed[0])) + mixed[1:]
				if isSimpleIdentifier(mixed) {
					t.Errorf("SQL keyword %q should be rejected (mixed case)", mixed)
				}
			}
		})
	}
}

// TestOracleInjectionAttempts tests specific Oracle SQL injection patterns
func TestOracleInjectionAttempts(t *testing.T) {
	injectionAttempts := []string{
		"users; DROP TABLE accounts",
		"users' OR '1'='1",
		"users-- comment",
		"users/* comment */",
		"users UNION SELECT * FROM passwords",
		"users)) UNION SELECT NULL--",
		"'; DROP TABLE users; --",
		"1=1",
		"admin'--",
		"' OR 1=1--",
		"users||'x'||'y'",
		"CHR(39)||CHR(39)",
		"users;EXEC sp_executesql",
	}

	for _, attempt := range injectionAttempts {
		t.Run("injection_"+attempt, func(t *testing.T) {
			if isSimpleIdentifier(attempt) {
				t.Errorf("Injection attempt %q should be rejected", attempt)
			}
		})
	}
}

// TestTableNameExtraction tests that table names extracted from SQL
// are properly validated (simulating Oracle ID retrieval scenario)
func TestTableNameExtraction(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		extracted string
		valid     bool
	}{
		{
			name:      "simple INSERT",
			sql:       "INSERT INTO users (name) VALUES (?)",
			extracted: "users",
			valid:     true,
		},
		{
			name:      "quoted table name",
			sql:       `INSERT INTO "user_accounts" (name) VALUES (?)`,
			extracted: "user_accounts",
			valid:     true,
		},
		{
			name:      "uppercase table",
			sql:       "INSERT INTO ORDERS (id) VALUES (?)",
			extracted: "ORDERS",
			valid:     true,
		},
		{
			name:      "injection attempt in table name",
			sql:       "INSERT INTO users;DROP TABLE accounts (name) VALUES (?)",
			extracted: "users;DROP",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate extraction
			parts := strings.Fields(tt.sql)
			var tableName string
			for i, part := range parts {
				if strings.ToUpper(part) == "INTO" && i+1 < len(parts) {
					tableName = strings.Trim(parts[i+1], `"`)
					break
				}
			}

			if tableName != tt.extracted {
				t.Fatalf("extracted wrong table name: got %q, want %q", tableName, tt.extracted)
			}

			valid := isSimpleIdentifier(tableName)
			if valid != tt.valid {
				t.Errorf("isSimpleIdentifier(%q) = %v, want %v", tableName, valid, tt.valid)
			}
		})
	}
}
