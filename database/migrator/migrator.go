package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/support/uid"
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
	recordID    string
	signature   string
	description string
	batch       int
	status      MigrationStatus // failed (Up) or rollback_failed (Down)
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
		recID := uid.GenerateShortID()
		tracker := MigrationTracker{
			ID:          recID,
			Migration:   signature,
			Batch:       batch,
			Description: migration.Description(),
			Status:      MigrationStatusRunning,
			StartedAt:   startedAt,
			CompletedAt: nullTime,
		}
		if err := s.createTracker(query, tracker); err != nil {
			return fmt.Errorf("failed to record migration %s start: %w", signature, err)
		}

		// Run migration
		if err := migration.Up(); err != nil {
			if failure != nil {
				*failure = &migrationFailure{
					recordID:    recID,
					signature:   signature,
					description: migration.Description(),
					batch:       batch,
					status:      MigrationStatusFailed,
					err:         err,
					startedAt:   startedAt,
					completedAt: time.Now(),
				}
			}
			_ = s.updateTrackerStatus(query, recID, MigrationStatusFailed, err.Error(), time.Now())
			return fmt.Errorf("migration %s failed: %w", signature, err)
		}
		completedAt := time.Now()

		// Mark migration completed
		if err := s.updateTrackerStatus(query, recID, MigrationStatusCompleted, "", completedAt); err != nil {
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
	if steps <= 0 {
		return fmt.Errorf("steps must be positive, got %d", steps)
	}

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
			return fmt.Errorf("failed to rollback migration %s: %w", migration.MigrationName(), err)
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
			return fmt.Errorf("failed to rollback migration %s: %w", migration.MigrationName(), err)
		}
	}

	return nil
}

// Status returns migration status for all registered migrations.
// Includes both completed migrations (from the tracker table) and
// pending migrations (registered but not yet run).
func (s *Migrator) Status() ([]MigrationStatusResponse, error) {
	var statuses []MigrationStatusResponse

	latestTracker := make(map[string]MigrationTracker)
	if s.db.Schema().HasTable(s.tableName) {
		trackers, err := s.getMigrations()
		if err != nil {
			return nil, fmt.Errorf("failed to get migrations: %w", err)
		}
		for _, t := range latestTrackers(trackers) {
			latestTracker[t.MigrationName()] = t
		}
	}

	processed := make(map[string]bool)
	for _, migration := range s.migrations {
		sig := migration.Signature()
		processed[sig] = true

		if t, ok := latestTracker[sig]; ok {
			statuses = append(statuses, trackerStatus(t, migration.Description()))
		} else {
			statuses = append(statuses, MigrationStatusResponse{
				Signature:   sig,
				Description: migration.Description(),
				State:       "pending",
			})
		}
	}

	for sig, t := range latestTracker {
		if !processed[sig] {
			statuses = append(statuses, trackerStatus(t, t.Description))
		}
	}

	// Sort by signature for consistent output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Signature < statuses[j].Signature
	})

	return statuses, nil
}

