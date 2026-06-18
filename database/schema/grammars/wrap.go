package grammars

import (
	"fmt"

	"regexp"
	"strings"

	contractsdatabase "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/support/collect"
)

type Wrap struct {
	driver      contractsdatabase.Driver
	tablePrefix string
}

func NewWrap(driver contractsdatabase.Driver, tablePrefix string) *Wrap {
	return &Wrap{
		driver:      driver,
		tablePrefix: tablePrefix,
	}
}

func (r *Wrap) Column(column string) (string, error) {
	if strings.Contains(column, " as ") {
		return r.aliasedValue(column)
	}

	return r.Segments(strings.Split(column, "."))
}

func (r *Wrap) Columns(columns []string) ([]string, error) {
	formatedColumns := make([]string, len(columns))
	for i, column := range columns {
		var err error
		formatedColumns[i], err = r.Column(column)
		if err != nil {
			return nil, err
		}
	}

	return formatedColumns, nil
}

func (r *Wrap) Columnize(columns []string) (string, error) {
	formatedColumns, err := r.Columns(columns)
	if err != nil {
		return "", err
	}

	return strings.Join(formatedColumns, ", "), nil
}

func (r *Wrap) GetPrefix() string {
	return r.tablePrefix
}

func (r *Wrap) PrefixArray(prefix string, values []string) []string {
	return collect.Map(values, func(value string, _ int) string {
		return prefix + " " + value
	})
}

func (r *Wrap) Quote(value string) string {
	if value == "" {
		return "''"
	}

	// Escape single quotes by doubling them
	escaped := strings.ReplaceAll(value, "'", "''")

	return fmt.Sprintf("'%s'", escaped)
}

func (r *Wrap) Quotes(value []string) []string {
	return collect.Map(value, func(v string, _ int) string {
		if r.driver == contractsdatabase.DriverSqlserver {
			return "N" + r.Quote(v)
		}
		return r.Quote(v)
	})
}

func (r *Wrap) Segments(segments []string) (string, error) {
	for i, segment := range segments {
		if i == 0 && len(segments) > 1 {
			var err error
			segments[i], err = r.Table(segment)
			if err != nil {
				return "", err
			}
		} else {
			var err error
			segments[i], err = r.Value(segment)
			if err != nil {
				return "", err
			}
		}
	}

	return strings.Join(segments, "."), nil
}

func (r *Wrap) Table(table string) (string, error) {
	if strings.Contains(table, " as ") {
		return r.aliasedTable(table)
	}
	if strings.Contains(table, ".") {
		lastDotIndex := strings.LastIndex(table, ".")

		left, err := r.Value(table[:lastDotIndex])
		if err != nil {
			return "", err
		}
		right, err := r.Value(r.tablePrefix + table[lastDotIndex+1:])
		if err != nil {
			return "", err
		}

		return left + "." + right, nil
	}

	return r.Value(r.tablePrefix + table)
}

func (r *Wrap) Value(value string) (string, error) {
	if value == "*" {
		return value, nil
	}

	// Validate identifier to prevent injection
	if !r.isValidIdentifier(value) {
		return "", fmt.Errorf("invalid identifier: %s", value)
	}

	if r.driver == contractsdatabase.DriverMysql {
		return "`" + strings.ReplaceAll(value, "`", "``") + "`", nil
	}

	// For Oracle, uppercase identifiers to match default behavior
	// Oracle stores unquoted identifiers in uppercase
	if r.driver == contractsdatabase.DriverOracle {
		value = strings.ToUpper(value)
	}

	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`, nil
}

func (r *Wrap) aliasedTable(table string) (string, error) {
	segments := strings.Split(table, " as ")

	left, err := r.Table(segments[0])
	if err != nil {
		return "", err
	}
	right, err := r.Value(r.tablePrefix + segments[1])
	if err != nil {
		return "", err
	}

	return left + " as " + right, nil
}

func (r *Wrap) aliasedValue(value string) (string, error) {
	segments := strings.Split(value, " as ")

	left, err := r.Column(segments[0])
	if err != nil {
		return "", err
	}
	right, err := r.Value(r.tablePrefix + segments[1])
	if err != nil {
		return "", err
	}

	return left + " as " + right, nil
}

func (r *Wrap) isValidIdentifier(value string) bool {
	// Allow alphanumeric, underscores, and dots.
	// Dots are allowed for schema.table format, but should be handled by Segments/Table
	// for proper quoting. Here we allow them for backward compatibility.
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.]+$`, value)
	if !matched {
		return false
	}

	// Additional check: if value contains dots, ensure each segment is valid
	if strings.Contains(value, ".") {
		segments := strings.Split(value, ".")
		segmentRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		for _, segment := range segments {
			if segment == "" {
				return false // Empty segment (e.g., "table.")
			}
			// Recursively validate each segment without dots
			if !segmentRegex.MatchString(segment) { //nolint:staticcheck
				return false
			}
		}
	}

	return true
}

func (r *Wrap) IsValidAction(action string) bool {
	action = strings.ToUpper(action)
	switch action {
	case "CASCADE", "RESTRICT", "NO ACTION", "SET NULL", "SET DEFAULT":
		return true
	default:
		return false
	}
}
