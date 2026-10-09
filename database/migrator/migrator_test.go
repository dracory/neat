package migrator

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
)

// MockMigration is a test migration implementation
type MockMigration struct {
	signature      string
	description    string
	upCalled       bool
	downCalled     bool
	schema         contractsschema.Schema
	shouldFail     bool
	downShouldFail bool
}

func (m *MockMigration) Signature() string {
	return m.signature
}

func (m *MockMigration) Description() string {
	return m.description
}

func (m *MockMigration) Up() error {
	m.upCalled = true
	if m.shouldFail {
		return fmt.Errorf("mock migration failure")
	}
	return nil
}

func (m *MockMigration) Down() error {
	m.downCalled = true
	if m.downShouldFail {
		return fmt.Errorf("mock migration down failure")
	}
	return nil
}

func (m *MockMigration) SetSchema(schema contractsschema.Schema) {
	m.schema = schema
}

func (m *MockMigration) GetSchema() contractsschema.Schema {
	return m.schema
}

// TestNewMigrator verifies that NewMigrator returns a configured Migrator
// with the given database, an empty migration list, and the default tracker
// table name.
func TestNewMigrator(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	if migrator == nil {
		t.Error("Expected non-nil migrator")
	}

	impl, ok := migrator.(*Migrator)
	if !ok {
		t.Error("Expected Migrator type")
	}
	if impl.db != db {
		t.Error("Expected db to be set")
	}
	if len(impl.migrations) != 0 {
		t.Error("Expected empty migrations list")
	}
	if impl.tableName != defaultTableName {
		t.Errorf("Expected default table name '%s', got '%s'", defaultTableName, impl.tableName)
	}
}

// TestSetTableName_Valid verifies that SetTableName accepts identifier-style
// names (letters, digits, underscores; not starting with a digit) and stores
// the chosen tracker table name.
func TestSetTableName_Valid(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)

	validNames := []string{"migrations", "my_migrations", "schema_migrations", "_tracker"}
	for _, name := range validNames {
		err := migrator.SetTableName(name)
		if err != nil {
			t.Errorf("Expected SetTableName('%s') to succeed, got error: %v", name, err)
		}

		impl := migrator.(*Migrator)
		if impl.tableName != name {
			t.Errorf("Expected table name '%s', got '%s'", name, impl.tableName)
		}
	}
}

// TestSetTableName_Invalid verifies that SetTableName rejects empty names,
// names starting with a digit, names containing invalid characters, and SQL
// keywords — and leaves the previous table name unchanged on failure.
func TestSetTableName_Invalid(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)

	invalidNames := []string{"", "1migrations", "migration-tracker", "SELECT", "DROP", "TABLE"}
	for _, name := range invalidNames {
		err := migrator.SetTableName(name)
		if err == nil {
			t.Errorf("Expected SetTableName('%s') to fail, got nil", name)
		}

		// Table name should remain unchanged
		impl := migrator.(*Migrator)
		if impl.tableName != defaultTableName {
			t.Errorf("Expected table name to remain '%s' after failed SetTableName, got '%s'", defaultTableName, impl.tableName)
		}
	}
}

// TestSetTableName_UsedForTracking verifies that after SetTableName, Up
// creates and writes to the custom tracker table and does not create the
// default one.
func TestSetTableName_UsedForTracking(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	customTable := "my_migrations"
	if err := migrator.SetTableName(customTable); err != nil {
		t.Fatalf("SetTableName failed: %v", err)
	}

	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	// Verify the custom table was created, not the default one
	if !db.Schema().HasTable(customTable) {
		t.Errorf("Expected custom table '%s' to be created", customTable)
	}
	if db.Schema().HasTable(defaultTableName) {
		t.Error("Expected default table NOT to be created when custom name is set")
	}
}

// TestAddMigration verifies that a single migration is registered.
func TestAddMigration(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}

	err = migrator.AddMigration(migration)
	if err != nil {
		t.Errorf("AddMigration failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if len(impl.migrations) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(impl.migrations))
	}
}

// TestAddMigrations verifies that several migrations are registered in one call.
func TestAddMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
		&MockMigration{signature: "migration_3", description: "Third migration"},
	}

	err = migrator.AddMigrations(migrations)
	if err != nil {
		t.Errorf("AddMigrations failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if len(impl.migrations) != 3 {
		t.Errorf("Expected 3 migrations, got %d", len(impl.migrations))
	}
}

// TestAddMigration_Duplicate verifies that registering the same signature
// twice fails with a "duplicate migration signature" error and that only the
// first registration is kept.
func TestAddMigration_Duplicate(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}

	// First registration succeeds
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("first AddMigration should succeed: %v", err)
	}

	// Second registration with same signature fails
	err = migrator.AddMigration(migration)
	if err == nil {
		t.Fatalf("expected error on duplicate migration, got nil")
	}
	if err.Error() != "duplicate migration signature: test_migration" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Verify only one migration was added
	impl := migrator.(*Migrator)
	if len(impl.migrations) != 1 {
		t.Errorf("expected 1 migration, got %d", len(impl.migrations))
	}
}

// TestAddMigration_EmptySignature verifies that a migration with an empty
// signature is rejected at registration time.
func TestAddMigration_EmptySignature(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{
		signature:   "",
		description: "Test migration with empty signature",
	}

	err = migrator.AddMigration(migration)
	if err == nil {
		t.Fatalf("expected error on empty signature, got nil")
	}
	if err.Error() != "migration signature cannot be empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestAddMigrations_DuplicateInSlice verifies that AddMigrations fails when
// two entries in the same slice share a signature, and that entries added
// before the error stay registered.
func TestAddMigrations_DuplicateInSlice(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
		&MockMigration{signature: "migration_1", description: "Duplicate migration"},
	}

	err = migrator.AddMigrations(migrations)
	if err == nil {
		t.Fatalf("expected error on duplicate in slice, got nil")
	}
	if err.Error() != "duplicate migration signature: migration_1" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Verify only first two were added before error
	impl := migrator.(*Migrator)
	if len(impl.migrations) != 2 {
		t.Errorf("expected 2 migrations, got %d", len(impl.migrations))
	}
}

