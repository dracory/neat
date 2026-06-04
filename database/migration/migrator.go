package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dracory/neat/contracts/config"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/contracts/migration"
	"github.com/dracory/neat/database/schema"
)

// Migrator handles database migrations.
type Migrator struct {
	config     config.Config
	orm        contractsorm.Orm
	repository *Repository
	schema     *schema.Schema
	paths      []string
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(config config.Config, orm contractsorm.Orm, schema *schema.Schema, paths []string) *Migrator {
	return &Migrator{
		config:     config,
		orm:        orm,
		repository: NewRepository(config, orm),
		schema:     schema,
		paths:      paths,
	}
}

func (m *Migrator) Create(name string) error {
	if len(m.paths) == 0 {
		return fmt.Errorf("no migration paths configured")
	}

	// Generate timestamp prefix
	timestamp := fmt.Sprintf("%d", getTimestamp())
	migrationName := timestamp + "_" + strings.ReplaceAll(name, " ", "_")

	// Use the first path for creating migrations
	path := m.paths[0]

	// Create migration file
	filePath := filepath.Join(path, migrationName+".go")

	content := `package migrations

import (
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// Up applies the migration
func Up(schema contractsschema.Schema) error {
	// Add your migration logic here
	return nil
}

// Down rolls back the migration
func Down(schema contractsschema.Schema) error {
	// Add your rollback logic here
	return nil
}
`

	return os.WriteFile(filePath, []byte(content), 0600)
}

func (m *Migrator) Fresh() error {
	// Drop all tables
	tables := m.schema.GetTableListing()
	for _, table := range tables {
		if table != m.repository.table {
			if err := m.schema.DropIfExists(table); err != nil {
				return fmt.Errorf("failed to drop table %s: %w", table, err)
			}
		}
	}

	// Delete migration repository
	if err := m.repository.DeleteRepository(); err != nil {
		return fmt.Errorf("failed to delete migration repository: %w", err)
	}

	// Run all migrations
	return m.Run()
}

func (m *Migrator) Reset() error {
	// Ensure repository exists
	if err := m.repository.CreateRepository(); err != nil {
		return fmt.Errorf("failed to create migration repository: %w", err)
	}

	// Rollback all migrations
	maxIterations := 1000 // Safety limit to prevent infinite loops
	completed := false
	for i := 0; i < maxIterations; i++ {
		files, err := m.repository.GetLast()
		if err != nil {
			return fmt.Errorf("failed to get last migration: %w", err)
		}

		if len(files) == 0 {
			completed = true
			break
		}

		if err := m.Rollback(0, 0); err != nil {
			return fmt.Errorf("failed to rollback: %w", err)
		}
	}

	// Check if we hit the safety limit without completing
	if !completed {
		return fmt.Errorf("reset hit safety limit (%d iterations) without completing rollback", maxIterations)
	}

	// Run all migrations
	return m.Run()
}

func (m *Migrator) Rollback(step, batch int) error {
	// Ensure repository exists
	if err := m.repository.CreateRepository(); err != nil {
		return fmt.Errorf("failed to create migration repository: %w", err)
	}

	var files []migration.File
	var err error

	if batch > 0 {
		files, err = m.repository.GetMigrationsByBatch(batch)
	} else if step > 0 {
		files, err = m.repository.GetMigrationsByStep(step)
	} else {
		files, err = m.repository.GetLast()
	}

	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	// Get migration files from disk
	migrationFiles, err := m.getMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Rollback in reverse order
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]

		// Find the migration file
		migrationFunc, ok := migrationFiles[file.Migration]
		if !ok {
			return fmt.Errorf("migration file not found: %s", file.Migration)
		}

		// Run Down method
		if err := migrationFunc.Down(m.schema); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", file.Migration, err)
		}

		// Remove from repository
		if err := m.repository.Delete(file.Migration); err != nil {
			return fmt.Errorf("failed to delete migration from repository: %w", err)
		}
	}

	return nil
}

