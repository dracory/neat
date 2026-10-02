package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
)

const defaultTableName = "migration_tracker"

// MigratorInterface defines the contract for migration management
type MigratorInterface interface {
	AddMigration(migration MigrationInterface) error
	AddMigrations(migrations []MigrationInterface) error
	Up(ctx context.Context) error
	Down(ctx context.Context) error
	RollbackSteps(ctx context.Context, steps int) error
	RollbackToBatch(ctx context.Context, batch int) error
	Status() ([]MigrationStatusResponse, error)
	Fresh(ctx context.Context) error
	Reset(ctx context.Context) error
	SetTransactionsEnabled(enabled bool)
	SetTransactionIsolationLevel(level string)
	SetTableName(name string) error
	SetSignatureValidation(enabled bool, format SignatureFormat)
	SetLexicographicalOrdering(enabled bool)
}

// Migrator handles execution and tracking of interface-based migrations
type Migrator struct {
	db                      *database.Database
	migrations              []MigrationInterface
	useTransactions         bool
	isolationLevel          string
	tableName               string
	sigValidation           bool
	sigValidationFormat     SignatureFormat
	lexicographicalOrdering bool
}

// NewMigrator creates a new Migrator instance
// Takes neat db instance as dependency, extracts schema and orm internally
func NewMigrator(db *database.Database) MigratorInterface {
	return &Migrator{
		db:                      db,
		migrations:              []MigrationInterface{},
		useTransactions:         true, // Default to safe transaction behavior
		tableName:               defaultTableName,
		lexicographicalOrdering: true, // Default to lexicographical ordering
	}
}

// Options configures the Migrator at construction time.
type Options struct {
	TableName                  string
	TransactionIsolationLevel  string
	LexicographicalOrdering    bool
	SignatureValidationEnabled bool
	SignatureValidationFormat  SignatureFormat
}

// NewMigratorWithOptions creates a Migrator with the given options.
func NewMigratorWithOptions(db *database.Database, opts *Options) (MigratorInterface, error) {
	m := NewMigrator(db)

	if opts == nil {
		return m, nil
	}

	if opts.TableName != "" {
		if err := m.SetTableName(opts.TableName); err != nil {
			return nil, err
		}
	}

	if opts.TransactionIsolationLevel != "" {
		m.SetTransactionIsolationLevel(opts.TransactionIsolationLevel)
	}

	m.SetLexicographicalOrdering(opts.LexicographicalOrdering)

	if opts.SignatureValidationEnabled {
		m.SetSignatureValidation(true, opts.SignatureValidationFormat)
	}

	return m, nil
}

// AddMigration adds a new migration to the list.
// Returns an error if the migration signature is empty or already registered.
func (s *Migrator) AddMigration(migration MigrationInterface) error {
	signature := migration.Signature()
	if signature == "" {
		return fmt.Errorf("migration signature cannot be empty")
	}

	for _, existing := range s.migrations {
		if existing.Signature() == signature {
			return fmt.Errorf("duplicate migration signature: %s", signature)
		}
	}

	s.migrations = append(s.migrations, migration)
	return nil
}

// AddMigrations adds multiple migrations to the runner.
// Returns an error if any migration has an empty or duplicate signature.
func (s *Migrator) AddMigrations(migrations []MigrationInterface) error {
	for _, migration := range migrations {
		if err := s.AddMigration(migration); err != nil {
			return err
		}
	}
	return nil
}

// SetTransactionsEnabled enables or disables transaction wrapping for migration operations
func (s *Migrator) SetTransactionsEnabled(enabled bool) {
	s.useTransactions = enabled
}

// SetTransactionIsolationLevel sets the transaction isolation level for migration operations
func (s *Migrator) SetTransactionIsolationLevel(level string) {
	s.isolationLevel = level
}

// SetTableName sets the name of the migration tracking table.
// The name is validated to prevent SQL injection.
func (s *Migrator) SetTableName(name string) error {
	if !isValidTableName(name) {
		return fmt.Errorf("invalid migration table name: '%s'", name)
	}
	s.tableName = name
	return nil
}

