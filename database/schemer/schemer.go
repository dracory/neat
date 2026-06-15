package schemer

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
)

// SchemerInterface defines the contract for migration management
type SchemerInterface interface {
	AddMigration(migration contractsschema.MigrationInterface) error
	AddMigrations(migrations []contractsschema.MigrationInterface) error
	Up(ctx context.Context) error
	Down(ctx context.Context) error
	RollbackSteps(ctx context.Context, steps int) error
	RollbackToBatch(ctx context.Context, batch int) error
	Status() ([]MigrationStatus, error)
	Fresh(ctx context.Context) error
	Reset(ctx context.Context) error
	SetTransactionsEnabled(enabled bool)
	SetTransactionIsolationLevel(level string)
}

// SchemerImplementation handles execution and tracking of interface-based migrations
type SchemerImplementation struct {
	db              *database.Database
	migrations      []contractsschema.MigrationInterface
	useTransactions bool
	isolationLevel  string
}

// NewSchemer creates a new SchemerImplementation instance
// Takes neat db instance as dependency, extracts schema and orm internally
func NewSchemer(db *database.Database) SchemerInterface {
	return &SchemerImplementation{
		db:              db,
		migrations:      []contractsschema.MigrationInterface{},
		useTransactions: true, // Default to safe transaction behavior
	}
}

// AddMigration adds a new migration to the list
func (s *SchemerImplementation) AddMigration(migration contractsschema.MigrationInterface) error {
	s.migrations = append(s.migrations, migration)
	return nil
}

// AddMigrations adds multiple migrations to the runner
func (s *SchemerImplementation) AddMigrations(migrations []contractsschema.MigrationInterface) error {
	s.migrations = append(s.migrations, migrations...)
	return nil
}

// SetTransactionsEnabled enables or disables transaction wrapping for migration operations
func (s *SchemerImplementation) SetTransactionsEnabled(enabled bool) {
	s.useTransactions = enabled
}

// SetTransactionIsolationLevel sets the transaction isolation level for migration operations
func (s *SchemerImplementation) SetTransactionIsolationLevel(level string) {
	s.isolationLevel = level
}

// Up runs all pending migrations
// Automatically injects schema into each migration before execution
func (s *SchemerImplementation) Up(ctx context.Context) error {
	if s.useTransactions {
		return s.db.Schema().Orm().Transaction(func(tx contractsorm.Query) error {
			schema := s.db.Schema().WithTransaction(tx)
			return s.runUp(ctx, schema, tx)
		}, s.txOptions())
	}
	return s.up(ctx)
}

// up contains the actual migration execution logic
func (s *SchemerImplementation) up(ctx context.Context) error {
	schema := s.db.Schema()
	query := schema.Orm().Query()
	return s.runUp(ctx, schema, query)
}

// runUp contains the shared migration execution logic
func (s *SchemerImplementation) runUp(ctx context.Context, schema contractsschema.Schema, query contractsorm.Query) error {
	// Create migration_tracker table if it doesn't exist
	if !schema.HasTable("migration_tracker") {
		err := schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.Integer("batch")
			table.String("description", 255)
			table.DateTime("started_at")
			table.DateTime("completed_at")
		})
		if err != nil {
			return fmt.Errorf("failed to create migration_tracker table: %w", err)
		}
	}

	// Get the next batch number
	batch, err := s.getNextBatchNumber()
	if err != nil {
		return fmt.Errorf("failed to get next batch number: %w", err)
	}

	// Get already run migrations
	ranMigrations, err := s.getRanMigrations(query)
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range s.migrations {
		signature := migration.Signature()

		// Skip if already run
		if s.isMigrationRan(signature, ranMigrations) {
			continue
		}

		// Validate signature (basic validation)
		if len(signature) == 0 {
			return fmt.Errorf("migration signature cannot be empty")
		}
		if len(signature) > 255 {
			return fmt.Errorf("migration signature too long (max 255 characters)")
		}

		// Inject transaction-aware schema into migration
		migration.SetSchema(schema)

		// Run migration
		startedAt := time.Now()
		if err := migration.Up(); err != nil {
			return fmt.Errorf("migration %s failed: %w", signature, err)
		}
		completedAt := time.Now()

		// Log migration
		if err := s.logMigration(query, signature, migration.Description(), batch, startedAt, completedAt); err != nil {
			return fmt.Errorf("failed to log migration %s: %w", signature, err)
		}
	}

	return nil
}

// Down rolls back the last migration
func (s *SchemerImplementation) Down(ctx context.Context) error {
	if s.useTransactions {
		return s.db.Schema().Orm().Transaction(func(tx contractsorm.Query) error {
			schema := s.db.Schema().WithTransaction(tx)
			return s.runRollbackSteps(ctx, schema, tx, 1)
		}, s.txOptions())
	}
	schema := s.db.Schema()
	query := schema.Orm().Query()
	return s.runRollbackSteps(ctx, schema, query, 1)
}

