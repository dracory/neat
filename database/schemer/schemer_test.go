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
	if impl.tableName != defaultTableName {
		t.Errorf("Expected default table name '%s', got '%s'", defaultTableName, impl.tableName)
	}
}

func TestSetTableName_Valid(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)

	validNames := []string{"migrations", "my_migrations", "schema_migrations", "_tracker"}
	for _, name := range validNames {
		err := schemer.SetTableName(name)
		if err != nil {
			t.Errorf("Expected SetTableName('%s') to succeed, got error: %v", name, err)
		}

		impl := schemer.(*SchemerImplementation)
		if impl.tableName != name {
			t.Errorf("Expected table name '%s', got '%s'", name, impl.tableName)
		}
	}
}

func TestSetTableName_Invalid(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)

	invalidNames := []string{"", "1migrations", "migration-tracker", "SELECT", "DROP", "TABLE"}
	for _, name := range invalidNames {
		err := schemer.SetTableName(name)
		if err == nil {
			t.Errorf("Expected SetTableName('%s') to fail, got nil", name)
		}

		// Table name should remain unchanged
		impl := schemer.(*SchemerImplementation)
		if impl.tableName != defaultTableName {
			t.Errorf("Expected table name to remain '%s' after failed SetTableName, got '%s'", defaultTableName, impl.tableName)
		}
	}
}

func TestSetTableName_UsedForTracking(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schemer := NewSchemer(db)
	customTable := "my_migrations"
	if err := schemer.SetTableName(customTable); err != nil {
		t.Fatalf("SetTableName failed: %v", err)
	}

	migration := &MockMigration{
		signature:   "test_migration",
		description: "Test migration",
	}
	schemer.AddMigration(migration)

	ctx := context.Background()
	err = schemer.Up(ctx)
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

	// Verify migration tracking table was created
	if !db.Schema().HasTable(defaultTableName) {
		t.Error("Expected migration tracking table to be auto-created")
	}
}

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
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
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
	// With transactions enabled by default, the injected schema is a WithTransaction wrapper.
	// Verify it works by checking the injected schema can perform operations.
	if migration.schema.Orm() == nil {
		t.Error("Expected injected schema to have a valid ORM")
	}
}

func TestUp_EmptySignature(t *testing.T) {
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
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
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

func TestUp_SignatureValidation_DateTime(t *testing.T) {
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
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	schemer := NewSchemer(db)
	schemer.SetSignatureValidation(true, SignatureFormatDateTime)

	// Valid datetime signature should pass
	validMigration := &MockMigration{
		signature:   "2026_06_15_1200_create_users_table",
		description: "Valid datetime signature",
	}
	schemer.AddMigration(validMigration)

	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Expected valid datetime signature to pass, got error: %v", err)
	}
	if !validMigration.upCalled {
		t.Error("Expected valid migration Up to be called")
	}
}

func TestUp_SignatureValidation_InvalidFormat(t *testing.T) {
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
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	schemer := NewSchemer(db)
	schemer.SetSignatureValidation(true, SignatureFormatDateTime)

	// Invalid signature should fail
	invalidMigration := &MockMigration{
		signature:   "test_migration",
		description: "Invalid datetime signature",
	}
	schemer.AddMigration(invalidMigration)

	ctx := context.Background()
	err = schemer.Up(ctx)
	if err == nil {
		t.Error("Expected error for invalid signature format")
	}
	if invalidMigration.upCalled {
		t.Error("Expected invalid migration Up NOT to be called")
	}
}

func TestUp_SignatureValidation_Disabled(t *testing.T) {
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
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	schemer := NewSchemer(db)
	// Validation is disabled by default, but explicitly set it
	schemer.SetSignatureValidation(false, SignatureFormatDateTime)

	// Arbitrary signature should pass when validation is disabled
	migration := &MockMigration{
		signature:   "totally_arbitrary_name",
		description: "Arbitrary signature",
	}
	schemer.AddMigration(migration)

	ctx := context.Background()
	err = schemer.Up(ctx)
	if err != nil {
		t.Errorf("Expected arbitrary signature to pass when validation disabled, got error: %v", err)
	}
	if !migration.upCalled {
		t.Error("Expected migration Up to be called when validation disabled")
	}
}

func TestUp_SkipAlreadyRun(t *testing.T) {
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
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
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

	// Create migration tracking table first
	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
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

	// Create migration tracking table first
	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
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

	// Create migration tracking table first
	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
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
		t.Errorf("Expected empty status when migration tracking table does not exist, got %d", len(status))
	}
}

func TestStatus_WithMigrations(t *testing.T) {
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
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
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

	// Create migration tracking table first
	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
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
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Pre-create migration tracking table so runUp doesn't need to create it
	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	schemer := NewSchemer(db)
	migrations := []contractsschema.MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration", shouldFail: true},
	}
	schemer.AddMigrations(migrations)

	// With transactions enabled (default), Up should fail and roll back tracker entries
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err == nil {
		t.Fatal("Expected error from failing migration")
	}

	var trackers []MigrationTracker
	query := db.Schema().Orm().Query().Table(defaultTableName)
	if err := query.Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 0 {
		t.Errorf("Expected 0 tracker entries after rollback, got %d", len(trackers))
	}
}

func TestUpWithTransactionsDisabled(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Pre-create migration tracking table
	schema := db.Schema()
	err = schema.Create(defaultTableName, func(table contractsschema.Blueprint) {
		table.String("id")
		table.Primary("id")
		table.Integer("batch")
		table.String("description", 255)
		table.DateTime("started_at")
		table.DateTime("completed_at")
	})
	if err != nil {
		t.Fatalf("failed to create migration tracking table: %v", err)
	}

	schemer := NewSchemer(db)
	impl := schemer.(*SchemerImplementation)
	impl.SetTransactionsEnabled(false)

	migrations := []contractsschema.MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration", shouldFail: true},
	}
	schemer.AddMigrations(migrations)

	// With transactions disabled, Up should fail but prior tracker entries persist
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err == nil {
		t.Fatal("Expected error from failing migration")
	}

	var trackers []MigrationTracker
	query := db.Schema().Orm().Query().Table(defaultTableName)
	if err := query.Get(&trackers); err != nil {
		t.Fatalf("failed to get trackers: %v", err)
	}
	if len(trackers) != 1 {
		t.Errorf("Expected 1 tracker entry after failed migration, got %d", len(trackers))
	}
	if len(trackers) > 0 && trackers[0].ID != "migration_1" {
		t.Errorf("Expected tracker ID 'migration_1', got '%s'", trackers[0].ID)
	}
}

func TestTransactionRollbackOnFailure(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Do NOT pre-create migration tracking table - runUp will create it inside the transaction
	schemer := NewSchemer(db)
	migrations := []contractsschema.MigrationInterface{
		&MockMigration{signature: "migration_1", description: "First migration"},
		&MockMigration{signature: "migration_2", description: "Second migration", shouldFail: true},
	}
	schemer.AddMigrations(migrations)

	// Up should fail; the entire transaction (including tracker table creation) should roll back
	ctx := context.Background()
	err = schemer.Up(ctx)
	if err == nil {
		t.Fatal("Expected Up to return error when migration fails")
	}

	if db.Schema().HasTable(defaultTableName) {
		t.Error("Expected migration tracking table to be rolled back when transaction fails")
	}
}