// SetSignatureValidation enables or disables signature format validation.
// When enabled, each migration signature is validated against the specified
// format before execution. Default is disabled.
func (s *Migrator) SetSignatureValidation(enabled bool, format SignatureFormat) {
	s.sigValidation = enabled
	s.sigValidationFormat = format
}

// SetLexicographicalOrdering enables or disables lexicographical ordering of migrations.
// When enabled, migrations are sorted by signature before execution.
// Default is enabled.
func (s *Migrator) SetLexicographicalOrdering(enabled bool) {
	s.lexicographicalOrdering = enabled
}

// sortMigrations sorts migrations lexicographically by signature.
func (s *Migrator) sortMigrations() {
	sort.Slice(s.migrations, func(i, j int) bool {
		return s.migrations[i].Signature() < s.migrations[j].Signature()
	})
}

// migrationFailure carries the details of a failed migration attempt so the
// failure can be recorded in the tracker table after the surrounding
// transaction has rolled back.
type migrationFailure struct {
	signature   string
	description string
	batch       int
	status      string // failed (Up) or rollback_failed (Down)
	err         error
	startedAt   time.Time
	completedAt time.Time
}

// Up runs all pending migrations
// Automatically injects schema into each migration before execution
func (s *Migrator) Up(ctx context.Context) error {
	var failure *migrationFailure
	var err error
	if s.useTransactions {
		err = s.db.Schema().Orm().Transaction(func(tx orm.Query) error {
			schema := s.db.Schema().WithTransaction(tx)
			return s.runUp(ctx, schema, tx, &failure)
		}, s.txOptions())
	} else {
		err = s.up(ctx, &failure)
	}
	if failure != nil {
		s.recordFailure(failure)
	}
	return err
}

// up contains the actual migration execution logic
func (s *Migrator) up(ctx context.Context, failure **migrationFailure) error {
	schema := s.db.Schema()
	query := schema.Orm().Query()
	return s.runUp(ctx, schema, query, failure)
}

// runUp contains the shared migration execution logic
func (s *Migrator) runUp(ctx context.Context, schema contractsschema.Schema, query orm.Query, failure **migrationFailure) error {
	_ = ctx
	// Ensure migration tracking table exists and is up to date
	if err := s.ensureMigrationTracker(schema); err != nil {
		return fmt.Errorf("failed to ensure migration tracker: %w", err)
	}

	// Get the next batch number
	batch, err := s.getNextBatchNumber(query)
	if err != nil {
		return fmt.Errorf("failed to get next batch number: %w", err)
	}

	// Get already run migrations
	ranMigrations, err := s.getRanMigrations(query)
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %w", err)
	}

	// Sort migrations lexicographically if enabled
	if s.lexicographicalOrdering {
		s.sortMigrations()
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

		// Validate signature format if enabled
		if s.sigValidation {
			if err := ValidateMigrationSignature(signature, s.sigValidationFormat); err != nil {
				return fmt.Errorf("migration %s has invalid signature: %w", signature, err)
			}
		}

		// Inject transaction-aware schema into migration
		migration.SetSchema(schema)

		// Record the attempt as running before executing so crashes leave
		// a trace on drivers where DDL does not roll back
		startedAt := time.Now()
		tracker := MigrationTracker{
			ID:          signature,
			Batch:       batch,
			Description: migration.Description(),
			Status:      MigrationTrackerStatusRunning,
			StartedAt:   startedAt,
			CompletedAt: startedAt,
		}
		if err := s.upsertTracker(query, tracker); err != nil {
			return fmt.Errorf("failed to record migration %s start: %w", signature, err)
		}

		// Run migration
		if err := migration.Up(); err != nil {
			if failure != nil {
				*failure = &migrationFailure{
					signature:   signature,
					description: migration.Description(),
					batch:       batch,
					status:      MigrationTrackerStatusFailed,
					err:         err,
					startedAt:   startedAt,
					completedAt: time.Now(),
				}
			}
			return fmt.Errorf("migration %s failed: %w", signature, err)
		}
		completedAt := time.Now()

		// Mark migration completed
		tracker.Status = MigrationTrackerStatusCompleted
		tracker.CompletedAt = completedAt
		if err := s.upsertTracker(query, tracker); err != nil {
			return fmt.Errorf("failed to log migration %s: %w", signature, err)
		}
	}

	return nil
}