// TestAddMigrations_DuplicateAcrossCalls verifies that AddMigrations rejects
// signatures already registered by a previous call, while entries added
// before the duplicate are kept.
func TestAddMigrations_DuplicateAcrossCalls(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)

	firstBatch := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
	}
	secondBatch := []MigrationInterface{
		&MockMigration{signature: "migration_3", description: "Third migration"},
		&MockMigration{signature: "migration_1", description: "Duplicate of first"},
	}

	if err := migrator.AddMigrations(firstBatch); err != nil {
		t.Fatalf("first AddMigrations should succeed: %v", err)
	}

	err = migrator.AddMigrations(secondBatch)
	if err == nil {
		t.Fatalf("expected error on duplicate across calls, got nil")
	}
	if err.Error() != "duplicate migration signature: migration_1" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Verify first batch + only migration_3 from second batch
	impl := migrator.(*Migrator)
	if len(impl.migrations) != 3 {
		t.Errorf("expected 3 migrations, got %d", len(impl.migrations))
	}
}

// TestUp_AutoCreateMigrationTracker verifies that Up creates the migration
// tracker table on the first run when it does not exist yet.
func TestUp_AutoCreateMigrationTracker(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	// Verify migration tracking table was created
	if !db.Schema().HasTable(defaultTableName) {
		t.Error("Expected migration tracking table to be auto-created")
	}
}

// TestUp_SchemaInjection verifies that Up injects the database schema — with
// a working ORM — into each migration before running it.
func TestUp_SchemaInjection(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration tracking table first
	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	if !migration.upCalled {
		t.Error("Expected migration Up to be called")
	}
	if migration.schema == nil {
		t.Error("Expected schema to be injected")
	}
	if migration.schema.Orm() == nil {
		t.Error("Expected injected schema to have a valid ORM")
	}
}

// TestUp_EmptySignature verifies that an empty signature is rejected before
// any migration runs (registration itself fails).
func TestUp_EmptySignature(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migration := &MockMigration{
		signature:   "",
		description: "Test migration",
	}

	err = migrator.AddMigration(migration)
	if err == nil {
		t.Fatal("expected AddMigration to fail for empty signature")
	}
	if err.Error() != "migration signature cannot be empty" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestUp_SignatureValidation_DateTime verifies that with datetime signature
// validation enabled, a "YYYY_MM_DD_HHMM_name" signature is accepted and the
// migration runs.
func TestUp_SignatureValidation_DateTime(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrator.SetSignatureValidation(true, SignatureFormatDateTime)

	validMigration := &MockMigration{
		signature:   "2026_06_15_1200_create_users_table",
		description: "Valid datetime signature",
	}
	if err := migrator.AddMigration(validMigration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Expected valid datetime signature to pass, got error: %v", err)
	}
	if !validMigration.upCalled {
		t.Error("Expected valid migration Up to be called")
	}
}

// TestUp_SignatureValidation_InvalidFormat verifies that with datetime
// validation enabled, a non-conforming signature makes Up fail and the
// migration is not executed.
func TestUp_SignatureValidation_InvalidFormat(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrator.SetSignatureValidation(true, SignatureFormatDateTime)

	invalidMigration := &MockMigration{
		signature:   "test_migration",
		description: "Invalid datetime signature",
	}
	if err := migrator.AddMigration(invalidMigration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err == nil {
		t.Error("Expected error for invalid signature format")
	}
	if invalidMigration.upCalled {
		t.Error("Expected invalid migration Up NOT to be called")
	}
}

// TestUp_SignatureValidation_Disabled verifies that with validation disabled,
// any signature format is accepted and the migration runs.
func TestUp_SignatureValidation_Disabled(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrator.SetSignatureValidation(false, SignatureFormatDateTime)

	migration := &MockMigration{
		signature:   "totally_arbitrary_name",
		description: "Arbitrary signature",
	}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Expected arbitrary signature to pass when validation disabled, got error: %v", err)
	}
	if !migration.upCalled {
		t.Error("Expected migration Up to be called when validation disabled")
	}
}

// TestEnsureMigrationTracker_CreatesTable verifies that ensureMigrationTracker
// creates the tracker table with all expected columns when it is missing.
func TestEnsureMigrationTracker_CreatesTable(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()

	if schema.HasTable(defaultTableName) {
		t.Fatal("Expected table to NOT exist initially")
	}

	s := NewMigrator(db).(*Migrator)
	if err := s.ensureMigrationTracker(schema); err != nil {
		t.Fatalf("ensureMigrationTracker failed: %v", err)
	}

	if !schema.HasTable(defaultTableName) {
		t.Fatal("Expected table to exist after ensureMigrationTracker")
	}

	expectedColumns := []string{"id", "migration", "batch", "description", "started_at", "completed_at"}
	for _, col := range expectedColumns {
		if !schema.HasColumn(defaultTableName, col) {
			t.Errorf("Expected column '%s' to exist in '%s'", col, defaultTableName)
		}
	}
}

// TestEnsureMigrationTracker_UpgradesExistingTable verifies that
// ensureMigrationTracker adds the columns missing from a tracker table
// created by an older version (schema upgrade path).
func TestEnsureMigrationTracker_UpgradesExistingTable(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()

	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
	})
	if err != nil {
		t.Fatalf("failed to create old-style table: %v", err)
	}

	missingColumns := []string{"migration", "description", "started_at", "completed_at"}
	for _, col := range missingColumns {
		if schema.HasColumn(defaultTableName, col) {
			t.Fatalf("Expected column '%s' to NOT exist before upgrade", col)
		}
	}

	s := NewMigrator(db).(*Migrator)
	if err := s.ensureMigrationTracker(schema); err != nil {
		t.Fatalf("ensureMigrationTracker failed: %v", err)
	}

	allColumns := []string{"id", "migration", "batch", "description", "started_at", "completed_at"}
	for _, col := range allColumns {
		if !schema.HasColumn(defaultTableName, col) {
			t.Errorf("Expected column '%s' to exist after upgrade", col)
		}
	}
}

