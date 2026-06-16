package schemer

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SignatureFormat defines the format for migration signatures
type SignatureFormat string

const (
	// SignatureFormatDateTime uses timestamp-based format (default)
	// Example: 2026_06_14_1200_create_users_table
	SignatureFormatDateTime SignatureFormat = "datetime"

	// SignatureFormatDate uses sequence-based format
	// Example: 2026_06_14_001_create_users_table
	SignatureFormatDate SignatureFormat = "date"

	// SignatureFormatUnix uses unix timestamp format (legacy)
	// Example: 1717080000_create_users_table
	SignatureFormatUnix SignatureFormat = "unix"

	// SignatureFormatCustom uses no prefix format restriction
	SignatureFormatCustom SignatureFormat = "custom"
)

// ValidateMigrationSignature validates that a migration signature follows the specified format.
//
// Supported formats:
//   - YYYY_MM_DD_HHMM_description (for datetime format)
//   - YYYY_MM_DD_NNN_description (for date format)
//   - unix_timestamp_description (for unix format)
//   - custom (no prefix format restriction, only length and non-empty)
//
// Business Logic:
//   - Enforces maximum length of 255 characters
//   - Rejects empty signatures
//   - For "custom" format: only validates length and non-empty
//   - For other formats: requires at least 5 underscore-separated parts (or 2 for unix)
//   - Validates date part (YYYY_MM_DD) is a valid calendar date
//   - Validates time part (HHMM) or sequence part (NNN) based on format
//   - Validates description exists and is within length limits
func ValidateMigrationSignature(signature string, format SignatureFormat) error {
	if len(signature) > 255 {
		return fmt.Errorf("migration signature too long (max 255 characters)")
	}

	if len(signature) == 0 {
		return fmt.Errorf("migration signature cannot be empty")
	}

	// For "custom" format, only validate length and non-empty
	if format == SignatureFormatCustom {
		return nil
	}

	parts := strings.Split(signature, "_")

	switch format {
	case SignatureFormatDateTime:
		// Minimum 5 parts: YYYY, MM, DD, HHMM, description
		if len(parts) < 5 {
			return fmt.Errorf("datetime format must have at least 5 parts separated by underscores")
		}
		if err := validateDatePart(parts); err != nil {
			return err
		}
		if err := validateTimePart(parts[3]); err != nil {
			return err
		}
		return validateDescription(parts[4:])
	case SignatureFormatDate:
		// Minimum 5 parts: YYYY, MM, DD, NNN, description
		if len(parts) < 5 {
			return fmt.Errorf("date format must have at least 5 parts separated by underscores")
		}
		if err := validateDatePart(parts); err != nil {
			return err
		}
		if err := validateSequencePart(parts[3]); err != nil {
			return err
		}
		return validateDescription(parts[4:])
	case SignatureFormatUnix:
		// Minimum 2 parts: timestamp, description
		if len(parts) < 2 {
			return fmt.Errorf("unix format must be: timestamp_description")
		}
		if _, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
			return fmt.Errorf("invalid unix timestamp: %s", parts[0])
		}
		description := strings.Join(parts[1:], "_")
		if description == "" {
			return fmt.Errorf("description cannot be empty")
		}
		return nil
	default:
		return fmt.Errorf("unknown migration signature format: %s", format)
	}
}

// validateDatePart checks that the date part (YYYY, MM, DD) is valid.
//
// Business rules:
//   - YYYY must be 4 digits
//   - MM must be 2 digits (01-12)
//   - DD must be 2 digits (01-31)
//   - The date must be valid (e.g., 2025-02-30 is invalid)
//   - The date can be any valid date (past, present, or future)
func validateDatePart(parts []string) error {
	if len(parts[0]) != 4 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		return fmt.Errorf("date parts must be YYYY_MM_DD format")
	}

	_, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("year must be numeric: %w", err)
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("month must be numeric: %w", err)
	}
	if month < 1 || month > 12 {
		return fmt.Errorf("month must be between 01 and 12, got %02d", month)
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("day must be numeric: %w", err)
	}
	if day < 1 || day > 31 {
		return fmt.Errorf("day must be between 01 and 31, got %02d", day)
	}

	// Validate actual calendar date
	dateStr := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])
	_, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid calendar date: %w", err)
	}

	return nil
}

// validateTimePart checks that the time part (HHMM) is valid (0000-2359).
//
// Business Logic:
//   - Requires exactly 4 digits
//   - Parses as integer to extract hour and minute
//   - Hour must be between 00-23
//   - Minute must be between 00-59
func validateTimePart(part string) error {
	if len(part) != 4 {
		return fmt.Errorf("time part must be 4 digits (HHMM)")
	}

	num, err := strconv.Atoi(part)
	if err != nil {
		return fmt.Errorf("time part must be numeric: %w", err)
	}

	hour := num / 100
	minute := num % 100
	if hour < 0 || hour > 23 {
		return fmt.Errorf("hour must be between 00 and 23, got %02d", hour)
	}
	if minute < 0 || minute > 59 {
		return fmt.Errorf("minute must be between 00 and 59, got %02d", minute)
	}
	return nil
}

// validateSequencePart checks that the sequence part (NNN) is valid (000-999).
//
// Business Logic:
//   - Requires exactly 3 digits
//   - Parses as integer
//   - Sequence must be between 000-999
//   - Allows up to 999 migrations per day
func validateSequencePart(part string) error {
	if len(part) != 3 {
		return fmt.Errorf("sequence part must be 3 digits (NNN)")
	}

	sequence, err := strconv.Atoi(part)
	if err != nil {
		return fmt.Errorf("sequence part must be numeric: %w", err)
	}

	if sequence < 0 || sequence > 999 {
		return fmt.Errorf("sequence must be between 000 and 999, got %03d", sequence)
	}
	return nil
}

// validateDescription checks that the description exists and is within length limits.
//
// Business Logic:
//   - Joins all parts with underscores
//   - Description cannot be empty
//   - Maximum length is 200 characters
func validateDescription(parts []string) error {
	description := strings.Join(parts, "_")
	if len(description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}
	if len(description) > 200 {
		return fmt.Errorf("description too long (max 200 characters)")
	}
	return nil
}
