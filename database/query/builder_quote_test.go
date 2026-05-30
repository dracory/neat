package query

import (
	"context"
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
