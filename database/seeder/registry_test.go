package seeder

import (
	"testing"
)

// Mock seeder for testing
type mockSeeder struct {
	signature string
	runCalled bool
}

func (m *mockSeeder) Signature() string {
	return m.signature
}

func (m *mockSeeder) Run() error {
	m.runCalled = true
	return nil
}

func TestRegisterSeeder(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	seeder1 := &mockSeeder{signature: "test_seeder_1"}
	seeder2 := &mockSeeder{signature: "test_seeder_2"}

	RegisterSeeder("test_seeder_1", seeder1)
	RegisterSeeder("test_seeder_2", seeder2)

	retrieved1 := GetSeeder("test_seeder_1")
	if retrieved1 == nil {
		t.Fatal("Expected seeder to be registered")
	}
	if retrieved1.Signature() != "test_seeder_1" {
		t.Errorf("Expected signature 'test_seeder_1', got '%s'", retrieved1.Signature())
	}

	retrieved2 := GetSeeder("test_seeder_2")
	if retrieved2 == nil {
		t.Fatal("Expected seeder to be registered")
	}
	if retrieved2.Signature() != "test_seeder_2" {
		t.Errorf("Expected signature 'test_seeder_2', got '%s'", retrieved2.Signature())
	}
}

func TestGetSeeder(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	seeder1 := &mockSeeder{signature: "test_seeder"}
	RegisterSeeder("test_seeder", seeder1)

	// Test existing seeder
	retrieved := GetSeeder("test_seeder")
	if retrieved == nil {
		t.Fatal("Expected seeder to be found")
	}
	if retrieved.Signature() != "test_seeder" {
		t.Errorf("Expected signature 'test_seeder', got '%s'", retrieved.Signature())
	}

	// Test non-existent seeder
	nonExistent := GetSeeder("non_existent")
	if nonExistent != nil {
		t.Error("Expected nil for non-existent seeder")
	}
}

func TestGetSeeders(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	seeder1 := &mockSeeder{signature: "seeder_1"}
	seeder2 := &mockSeeder{signature: "seeder_2"}
	seeder3 := &mockSeeder{signature: "seeder_3"}

	RegisterSeeder("seeder_1", seeder1)
	RegisterSeeder("seeder_2", seeder2)
	RegisterSeeder("seeder_3", seeder3)

	seeders := GetSeeders()
	if len(seeders) != 3 {
		t.Errorf("Expected 3 seeders, got %d", len(seeders))
	}

	signatures := make(map[string]bool)
	for _, s := range seeders {
		signatures[s.Signature()] = true
	}

	if !signatures["seeder_1"] || !signatures["seeder_2"] || !signatures["seeder_3"] {
		t.Error("Expected all seeders to be present")
	}
}

func TestClearRegistry(t *testing.T) {
	ClearRegistry()

	seeder1 := &mockSeeder{signature: "test_seeder"}
	RegisterSeeder("test_seeder", seeder1)

	// Verify seeder is registered
	retrieved := GetSeeder("test_seeder")
	if retrieved == nil {
		t.Fatal("Expected seeder to be registered")
	}

	// Clear registry
	ClearRegistry()

	// Verify seeder is gone
	retrieved = GetSeeder("test_seeder")
	if retrieved != nil {
		t.Error("Expected seeder to be cleared")
	}
}

func TestRegisterSeederOverwrite(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	seeder1 := &mockSeeder{signature: "original"}
	seeder2 := &mockSeeder{signature: "updated"}

	RegisterSeeder("test_seeder", seeder1)
	RegisterSeeder("test_seeder", seeder2)

	retrieved := GetSeeder("test_seeder")
	if retrieved == nil {
		t.Fatal("Expected seeder to be registered")
	}
	if retrieved.Signature() != "updated" {
		t.Errorf("Expected signature 'updated' after overwrite, got '%s'", retrieved.Signature())
	}
}