// Down rolls back the last migration
func (s *Migrator) Down(ctx context.Context) error {
	var failure *migrationFailure
	err := s.runInTx(ctx, func(schema contractsschema.Schema, query orm.Query) error {
		return s.runRollbackSteps(ctx, schema, query, 1, &failure)
	})
	if failure != nil {
		s.recordFailure(failure)
	}
	return err
}

// RollbackSteps rolls back the specified number of migrations
func (s *Migrator) RollbackSteps(ctx context.Context, steps int) error {
	var failure *migrationFailure
	err := s.runInTx(ctx, func(schema contractsschema.Schema, query orm.Query) error {
		return s.runRollbackSteps(ctx, schema, query, steps, &failure)
	})
	if failure != nil {
		s.recordFailure(failure)
	}
	return err
}

// runInTx wraps fn in a transaction when enabled, otherwise runs it directly.
func (s *Migrator) runInTx(ctx context.Context, fn func(schema contractsschema.Schema, query orm.Query) error) error {
	if s.useTransactions {
		return s.db.Schema().Orm().Transaction(func(tx orm.Query) error {
			return fn(s.db.Schema().WithTransaction(tx), tx)
		}, s.txOptions())
	}
	schema := s.db.Schema()
	return fn(schema, schema.Orm().Query())
}

