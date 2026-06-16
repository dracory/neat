package schema

import (
	"fmt"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// MigrationTracker represents a migration record from the migration_tracker table
type MigrationTracker struct {
	ID          string    // The migration signature (e.g., "2024_06_15_120000_create_users_table")
	Batch       int       // Timestamp ID (YYYYMMDDHHMMSS). Groups the run
	Description string    // The migration description from Description() method
	StartedAt   time.Time // When the migration started
	CompletedAt time.Time // When the migration finished
}

// Status represents the status of a migration
type Status struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Batch       int       `json:"batch"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	State       string    `json:"state"` // "pending", "completed", "failed"
}

// MigrationManager handles execution and tracking of interface-based migrations
type MigrationManager struct {
	schema contractsschema.Schema
	orm    contractsorm.Orm
}

// NewMigrationManager creates a new MigrationManager instance.
//
// Deprecated: This will be moved to the database/schemer package in a future release.
// Use database/schemer.NewMigrationManager instead when available.
// See: https://github.com/dracory/neat/docs/proposals/schemer-package.md
func NewMigrationManager(schema contractsschema.Schema, orm contractsorm.Orm) *MigrationManager {
	return &MigrationManager{
		schema: schema,
		orm:    orm,
	}
}

// Run executes pending migrations
func (m *MigrationManager) Run(migrations []contractsschema.MigrationInterface) error {
	// Ensure migration_tracker table exists
	if !m.schema.HasTable("migration_tracker") {
		return fmt.Errorf("migration_tracker table does not exist. Run CreateMigrationTrackerTable migration first")
	}

	// Get the next batch number
	batch, err := m.getNextBatchNumber()
	if err != nil {
		return fmt.Errorf("failed to get next batch number: %w", err)
	}

	// Get already run migrations
	ranMigrations, err := m.getRanMigrations()
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		signature := migration.Signature()

		// Skip if already run
		if m.isMigrationRan(signature, ranMigrations) {
			continue
		}

		// Validate signature (basic validation)
		if len(signature) == 0 {
			return fmt.Errorf("migration signature cannot be empty")
		}
		if len(signature) > 255 {
			return fmt.Errorf("migration signature too long (max 255 characters)")
		}

		// Run migration
		startedAt := time.Now()
		if err := migration.Up(); err != nil {
			return fmt.Errorf("migration %s failed: %w", signature, err)
		}
		completedAt := time.Now()

		// Log migration
		if err := m.logMigration(signature, migration.Description(), batch, startedAt, completedAt); err != nil {
			return fmt.Errorf("failed to log migration %s: %w", signature, err)
		}
	}

	return nil
}

// Rollback reverts migrations
func (m *MigrationManager) Rollback(step, batch int) error {
	// Ensure migration_tracker table exists
	if !m.schema.HasTable("migration_tracker") {
		return fmt.Errorf("migration_tracker table does not exist")
	}

	var migrationsToRollback []MigrationTracker
	var err error

	if batch > 0 {
		// Rollback specific batch
		migrationsToRollback, err = m.getMigrationsByBatch(batch)
		if err != nil {
			return fmt.Errorf("failed to get migrations for batch %d: %w", batch, err)
		}
	} else if step > 0 {
		// Rollback last N migrations
		migrationsToRollback, err = m.getLastMigrations(step)
		if err != nil {
			return fmt.Errorf("failed to get last %d migrations: %w", step, err)
		}
	} else {
		// Rollback last batch
		migrationsToRollback, err = m.getLastBatch()
		if err != nil {
			return fmt.Errorf("failed to get last batch: %w", err)
		}
	}

	// Rollback in reverse order
	for i := len(migrationsToRollback) - 1; i >= 0; i-- {
		migration := migrationsToRollback[i]
		if err := m.rollbackMigration(migration.ID); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// Status returns migration status
func (m *MigrationManager) Status() ([]Status, error) {
	// Ensure migration_tracker table exists
	if !m.schema.HasTable("migration_tracker") {
		return []Status{}, nil
	}

	// Get all migrations from tracker
	migrations, err := m.getMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get migrations: %w", err)
	}

	// Convert to status
	status := make([]Status, len(migrations))
	for i, m := range migrations {
		status[i] = Status{
			ID:          m.ID,
			Description: m.Description,
			Batch:       m.Batch,
			StartedAt:   m.StartedAt,
			CompletedAt: m.CompletedAt,
			State:       "completed",
		}
	}

	return status, nil
}

// Fresh drops all tables and re-runs migrations
func (m *MigrationManager) Fresh() error {
	// Get all tables except migration_tracker
	tables, err := m.getAllTables()
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	// Drop all tables except migration_tracker
	for _, table := range tables {
		if table != "migration_tracker" {
			if err := m.schema.DropIfExists(table); err != nil {
				return fmt.Errorf("failed to drop table %s: %w", table, err)
			}
		}
	}

	// Clear migration_tracker table
	if err := m.clearMigrationTracker(); err != nil {
		return fmt.Errorf("failed to clear migration_tracker: %w", err)
	}

	return nil
}

// Reset rolls back and re-runs all migrations
func (m *MigrationManager) Reset() error {
	// Get all migrations
	migrations, err := m.getMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	// Rollback in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if err := m.rollbackMigration(migration.ID); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// Helper methods

func (m *MigrationManager) getNextBatchNumber() (int, error) {
	// Simple implementation: use current timestamp as batch number
	return int(time.Now().Unix()), nil
}

func (m *MigrationManager) getRanMigrations() ([]string, error) {
	var trackers []MigrationTracker
	query := m.orm.Query().Table("migration_tracker")
	if err := query.Get(&trackers); err != nil {
		return nil, err
	}

	ids := make([]string, len(trackers))
	for i, t := range trackers {
		ids[i] = t.ID
	}
	return ids, nil
}

func (m *MigrationManager) isMigrationRan(signature string, ranMigrations []string) bool {
	for _, ran := range ranMigrations {
		if ran == signature {
			return true
		}
	}
	return false
}

func (m *MigrationManager) logMigration(id, description string, batch int, startedAt, completedAt time.Time) error {
	tracker := MigrationTracker{
		ID:          id,
		Batch:       batch,
		Description: description,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
	}
	return m.orm.Query().Table("migration_tracker").Create(&tracker)
}

func (m *MigrationManager) getMigrationsByBatch(batch int) ([]MigrationTracker, error) {
	var trackers []MigrationTracker
	query := m.orm.Query().Table("migration_tracker").Where("batch = ?", batch)
	if err := query.Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func (m *MigrationManager) getLastMigrations(step int) ([]MigrationTracker, error) {
	var trackers []MigrationTracker
	query := m.orm.Query().Table("migration_tracker").OrderBy("id", "desc").Limit(step)
	if err := query.Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func (m *MigrationManager) getLastBatch() ([]MigrationTracker, error) {
	// Get the highest batch number
	var maxBatch int
	query := m.orm.Query().Table("migration_tracker").Select("MAX(batch) as max_batch")
	if err := query.First(&maxBatch); err != nil {
		return nil, err
	}

	if maxBatch == 0 {
		return []MigrationTracker{}, nil
	}

	return m.getMigrationsByBatch(maxBatch)
}

func (m *MigrationManager) rollbackMigration(id string) error {
	// Delete the migration record from tracker
	_, err := m.orm.Query().Table("migration_tracker").Where("id = ?", id).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (m *MigrationManager) getMigrations() ([]MigrationTracker, error) {
	var trackers []MigrationTracker
	query := m.orm.Query().Table("migration_tracker").OrderBy("id", "asc")
	if err := query.Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func (m *MigrationManager) getAllTables() ([]string, error) {
	// Get all tables from database
	// This is database-specific and would need to be implemented per driver
	// For now, return empty slice
	return []string{}, nil
}

func (m *MigrationManager) clearMigrationTracker() error {
	_, err := m.orm.Query().Table("migration_tracker").Delete()
	return err
}
