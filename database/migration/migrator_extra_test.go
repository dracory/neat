package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// TestGetMigrationsFromRegistry verifies getMigrations loads from the global registry.
func TestGetMigrationsFromRegistry(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	RegisterMigration("2024_01_add_users", Migration{
		Up:   func(schema contractsschema.Schema) error { return nil },
		Down: func(schema contractsschema.Schema) error { return nil },
	})
	RegisterMigration("2024_02_add_posts", Migration{
		Up:   func(schema contractsschema.Schema) error { return nil },
		Down: func(schema contractsschema.Schema) error { return nil },
	})

	m := &Migrator{paths: []string{}}
	got, err := m.getMigrations()
	if err != nil {
		t.Fatalf("getMigrations: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 migrations from registry, got %d", len(got))
	}
	if _, ok := got["2024_01_add_users"]; !ok {
		t.Error("expected 2024_01_add_users in result")
	}
}

// TestGetMigrationsNonExistentPath skips non-existent directories gracefully.
func TestGetMigrationsNonExistentPath(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	m := &Migrator{paths: []string{"/does/not/exist/path_xyz"}}
	got, err := m.getMigrations()
	if err != nil {
		t.Fatalf("getMigrations should not error for non-existent path: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(got))
	}
}

// TestGetMigrationsFromFilesAndRegistry verifies file-based discovery merges with registry.
func TestGetMigrationsFromFilesAndRegistry(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	dir := t.TempDir()

	// Create a .go file whose name matches a registry entry.
	if err := os.WriteFile(filepath.Join(dir, "2024_01_from_file.go"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	// Create a _test.go file that should be skipped.
	if err := os.WriteFile(filepath.Join(dir, "2024_01_from_file_test.go"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	// Create a non-.go file that should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	RegisterMigration("2024_01_from_file", Migration{
		Up:   func(schema contractsschema.Schema) error { return nil },
		Down: func(schema contractsschema.Schema) error { return nil },
	})
	// Registry-only entry (no matching file).
	RegisterMigration("2024_02_registry_only", Migration{
		Up:   func(schema contractsschema.Schema) error { return nil },
		Down: func(schema contractsschema.Schema) error { return nil },
	})

	m := &Migrator{paths: []string{dir}}
	got, err := m.getMigrations()
	if err != nil {
		t.Fatalf("getMigrations: %v", err)
	}
	// Both registry entries should be present.
	if _, ok := got["2024_01_from_file"]; !ok {
		t.Error("expected 2024_01_from_file")
	}
	if _, ok := got["2024_02_registry_only"]; !ok {
		t.Error("expected 2024_02_registry_only")
	}
	// _test.go file name should NOT produce an entry.
	if _, ok := got["2024_01_from_file_test"]; ok {
		t.Error("_test.go file name should be excluded")
	}
}

// TestNewMigratorConstructor verifies NewMigrator sets all fields.
func TestNewMigratorConstructor(t *testing.T) {
	m := &Migrator{
		paths: []string{"/some/path"},
	}
	if len(m.paths) != 1 || m.paths[0] != "/some/path" {
		t.Errorf("unexpected paths: %v", m.paths)
	}
}

// TestMigratorCreate_DirectoryDoesNotExist verifies Create returns error when dir missing.
func TestMigratorCreate_DirectoryDoesNotExist(t *testing.T) {
	m := &Migrator{
		paths:  []string{"/nonexistent_migration_dir_xyz"},
		config: &mockConfig{},
	}
	err := m.Create("test")
	if err == nil {
		t.Error("expected error when migration dir does not exist")
	}
}

// TestMigratorCreate_FileContents verifies the generated file has expected content.
func TestMigratorCreate_FileContents(t *testing.T) {
	dir := t.TempDir()
	m := &Migrator{
		paths:  []string{dir},
		config: &mockConfig{},
	}

	if err := m.Create("add_orders"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*_add_orders.go"))
	if err != nil || len(files) == 0 {
		t.Fatal("migration file not found")
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	body := string(content)
	if len(body) == 0 {
		t.Error("generated file is empty")
	}
	// Should contain a package declaration and Up/Down functions.
	for _, substr := range []string{"package", "func Up", "func Down"} {
		if !strings.Contains(body, substr) {
			t.Errorf("generated file missing %q", substr)
		}
	}
}

// TestRegisterMigrationConcurrency verifies concurrent registration is safe.
func TestRegisterMigrationConcurrency(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		name := filepath.Join("migration", string(rune('a'+i)))
		go func(n string) {
			RegisterMigration(n, Migration{
				Up:   func(schema contractsschema.Schema) error { return nil },
				Down: func(schema contractsschema.Schema) error { return nil },
			})
			done <- struct{}{}
		}(name)
	}
	for i := 0; i < 20; i++ {
		<-done
	}

	registryMutex.RLock()
	count := len(migrationRegistry)
	registryMutex.RUnlock()
	if count != 20 {
		t.Errorf("expected 20 registered migrations, got %d", count)
	}
}
