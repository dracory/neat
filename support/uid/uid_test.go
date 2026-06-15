package uid

import (
	"testing"
)

func TestGenerateShortID(t *testing.T) {
	id := GenerateShortID()
	if id == "" {
		t.Fatal("GenerateShortID returned empty string")
	}
	if len(id) != 11 {
		t.Fatalf("expected ID length 11, got %d: %s", len(id), id)
	}

	// Generate another ID and ensure they are different
	id2 := GenerateShortID()
	if id == id2 {
		t.Fatal("GenerateShortID returned duplicate IDs")
	}
}

func TestGenerateShortIDNoDuplicates(t *testing.T) {
	const count = 1000
	seen := make(map[string]struct{}, count)
	for i := 0; i < count; i++ {
		id := GenerateShortID()
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate ID generated at iteration %d: %s", i, id)
		}
		seen[id] = struct{}{}
	}
}

func TestEncodeCrockford(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{31, "z"},
		{32, "10"},
		{1024, "100"},
	}

	for _, tt := range tests {
		result := encodeCrockford(tt.input)
		if result != tt.expected {
			t.Errorf("encodeCrockford(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeID(t *testing.T) {
	if NormalizeID("  ABC123  ") != "abc123" {
		t.Error("NormalizeID failed to normalize")
	}
}

func TestIsShortID(t *testing.T) {
	if !IsShortID("abc123def45") {
		t.Error("expected 11-char ID to be short")
	}
	if !IsShortID("abc123def45abc123def4") {
		t.Error("expected 21-char ID to be short")
	}
	if IsShortID("abc123") {
		t.Error("expected 6-char ID to not be short")
	}
	if IsShortID("abc123def45abc123def45") {
		t.Error("expected 22-char ID to not be short")
	}
}
