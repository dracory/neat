package query

import (
	"testing"
)

func TestNewBuilder(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	b := NewBuilder(q)
	if b == nil {
		t.Error("Expected non-nil Builder")
	}
	if b.query != q {
		t.Error("Expected Builder to have the provided query")
	}
}

func TestQuoteIdentifier(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
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
