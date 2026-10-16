package uid

import (
	"testing"
)

func TestGenerateShortID(t *testing.T) {
	id := GenerateShortID()
	if id == "" {
		t.Fatal("GenerateShortID returned empty string")
	}
	if len(id) != 14 {
		t.Fatalf("expected ID length 14, got %d: %s", len(id), id)
	}

	// Generate another ID and ensure they are different
	id2 := GenerateShortID()
	if id == id2 {
		t.Fatal("GenerateShortID returned duplicate IDs")
	}
}

func TestSafeShortID(t *testing.T) {
	id := SafeShortID()
	if id == "" {
		t.Fatal("SafeShortID returned empty string")
	}
	if len(id) != 14 {
		t.Fatalf("expected ID length 14, got %d: %s", len(id), id)
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

func TestNormalizeID(t *testing.T) {
	if NormalizeID("  ABC123  ") != "abc123" {
		t.Error("NormalizeID failed to normalize")
	}
}

func TestIsShortID(t *testing.T) {
	if !IsShortID("abc123def45") {
		t.Error("expected 11-char ID to be short")
	}
	if !IsShortID("05fr88tc139r8t") {
		t.Error("expected 14-char ID to be short")
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
