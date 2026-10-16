package uid

import (
	"testing"
	"time"
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

// TestGenerateShortIDBackwardsClock verifies that generated IDs remain
// strictly increasing even when the wall clock steps backwards. It overrides
// the nowMicro seam, so it must not run in parallel and must restore both
// nowMicro and shortIDAtomicState.
func TestGenerateShortIDBackwardsClock(t *testing.T) {
	oldNow := nowMicro
	oldState := shortIDAtomicState.Swap(0)
	defer func() {
		nowMicro = oldNow
		shortIDAtomicState.Store(oldState)
	}()

	now := time.Now().UnixMicro()
	nowMicro = func() int64 { return now }

	prev := GenerateShortID()
	for i := 0; i < 100; i++ {
		now -= 1000 // step the clock backwards on every iteration
		id := GenerateShortID()
		if id <= prev {
			t.Fatalf("IDs regressed after a backwards clock step: %q then %q", prev, id)
		}
		prev = id
	}
}

// TestGenerateShortIDCounterOverflow verifies that when more than 16 IDs are
// generated within the same microsecond, the synthetic timestamp bump keeps
// IDs strictly increasing instead of stalling or colliding.
func TestGenerateShortIDCounterOverflow(t *testing.T) {
	oldNow := nowMicro
	oldState := shortIDAtomicState.Swap(0)
	defer func() {
		nowMicro = oldNow
		shortIDAtomicState.Store(oldState)
	}()

	nowMicro = func() int64 { return 1_000_000 } // frozen clock forces same-microsecond IDs

	prev := GenerateShortID()
	for i := 0; i < 40; i++ { // more than the 16 counter slots per microsecond
		id := GenerateShortID()
		if id <= prev {
			t.Fatalf("IDs regressed after counter overflow: %q then %q", prev, id)
		}
		prev = id
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
	if !IsShortID("05fh88gd139f8h") {
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
	if IsShortID("!!!!!!!!!!!") {
		t.Error("expected 11-char ID with invalid chars to not be short")
	}
	if IsShortID("05ae88aa139a8e") {
		t.Error("expected 14-char ID with vowels 'a'/'e' to not be short")
	}
	if !IsShortID("V1StGXR8_Z5jdHi6B-myT") {
		t.Error("expected 21-char NanoID to be short")
	}
	if IsShortID("V1StGXR8_Z5jdHi6B-my!") {
		t.Error("expected 21-char ID with invalid char to not be short")
	}
}
