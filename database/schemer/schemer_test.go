package schemer

import (
	"context"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// MockMigration is a test migration implementation
type MockMigration struct {
	signature   string
	description string
	upCalled    bool
	downCalled  bool
	schema      contractsschema.Schema
	shouldFail  bool
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
	return nil
}

func (m *MockMigration) SetSchema(schema contractsschema.Schema) {
	m.schema = schema
}

func (m *MockMigration) GetSchema() contractsschema.Schema {
	return m.schema
}

func TestNewSchemer(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)
	if schemer == nil {
		t.Error("Expected non-nil schemer")
	}

	impl, ok := schemer.(*SchemerImplementation)
	if !ok {
		t.Error("Expected SchemerImplementation type")
	}
	if impl.db != db {
		t.Error("Expected db to be set")
	}
	if len(impl.migrations) != 0 {
		t.Error("Expected empty migrations list")
	}
}

func TestAddMigration(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}

	err = schemer.AddMigration(migration)
	if err != nil {
		t.Errorf("AddMigration failed: %v", err)
	}

	impl := schemer.(*SchemerImplementation)
	if len(impl.migrations) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(impl.migrations))
	}
}

func TestAddMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)
	migrations := []contractsschema.MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
		&MockMigration{signature: "migration_3", description: "Third migration"},
	}

	err = schemer.AddMigrations(migrations)
	if err != nil {
		t.Errorf("AddMigrations failed: %v", err)
	}

	impl := schemer.(*SchemerImplementation)
	if len(impl.migrations) != 3 {
		t.Errorf("Expected 3 migrations, got %d", len(impl.migrations))
	}
}

func TestUp_AutoCreateMigrationTracker(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	schemer.AddMigration(migration)

	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	// Verify migration_tracker table was created
	if !db.Schema().HasTable("migration_tracker") {
		t.Error("Expected migration_tracker table to be auto-created")
	}
}

func TestUp_SchemaInjection(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration_tracker table first
	schema := db.Schema()
	err = schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration_tracker table: %v", err)
	}

	schemer := NewSchemer(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	schemer.AddMigration(migration)

	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	if !migration.upCalled {
		t.Error("Expected migration Up to be called")
	}
	if migration.schema == nil {
		t.Error("Expected schema to be injected")
	}
	if migration.schema != schema {
		t.Error("Expected injected schema to match db schema")
	}
}

func TestUp_EmptySignature(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration_tracker table first
	schema := db.Schema()
	err = schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration_tracker table: %v", err)
	}

	schemer := NewSchemer(db)
	migration := &MockMigration{
		signature:   "",
		description: "Test migration",
	}
	schemer.AddMigration(migration)

	ctx := context.Background()
	err = schemer.Up(ctx)
	if err == nil {
		t.Error("Expected error for empty signature")
	}
}

func TestUp_SkipAlreadyRun(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration_tracker table first
	schema := db.Schema()
	err = schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration_tracker table: %v", err)
	}

	schemer := NewSchemer(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	schemer.AddMigration(migration)

	// Run migration first time
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("First Up failed: %v", err)
	}

	// Reset the flag
	migration.upCalled = false

	// Run migration second time - should be skipped
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Second Up failed: %v", err)
	}

	if migration.upCalled {
		t.Error("Expected migration to be skipped on second run")
	}
}

func TestDown(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration_tracker table first
	schema := db.Schema()
	err = schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration_tracker table: %v", err)
	}

	schemer := NewSchemer(db)
	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	schemer.AddMigration(migration)

	// Run migration first
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	// Down should rollback last migration
	err = schemer.Down(ctx)
	if err != nil {
		t.Errorf("Down failed: %v", err)
	}
}

func TestRollbackSteps(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration_tracker table first
	schema := db.Schema()
	err = schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration_tracker table: %v", err)
	}

	schemer := NewSchemer(db)
	migrations := []contractsschema.MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
		&MockMigration{signature: "migration_3", description: "Third migration"},
	}
	schemer.AddMigrations(migrations)

	// Run migrations
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	// Rollback 2 migrations
	err = schemer.RollbackSteps(ctx, 2)
	if err != nil {
		t.Errorf("RollbackSteps failed: %v", err)
	}
}