func (m *Migrator) Run() error {
	// Ensure repository exists
	if err := m.repository.CreateRepository(); err != nil {
		return fmt.Errorf("failed to create migration repository: %w", err)
	}

	// Get ran migrations
	ran, err := m.repository.GetRan()
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %w", err)
	}

	// Get migration files from disk
	migrationFiles, err := m.getMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get next batch number
	batch, err := m.repository.GetNextBatchNumber()
	if err != nil {
		return fmt.Errorf("failed to get next batch number: %w", err)
	}

	// Sort migration names to ensure correct execution order
	var names []string
	for name := range migrationFiles {
		names = append(names, name)
	}
	sort.Strings(names)

	// Run pending migrations
	for _, name := range names {
		// Skip if already ran
		if contains(ran, name) {
			continue
		}

		migrationFunc := migrationFiles[name]

		// Run Up method
		if err := migrationFunc.Up(m.schema); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", name, err)
		}

		// Log migration
		if err := m.repository.Log(name, batch); err != nil {
			return fmt.Errorf("failed to log migration: %w", err)
		}
	}

	return nil
}

func (m *Migrator) Status() ([]migration.Status, error) {
	// Ensure repository exists
	if err := m.repository.CreateRepository(); err != nil {
		return nil, fmt.Errorf("failed to create migration repository: %w", err)
	}

	// Get ran migrations
	ran, err := m.repository.GetRan()
	if err != nil {
		return nil, fmt.Errorf("failed to get ran migrations: %w", err)
	}

	// Get migration files from disk
	migrationFiles, err := m.getMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get all migrations with their batch numbers
	files, err := m.repository.GetMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get migrations: %w", err)
	}

	// Create batch lookup
	batchLookup := make(map[string]int)
	for _, file := range files {
		batchLookup[file.Migration] = file.Batch
	}

	// Build status list
	var status []migration.Status
	for name := range migrationFiles {
		s := migration.Status{
			Name: name,
			Ran:  contains(ran, name),
		}
		if s.Ran {
			s.Batch = batchLookup[name]
		}
		status = append(status, s)
	}

	// Sort by name
	sort.Slice(status, func(i, j int) bool {
		return status[i].Name < status[j].Name
	})

	return status, nil
}

// getMigrations loads migrations from the global registry and optionally from file paths.
// It first loads all migrations registered in the global registry (works without files on disk),
// then supplements with file-based discovery if paths are provided and directories exist.
// Returns a map of migration names to Migration functions.
func (m *Migrator) getMigrations() (map[string]Migration, error) {
	migrations := make(map[string]Migration)

	// First, load from global registry (works without files on disk)
	registryMutex.RLock()
	for name, migrationFunc := range migrationRegistry {
		migrations[name] = migrationFunc
	}
	registryMutex.RUnlock()

	// Then, if paths are provided, try to load from files
	for _, path := range m.paths {
		// Check if directory exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Directory doesn't exist, skip it
			continue
		}

		// Walk the directory to find migration files
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Only process .go files
			if !strings.HasSuffix(filePath, ".go") {
				return nil
			}

			// Extract migration name from filename
			filename := filepath.Base(filePath)
			name := strings.TrimSuffix(filename, ".go")

			// Skip test files
			if strings.HasSuffix(name, "_test") {
				return nil
			}

			// Load migration function from registry (protected by read lock)
			registryMutex.RLock()
			migrationFunc, ok := migrationRegistry[name]
			registryMutex.RUnlock()

			if ok {
				migrations[name] = migrationFunc
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk migration path %s: %w", path, err)
		}
	}

	return migrations, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Migration represents a migration with Up and Down methods
type Migration struct {
	Up   func(schema contractsschema.Schema) error
	Down func(schema contractsschema.Schema) error
}

// Global migration registry with mutex for thread safety
var (
	migrationRegistry = make(map[string]Migration)
	registryMutex     sync.RWMutex
)

// RegisterMigration registers a migration in the global registry
func RegisterMigration(name string, migration Migration) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	migrationRegistry[name] = migration
}

// ClearRegistry clears the global migration registry (useful for testing)
func ClearRegistry() {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	migrationRegistry = make(map[string]Migration)
}

func getTimestamp() int64 {
	return time.Now().Unix()
}
