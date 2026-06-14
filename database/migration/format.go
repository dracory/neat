package migration

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// MigrationIDFormat defines the format for migration IDs
type MigrationIDFormat string

const (
	// MigrationIDFormatDateTime uses timestamp-based format (default)
	MigrationIDFormatDateTime MigrationIDFormat = "datetime" // 2026_06_14_1200_description
	// MigrationIDFormatDate uses sequence-based format
	MigrationIDFormatDate MigrationIDFormat = "date" // 2026_06_14_001_description
	// MigrationIDFormatUnix uses unix timestamp format (legacy)
	MigrationIDFormatUnix MigrationIDFormat = "unix" // 1717080000_description
	// MigrationIDFormatCustom uses no prefix format restriction
	MigrationIDFormatCustom MigrationIDFormat = "custom" // any format
)

// ValidateTableName ensures the table name contains only safe characters
// This function is exported to allow external validation of table names
// before creating a migrator instance.
//
// Business Logic:
// - Rejects empty table names
// - Enforces maximum length of 64 characters
// - First character must be a letter or underscore (not a digit)
// - All characters must be alphanumeric or underscore
// - Prevents SQL injection and naming conflicts
func ValidateTableName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("table name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("table name too long (max 64 characters)")
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
	return nil
}

// ValidateMigrationID validates that the migration ID follows the specified format
// Supported formats:
//   - YYYY_MM_DD_HHMM_description (for datetime format)
//   - YYYY_MM_DD_NNN_description (for date format)
//   - unix_timestamp_description (for unix format)
//   - custom (no prefix format restriction)
//
// Business Logic:
// - Enforces maximum length of 255 characters
// - Rejects empty IDs
// - For "custom" format: only validates length and non-empty
// - For other formats: requires at least 5 underscore-separated parts (or 2 for unix)
// - Validates date part (YYYY_MM_DD) is a valid calendar date
// - Validates time part (HHMM) or sequence part (NNN) based on format
// - Validates description exists and is within length limits
// - Ensures lexicographical ordering by date/time
func ValidateMigrationID(id string, format MigrationIDFormat) error {
	if len(id) > 255 {
		return fmt.Errorf("migration ID too long (max 255 characters)")
	}

	if len(id) == 0 {
		return fmt.Errorf("migration ID cannot be empty")
	}

	// For "custom" format, only validate length and non-empty
	if format == MigrationIDFormatCustom {
		return nil
	}

	parts := strings.Split(id, "_")

	switch format {
	case MigrationIDFormatDateTime:
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
	case MigrationIDFormatDate:
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
	case MigrationIDFormatUnix:
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
		return fmt.Errorf("unknown migration ID format: %s", format)
	}
}

// validateDatePart checks that the date part (YYYY, MM, DD) is valid
//
// Business rules:
// - YYYY must be 4 digits
// - MM must be 2 digits (01-12)
// - DD must be 2 digits (01-31)
// - The date must be valid (e.g., 2025-02-30 is invalid)
// - The date can be any valid date (past, present, or future)
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

// validateTimePart checks that the time part (HHMM) is valid (00:00-23:59)
//
// Business Logic:
// - Requires exactly 4 digits
// - Parses as integer to extract hour and minute
// - Hour must be between 00-23
// - Minute must be between 00-59
// - Ensures valid 24-hour time format
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

// validateSequencePart checks that the sequence part (NNN) is valid (000-999)
//
// Business Logic:
// - Requires exactly 3 digits
// - Parses as integer
// - Sequence must be between 000-999
// - Allows up to 999 migrations per day
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

// validateDescription checks that the description exists and is within length limits
//
// Business Logic:
// - Joins all parts with underscores
// - Description cannot be empty
// - Maximum length is 200 characters
// - Ensures meaningful migration identifiers
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

// generateDateSequence generates a date-based sequence ID (YYYY_MM_DD_NNN)
// by checking existing migrations for the current date and incrementing
func (m *Migrator) generateDateSequence() (string, error) {
	now := time.Now()
	datePrefix := now.Format("2006_01_02")

	// Get existing migrations
	migrationFiles, err := m.getMigrations()
	if err != nil {
		return "", fmt.Errorf("failed to get migrations: %w", err)
	}

	// Find highest sequence number for today
	maxSequence := -1
	for name := range migrationFiles {
		if strings.HasPrefix(name, datePrefix) {
			parts := strings.Split(name, "_")
			if len(parts) >= 4 {
				if seq, err := strconv.Atoi(parts[3]); err == nil {
					if seq > maxSequence {
						maxSequence = seq
					}
				}
			}
		}
	}

	nextSequence := maxSequence + 1
	if nextSequence > 999 {
		return "", fmt.Errorf("too many migrations for date %s (max 999)", datePrefix)
	}

	return fmt.Sprintf("%s_%03d", datePrefix, nextSequence), nil
}