// runRollbackSteps contains the shared rollback logic
func (s *Migrator) runRollbackSteps(ctx context.Context, schema contractsschema.Schema, query orm.Query, steps int, failure **migrationFailure) error {
	_ = ctx
	// Ensure migration tracking table exists
	if !schema.HasTable(s.tableName) {
		return fmt.Errorf("%s table does not exist", s.tableName)
	}

	// Rollback last N migrations
	migrationsToRollback, err := s.getLastMigrations(query, steps)
	if err != nil {
		return fmt.Errorf("failed to get last %d migrations: %w", steps, err)
	}

	// Rollback in reverse order
	for i := len(migrationsToRollback) - 1; i >= 0; i-- {
		migration := migrationsToRollback[i]
		if err := s.rollbackMigration(schema, query, migration, failure); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// RollbackToBatch rolls back all migrations to the specified batch
func (s *Migrator) RollbackToBatch(ctx context.Context, batch int) error {
	var failure *migrationFailure
	err := s.runInTx(ctx, func(schema contractsschema.Schema, query orm.Query) error {
		return s.runRollbackToBatch(ctx, schema, query, batch, &failure)
	})
	if failure != nil {
		s.recordFailure(failure)
	}
	return err
}

// runRollbackToBatch contains the shared batch rollback logic
func (s *Migrator) runRollbackToBatch(ctx context.Context, schema contractsschema.Schema, query orm.Query, batch int, failure **migrationFailure) error {
	_ = ctx
	// Ensure migration tracking table exists
	if !schema.HasTable(s.tableName) {
		return fmt.Errorf("%s table does not exist", s.tableName)
	}

	// Rollback specific batch
	migrationsToRollback, err := s.getMigrationsByBatch(query, batch)
	if err != nil {
		return fmt.Errorf("failed to get migrations for batch %d: %w", batch, err)
	}

	// Rollback in reverse order
	for i := len(migrationsToRollback) - 1; i >= 0; i-- {
		migration := migrationsToRollback[i]
		if err := s.rollbackMigration(schema, query, migration, failure); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// Status returns migration status for all registered migrations.
// Includes both completed migrations (from the tracker table) and
// pending migrations (registered but not yet run).
func (s *Migrator) Status() ([]MigrationStatusResponse, error) {
	var statuses []MigrationStatusResponse

	// Collect recorded migrations from tracker
	recorded := make(map[string]bool)
	if s.db.Schema().HasTable(s.tableName) {
		trackers, err := s.getMigrations()
		if err != nil {
			return nil, fmt.Errorf("failed to get migrations: %w", err)
		}
		for _, t := range trackers {
			recorded[t.ID] = true
			state := t.Status
			if state == "" {
				// Legacy rows written before status tracking are completed
				state = MigrationTrackerStatusCompleted
			}
			statuses = append(statuses, MigrationStatusResponse{
				ID:          t.ID,
				Description: t.Description,
				Batch:       t.Batch,
				StartedAt:   t.StartedAt,
				CompletedAt: t.CompletedAt,
				State:       state,
				Error:       t.ErrorMessage,
			})
		}
	}

	// Add pending migrations from registered list
	for _, migration := range s.migrations {
		sig := migration.Signature()
		if !recorded[sig] {
			statuses = append(statuses, MigrationStatusResponse{
				ID:          sig,
				Description: migration.Description(),
				State:       "pending",
			})
		}
	}

	// Sort by signature for consistent output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].ID < statuses[j].ID
	})

	return statuses, nil
}

// Fresh drops all tables and re-runs migrations
func (s *Migrator) Fresh(ctx context.Context) error {
	var failure *migrationFailure
	err := s.runInTx(ctx, func(schema contractsschema.Schema, query orm.Query) error {
		return s.runFresh(ctx, schema, query, &failure)
	})
	if failure != nil {
		s.recordFailure(failure)
	}
	return err
}

// runFresh contains the shared fresh logic
func (s *Migrator) runFresh(ctx context.Context, schema contractsschema.Schema, query orm.Query, failure **migrationFailure) error {
	// Note: DDL operations (DROP TABLE) may cause implicit commits in some databases
	// (MySQL, PostgreSQL). This means the transaction wrapper may not provide full atomicity
	// for Fresh operations. However, it's still useful for the migration tracking table cleanup.

	// Get all tables except the migration tracking table
	tables, err := s.getAllTables(schema)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	// Drop all tables except the migration tracking table
	for _, table := range tables {
		if table != s.tableName {
			if err := schema.DropIfExists(table); err != nil {
				return fmt.Errorf("failed to drop table %s: %w", table, err)
			}
		}
	}

	// Clear migration tracking table (it may not exist on a brand-new database)
	if schema.HasTable(s.tableName) {
		if err := s.clearMigrationTracker(query); err != nil {
			return fmt.Errorf("failed to clear %s: %w", s.tableName, err)
		}
	}

	// Re-run all migrations
	if err := s.runUp(ctx, schema, query, failure); err != nil {
		return fmt.Errorf("failed to re-run migrations: %w", err)
	}

	return nil
}

// Reset rolls back all applied migrations
func (s *Migrator) Reset(ctx context.Context) error {
	var failure *migrationFailure
	err := s.runInTx(ctx, func(schema contractsschema.Schema, query orm.Query) error {
		return s.runReset(ctx, schema, query, &failure)
	})
	if failure != nil {
		s.recordFailure(failure)
	}
	return err
}

const maxResetIterations = 1000

// runReset contains the shared reset logic
func (s *Migrator) runReset(ctx context.Context, schema contractsschema.Schema, query orm.Query, failure **migrationFailure) error {
	_ = ctx
	// Get all migrations
	migrations, err := s.getMigrationsWithQuery(query)
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	// Safety guard against unexpectedly large rollback sets
	if len(migrations) > maxResetIterations {
		return fmt.Errorf("too many migrations to reset (%d > max %d)", len(migrations), maxResetIterations)
	}

	// Rollback in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if err := s.rollbackMigration(schema, query, migration, failure); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// Helper methods

func (s *Migrator) txOptions() *sql.TxOptions {
	if s.isolationLevel != "" {
		return &sql.TxOptions{
			Isolation: s.parseIsolationLevel(s.isolationLevel),
		}
	}
	return nil
}

func (s *Migrator) getNextBatchNumber(query orm.Query) (int, error) {
	var maxBatch struct {
		Max sql.NullInt64
	}

	batchSQL := fmt.Sprintf("SELECT MAX(batch) as max FROM %s", s.tableName)
	if err := query.Raw(batchSQL).Scan(&maxBatch); err != nil {
		return 0, fmt.Errorf("failed to get max batch: %w", err)
	}

	if !maxBatch.Max.Valid {
		return 1, nil
	}

	return int(maxBatch.Max.Int64) + 1, nil
}

func (s *Migrator) getRanMigrations(query orm.Query) ([]string, error) {
	var trackers []MigrationTracker
	if err := cloneQuery(query).Table(s.tableName).Get(&trackers); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(trackers))
	for _, t := range trackers {
		// Completed migrations count as ran; so do rollback_failed rows,
		// whose schema change is still applied. Failed or interrupted Up()
		// attempts must be retried. Legacy rows (empty status) are completed.
		if t.Status == "" || t.Status == MigrationTrackerStatusCompleted ||
			t.Status == MigrationTrackerStatusRollbackFailed {
			ids = append(ids, t.ID)
		}
	}
	return ids, nil
}

func (s *Migrator) isMigrationRan(signature string, ranMigrations []string) bool {
	for _, ran := range ranMigrations {
		if ran == signature {
			return true
		}
	}
	return false
}

// upsertTracker inserts the tracker row, or updates the existing row when a
// record with the same signature already exists (e.g. retrying a failed
// migration). One row per migration signature always reflects the latest
// attempt.
// appliedStatusFilter limits queries to tracker rows whose migration was
// actually applied: completed rows, legacy rows (NULL or empty status), and
// rollback_failed rows whose schema change is still present. Rows from
// failed or interrupted Up() attempts are excluded — their migrations
// never ran and must not be rolled back.
const appliedStatusFilter = "(status IS NULL OR status IN ('', '" +
	MigrationTrackerStatusCompleted + "', '" + MigrationTrackerStatusRollbackFailed + "'))"

// cloneQuery returns an independent copy of q so that builder methods
// (Table/Where/OrderBy/Limit, which mutate the receiver) do not accumulate
// clauses on the shared query object. This also prevents a check-then-insert
// mismatch when the same query is reused across loop iterations.
func cloneQuery(q orm.Query) orm.Query {
	if c, ok := q.(interface{ Clone() orm.Query }); ok {
		return c.Clone()
	}
	return q
}

func (s *Migrator) upsertTracker(query orm.Query, tracker MigrationTracker) error {
	// Note: check-then-insert is not atomic — two concurrent migrators could
	// race into a PK violation on Create. Migrations are not expected to run
	// concurrently; revisit if that assumption changes.
	var existing []MigrationTracker
	if err := cloneQuery(query).Table(s.tableName).Where("id = ?", tracker.ID).Get(&existing); err != nil {
		return err
	}
	if len(existing) == 0 {
		return cloneQuery(query).Table(s.tableName).Create(&tracker)
	}
	_, err := cloneQuery(query).Table(s.tableName).Where("id = ?", tracker.ID).Update(map[string]any{
		"batch":         tracker.Batch,
		"description":   tracker.Description,
		"status":        tracker.Status,
		"error_message": tracker.ErrorMessage,
		"started_at":    tracker.StartedAt,
		"completed_at":  tracker.CompletedAt,
	})
	return err
}

// recordFailure persists a failed migration attempt in the tracker table.
// It runs outside the migration transaction, since the rollback removes any
// tracker rows written inside it. Best-effort: errors are swallowed because
// the migration error itself is already being returned to the caller.
func (s *Migrator) recordFailure(failure *migrationFailure) {
	schema := s.db.Schema()
	if err := s.ensureMigrationTracker(schema); err != nil {
		return
	}
	_ = s.upsertTracker(schema.Orm().Query(), MigrationTracker{
		ID:           failure.signature,
		Batch:        failure.batch,
		Description:  failure.description,
		Status:       failure.status,
		ErrorMessage: failure.err.Error(),
		StartedAt:    failure.startedAt,
		CompletedAt:  failure.completedAt,
	})
}

func (s *Migrator) getMigrationsByBatch(query orm.Query, batch int) ([]MigrationTracker, error) {
	trackers := make([]MigrationTracker, 0)
	if err := cloneQuery(query).Table(s.tableName).
		Where("batch = ?", batch).Where(appliedStatusFilter).Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func (s *Migrator) getLastMigrations(query orm.Query, step int) ([]MigrationTracker, error) {
	trackers := make([]MigrationTracker, 0)
	if err := cloneQuery(query).Table(s.tableName).
		Where(appliedStatusFilter).OrderBy("id", "desc").Limit(step).Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func (s *Migrator) rollbackMigration(schema contractsschema.Schema, query orm.Query, tracker MigrationTracker, failure **migrationFailure) error {
	// Find the migration by signature
	var migration MigrationInterface
	for _, m := range s.migrations {
		if m.Signature() == tracker.ID {
			migration = m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found in registered migrations", tracker.ID)
	}

	// Inject transaction-aware schema
	migration.SetSchema(schema)

	// Run the Down migration
	startedAt := time.Now()
	if err := migration.Down(); err != nil {
		if failure != nil {
			*failure = &migrationFailure{
				signature:   tracker.ID,
				description: migration.Description(),
				batch:       tracker.Batch,
				status:      MigrationTrackerStatusRollbackFailed,
				err:         err,
				startedAt:   startedAt,
				completedAt: time.Now(),
			}
		}
		return fmt.Errorf("failed to rollback migration %s: %w", tracker.ID, err)
	}

	// Delete the migration record from tracker
	_, err := cloneQuery(query).Table(s.tableName).Where("id = ?", tracker.ID).Delete()
	if err != nil {
		return err
	}
	return nil
}

// getMigrations returns all tracker rows regardless of status — used by
// Status(), which must report failed and running rows too.
func (s *Migrator) getMigrations() ([]MigrationTracker, error) {
	trackers := make([]MigrationTracker, 0)
	err := s.db.Schema().Orm().Query().Table(s.tableName).OrderBy("id", "asc").Get(&trackers)
	return trackers, err
}

func (s *Migrator) getMigrationsWithQuery(query orm.Query) ([]MigrationTracker, error) {
	trackers := make([]MigrationTracker, 0)
	if err := cloneQuery(query).Table(s.tableName).
		Where(appliedStatusFilter).OrderBy("id", "asc").Get(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

// ensureMigrationTracker creates the migration tracking table if it doesn't exist,
// and upgrades the schema by adding any missing columns.
func (s *Migrator) ensureMigrationTracker(schema contractsschema.Schema) error {
	// Create table if it doesn't exist
	if !schema.HasTable(s.tableName) {
		err := schema.Create(s.tableName, func(table contractsschema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.Integer("batch")
			table.String("description", 255)
			table.String("status", 16).Nullable()
			table.Text("error_message").Nullable()
			table.DateTime("started_at")
			table.DateTime("completed_at")
		})
		if err != nil {
			return fmt.Errorf("failed to create %s table: %w", s.tableName, err)
		}
		return nil
	}

	// Upgrade: add missing columns to existing table
	// Each addition is independent - if one fails, we still try the others.
	columnsToAdd := []struct {
		name string
		add  func(table contractsschema.Blueprint)
	}{
		{
			name: "description",
			add: func(table contractsschema.Blueprint) {
				table.String("description", 255)
			},
		},
		{
			name: "started_at",
			add: func(table contractsschema.Blueprint) {
				table.DateTime("started_at")
			},
		},
		{
			name: "completed_at",
			add: func(table contractsschema.Blueprint) {
				table.DateTime("completed_at")
			},
		},
		{
			name: "status",
			add: func(table contractsschema.Blueprint) {
				table.String("status", 16).Nullable()
			},
		},
		{
			name: "error_message",
			add: func(table contractsschema.Blueprint) {
				table.Text("error_message").Nullable()
			},
		},
	}

	for _, col := range columnsToAdd {
		if !schema.HasColumn(s.tableName, col.name) {
			_ = schema.Table(s.tableName, func(table contractsschema.Blueprint) {
				col.add(table)
			})
			// Intentionally ignoring error: column may already exist
			// (race condition or driver-specific behavior).
		}
	}

	return nil
}

func (s *Migrator) getAllTables(schema contractsschema.Schema) ([]string, error) {
	// Use schema.GetTableListing for driver-agnostic table discovery
	tables := schema.GetTableListing()

	// Filter out the migration tracking table
	var result []string
	for _, table := range tables {
		if table != s.tableName {
			result = append(result, table)
		}
	}
	return result, nil
}

func (s *Migrator) clearMigrationTracker(query orm.Query) error {
	_, err := cloneQuery(query).Table(s.tableName).Delete()
	return err
}

// parseIsolationLevel converts string isolation level to sql.IsolationLevel
func (s *Migrator) parseIsolationLevel(level string) sql.IsolationLevel {
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
