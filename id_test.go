package neat_test

import (
	"testing"

	"github.com/dracory/neat"
)

func TestGenerateID(t *testing.T) {
	id1 := neat.GenerateID()
	if len(id1) != 14 {
		t.Fatalf("expected ID length 14, got %d (%s)", len(id1), id1)
	}

	id2 := neat.GenerateID()
	if len(id2) != 14 {
		t.Fatalf("expected ID length 14, got %d (%s)", len(id2), id2)
	}

	if id1 == id2 {
		t.Fatalf("expected unique IDs, got duplicate %s", id1)
	}
}
