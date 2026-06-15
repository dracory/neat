package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
func NewMigrator(
	config config.Config,
	orm contractsorm.Orm,
	schema *schema.Schema,
	paths []string,
) migration.MigratorInterface {
	return &Migrator{
		config:     config,
		orm:        orm,
		repository: NewRepository(config, orm, schema),
		schema:     schema,
		paths:      paths,
	}
}

func (m *Migrator) Create(name string) error {
	if len(m.paths) == 0 {
		return fmt.Errorf("no migration paths configured")
	}

	// Get ID format from config
	format := MigrationIDFormat(m.config.GetString("database.migrations.id_format", "datetime"))

	// Generate prefix based on format
	var prefix string
	var err error

	switch format {
	case MigrationIDFormatDateTime:
		prefix = time.Now().Format("2006_01_02_1504") // YYYY_MM_DD_HHMM
	case MigrationIDFormatDate:
		prefix, err = m.generateDateSequence()
		if err != nil {
			return err
		}
	case MigrationIDFormatUnix:
		prefix = fmt.Sprintf("%d", time.Now().Unix())
	case MigrationIDFormatCustom:
		// For custom format, still use datetime as base
		prefix = time.Now().Format("2006_01_02_1504")
	default:
		// Default to datetime
		prefix = time.Now().Format("2006_01_02_1504")
	}

	migrationName := prefix + "_" + strings.ReplaceAll(name, " ", "_")

	// Validate format if enabled
	if m.config.GetBool("database.migrations.validate_format", true) {
		if err := ValidateMigrationID(migrationName, format); err != nil {
			return fmt.Errorf("invalid migration ID: %w", err)
		}
	}

	// Use the first path for creating migrations
	path := m.paths[0]

	// Create migration file
	filePath := filepath.Join(path, migrationName+".go")

	content := `package migrations

import (
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// Description: ` + name + `

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

	useTransactions := m.config.GetBool("database.migrations.use_transactions", true)

	// Run pending migrations
	for _, name := range names {
		// Skip if already ran
		if slices.Contains(ran, name) {
			continue
		}

		migrationFunc := migrationFiles[name]
		description := m.extractDescription(name)

		m.log("info", "Running migration", map[string]any{
			"migration": name,
			"batch":     batch,
		})

		startedAt := time.Now()

		if useTransactions {
			// Run migration in transaction
			tx, err := m.orm.Query().Begin()
			if err != nil {
				return fmt.Errorf("failed to start transaction for migration %s: %w", name, err)
			}

			// Note: We cannot easily create a new schema with the transaction
			// because the schema builder doesn't support transaction-scoped instances yet.
			// For now, run migrations without transaction-scoped schema.
			// TODO: Enhance schema builder to support transaction contexts

			// Run Up method
			if err := migrationFunc.Up(m.schema); err != nil {
				tx.Rollback()
				m.log("error", "Migration failed, rolled back", map[string]any{
					"migration": name,
					"error":     err.Error(),
				})
				return fmt.Errorf("failed to run migration %s: %w", name, err)
			}

			completedAt := time.Now()
			duration := completedAt.Sub(startedAt)

			// Log migration within same transaction
			if err := m.repository.Log(name, batch, description, startedAt, completedAt); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to log migration: %w", err)
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction for migration %s: %w", name, err)
			}

			m.log("info", "Migration completed", map[string]any{
				"migration": name,
				"duration":  duration.String(),
				"batch":     batch,
			})
		} else {
			// Run migration without transaction
			if err := migrationFunc.Up(m.schema); err != nil {
				m.log("error", "Migration failed", map[string]any{
					"migration": name,
					"error":     err.Error(),
				})
				return fmt.Errorf("failed to run migration %s: %w", name, err)
			}

			completedAt := time.Now()
			duration := completedAt.Sub(startedAt)

			// Log migration
			if err := m.repository.Log(name, batch, description, startedAt, completedAt); err != nil {
				return fmt.Errorf("failed to log migration: %w", err)
			}

			m.log("info", "Migration completed", map[string]any{
				"migration": name,
				"duration":  duration.String(),
				"batch":     batch,
			})
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

	// Get all migrations with their details
	files, err := m.repository.GetHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to get migrations: %w", err)
	}

	// Create lookup maps
	batchLookup := make(map[string]int)
	descriptionLookup := make(map[string]string)
	durationLookup := make(map[string]time.Duration)
	lastRunLookup := make(map[string]time.Time)

	for _, file := range files {
		batchLookup[file.Migration] = file.Batch
		descriptionLookup[file.Migration] = file.Description
		if !file.StartedAt.IsZero() && !file.CompletedAt.IsZero() {
			durationLookup[file.Migration] = file.CompletedAt.Sub(file.StartedAt)
		}
		if !file.CompletedAt.IsZero() {
			lastRunLookup[file.Migration] = file.CompletedAt
		}
	}

	// Build status list
	var status []migration.Status
	for name := range migrationFiles {
		s := migration.Status{
			Name:        name,
			Ran:         slices.Contains(ran, name),
			Description: descriptionLookup[name],
		}
		if s.Ran {
			s.Batch = batchLookup[name]
			s.Duration = durationLookup[name]
			s.LastRun = lastRunLookup[name]
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
// Returns error if migration with the same name already exists
func RegisterMigration(name string, migration Migration) error {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if _, exists := migrationRegistry[name]; exists {
		return fmt.Errorf("migration '%s' is already registered", name)
	}

	migrationRegistry[name] = migration
	return nil
}

// MustRegisterMigration registers a migration and panics on error
// Use this for init() functions where errors should be fatal
func MustRegisterMigration(name string, migration Migration) {
	if err := RegisterMigration(name, migration); err != nil {
		panic(err)
	}
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

// extractDescription extracts description from migration name or file comments
func (m *Migrator) extractDescription(migrationName string) string {
	// Try to extract from migration name (everything after the prefix)
	parts := strings.Split(migrationName, "_")
	if len(parts) >= 5 {
		// For datetime/date formats: YYYY_MM_DD_HHMM/NNN_description
		return strings.Join(parts[4:], " ")
	} else if len(parts) >= 2 {
		// For unix format: timestamp_description
		return strings.Join(parts[1:], " ")
	}
	return ""
}

// log logs a message if logging is enabled
func (m *Migrator) log(level, message string, data map[string]any) {
	if !m.config.GetBool("database.migrations.logging.enabled", false) {
		return
	}

	configLevel := m.config.GetString("database.migrations.logging.level", "info")

	// Build message with data
	msg := message
	if len(data) > 0 {
		parts := []string{message}
		for k, v := range data {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		msg = strings.Join(parts, " ")
	}

	// Get logger from config if available
	// For now, we'll skip actual logging since we need to integrate with Neat's log system
	// This can be enhanced later when log.Log is available in the migrator

	_ = level
	_ = configLevel
	_ = msg
}
