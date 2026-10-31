package db

// Security-review proof tests for docs/security-review.md Findings #2, #3, #6.
// These tests FAIL on the current implementation: DSN builders interpolate
// credentials/parameters without escaping, allowing connection-parameter
// injection, and the exported Quote*/Value helpers do not escape the
// dialect's quote/escape characters.

import (
	"strings"
	"testing"

	contracts "github.com/dracory/neat/contracts/database"
)

// Finding #2: a PostgreSQL password containing a space injects arbitrary
// libpq connection parameters — e.g. a TLS downgrade via sslmode=disable.
func TestSecurityReviewPostgresDSNParamInjectionViaPassword(t *testing.T) {
	config := ConnectionConfig{
		Driver:   contracts.DriverPostgres,
		Host:     "localhost",
		Port:     5432,
		Database: "appdb",
		Username: "app",
		Password: "realpw sslmode=disable",
		SSLMode:  "verify-full",
	}

	dsn, err := NewConfigBuilder(config).BuildDSN()
	if err != nil {
		t.Fatalf("BuildDSN: %v", err)
	}

	// libpq treats the DSN as space-separated key=value pairs. An injected
	// "sslmode=disable" must never appear as a standalone parameter —
	// it would override the configured verify-full.
	if strings.Contains(dsn, " sslmode=disable") {
		t.Errorf("password broke out of its key=value pair and injected a parameter: %q", dsn)
	}
}

// Finding #2 (variant): Schema is appended as search_path without quoting;
// a hostile tenant schema value injects parameters.
func TestSecurityReviewPostgresDSNParamInjectionViaSchema(t *testing.T) {
	config := ConnectionConfig{
		Driver:   contracts.DriverPostgres,
		Host:     "localhost",
		Port:     5432,
		Database: "appdb",
		Username: "app",
		Password: "pw",
		Schema:   "public sslmode=disable",
	}

	dsn, err := NewConfigBuilder(config).BuildDSN()
	if err != nil {
		t.Fatalf("BuildDSN: %v", err)
	}

	if strings.Contains(dsn, "search_path=public sslmode=disable") {
		t.Errorf("schema value injected an extra libpq parameter: %q", dsn)
	}
}

// Finding #3: a MySQL password containing "@tcp(" injects a second network
// address into the go-sql-driver DSN. The driver format has no escaping for
// such characters, so the builder must either reject the credentials or emit
// a DSN with exactly one network address.
func TestSecurityReviewMySQLDSNAddressInjection(t *testing.T) {
	config := ConnectionConfig{
		Driver:   contracts.DriverMysql,
		Host:     "db.internal",
		Port:     3306,
		Database: "appdb",
		Username: "app",
		Password: "pw@tcp(evil.example:3307)/x",
	}

	dsn, err := NewConfigBuilder(config).BuildDSN()
	if err != nil {
		// Rejecting unencodable credentials is also a correct fix.
		return
	}

	// go-sql-driver scans for the LAST '/' and LAST '@', so the injected
	// network text must land inside the password — the real address must
	// win and the password must round-trip intact. Replicated here without
	// importing the driver (it must only be referenced from the MySQL
	// integration tests).
	slash := strings.LastIndex(dsn, "/")
	at := strings.LastIndex(dsn[:slash], "@")
	userinfo, rest := dsn[:at], dsn[at+1:]
	pass := userinfo[strings.Index(userinfo, ":")+1:]
	addr := rest[strings.Index(rest, "(")+1 : strings.LastIndex(rest, ")")]

	if addr != "db.internal:3306" {
		t.Errorf("injected network address won: addr=%q (dsn %q)", addr, dsn)
	}
	if pass != "pw@tcp(evil.example:3307)/x" {
		t.Errorf("password did not round-trip: %q", pass)
	}
}

// Finding #3 (variant): SQL Server database name containing "&encrypt="
// appends a query parameter that disables encryption.
func TestSecurityReviewSQLServerDSNParamInjection(t *testing.T) {
	config := ConnectionConfig{
		Driver:   contracts.DriverSqlserver,
		Host:     "db.internal",
		Port:     1433,
		Database: "appdb&encrypt=disable",
		Username: "app",
		Password: "pw",
	}

	dsn, err := NewConfigBuilder(config).BuildDSN()
	if err != nil {
		t.Fatalf("BuildDSN: %v", err)
	}

	if strings.Contains(dsn, "&encrypt=disable") {
		t.Errorf("database name injected an encrypt=disable parameter: %q", dsn)
	}
}

// Finding #6: QuoteValue does not escape backslashes, which are escape
// characters inside MySQL string literals — \' escapes the quote and the
// remainder becomes raw SQL.
func TestSecurityReviewQuoteValueMySQLBackslash(t *testing.T) {
	got := QuoteValue(`x\' OR 1=1 -- `, "mysql")

	// Safe MySQL escaping doubles backslashes and quotes: \→\\ then '→'',
	// producing 'x\\'' OR 1=1 -- ' (the \\ is an escaped backslash, the ''
	// an escaped quote — the literal cannot be terminated early).
	want := `'x\\'' OR 1=1 -- '`
	if got != want {
		t.Errorf("QuoteValue did not safely escape MySQL literal: got %q, want %q", got, want)
	}
}

// Finding #6 (variant): QuoteIdentifier does not escape the SQL Server
// closing bracket — [a]b produces unbalanced quoting.
func TestSecurityReviewQuoteIdentifierSQLServerBracket(t *testing.T) {
	got := QuoteIdentifier("a]b", "sqlserver")

	// Expect the inner ] to be escaped as ]]: [a]]b]
	if got != "[a]]b]" {
		t.Errorf("QuoteIdentifier produced unbalanced bracket quoting: %q", got)
	}
}