// TestUp_SkipAlreadyRun verifies that a second Up does not re-run a migration
// that is already recorded as applied.
func TestUp_SkipAlreadyRun(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("First Up failed: %v", err)
	}

	migration.upCalled = false

	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Second Up failed: %v", err)
	}

	if migration.upCalled {
		t.Error("Expected migration to be skipped on second run")
	}
}

// TestDown verifies that Down rolls back the most recently applied migration
// without error.
func TestDown(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	err = migrator.Down(ctx)
	if err != nil {
		t.Errorf("Down failed: %v", err)
	}
}

// TestRollbackSteps verifies that RollbackSteps rolls back the given number
// of most recently applied migrations without error.
func TestRollbackSteps(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
		&MockMigration{signature: "migration_3", description: "Third migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	err = migrator.RollbackSteps(ctx, 2)
	if err != nil {
		t.Errorf("RollbackSteps failed: %v", err)
	}
}

// TestRollbackToBatch verifies that RollbackToBatch rolls back every
// migration belonging to the given batch without error.
func TestRollbackToBatch(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	status, err := migrator.Status()
	if err != nil {
		t.Errorf("Status failed: %v", err)
	}
	if len(status) == 0 {
		t.Fatal("Expected at least one migration status")
	}
	batch := status[0].Batch

	err = migrator.RollbackToBatch(ctx, batch)
	if err != nil {
		t.Errorf("RollbackToBatch failed: %v", err)
	}
}

// TestStatus_NoMigrationTrackerTable verifies that Status returns an empty
// list when the tracker table does not exist.
func TestStatus_NoMigrationTrackerTable(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)

	status, err := migrator.Status()
	if err != nil {
		t.Errorf("Status failed: %v", err)
	}
	if len(status) != 0 {
		t.Errorf("Expected empty status when migration tracking table does not exist, got %d", len(status))
	}
}

// TestStatus_WithMigrations verifies that Status reports each applied
// migration as "completed", keyed by its signature.
func TestStatus_WithMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	status, err := migrator.Status()
	if err != nil {
		t.Errorf("Status failed: %v", err)
	}
	if len(status) != 2 {
		t.Errorf("Expected 2 migration statuses, got %d", len(status))
	}

	for i, s := range status {
		if s.State != "completed" {
			t.Errorf("Expected state 'completed' for migration %d, got '%s'", i, s.State)
		}
		if s.ID != migrations[i].Signature() {
			t.Errorf("Expected ID '%s' for migration %d, got '%s'", migrations[i].Signature(), i, s.ID)
		}
	}
}

