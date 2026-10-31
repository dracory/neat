package query

import (
	"context"
	"regexp"
	"strings"
	"testing"
)

func TestQuoteIdentifier(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	// Test with nil driver
	result := b.quoteIdentifier("test")
	if result != "test" {
		t.Errorf("Expected 'test', got %q", result)
	}

	// Test with wildcard
	result = b.quoteIdentifier("*")
	if result != "*" {
		t.Errorf("Expected '*', got %q", result)
	}

	// Test with empty string
	result = b.quoteIdentifier("")
	if result != "" {
		t.Errorf("Expected '', got %q", result)
	}
}

func TestQuoteWhereIdentifiers(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple equals", "name = ?", "name = ?"},
		{"is null", "deleted_at IS NULL", "deleted_at IS NULL"},
		{"sql keyword", "AND name = ?", "AND name = ?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := b.quoteWhereIdentifiers(tt.input)
			if result != tt.expected {
				t.Errorf("quoteWhereIdentifiers(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Complete pre-quoted identifiers — including per-part quoted dotted
// names and doubled-quote escapes — keep passing through untouched.
func TestQuoteIdentifierPreQuotedPassthrough(t *testing.T) {
	drv := &FakeDriver{DialectName: "postgres"}
	q := NewQuery(context.TODO(), nil, drv, "", nil, nil)
	b := NewBuilder(q)

	for _, name := range []string{`"name"`, `"a"."b"`, `"a""b"`, `"a.b"`} {
		if got := b.quoteIdentifier(name); got != name {
			t.Errorf("quoteIdentifier(%q) = %q, expected passthrough", name, got)
		}
	}
}

// A string that merely starts and ends with a quote character is not a
// quoted identifier. quoteIdentifier must not emit it raw; every fragment
// it produces must stay inside quotes so nothing can break out.
func TestQuoteIdentifierRejectsMalformedPreQuoted(t *testing.T) {
	drv := &FakeDriver{DialectName: "postgres"}
	q := NewQuery(context.TODO(), nil, drv, "", nil, nil)
	b := NewBuilder(q)

	quotedSpan := regexp.MustCompile(`"([^"]|"")*"`)
	payloads := []string{
		`"a" = 1, "b"`,
		`"nickname" = (SELECT password FROM users WHERE id = 1), "bio"`,
		`"a", "b"`,
	}
	for _, payload := range payloads {
		got := b.quoteIdentifier(payload)
		if got == payload {
			t.Fatalf("crafted pre-quoted name passed through raw: %s", got)
		}
		if stripped := quotedSpan.ReplaceAllString(got, ""); strings.Contains(stripped, "=") {
			t.Fatalf("unquoted '=' leaked through quoting of %q: %s", payload, got)
		}
	}
}
