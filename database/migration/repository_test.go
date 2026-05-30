package migration

import (
	"testing"

	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

func TestGetTimestamp(t *testing.T) {
	timestamp := getTimestamp()
	if timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !contains(slice, "a") {
		t.Error("Expected true for existing element")
	}

	if contains(slice, "d") {
		t.Error("Expected false for non-existing element")
	}

	if contains([]string{}, "a") {
		t.Error("Expected false for empty slice")
	}
}

func TestRegisterMigration(t *testing.T) {
	// Clear registry before test
	ClearRegistry()
	defer ClearRegistry()

	name := "test_migration"
	migration := Migration{
		Up:   func(schema contractsschema.Schema) error { return nil },
		Down: func(schema contractsschema.Schema) error { return nil },
	}

	RegisterMigration(name, migration)

	registryMutex.RLock()
	_, ok := migrationRegistry[name]
	registryMutex.RUnlock()

	if !ok {
		t.Error("Expected migration to be registered")
	}
}