// TestStatus_WithPendingMigrations verifies that Status merges tracker rows
// with registered migrations: applied ones report "completed" with their
// batch, registered-but-not-run ones report "pending" with batch 0, and the
// ordering stays stable.
func TestStatus_WithPendingMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)

	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
		&MockMigration{signature: "migration_3", description: "Third migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	pendingMigration := &MockMigration{signature: "migration_4", description: "Fourth migration"}
	if err := migrator.AddMigration(pendingMigration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	status, err := migrator.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if len(status) != 4 {
		t.Fatalf("expected 4 statuses (3 completed + 1 pending), got %d", len(status))
	}

	completed := 0
	pending := 0
	for _, s := range status {
		switch s.State {
		case "completed":
			completed++
			if s.Batch == 0 {
				t.Errorf("completed migration %s should have a batch number", s.ID)
			}
		case "pending":
			pending++
			if s.Batch != 0 {
				t.Errorf("pending migration %s should have batch 0, got %d", s.ID, s.Batch)
			}
		default:
			t.Errorf("unexpected state '%s' for migration %s", s.State, s.ID)
		}
	}

	if completed != 3 {
		t.Errorf("expected 3 completed migrations, got %d", completed)
	}
	if pending != 1 {
		t.Errorf("expected 1 pending migration, got %d", pending)
	}

	expectedOrder := []string{"migration_1", "migration_2", "migration_3", "migration_4"}
	for i, expected := range expectedOrder {
		if status[i].ID != expected {
			t.Errorf("expected status[%d].ID = '%s', got '%s'", i, expected, status[i].ID)
		}
	}
}

// TestStatus_AllPending verifies that with no tracker data, every registered
// migration reports "pending".
func TestStatus_AllPending(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_a", description: "A migration"},
		&MockMigration{signature: "migration_b", description: "B migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	status, err := migrator.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if len(status) != 2 {
		t.Fatalf("expected 2 pending statuses, got %d", len(status))
	}

	for _, s := range status {
		if s.State != "pending" {
			t.Errorf("expected state 'pending', got '%s'", s.State)
		}
	}
}

// TestStatus_AllCompleted verifies that once all migrations ran, every entry
// reports "completed".
func TestStatus_AllCompleted(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	status, err := migrator.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if len(status) != 2 {
		t.Fatalf("expected 2 completed statuses, got %d", len(status))
	}

	for _, s := range status {
		if s.State != "completed" {
			t.Errorf("expected state 'completed', got '%s'", s.State)
		}
	}
}

// TestNewMigratorWithOptions_NilOpts verifies that a nil Options produces a
// Migrator with default settings (default table name, lexicographical
// ordering enabled).
func TestNewMigratorWithOptions_NilOpts(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator, err := NewMigratorWithOptions(db, nil)
	if err != nil {
		t.Fatalf("NewMigratorWithOptions(nil) failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if impl.tableName != defaultTableName {
		t.Errorf("expected default table name, got '%s'", impl.tableName)
	}
	if !impl.lexicographicalOrdering {
		t.Error("expected lexicographical ordering to be enabled by default")
	}
}

// TestNewMigratorWithOptions_EmptyOpts verifies that an empty Options struct
// produces a Migrator with default settings.
func TestNewMigratorWithOptions_EmptyOpts(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator, err := NewMigratorWithOptions(db, &Options{})
	if err != nil {
		t.Fatalf("NewMigratorWithOptions(&Options{}) failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if impl.tableName != defaultTableName {
		t.Errorf("expected default table name, got '%s'", impl.tableName)
	}
}

// TestNewMigratorWithOptions_TableName verifies that Options.TableName
// overrides the default tracker table name.
func TestNewMigratorWithOptions_TableName(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator, err := NewMigratorWithOptions(db, &Options{
		TableName: "my_migrations",
	})
	if err != nil {
		t.Fatalf("NewMigratorWithOptions failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if impl.tableName != "my_migrations" {
		t.Errorf("expected table name 'my_migrations', got '%s'", impl.tableName)
	}
}

// TestNewMigratorWithOptions_InvalidTableName verifies that an invalid
// Options.TableName makes NewMigratorWithOptions fail instead of producing
// a migrator.
func TestNewMigratorWithOptions_InvalidTableName(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = NewMigratorWithOptions(db, &Options{
		TableName: "1invalid",
	})
	if err == nil {
		t.Error("expected error for invalid table name, got nil")
	}
}

// TestNewMigratorWithOptions_IsolationLevel verifies that
// Options.TransactionIsolationLevel is stored on the migrator.
func TestNewMigratorWithOptions_IsolationLevel(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator, err := NewMigratorWithOptions(db, &Options{
		TransactionIsolationLevel: "SERIALIZABLE",
	})
	if err != nil {
		t.Fatalf("NewMigratorWithOptions failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if impl.isolationLevel != "SERIALIZABLE" {
		t.Errorf("expected isolation level 'SERIALIZABLE', got '%s'", impl.isolationLevel)
	}
}

// TestNewMigratorWithOptions_LexicographicalOrdering verifies that
// Options.LexicographicalOrdering can disable signature ordering.
func TestNewMigratorWithOptions_LexicographicalOrdering(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator, err := NewMigratorWithOptions(db, &Options{
		LexicographicalOrdering: false,
	})
	if err != nil {
		t.Fatalf("NewMigratorWithOptions failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if impl.lexicographicalOrdering {
		t.Error("expected lexicographical ordering to be disabled")
	}
}

// TestNewMigratorWithOptions_SignatureValidation verifies that the
// Options signature-validation fields are stored on the migrator.
func TestNewMigratorWithOptions_SignatureValidation(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator, err := NewMigratorWithOptions(db, &Options{
		SignatureValidationEnabled: true,
		SignatureValidationFormat:  SignatureFormatDateTime,
	})
	if err != nil {
		t.Fatalf("NewMigratorWithOptions failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if !impl.sigValidation {
		t.Error("expected signature validation to be enabled")
	}
	if impl.sigValidationFormat != SignatureFormatDateTime {
		t.Errorf("expected signature format 'datetime', got '%s'", impl.sigValidationFormat)
	}
}

// TestNewMigratorWithOptions_Full verifies that all Options fields are
// applied together on a single migrator.
func TestNewMigratorWithOptions_Full(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator, err := NewMigratorWithOptions(db, &Options{
		TableName:                  "custom_migrations",
		TransactionIsolationLevel:  "READ COMMITTED",
		LexicographicalOrdering:    true,
		SignatureValidationEnabled: true,
		SignatureValidationFormat:  SignatureFormatDate,
	})
	if err != nil {
		t.Fatalf("NewMigratorWithOptions failed: %v", err)
	}

	impl := migrator.(*Migrator)
	if impl.tableName != "custom_migrations" {
		t.Errorf("expected table name 'custom_migrations', got '%s'", impl.tableName)
	}
	if impl.isolationLevel != "READ COMMITTED" {
		t.Errorf("expected isolation level 'READ COMMITTED', got '%s'", impl.isolationLevel)
	}
	if !impl.lexicographicalOrdering {
		t.Error("expected lexicographical ordering to be enabled")
	}
	if !impl.sigValidation {
		t.Error("expected signature validation to be enabled")
	}
	if impl.sigValidationFormat != SignatureFormatDate {
		t.Errorf("expected signature format 'date', got '%s'", impl.sigValidationFormat)
	}
}

// TestFresh verifies that Fresh drops the tables, records the rollback rows,
// re-runs Up, and leaves the applied migrations tracked as completed.
func TestFresh(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)

	userMigration := &MockMigration{
		signature:   "2026_06_15_1200_create_users",
		description: "Create users table",
	}
	if err := migrator.AddMigration(userMigration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()

	err = migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	status, err := migrator.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if len(status) != 1 {
		t.Fatalf("Expected 1 tracked migration, got %d", len(status))
	}

	err = migrator.Fresh(ctx)
	if err != nil {
		t.Fatalf("Fresh failed: %v", err)
	}

	if !db.Schema().HasTable(defaultTableName) {
		t.Error("Expected migration tracking table to exist after Fresh")
	}

	status, err = migrator.Status()
	if err != nil {
		t.Fatalf("Status after Fresh failed: %v", err)
	}
	if len(status) != 1 {
		t.Errorf("Expected 1 tracked migration after Fresh, got %d", len(status))
	}
}

// TestReset verifies that Reset rolls back all applied migrations without
// error.
func TestReset(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	err = migrator.Reset(ctx)
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}
}

// TestReset_SafetyLimit verifies that Reset aborts with a "too many
// migrations" error when the number of active tracker rows exceeds
// maxResetIterations, guarding against runaway rollback loops.
func TestReset_SafetyLimit(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
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
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)

	query := db.Schema().Orm().Query()
	for i := 0; i < maxResetIterations+1; i++ {
		tracker := MigrationTracker{
			ID:        fmt.Sprintf("rec_%d", i),
			Migration: fmt.Sprintf("migration_%d", i),
			Batch:     1,
		}
		if err := query.Table(defaultTableName).Create(&tracker); err != nil {
			t.Fatalf("failed to seed tracker: %v", err)
		}
	}

	ctx := context.Background()
	err = migrator.Reset(ctx)
	if err == nil {
		t.Fatal("Expected Reset to fail with safety limit exceeded")
	}
	if !containsSubstringHelper(err.Error(), "too many migrations") {
		t.Errorf("Expected 'too many migrations' error, got '%s'", err.Error())
	}
}

// TestSetTransactionsEnabled verifies that transactions are enabled by
// default and that SetTransactionsEnabled toggles the flag.
func TestSetTransactionsEnabled(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	impl := migrator.(*Migrator)

	if !impl.useTransactions {
		t.Error("Expected transactions to be enabled by default")
	}

	impl.SetTransactionsEnabled(false)
	if impl.useTransactions {
		t.Error("Expected transactions to be disabled")
	}

	impl.SetTransactionsEnabled(true)
	if !impl.useTransactions {
		t.Error("Expected transactions to be enabled")
	}
}

// TestSetTransactionIsolationLevel verifies that SetTransactionIsolationLevel
// stores each known level string verbatim.
func TestSetTransactionIsolationLevel(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	impl := migrator.(*Migrator)

	impl.SetTransactionIsolationLevel("SERIALIZABLE")
	if impl.isolationLevel != "SERIALIZABLE" {
		t.Error("Expected isolation level to be SERIALIZABLE")
	}

	levels := []string{"READ UNCOMMITTED", "READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"}
	for _, level := range levels {
		impl.SetTransactionIsolationLevel(level)
		if impl.isolationLevel != level {
			t.Errorf("Expected isolation level to be %s, got %s", level, impl.isolationLevel)
		}
	}
}

// TestUpWithTransactionsEnabled verifies that when a migration fails inside
// the transaction, the earlier tracker rows roll back with it and only a
// single "failed" row — with error message and timestamps — is persisted for
// the failed migration.
func TestUpWithTransactionsEnabled(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration", shouldFail: true},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err == nil {
		t.Fatal("Expected error from failing migration")
	}

	var trackers []MigrationTracker
	query := db.Schema().Orm().Query().Table(defaultTableName)
	if err := query.Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 1 {
		t.Fatalf("Expected 1 tracker entry (the failed migration), got %d", len(trackers))
	}
	if trackers[0].MigrationName() != "migration_2" || trackers[0].Status != MigrationTrackerStatusFailed {
		t.Errorf("Expected migration_2 with status 'failed', got %s/%s", trackers[0].MigrationName(), trackers[0].Status)
	}
	if trackers[0].ErrorMessage == "" {
		t.Error("Expected error message to be recorded for failed migration")
	}
	if trackers[0].StartedAt.IsZero() || trackers[0].CompletedAt.IsZero() {
		t.Error("Expected started_at and completed_at to be set on failed migration")
	}
}

// TestUpWithTransactionsDisabled verifies that without a transaction, the
// earlier successful migration stays recorded as "completed" and the failure
// is appended as a second "failed" row.
func TestUpWithTransactionsDisabled(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	impl := migrator.(*Migrator)
	impl.SetTransactionsEnabled(false)

	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration", shouldFail: true},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err == nil {
		t.Fatal("Expected error from failing migration")
	}

	var trackers []MigrationTracker
	query := db.Schema().Orm().Query().Table(defaultTableName)
	if err := query.Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 2 {
		t.Fatalf("Expected 2 tracker entries after failed migration, got %d", len(trackers))
	}
	statusByID := map[string]MigrationTracker{}
	for _, tr := range trackers {
		statusByID[tr.MigrationName()] = tr
	}
	if statusByID["migration_1"].Status != MigrationTrackerStatusCompleted {
		t.Errorf("Expected migration_1 status 'completed', got '%s'", statusByID["migration_1"].Status)
	}
	if statusByID["migration_2"].Status != MigrationTrackerStatusFailed {
		t.Errorf("Expected migration_2 status 'failed', got '%s'", statusByID["migration_2"].Status)
	}
	if statusByID["migration_2"].ErrorMessage == "" {
		t.Error("Expected error message to be recorded for failed migration")
	}
}

// TestLexicographicalOrdering_Default verifies that lexicographical signature
// ordering is enabled by default.
func TestLexicographicalOrdering_Default(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	impl := migrator.(*Migrator)

	if !impl.lexicographicalOrdering {
		t.Error("Expected lexicographical ordering to be enabled by default")
	}
}

// TestLexicographicalOrdering_Enabled verifies that when ordering is enabled,
// Up runs migrations in signature order regardless of registration order.
func TestLexicographicalOrdering_Enabled(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrator.SetLexicographicalOrdering(true)

	migrations := []MigrationInterface{
		&MockMigration{signature: "2026_06_15_1400_third", description: "Third migration"},
		&MockMigration{signature: "2026_06_15_1200_first", description: "First migration"},
		&MockMigration{signature: "2026_06_15_1300_second", description: "Second migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	var trackers []MigrationTracker
	query := db.Schema().Orm().Query().Table(defaultTableName).OrderBy("started_at", "asc")
	if err := query.Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 3 {
		t.Fatalf("Expected 3 tracker entries, got %d", len(trackers))
	}

	expectedOrder := []string{"2026_06_15_1200_first", "2026_06_15_1300_second", "2026_06_15_1400_third"}
	for i, expected := range expectedOrder {
		if trackers[i].MigrationName() != expected {
			t.Errorf("Expected migration at position %d to be '%s', got '%s'", i, expected, trackers[i].MigrationName())
		}
	}
}

// TestLexicographicalOrdering_Disabled verifies that when ordering is
// disabled, Up runs migrations in registration order.
func TestLexicographicalOrdering_Disabled(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	migrator := NewMigrator(db)
	migrator.SetLexicographicalOrdering(false)

	migrations := []MigrationInterface{
		&MockMigration{signature: "2026_06_15_1400_third", description: "Third migration"},
		&MockMigration{signature: "2026_06_15_1200_first", description: "First migration"},
		&MockMigration{signature: "2026_06_15_1300_second", description: "Second migration"},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	var trackers []MigrationTracker
	query := db.Schema().Orm().Query().Table(defaultTableName).OrderBy("started_at", "asc")
	if err := query.Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 3 {
		t.Fatalf("Expected 3 tracker entries, got %d", len(trackers))
	}

	expectedOrder := []string{"2026_06_15_1400_third", "2026_06_15_1200_first", "2026_06_15_1300_second"}
	for i, expected := range expectedOrder {
		if trackers[i].MigrationName() != expected {
			t.Errorf("Expected migration at position %d to be '%s', got '%s'", i, expected, trackers[i].MigrationName())
		}
	}
}

// TestTransactionRollbackOnFailure verifies that on a transactional failure
// the whole batch rolls back and a single "failed" tracker row is
// re-recorded afterwards outside the transaction.
func TestTransactionRollbackOnFailure(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migrations := []MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration", shouldFail: true},
	}
	if err := migrator.AddMigrations(migrations); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	err = migrator.Up(ctx)
	if err == nil {
		t.Fatal("Expected Up to return error when migration fails")
	}

	if !db.Schema().HasTable(defaultTableName) {
		t.Fatal("Expected migration tracking table to exist after failure is recorded")
	}
	var trackers []MigrationTracker
	query := db.Schema().Orm().Query().Table(defaultTableName)
	if err := query.Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 1 {
		t.Fatalf("Expected 1 tracker entry (the failed migration), got %d", len(trackers))
	}
	if trackers[0].MigrationName() != "migration_2" || trackers[0].Status != MigrationTrackerStatusFailed {
		t.Errorf("Expected migration_2 with status 'failed', got %s/%s", trackers[0].MigrationName(), trackers[0].Status)
	}
	if trackers[0].ErrorMessage == "" {
		t.Error("Expected error message to be recorded for failed migration")
	}
}

// TestFailedMigration_RetriedOnNextUp verifies that a failed migration is
// retried by the next Up, and that both the failed and the successful
// attempts are kept in the append-only history.
func TestFailedMigration_RetriedOnNextUp(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{signature: "migration_1", description: "First migration", shouldFail: true}
	if err := migrator.AddMigrations([]MigrationInterface{migration}); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err == nil {
		t.Fatal("Expected error from failing migration")
	}

	// Fix the migration and retry - a new row should be added, preserving the failed attempt in history
	migration.shouldFail = false
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Expected retry to succeed, got: %v", err)
	}

	var trackers []MigrationTracker
	if err := db.Schema().Orm().Query().Table(defaultTableName).OrderBy("started_at", "asc").Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 2 {
		t.Fatalf("Expected 2 tracker entries after retry (history preserved), got %d", len(trackers))
	}
	if trackers[0].Status != MigrationTrackerStatusFailed {
		t.Errorf("Expected first entry status 'failed', got '%s'", trackers[0].Status)
	}
	if trackers[1].Status != MigrationTrackerStatusCompleted {
		t.Errorf("Expected second entry status 'completed' after retry, got '%s'", trackers[1].Status)
	}
}

// TestStatus_ReportsFailedMigration verifies that Status reports the failed
// migration as "failed" with its error message, while the sibling migration
// reports "completed" or "pending".
func TestStatus_ReportsFailedMigration(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	if err := migrator.AddMigrations([]MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration", shouldFail: true},
	}); err != nil {
		t.Fatalf("AddMigrations failed: %v", err)
	}

	if err := migrator.Up(context.Background()); err == nil {
		t.Fatal("Expected error from failing migration")
	}

	statuses, err := migrator.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	byID := map[string]MigrationStatusResponse{}
	for _, st := range statuses {
		byID[st.ID] = st
	}
	if byID["migration_2"].State != MigrationTrackerStatusFailed {
		t.Errorf("Expected migration_2 state 'failed', got '%s'", byID["migration_2"].State)
	}
	if byID["migration_2"].Error == "" {
		t.Error("Expected Error to be populated for failed migration")
	}
	if s := byID["migration_1"].State; s != MigrationTrackerStatusCompleted && s != "pending" {
		t.Errorf("Unexpected state for migration_1: '%s'", s)
	}
}

// TestStatus_LegacyRowTreatedAsCompleted verifies that a tracker row with an
// empty status (data from before the status column existed) is treated as
// completed: Status reports it as such and Up does not re-run the migration.
func TestStatus_LegacyRowTreatedAsCompleted(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()
	if err := schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.String("migration", 255).Nullable()
		table.Integer("batch")
		table.String("description", 255)
		table.String("status", 16).Nullable()
		table.Text("error_message").Nullable()
		table.DateTime("started_at")
		table.DateTime("completed_at")
	}); err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}
	legacy := MigrationTracker{ID: "legacy_migration", Migration: "legacy_migration", Batch: 1, Status: ""}
	if err := db.Schema().Orm().Query().Table(defaultTableName).Create(&legacy); err != nil {
		t.Fatalf("failed to seed tracker: %v", err)
	}

	migrator := NewMigrator(db)
	statuses, err := migrator.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if len(statuses) != 1 || statuses[0].State != MigrationTrackerStatusCompleted {
		t.Fatalf("Expected legacy row to report 'completed', got %+v", statuses)
	}

	if err := migrator.AddMigration(&MockMigration{signature: "legacy_migration"}); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}
	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Up failed: %v", err)
	}
	var trackers []MigrationTracker
	if err := db.Schema().Orm().Query().Table(defaultTableName).Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 1 {
		t.Fatalf("Expected legacy migration to be skipped (1 row), got %d", len(trackers))
	}
}