// trackerStatus builds the Status entry for a migration's latest tracker row.
// A rolled-back migration is pending again, so it is reported like a
// never-run migration (no batch, timestamps or error); its past attempts
// remain in the tracker table.
func trackerStatus(t MigrationTracker, pendingDescription string) MigrationStatusResponse {
	switch t.Status {
	case MigrationStatusRolledBack:
		return MigrationStatusResponse{
			Signature:   t.MigrationName(),
			Description: pendingDescription,
			State:       "pending",
		}
	case "":
		// Legacy row from before the status column existed; remove after
		// October 2027 (see MigrationTracker.Status).
		t.Status = MigrationStatusCompleted
	}
	return MigrationStatusResponse{
		Signature:   t.MigrationName(),
		Description: t.Description,
		Batch:       t.Batch,
		StartedAt:   t.StartedAt,
		CompletedAt: t.CompletedAt,
		State:       string(t.Status),
		Error:       t.ErrorMessage,
	}
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
	// for Fresh operations. However, it's still useful for the migration tracker updates.

	// Applied migrations are marked rolled back rather than deleted, so the
	// tracker keeps its full history across Fresh runs
	var applied []MigrationTracker
	if schema.HasTable(s.tableName) {
		var err error
		if applied, err = s.activeTrackers(query); err != nil {
			return fmt.Errorf("failed to get applied migrations: %w", err)
		}
	}

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

	droppedAt := time.Now()
	for i := len(applied) - 1; i >= 0; i-- {
		t := applied[i]
		if err := s.createTracker(query, MigrationTracker{
			ID:          uid.GenerateShortID(),
			Migration:   t.MigrationName(),
			Batch:       t.Batch,
			Description: t.Description,
			Status:      MigrationStatusRolledBack,
			StartedAt:   droppedAt,
			CompletedAt: droppedAt,
		}); err != nil {
			return fmt.Errorf("failed to record drop of migration %s: %w", t.MigrationName(), err)
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
			return fmt.Errorf("failed to rollback migration %s: %w", migration.MigrationName(), err)
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
	trackers, err := s.activeTrackers(query)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(trackers))
	for _, t := range trackers {
		ids = append(ids, t.MigrationName())
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

func (s *Migrator) createTracker(query orm.Query, tracker MigrationTracker) error {
	return cloneQuery(query).Table(s.tableName).Create(&tracker)
}

func (s *Migrator) updateTrackerStatus(query orm.Query, recordID string, status MigrationStatus, errMsg string, completedAt time.Time) error {
	updates := map[string]any{
		"status":       status,
		"completed_at": completedAt,
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	} else {
		updates["error_message"] = nil
	}
	_, err := cloneQuery(query).Table(s.tableName).Where("id = ?", recordID).Update(updates)
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
	query := schema.Orm().Query()
	recID := failure.recordID
	if recID != "" {
		var existing []MigrationTracker
		if err := cloneQuery(query).Table(s.tableName).Where("id = ?", recID).Get(&existing); err == nil && len(existing) > 0 {
			switch existing[0].Status {
			case "", MigrationStatusCompleted, MigrationStatusRolledBack:
				// "" is a legacy row from before the status column existed;
				// remove after October 2027 (see MigrationTracker.Status).
				// The existing row is already in a terminal state; don't overwrite it.
				// Insert a new failure row with a fresh ID below.
				recID = ""
			default:
				_ = s.updateTrackerStatus(query, recID, failure.status, failure.err.Error(), failure.completedAt)
				return
			}
		}
	}
	if recID == "" {
		recID = uid.GenerateShortID()
	}
	_ = s.createTracker(query, MigrationTracker{
		ID:           recID,
		Migration:    failure.signature,
		Batch:        failure.batch,
		Description:  failure.description,
		Status:       failure.status,
		ErrorMessage: failure.err.Error(),
		StartedAt:    failure.startedAt,
		CompletedAt:  failure.completedAt,
	})
}

func (s *Migrator) getMigrationsByBatch(query orm.Query, batch int) ([]MigrationTracker, error) {
	trackers, err := s.activeTrackers(query)
	if err != nil {
		return nil, err
	}
	inBatch := make([]MigrationTracker, 0)
	for _, t := range trackers {
		if t.Batch == batch {
			inBatch = append(inBatch, t)
		}
	}
	return inBatch, nil
}

func (s *Migrator) getLastMigrations(query orm.Query, step int) ([]MigrationTracker, error) {
	trackers, err := s.activeTrackers(query)
	if err != nil {
		return nil, err
	}
	if len(trackers) <= step {
		return trackers, nil
	}
	return trackers[len(trackers)-step:], nil
}

func (s *Migrator) rollbackMigration(schema contractsschema.Schema, query orm.Query, tracker MigrationTracker, failure **migrationFailure) error {
	sig := tracker.MigrationName()
	// Find the migration by signature
	var migration MigrationInterface
	for _, m := range s.migrations {
		if m.Signature() == sig {
			migration = m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found in registered migrations", sig)
	}

	// Inject transaction-aware schema
	migration.SetSchema(schema)

	// Run the Down migration. Its outcome is appended as a new tracker row,
	// leaving the rows of earlier attempts untouched.
	record := MigrationTracker{
		ID:          uid.GenerateShortID(),
		Migration:   sig,
		Batch:       tracker.Batch,
		Description: migration.Description(),
		StartedAt:   time.Now(),
	}
	if err := migration.Down(); err != nil {
		record.Status = MigrationStatusRollbackFailed
		record.ErrorMessage = err.Error()
		record.CompletedAt = time.Now()
		if failure != nil {
			*failure = &migrationFailure{
				recordID:    record.ID,
				signature:   sig,
				description: record.Description,
				batch:       record.Batch,
				status:      record.Status,
				err:         err,
				startedAt:   record.StartedAt,
				completedAt: record.CompletedAt,
			}
		}
		_ = s.createTracker(query, record)
		return fmt.Errorf("failed to rollback migration %s: %w", sig, err)
	}

	record.Status = MigrationStatusRolledBack
	record.CompletedAt = time.Now()
	if err := s.createTracker(query, record); err != nil {
		return fmt.Errorf("failed to record rollback of migration %s: %w", sig, err)
	}
	return nil
}

// getMigrations returns all tracker rows regardless of status — used by
// Status(), which must report failed and running rows too.
func (s *Migrator) getMigrations() ([]MigrationTracker, error) {
	return s.trackerHistory(s.db.Schema().Orm().Query())
}

func (s *Migrator) getMigrationsWithQuery(query orm.Query) ([]MigrationTracker, error) {
	return s.activeTrackers(query)
}

// trackerHistory returns every tracker row in execution order. started_at has
// only second precision on SQLite, so the record ID breaks ties between attempts
// recorded within the same second. New time-ordered short IDs sort after legacy
// signature IDs so that a post-upgrade attempt in the same second is treated as
// later; within the same ID scheme, lexicographic ID order is chronological.
func (s *Migrator) trackerHistory(query orm.Query) ([]MigrationTracker, error) {
	trackers := make([]MigrationTracker, 0)
	if err := cloneQuery(query).Table(s.tableName).Get(&trackers); err != nil {
		return nil, err
	}
	sort.Slice(trackers, func(i, j int) bool {
		if !trackers[i].StartedAt.Equal(trackers[j].StartedAt) {
			return trackers[i].StartedAt.Before(trackers[j].StartedAt)
		}
		iGen := isGeneratedTrackerID(trackers[i].ID)
		jGen := isGeneratedTrackerID(trackers[j].ID)
		if iGen != jGen {
			// Legacy signature IDs predate generated short IDs, so they sort first.
			return !iGen
		}
		return trackers[i].ID < trackers[j].ID
	})
	return trackers, nil
}

// isGeneratedTrackerID reports whether id looks like a generated short ID
// produced by uid.GenerateShortID (11 characters from the Crockford alphabet).
// It is used as a tie-breaker when legacy signature IDs and generated IDs share
// the same started_at second.
func isGeneratedTrackerID(id string) bool {
	const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"
	if len(id) != 11 {
		return false
	}
	for i := 0; i < len(id); i++ {
		if !strings.ContainsRune(alphabet, rune(id[i])) {
			return false
		}
	}
	return true
}

// latestTrackers reduces the append-only tracker history to the most recent
// row of each migration, preserving execution order.
func latestTrackers(history []MigrationTracker) []MigrationTracker {
	latest := make(map[string]string, len(history))
	for _, t := range history {
		latest[t.MigrationName()] = t.ID
	}
	result := make([]MigrationTracker, 0, len(latest))
	for _, t := range history {
		if latest[t.MigrationName()] == t.ID {
			result = append(result, t)
		}
	}
	return result
}

// activeTrackers returns the latest row of every migration whose schema
// change is currently applied, in execution order. A failed rollback leaves
// the change applied, so it stays active and eligible for a rollback retry.
func (s *Migrator) activeTrackers(query orm.Query) ([]MigrationTracker, error) {
	history, err := s.trackerHistory(query)
	if err != nil {
		return nil, err
	}
	active := make([]MigrationTracker, 0)
	for _, t := range latestTrackers(history) {
		switch t.Status {
		case "", MigrationStatusCompleted, MigrationStatusRollbackFailed:
			// "" is a legacy row from before the status column existed;
			// remove after October 2027 (see MigrationTracker.Status).
			active = append(active, t)
		}
	}
	return active, nil
}

// ensureMigrationTracker creates the migration tracking table if it doesn't exist,
// and upgrades the schema by adding any missing columns.
func (s *Migrator) ensureMigrationTracker(schema contractsschema.Schema) error {
	// Create table if it doesn't exist
	if !schema.HasTable(s.tableName) {
		err := schema.Create(s.tableName, func(table contractsschema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("migration", 255).Nullable()
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
			name: "migration",
			add: func(table contractsschema.Blueprint) {
				table.String("migration", 255).Nullable()
			},
		},
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