// RollbackSteps rolls back the specified number of migrations
func (s *SchemerImplementation) RollbackSteps(ctx context.Context, steps int) error {
	if s.useTransactions {
		return s.db.Schema().Orm().Transaction(func(tx contractsorm.Query) error {
			schema := s.db.Schema().WithTransaction(tx)
			return s.runRollbackSteps(ctx, schema, tx, steps)
		}, s.txOptions())
	}
	schema := s.db.Schema()
	query := schema.Orm().Query()
	return s.runRollbackSteps(ctx, schema, query, steps)
}

// runRollbackSteps contains the shared rollback logic
func (s *SchemerImplementation) runRollbackSteps(ctx context.Context, schema contractsschema.Schema, query contractsorm.Query, steps int) error {
	// Ensure migration_tracker table exists
	if !schema.HasTable("migration_tracker") {
		return fmt.Errorf("migration_tracker table does not exist")
	}

	// Rollback last N migrations
	migrationsToRollback, err := s.getLastMigrations(query, steps)
	if err != nil {
		return fmt.Errorf("failed to get last %d migrations: %w", steps, err)
	}

	// Rollback in reverse order
	for i := len(migrationsToRollback) - 1; i >= 0; i-- {
		migration := migrationsToRollback[i]
		if err := s.rollbackMigration(schema, query, migration.ID); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// RollbackToBatch rolls back all migrations to the specified batch
func (s *SchemerImplementation) RollbackToBatch(ctx context.Context, batch int) error {
	if s.useTransactions {
		return s.db.Schema().Orm().Transaction(func(tx contractsorm.Query) error {
			schema := s.db.Schema().WithTransaction(tx)
			return s.runRollbackToBatch(ctx, schema, tx, batch)
		}, s.txOptions())
	}
	schema := s.db.Schema()
	query := schema.Orm().Query()
	return s.runRollbackToBatch(ctx, schema, query, batch)
}

// runRollbackToBatch contains the shared batch rollback logic
func (s *SchemerImplementation) runRollbackToBatch(ctx context.Context, schema contractsschema.Schema, query contractsorm.Query, batch int) error {
	// Ensure migration_tracker table exists
	if !schema.HasTable("migration_tracker") {
		return fmt.Errorf("migration_tracker table does not exist")
	}

	// Rollback specific batch
	migrationsToRollback, err := s.getMigrationsByBatch(query, batch)
	if err != nil {
		return fmt.Errorf("failed to get migrations for batch %d: %w", batch, err)
	}

	// Rollback in reverse order
	for i := len(migrationsToRollback) - 1; i >= 0; i-- {
		migration := migrationsToRollback[i]
		if err := s.rollbackMigration(schema, query, migration.ID); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// Status returns migration status
func (s *SchemerImplementation) Status() ([]MigrationStatus, error) {
	// Ensure migration_tracker table exists
	if !s.db.Schema().HasTable("migration_tracker") {
		return []MigrationStatus{}, nil
	}

	// Get all migrations from tracker
	migrations, err := s.getMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get migrations: %w", err)
	}

	// Convert to status
	status := make([]MigrationStatus, len(migrations))
	for i, m := range migrations {
		status[i] = MigrationStatus{
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
func (s *SchemerImplementation) Fresh(ctx context.Context) error {
	if s.useTransactions {
		return s.db.Schema().Orm().Transaction(func(tx contractsorm.Query) error {
			schema := s.db.Schema().WithTransaction(tx)
			return s.runFresh(ctx, schema, tx)
		}, s.txOptions())
	}
	schema := s.db.Schema()
	query := schema.Orm().Query()
	return s.runFresh(ctx, schema, query)
}

// runFresh contains the shared fresh logic
func (s *SchemerImplementation) runFresh(ctx context.Context, schema contractsschema.Schema, query contractsorm.Query) error {
	// Note: DDL operations (DROP TABLE) may cause implicit commits in some databases
	// (MySQL, PostgreSQL). This means the transaction wrapper may not provide full atomicity
	// for Fresh operations. However, it's still useful for the migration_tracker cleanup.

	// Get all tables except migration_tracker
	tables, err := s.getAllTables()
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	// Drop all tables except migration_tracker
	for _, table := range tables {
		if table != "migration_tracker" {
			if err := schema.DropIfExists(table); err != nil {
				return fmt.Errorf("failed to drop table %s: %w", table, err)
			}
		}
	}

	// Clear migration_tracker table
	if err := s.clearMigrationTracker(query); err != nil {
		return fmt.Errorf("failed to clear migration_tracker: %w", err)
	}

	// Re-run all migrations
	if err := s.runUp(ctx, schema, query); err != nil {
		return fmt.Errorf("failed to re-run migrations: %w", err)
	}

	return nil
}

// Reset rolls back and re-runs all migrations
func (s *SchemerImplementation) Reset(ctx context.Context) error {
	if s.useTransactions {
		return s.db.Schema().Orm().Transaction(func(tx contractsorm.Query) error {
			schema := s.db.Schema().WithTransaction(tx)
			return s.runReset(ctx, schema, tx)
		}, s.txOptions())
	}
	schema := s.db.Schema()
	query := schema.Orm().Query()
	return s.runReset(ctx, schema, query)
}

// runReset contains the shared reset logic
func (s *SchemerImplementation) runReset(ctx context.Context, schema contractsschema.Schema, query contractsorm.Query) error {
	// Get all migrations
	migrations, err := s.getMigrationsWithQuery(query)
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	// Rollback in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if err := s.rollbackMigration(schema, query, migration.ID); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// Helper methods

func (s *SchemerImplementation) txOptions() *sql.TxOptions {
	if s.isolationLevel != "" {
		return &sql.TxOptions{
			Isolation: s.parseIsolationLevel(s.isolationLevel),
		}
	}
	return nil
}

func (s *SchemerImplementation) getNextBatchNumber() (int, error) {
	// Simple implementation: use current timestamp as batch number
	return int(time.Now().Unix()), nil
}

func (s *SchemerImplementation) getRanMigrations(query contractsorm.Query) ([]string, error) {
	var trackers []MigrationTracker
	if err := query.Table("migration_tracker").Get(&trackers); err != nil {
		return nil, err
	}

	ids := make([]string, len(trackers))
	for i, t := range trackers {
		ids[i] = t.ID
	}
	return ids, nil
}

func (s *SchemerImplementation) isMigrationRan(signature string, ranMigrations []string) bool {
	for _, ran := range ranMigrations {
		if ran == signature {
			return true
		}
	}
	return false
}

func (s *SchemerImplementation) logMigration(query contractsorm.Query, id, description string, batch int, startedAt, completedAt time.Time) error {
	tracker := MigrationTracker{
		ID:          id,
		Batch:       batch,
		Description: description,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
	}
	return query.Table("migration_tracker").Create(&tracker)
}

func (s *SchemerImplementation) getMigrationsByBatch(query contractsorm.Query, batch int) ([]MigrationTracker, error) {
	var trackers []MigrationTracker
	if err := query.Table("migration_tracker").Where("batch = ?", batch).Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func (s *SchemerImplementation) getLastMigrations(query contractsorm.Query, step int) ([]MigrationTracker, error) {
	var trackers []MigrationTracker
	if err := query.Table("migration_tracker").OrderBy("id", "desc").Limit(step).Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func (s *SchemerImplementation) rollbackMigration(schema contractsschema.Schema, query contractsorm.Query, id string) error {
	// Find the migration by signature
	var migration contractsschema.MigrationInterface
	for _, m := range s.migrations {
		if m.Signature() == id {
			migration = m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found in registered migrations", id)
	}

	// Inject transaction-aware schema
	migration.SetSchema(schema)

	// Run the Down migration
	if err := migration.Down(); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", id, err)
	}

	// Delete the migration record from tracker
	_, err := query.Table("migration_tracker").Where("id = ?", id).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (s *SchemerImplementation) getMigrations() ([]MigrationTracker, error) {
	return s.getMigrationsWithQuery(s.db.Schema().Orm().Query().Table("migration_tracker").OrderBy("id", "asc"))
}

func (s *SchemerImplementation) getMigrationsWithQuery(query contractsorm.Query) ([]MigrationTracker, error) {
	var trackers []MigrationTracker
	if err := query.Table("migration_tracker").OrderBy("id", "asc").Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func (s *SchemerImplementation) getAllTables() ([]string, error) {
	// Get all tables from database
	// This is database-specific and would need to be implemented per driver
	// For now, return empty slice
	return []string{}, nil
}

func (s *SchemerImplementation) clearMigrationTracker(query contractsorm.Query) error {
	_, err := query.Table("migration_tracker").Delete()
	return err
}

// parseIsolationLevel converts string isolation level to sql.IsolationLevel
func (s *SchemerImplementation) parseIsolationLevel(level string) sql.IsolationLevel {
	switch level {
	case "READ UNCOMMITTED":
		return sql.LevelReadUncommitted
	case "READ COMMITTED":
		return sql.LevelReadCommitted
	case "REPEATABLE READ":
		return sql.LevelRepeatableRead
	case "SERIALIZABLE":
		return sql.LevelSerializable
	case "SNAPSHOT":
		return sql.LevelSnapshot
	default:
		return sql.LevelDefault
	}
}