// TestDown_FailureMarkedFailed verifies that a failing Down appends a
// "rollback_failed" row carrying the error, leaves the original Up row
// untouched in the same batch, and that the rollback can be retried
// successfully afterwards.
func TestDown_FailureMarkedFailed(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{signature: "migration_1", description: "First migration", downShouldFail: true}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	if err := migrator.Down(ctx); err == nil {
		t.Fatal("Expected error from failing Down migration")
	}

	trackers := trackerHistoryForTest(t, db)
	if len(trackers) != 2 {
		t.Fatalf("Expected 2 tracker entries (completed Up, failed rollback), got %d", len(trackers))
	}
	if trackers[0].Status != MigrationTrackerStatusCompleted || trackers[0].ErrorMessage != "" {
		t.Errorf("Expected Up record to stay 'completed' without error, got '%s' (%q)", trackers[0].Status, trackers[0].ErrorMessage)
	}
	if trackers[1].Status != MigrationTrackerStatusRollbackFailed {
		t.Errorf("Expected status 'rollback_failed' after failed rollback, got '%s'", trackers[1].Status)
	}
	if trackers[1].ErrorMessage == "" {
		t.Error("Expected error message to be recorded for failed rollback")
	}
	if trackers[1].Batch != trackers[0].Batch {
		t.Errorf("Expected rollback record in batch %d, got %d", trackers[0].Batch, trackers[1].Batch)
	}

	// The schema change is still applied, so the rollback can be retried.
	migration.downShouldFail = false
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("Down retry failed: %v", err)
	}
	trackers = trackerHistoryForTest(t, db)
	if len(trackers) != 3 || trackers[2].Status != MigrationTrackerStatusRolledBack {
		t.Fatalf("Expected a third 'rolled_back' record after retry, got %+v", trackers)
	}
}

