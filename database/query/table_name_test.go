package query

import (
	"context"
	"strings"
	"testing"
)

// newTestQuery builds a Query with a FakeDriver of the given dialect.
func newTableTestQuery(dialect string) *Query {
	return NewQuery(context.TODO(), nil, &FakeDriver{DialectName: dialect}, "default", nil, nil)
}

// TestTableReservedKeywords verifies that reserved SQL keywords are accepted
// as table names in Table() and are emitted quoted in the built SQL.
func TestTableReservedKeywords(t *testing.T) {
	keywords := []string{"group", "order", "select", "where", "user", "table"}

	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			q := newTableTestQuery("sqlite")
			q.Table(kw)
			if q.table != kw {
				t.Fatalf("Table(%q) was rejected, q.table = %q", kw, q.table)
			}
			if q.buildError != nil {
				t.Fatalf("Table(%q) set buildError: %v", kw, q.buildError)
			}
			sql, _ := NewBuilder(q).BuildSelect()
			if !strings.Contains(sql, `FROM "`+kw+`"`) {
				t.Errorf("expected quoted table name in SQL, got: %s", sql)
			}
		})
	}
}

// TestTableReservedKeywordMySQL verifies backtick quoting for MySQL.
func TestTableReservedKeywordMySQL(t *testing.T) {
	q := newTableTestQuery("mysql")
	q.Table("group")
	sql, _ := NewBuilder(q).BuildSelect()
	if !strings.Contains(sql, "FROM `group`") {
		t.Errorf("expected backtick-quoted table name, got: %s", sql)
	}
}

// TestTableReservedKeywordMixedCase verifies case variants are accepted.
func TestTableReservedKeywordMixedCase(t *testing.T) {
	q := newTableTestQuery("sqlite")
	q.Table("Group")
	if q.table != "Group" {
		t.Fatalf("Table(\"Group\") was rejected, q.table = %q", q.table)
	}
}

// TestTableReservedKeywordWithAlias verifies keyword names work with aliases.
func TestTableReservedKeywordWithAlias(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"group g", `FROM "group" "g"`},
		{"group AS g", `FROM "group" AS "g"`},
		{"group grp", `FROM "group" "grp"`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			q := newTableTestQuery("sqlite")
			q.Table(tt.input)
			if q.buildError != nil {
				t.Fatalf("Table(%q) set buildError: %v", tt.input, q.buildError)
			}
			if q.table != tt.input {
				t.Fatalf("Table(%q) was rejected, q.table = %q", tt.input, q.table)
			}
			sql, _ := NewBuilder(q).BuildSelect()
			if !strings.Contains(sql, tt.expected) {
				t.Errorf("expected %q in SQL, got: %s", tt.expected, sql)
			}
		})
	}
}

// TestTableInvalidNamesSetBuildError verifies that genuinely invalid table
// names are rejected AND surface a buildError instead of silently keeping a
// stale (e.g. model-derived) table name.
func TestTableInvalidNamesSetBuildError(t *testing.T) {
	invalid := []string{
		"",
		"users; DROP TABLE users",
		"user' OR '1'='1",
		"users--",
		"users/*comment*/",
		"schema.users",
		"123table",
		"users u extra junk",
		`users"`,
		"users,accounts",
	}
	for _, name := range invalid {
		t.Run(name, func(t *testing.T) {
			q := newTableTestQuery("sqlite")
			q.table = "stale_table"
			q.Table(name)
			if q.buildError == nil {
				t.Errorf("Table(%q) should set buildError", name)
			}
			if q.table != "stale_table" {
				t.Errorf("Table(%q) should not overwrite q.table, got %q", name, q.table)
			}
		})
	}
}

// TestTableOverridesModelDerivedName verifies a keyword table name overrides
// the name resolved from Model() (previously it silently fell back).
func TestTableOverridesModelDerivedName(t *testing.T) {
	type Group struct{ ID int }
	q := newTableTestQuery("sqlite")
	q.Model(&Group{})
	q.Table("group")
	if q.table != "group" {
		t.Fatalf("expected q.table = \"group\", got %q", q.table)
	}
	sql, _ := NewBuilder(q).BuildSelect()
	if !strings.Contains(sql, `FROM "group"`) {
		t.Errorf("expected quoted table name in SQL, got: %s", sql)
	}
}

// TestIsTableIdentifier unit-tests the relaxed table-name validator.
func TestIsTableIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"users", true},
		{"group", true},
		{"ORDER", true},
		{"Group", true},
		{"_temp", true},
		{"users_2024", true},
		{"", false},
		{"123table", false},
		{"schema.users", false},
		{"users;DROP", false},
		{"user accounts", false},
		{"users'", false},
		{`users"`, false},
		{"COUNT()", false},
		{"users--", false},
	}
	for _, tt := range tests {
		if got := isTableIdentifier(tt.input); got != tt.want {
			t.Errorf("isTableIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
