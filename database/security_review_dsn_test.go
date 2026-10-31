package database

// Security-review proof tests for docs/security-review.md Finding #1.
// These tests FAIL on the current implementation: redactDSN only handles
// URL-style DSNs containing "://" and leaves keyword/value (PostgreSQL) and
// driver-native MySQL DSNs unredacted, so passwords leak into error messages
// produced by NewFromDSN (database/db.go:210, :514).

import (
	"strings"
	"testing"
)

// Finding #1a: PostgreSQL keyword/value DSNs are returned unredacted.
func TestSecurityReviewRedactDSNPostgresKeywordValue(t *testing.T) {
	dsn := "host=localhost port=5432 user=app password=S3cr3tPW dbname=appdb sslmode=require"

	redacted := redactDSN(dsn)

	if strings.Contains(redacted, "S3cr3tPW") {
		t.Errorf("redactDSN leaked password in keyword/value DSN: %q", redacted)
	}
}

// Finding #1b: go-sql-driver native MySQL DSNs are returned unredacted.
// Format: user:pass@tcp(host:port)/db?params — no "://" scheme.
func TestSecurityReviewRedactDSNMySQLNative(t *testing.T) {
	dsn := "appuser:S3cr3tPW@tcp(db.internal:3306)/appdb?parseTime=true"

	redacted := redactDSN(dsn)

	if strings.Contains(redacted, "S3cr3tPW") {
		t.Errorf("redactDSN leaked password in native MySQL DSN: %q", redacted)
	}
}

// Finding #1c: NewFromDSN embeds the (unredacted) DSN in parse errors.
// A malformed native MySQL DSN produces an error containing the password.
func TestSecurityReviewNewFromDSNErrorLeaksPassword(t *testing.T) {
	// Unparseable DSN that still carries a credential.
	dsn := "appuser:S3cr3tPW@tcp(%%%invalid%%%)/appdb"

	_, err := NewFromDSN(dsn)
	if err == nil {
		t.Skip("NewFromDSN unexpectedly succeeded on malformed DSN")
	}
	if strings.Contains(err.Error(), "S3cr3tPW") {
		t.Errorf("NewFromDSN error leaked password: %q", err.Error())
	}
}

// Control: URL-style DSNs ARE redacted today — this test should already pass
// and guards against regressions while the gaps above are fixed.
func TestSecurityReviewRedactDSNUrlStyleAlreadyRedacted(t *testing.T) {
	dsn := "postgres://appuser:S3cr3tPW@db.internal:5432/appdb"

	redacted := redactDSN(dsn)

	if strings.Contains(redacted, "S3cr3tPW") {
		t.Errorf("redactDSN leaked password in URL DSN: %q", redacted)
	}
}

// Regression: a password containing '@' must be redacted in full — matching
// the last-'@' boundary both libpq URLs and go-sql-driver use.
func TestSecurityReviewRedactDSNPasswordContainingAt(t *testing.T) {
	for _, dsn := range []string{
		"appuser:pa@ss@tcp(db.internal:3306)/appdb",
		"postgres://appuser:pa@ss@db.internal:5432/appdb",
	} {
		redacted := redactDSN(dsn)
		if strings.Contains(redacted, "ss@") && !strings.Contains(redacted, "[REDACTED]@") {
			t.Errorf("redactDSN leaked password tail in %q -> %q", dsn, redacted)
		}
		if strings.Contains(redacted, "pa@ss") {
			t.Errorf("redactDSN leaked password in %q -> %q", dsn, redacted)
		}
	}
}