// TestTrackerTimestampsStoredAsPlainDatetime verifies that started_at and
// completed_at are stored as plain "YYYY-MM-DD HH:MM:SS" text in SQLite,
// which keeps the column comparable and portable across drivers.
func TestTrackerTimestampsStoredAsPlainDatetime(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	if err := migrator.AddMigration(&MockMigration{signature: "2026_10_03_054245_test", description: "test"}); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}
	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get DB: %v", err)
	}
	var startedAt, completedAt string
	err = sqlDB.QueryRow("SELECT quote(started_at), quote(completed_at) FROM "+defaultTableName).Scan(&startedAt, &completedAt)
	if err != nil {
		t.Fatalf("failed to query tracker: %v", err)
	}
	startedAt = strings.Trim(startedAt, "'")
	completedAt = strings.Trim(completedAt, "'")

	for name, value := range map[string]string{"started_at": startedAt, "completed_at": completedAt} {
		if _, err := time.Parse("2006-01-02 15:04:05", value); err != nil {
			t.Errorf("Expected %s in 'YYYY-MM-DD HH:MM:SS' format, got %q", name, value)
		}
	}
}

// TestHistory_PreservedOnRetry verifies that a failed Up followed by a
// successful retry leaves two history rows (failed then completed) — earlier
// attempts are never overwritten.
func TestHistory_PreservedOnRetry(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{signature: "2026_06_15_120000_create_users", description: "Create users", shouldFail: true}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()

	// Run 1: Fails
	if err := migrator.Up(ctx); err == nil {
		t.Fatal("Expected Up to fail")
	}

	// Run 2: Succeeded after fixing migration
	migration.shouldFail = false
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Expected Up retry to succeed, got %v", err)
	}

	var trackers []MigrationTracker
	if err := db.Schema().Orm().Query().Table(defaultTableName).OrderBy("started_at", "asc").Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}

	if len(trackers) != 2 {
		t.Fatalf("Expected 2 history records in migration_tracker, got %d", len(trackers))
	}

	if trackers[0].MigrationName() != "2026_06_15_120000_create_users" || trackers[0].Status != MigrationTrackerStatusFailed {
		t.Errorf("Expected first history record to be failed, got status %s", trackers[0].Status)
	}

	if trackers[1].MigrationName() != "2026_06_15_120000_create_users" || trackers[1].Status != MigrationTrackerStatusCompleted {
		t.Errorf("Expected second history record to be completed, got status %s", trackers[1].Status)
	}
}