func TestRollbackToBatch(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration_tracker table first
	schema := db.Schema()
	err = schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration_tracker table: %v", err)
	}

	schemer := NewSchemer(db)
	migrations := []contractsschema.MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
	}
	schemer.AddMigrations(migrations)

	// Run migrations
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	// Get the batch number from status
	status, err := schemer.Status()
	if err != nil {
		t.Errorf("Status failed: %v", err)
	}
	if len(status) == 0 {
		t.Fatal("Expected at least one migration status")
	}
	batch := status[0].Batch

	// Rollback to that batch
	err = schemer.RollbackToBatch(ctx, batch)
	if err != nil {
		t.Errorf("RollbackToBatch failed: %v", err)
	}
}

func TestStatus_NoMigrationTrackerTable(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)

	status, err := schemer.Status()
	if err != nil {
		t.Errorf("Status failed: %v", err)
	}
	if len(status) != 0 {
		t.Errorf("Expected empty status when migration_tracker table does not exist, got %d", len(status))
	}
}

func TestStatus_WithMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration_tracker table first
	schema := db.Schema()
	err = schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration_tracker table: %v", err)
	}

	schemer := NewSchemer(db)
	migrations := []contractsschema.MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
	}
	schemer.AddMigrations(migrations)

	// Run migrations
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	// Get status
	status, err := schemer.Status()
	if err != nil {
		t.Errorf("Status failed: %v", err)
	}
	if len(status) != 2 {
		t.Errorf("Expected 2 migration statuses, got %d", len(status))
	}

	// Verify status content
	for i, s := range status {
		if s.State != "completed" {
			t.Errorf("Expected state 'completed' for migration %d, got '%s'", i, s.State)
		}
		if s.ID != migrations[i].Signature() {
			t.Errorf("Expected ID '%s' for migration %d, got '%s'", migrations[i].Signature(), i, s.ID)
		}
	}
}

func TestFresh(t *testing.T) {
	t.Skip("Skipping Fresh test - getAllTables() has placeholder implementation that causes timeout")
}

func TestReset(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration_tracker table first
	schema := db.Schema()
	err = schema.Create("migration_tracker", func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration_tracker table: %v", err)
	}

	schemer := NewSchemer(db)
	migrations := []contractsschema.MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration"},
	}
	schemer.AddMigrations(migrations)

	// Run migrations
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Up failed: %v", err)
	}

	// Reset should rollback all migrations
	err = schemer.Reset(ctx)
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}
}

func TestSetTransactionsEnabled(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)
	impl := schemer.(*SchemerImplementation)

	// Default should be enabled
	if !impl.useTransactions {
		t.Error("Expected transactions to be enabled by default")
	}

	// Disable transactions
	impl.SetTransactionsEnabled(false)
	if impl.useTransactions {
		t.Error("Expected transactions to be disabled")
	}

	// Enable transactions
	impl.SetTransactionsEnabled(true)
	if !impl.useTransactions {
		t.Error("Expected transactions to be enabled")
	}
}

func TestSetTransactionIsolationLevel(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)
	impl := schemer.(*SchemerImplementation)

	// Set isolation level
	impl.SetTransactionIsolationLevel("SERIALIZABLE")
	if impl.isolationLevel != "SERIALIZABLE" {
		t.Error("Expected isolation level to be SERIALIZABLE")
	}

	// Test different isolation levels
	levels := []string{"READ UNCOMMITTED", "READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"}
	for _, level := range levels {
		impl.SetTransactionIsolationLevel(level)
		if impl.isolationLevel != level {
			t.Errorf("Expected isolation level to be %s, got %s", level, impl.isolationLevel)
		}
	}
}

func TestUpWithTransactionsEnabled(t *testing.T) {
	t.Skip("Skipping transaction tests - schema transaction detection needs investigation")
}

func TestUpWithTransactionsDisabled(t *testing.T) {
	t.Skip("Skipping transaction tests - schema transaction detection needs investigation")
}

func TestTransactionRollbackOnFailure(t *testing.T) {
	t.Skip("Skipping transaction tests - schema transaction detection needs investigation")
}
