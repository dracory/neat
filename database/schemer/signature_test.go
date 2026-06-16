package schemer

import (
	"testing"
)

func TestValidateMigrationSignature_DateTime_Valid(t *testing.T) {
	validSignatures := []string{
		"2026_06_15_1200_create_users_table",
		"2024_01_01_0000_init",
		"2024_12_31_2359_final_cleanup",
		"2025_02_28_1430_add_posts_table",
		"2020_02_29_1200_leap_year_migration",
	}

	for _, sig := range validSignatures {
		if err := ValidateMigrationSignature(sig, SignatureFormatDateTime); err != nil {
			t.Errorf("Expected '%s' to be valid datetime signature, got error: %v", sig, err)
		}
	}
}

func TestValidateMigrationSignature_DateTime_Invalid(t *testing.T) {
	tests := []struct {
		signature string
		wantErr   string
	}{
		{"", "migration signature cannot be empty"},
		{"2026_06_15_create_users", "time part must be 4 digits"},
		{"2026_06_15_12_create_users", "time part must be 4 digits"},
		{"2026_6_15_1200_create_users", "date parts must be YYYY_MM_DD format"},
		{"2026_06_5_1200_create_users", "date parts must be YYYY_MM_DD format"},
		{"2026_06_15_2400_create_users", "hour must be between 00 and 23"},
		{"2026_06_15_1260_create_users", "minute must be between 00 and 59"},
		{"2026_13_15_1200_create_users", "month must be between 01 and 12"},
		{"2026_06_32_1200_create_users", "day must be between 01 and 31"},
		{"2026_02_30_1200_create_users", "invalid calendar date"},
		{"2026_06_15_120", "datetime format must have at least 5 parts"},
		{"2026_06_15_12000_create_users", "time part must be 4 digits"},
		{"YYYY_MM_DD_1200_create_users", "year must be numeric"},
	}

	for _, tc := range tests {
		err := ValidateMigrationSignature(tc.signature, SignatureFormatDateTime)
		if err == nil {
			t.Errorf("Expected '%s' to be invalid, got nil", tc.signature)
			continue
		}
		if !containsSubstringHelper(err.Error(), tc.wantErr) {
			t.Errorf("Expected error containing '%s' for '%s', got '%s'", tc.wantErr, tc.signature, err.Error())
		}
	}
}

func TestValidateMigrationSignature_Date_Valid(t *testing.T) {
	validSignatures := []string{
		"2026_06_15_001_create_users_table",
		"2024_01_01_000_init",
		"2024_12_31_999_final_cleanup",
	}

	for _, sig := range validSignatures {
		if err := ValidateMigrationSignature(sig, SignatureFormatDate); err != nil {
			t.Errorf("Expected '%s' to be valid date signature, got error: %v", sig, err)
		}
	}
}

func TestValidateMigrationSignature_Date_Invalid(t *testing.T) {
	tests := []struct {
		signature string
		wantErr   string
	}{
		{"2026_06_15_01_create_users", "sequence part must be 3 digits"},
		{"2026_06_15_0001_create_users", "sequence part must be 3 digits"},
		{"2026_06_15_1000_create_users", "sequence part must be 3 digits"},
		{"2026_06_15_ABC_create_users", "sequence part must be numeric"},
	}

	for _, tc := range tests {
		err := ValidateMigrationSignature(tc.signature, SignatureFormatDate)
		if err == nil {
			t.Errorf("Expected '%s' to be invalid, got nil", tc.signature)
			continue
		}
		if !containsSubstringHelper(err.Error(), tc.wantErr) {
			t.Errorf("Expected error containing '%s' for '%s', got '%s'", tc.wantErr, tc.signature, err.Error())
		}
	}
}

func TestValidateMigrationSignature_Unix_Valid(t *testing.T) {
	validSignatures := []string{
		"1717080000_create_users_table",
		"0_init",
		"9999999999_final_cleanup",
	}

	for _, sig := range validSignatures {
		if err := ValidateMigrationSignature(sig, SignatureFormatUnix); err != nil {
			t.Errorf("Expected '%s' to be valid unix signature, got error: %v", sig, err)
		}
	}
}

func TestValidateMigrationSignature_Unix_Invalid(t *testing.T) {
	tests := []struct {
		signature string
		wantErr   string
	}{
		{"create_users_table", "invalid unix timestamp"},
		{"abc_create_users", "invalid unix timestamp"},
		{"1717080000_", "description cannot be empty"},
	}

	for _, tc := range tests {
		err := ValidateMigrationSignature(tc.signature, SignatureFormatUnix)
		if err == nil {
			t.Errorf("Expected '%s' to be invalid, got nil", tc.signature)
			continue
		}
		if !containsSubstringHelper(err.Error(), tc.wantErr) {
			t.Errorf("Expected error containing '%s' for '%s', got '%s'", tc.wantErr, tc.signature, err.Error())
		}
	}
}

func TestValidateMigrationSignature_Custom_Valid(t *testing.T) {
	validSignatures := []string{
		"anything_goes_here",
		"v1.0.0",
		"just_a_name",
		"2026-06-15-create-users",
		"A",
	}

	for _, sig := range validSignatures {
		if err := ValidateMigrationSignature(sig, SignatureFormatCustom); err != nil {
			t.Errorf("Expected '%s' to be valid custom signature, got error: %v", sig, err)
		}
	}
}

func TestValidateMigrationSignature_Custom_Invalid(t *testing.T) {
	tests := []struct {
		signature string
		wantErr   string
	}{
		{"", "migration signature cannot be empty"},
		{string(make([]byte, 256)), "migration signature too long"},
	}

	for _, tc := range tests {
		err := ValidateMigrationSignature(tc.signature, SignatureFormatCustom)
		if err == nil {
			t.Errorf("Expected '%s' to be invalid, got nil", tc.signature)
			continue
		}
		if !containsSubstringHelper(err.Error(), tc.wantErr) {
			t.Errorf("Expected error containing '%s' for '%s', got '%s'", tc.wantErr, tc.signature, err.Error())
		}
	}
}

func TestValidateMigrationSignature_UnknownFormat(t *testing.T) {
	err := ValidateMigrationSignature("test", SignatureFormat("unknown"))
	if err == nil {
		t.Error("Expected error for unknown format")
	}
	if !containsSubstringHelper(err.Error(), "unknown migration signature format") {
		t.Errorf("Expected 'unknown migration signature format' error, got '%s'", err.Error())
	}
}

func TestSetSignatureValidation(t *testing.T) {
	// This is tested indirectly via the schemer integration tests below
	// Here we verify the setter works on the struct
	s := &SchemerImplementation{}

	if s.sigValidation {
		t.Error("Expected signature validation to be disabled by default")
	}

	s.SetSignatureValidation(true, SignatureFormatDateTime)
	if !s.sigValidation {
		t.Error("Expected signature validation to be enabled")
	}
	if s.sigValidationFormat != SignatureFormatDateTime {
		t.Errorf("Expected format '%s', got '%s'", SignatureFormatDateTime, s.sigValidationFormat)
	}

	s.SetSignatureValidation(false, SignatureFormatCustom)
	if s.sigValidation {
		t.Error("Expected signature validation to be disabled")
	}
	if s.sigValidationFormat != SignatureFormatCustom {
		t.Errorf("Expected format '%s', got '%s'", SignatureFormatCustom, s.sigValidationFormat)
	}
}