// TestHistory_PreservedOnRollbackAndReUp verifies that Up → Down → Up appends
// three rows (completed, rolled_back, completed), that the rollback row keeps
// the batch it reversed, and that the re-run lands in a new batch.
func TestHistory_PreservedOnRollbackAndReUp(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{signature: "2026_06_15_120000_create_posts", description: "Create posts"}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()

	// Step 1: Up
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	// Step 2: Down (Rollback)
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("Down failed: %v", err)
	}

	// Step 3: Up again
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up second time failed: %v", err)
	}

	trackers := trackerHistoryForTest(t, db)
	if len(trackers) != 3 {
		t.Fatalf("Expected 3 history records (completed, rolled_back, completed), got %d", len(trackers))
	}

	want := []string{MigrationTrackerStatusCompleted, MigrationTrackerStatusRolledBack, MigrationTrackerStatusCompleted}
	for i, status := range want {
		if trackers[i].Status != status {
			t.Errorf("Expected record %d status '%s', got '%s'", i, status, trackers[i].Status)
		}
	}
	if trackers[1].Batch != trackers[0].Batch {
		t.Errorf("Expected rollback record in batch %d, got %d", trackers[0].Batch, trackers[1].Batch)
	}
	if trackers[2].Batch <= trackers[0].Batch {
		t.Errorf("Expected re-run in a new batch after %d, got %d", trackers[0].Batch, trackers[2].Batch)
	}
	if trackers[0].CompletedAt.After(trackers[1].StartedAt) {
		t.Errorf("Expected first Up completion %v to be preserved (before rollback start %v)", trackers[0].CompletedAt, trackers[1].StartedAt)
	}
}

