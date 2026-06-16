package migrator

import (
	"testing"
)

func TestIsValidTableName_ValidNames(t *testing.T) {
	validNames := []string{
		"migration_tracker",
		"migrations",
		"my_migrations",
		"_migrations",
		"schema_migrations",
		"a",
		"migration_tracker_2024",
	}

	for _, name := range validNames {
		if !isValidTableName(name) {
			t.Errorf("Expected '%s' to be a valid table name", name)
		}
	}
}

func TestIsValidTableName_InvalidNames(t *testing.T) {
	invalidNames := []struct {
		name   string
		reason string
	}{
		{"", "empty string"},
		{"1migrations", "starts with digit"},
		{"migration-tracker", "contains hyphen"},
		{"migration tracker", "contains space"},
		{"migration!tracker", "contains special character"},
		{"SELECT", "SQL keyword"},
		{"DROP", "SQL keyword"},
		{"TABLE", "SQL keyword"},
		{"FROM", "SQL keyword"},
		{"migration_tracker_with_a_very_long_name_that_exceeds_the_maximum_allowed_length_of_128_characters_and_should_be_rejected_by_the_validator", "too long (>128)"},
	}

	for _, tc := range invalidNames {
		if isValidTableName(tc.name) {
			t.Errorf("Expected '%s' to be invalid (%s)", tc.name, tc.reason)
		}
	}
}

func TestValidateTableName_Valid(t *testing.T) {
	validNames := []string{
		"migration_tracker",
		"migrations",
		"my_migrations",
		"_migrations",
	}

	for _, name := range validNames {
		if err := ValidateTableName(name); err != nil {
			t.Errorf("Expected '%s' to be valid, got error: %v", name, err)
		}
	}
}

func TestValidateTableName_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		expectError string
	}{
		{"", "table name cannot be empty"},
		{"1migrations", "table name must start with a letter or underscore"},
		{"migration-tracker", "table name contains invalid characters"},
		{"migration tracker", "table name contains invalid characters"},
		{"SELECT", "table name cannot be an SQL keyword"},
		{"DROP", "table name cannot be an SQL keyword"},
	}

	for _, tc := range tests {
		err := ValidateTableName(tc.name)
		if err == nil {
			t.Errorf("Expected error for '%s', got nil", tc.name)
			continue
		}
		if !containsSubstring(err.Error(), tc.expectError) {
			t.Errorf("Expected error containing '%s' for '%s', got '%s'", tc.expectError, tc.name, err.Error())
		}
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
