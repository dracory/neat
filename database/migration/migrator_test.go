package migration

import (
	"os"
	"path/filepath"
	"testing"

	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

func TestMigrator_Create(t *testing.T) {
	testDir := "./test_migrations"
	migrator := &Migrator{
		paths: []string{testDir},
	}

	// Create directory before test
	os.MkdirAll(testDir, 0755)

	// Cleanup after test
	defer os.RemoveAll(testDir)

	err := migrator.Create("test_migration")
	if err != nil {
		t.Errorf("Create() failed: %v", err)
	}
}

func TestMigrator_Create_NoPaths(t *testing.T) {
	migrator := &Migrator{
		paths: []string{},
	}

	err := migrator.Create("test_migration")
	if err == nil {
		t.Error("Expected error when no paths configured")
	}
}

func TestMigrator_Create_WithSpaces(t *testing.T) {
	testDir := "./test_migrations"
	migrator := &Migrator{
		paths: []string{testDir},
	}

	// Create directory before test
	os.MkdirAll(testDir, 0755)

	// Cleanup after test
	defer os.RemoveAll(testDir)

	err := migrator.Create("test migration with spaces")
	if err != nil {
		t.Errorf("Create() with spaces failed: %v", err)
	}

	// Verify file was created with underscores instead of spaces
	files, err := filepath.Glob(filepath.Join(testDir, "*_test_migration_with_spaces.go"))
	if err != nil {
		t.Errorf("Failed to glob files: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected migration file to be created with underscores")
	}
}

func TestClearRegistry(t *testing.T) {
	// Register a migration
	migration := Migration{
		Up:   func(schema contractsschema.Schema) error { return nil },
		Down: func(schema contractsschema.Schema) error { return nil },
	}
	RegisterMigration("test", migration)

	// Clear registry
	ClearRegistry()

	// Verify registry is empty
	registryMutex.RLock()
	_, ok := migrationRegistry["test"]
	registryMutex.RUnlock()

	if ok {
		t.Error("Expected registry to be cleared")
	}
}

func TestMigration_Struct(t *testing.T) {
	migration := Migration{
		Up: func(schema contractsschema.Schema) error {
			return nil
		},
		Down: func(schema contractsschema.Schema) error {
			return nil
		},
	}

	if migration.Up == nil {
		t.Error("Expected Up function to be set")
	}
	if migration.Down == nil {
		t.Error("Expected Down function to be set")
	}
}
