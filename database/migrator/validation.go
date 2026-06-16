package migrator

import (
	"fmt"
	"strings"
	"unicode"
)

// isValidTableName validates that a table name is safe to use in SQL queries.
// It checks for:
//   - Only alphanumeric characters and underscores
//   - Does not start with a number
//   - Not empty
//   - Not a known SQL keyword
//   - Reasonable length (max 128)
//
// This prevents SQL injection attacks through malicious table names.
func isValidTableName(tableName string) bool {
	if tableName == "" {
		return false
	}

	if len(tableName) > 128 {
		return false
	}

	// Must start with a letter or underscore (not a number)
	first := rune(tableName[0])
	if !unicode.IsLetter(first) && first != '_' {
		return false
	}

	// Must contain only alphanumeric characters and underscores
	for _, r := range tableName {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}

	// Reject SQL keywords to prevent injection attempts
	upperTableName := strings.ToUpper(tableName)
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"ALTER", "TRUNCATE", "REPLACE", "MERGE", "UNION", "EXCEPT",
		"INTERSECT", "WHERE", "FROM", "JOIN", "INNER", "OUTER",
		"LEFT", "RIGHT", "FULL", "CROSS", "ON", "USING", "AND",
		"OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS",
		"NULL", "TRUE", "FALSE", "CASE", "WHEN", "THEN", "ELSE",
		"END", "GROUP", "HAVING", "ORDER", "BY", "LIMIT", "OFFSET",
		"DISTINCT", "ALL", "AS", "TABLE", "VIEW", "INDEX", "TRIGGER",
		"PROCEDURE", "FUNCTION", "DATABASE", "SCHEMA", "GRANT", "REVOKE",
		"EXEC", "EXECUTE",
	}

	for _, keyword := range sqlKeywords {
		if upperTableName == keyword {
			return false
		}
	}

	return true
}

// ValidateTableName ensures the table name contains only safe characters.
// Exported to allow external validation of table names before creating
// a migrator instance.
func ValidateTableName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("table name cannot be empty")
	}
	if len(name) > 128 {
		return fmt.Errorf("table name too long (max 128 characters)")
	}

	// First character must be a letter or underscore (not a digit)
	firstRune := rune(name[0])
	if !unicode.IsLetter(firstRune) && firstRune != '_' {
		return fmt.Errorf("table name must start with a letter or underscore")
	}

	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("table name contains invalid characters (only alphanumeric and underscore allowed)")
		}
	}

	// SQL keyword check
	upperName := strings.ToUpper(name)
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"ALTER", "TRUNCATE", "REPLACE", "MERGE", "UNION", "EXCEPT",
		"INTERSECT", "WHERE", "FROM", "JOIN", "INNER", "OUTER",
		"LEFT", "RIGHT", "FULL", "CROSS", "ON", "USING", "AND",
		"OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS",
		"NULL", "TRUE", "FALSE", "CASE", "WHEN", "THEN", "ELSE",
		"END", "GROUP", "HAVING", "ORDER", "BY", "LIMIT", "OFFSET",
		"DISTINCT", "ALL", "AS", "TABLE", "VIEW", "INDEX", "TRIGGER",
		"PROCEDURE", "FUNCTION", "DATABASE", "SCHEMA", "GRANT", "REVOKE",
		"EXEC", "EXECUTE",
	}

	for _, keyword := range sqlKeywords {
		if upperName == keyword {
			return fmt.Errorf("table name cannot be an SQL keyword: %s", name)
		}
	}

	return nil
}