// trackerHistoryForTest returns all tracker rows ordered by started_at then
// ID — the same execution order the migrator uses.
func trackerHistoryForTest(t *testing.T, db *database.Database) []MigrationTracker {
	t.Helper()
	var trackers []MigrationTracker
	if err := db.Schema().Orm().Query().Table(defaultTableName).OrderBy("started_at", "asc").OrderBy("id", "asc").Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	return trackers
}

// TestHistory_SameSecondAttemptsOrderedByID verifies that when two attempts
// of the same migration share a started_at second, the later row ID wins —
// latest-wins logic must see the "completed" batch-2 row even though it was
// inserted before the older "rolled_back" row.
func TestHistory_SameSecondAttemptsOrderedByID(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	migration := &MockMigration{signature: "2026_06_15_120000_create_tags", description: "Create tags"}
	if err := migrator.AddMigration(migration); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}
	if err := migrator.(*Migrator).ensureMigrationTracker(db.Schema()); err != nil {
		t.Fatalf("ensureMigrationTracker failed: %v", err)
	}

	// Two attempts within the same second: the newer completed row is
	// inserted first so that insertion order disagrees with execution order.
	at := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	rows := []MigrationTracker{
		{ID: "t0000000002", Migration: migration.signature, Batch: 2, Status: MigrationTrackerStatusCompleted, StartedAt: at, CompletedAt: at},
		{ID: "t0000000001", Migration: migration.signature, Batch: 1, Status: MigrationTrackerStatusRolledBack, StartedAt: at, CompletedAt: at},
	}
	for _, row := range rows {
		if err := db.Schema().Orm().Query().Table(defaultTableName).Create(&row); err != nil {
			t.Fatalf("failed to seed tracker: %v", err)
		}
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Up failed: %v", err)
	}
	if migration.upCalled {
		t.Error("Expected migration to be treated as applied, but Up ran it again")
	}

	statuses, err := migrator.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if len(statuses) != 1 || statuses[0].State != MigrationTrackerStatusCompleted || statuses[0].Batch != 2 {
		t.Fatalf("Expected completed status from batch 2, got %+v", statuses)
	}
}

// TestStatus_RolledBackReportedAsUnassignedPending verifies that a rolled-back
// migration is reported exactly like a never-run one — "pending" with no
// batch or timestamps — for both registered and unregistered tracker rows.
func TestStatus_RolledBackReportedAsUnassignedPending(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	if err := migrator.AddMigration(&MockMigration{signature: "2026_06_15_120000_create_users", description: "Create users"}); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up failed: %v", err)
	}
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("Down failed: %v", err)
	}

	statuses, err := migrator.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	want := MigrationStatusResponse{ID: "2026_06_15_120000_create_users", Description: "Create users", State: "pending"}
	if len(statuses) != 1 || statuses[0] != want {
		t.Fatalf("Expected %+v, got %+v", want, statuses)
	}

	// Unregistered rolled-back rows follow the same policy.
	statuses, err = NewMigrator(db).Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if len(statuses) != 1 || statuses[0] != want {
		t.Fatalf("Expected unregistered %+v, got %+v", want, statuses)
	}
}

// TestFresh_PreservesHistory verifies that Fresh appends a "rolled_back" row
// for the dropped migration and a new "completed" row in a higher batch,
// keeping the earlier attempt in history.
func TestFresh_PreservesHistory(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	migrator := NewMigrator(db)
	if err := migrator.AddMigration(&MockMigration{signature: "2026_06_15_120000_create_users", description: "Create users"}); err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up failed: %v", err)
	}
	if err := migrator.Fresh(ctx); err != nil {
		t.Fatalf("Fresh failed: %v", err)
	}

	trackers := trackerHistoryForTest(t, db)
	want := []string{MigrationTrackerStatusCompleted, MigrationTrackerStatusRolledBack, MigrationTrackerStatusCompleted}
	if len(trackers) != len(want) {
		t.Fatalf("Expected %d history records after Fresh, got %+v", len(want), trackers)
	}
	for i, status := range want {
		if trackers[i].Status != status {
			t.Errorf("Expected record %d status '%s', got '%s'", i, status, trackers[i].Status)
		}
	}
	if trackers[2].Batch <= trackers[0].Batch {
		t.Errorf("Expected Fresh re-run in a new batch after %d, got %d", trackers[0].Batch, trackers[2].Batch)
	}
}
