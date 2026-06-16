package main

import (
	"context"
	"fmt"
	"log"

	_ "modernc.org/sqlite"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
		"github.com/dracory/neat/database/migrator"
)

// This example demonstrates transaction failure behavior
// showing how failed migrations automatically roll back when transactions are enabled
func main() {
	if err := RunTransactionFailureExample("sqlite://./example_transaction_failure.db"); err != nil {
		log.Fatalf("Transaction failure example failed: %v", err)
	}
}

// RunTransactionFailureExample demonstrates transaction failure and automatic rollback
func RunTransactionFailureExample(dsn string) error {
	fmt.Println("=== Transaction Failure and Automatic Rollback Example ===")
	fmt.Println("This demonstrates how failed migrations automatically roll back")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Example 1: Migration failure without transactions
	fmt.Println("=== Example 1: Migration Failure WITHOUT Transactions ===")
	migratorNoTx := migrator.NewMigrator(db)
	migratorNoTx.SetTransactionsEnabled(false)
	fmt.Println("Transactions enabled: false")

	if err := migratorNoTx.AddMigration(&CreateMigrationTrackerTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}
	if err := migratorNoTx.AddMigration(&CreateUsersTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}
	if err := migratorNoTx.AddMigration(&FailingMigration{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}

	ctx := context.Background()
	err = migratorNoTx.Up(ctx)
	if err != nil {
		fmt.Printf("Migration failed (expected): %v\n", err)
	}

	// Check what tables exist - users table might still exist
	fmt.Println("\n=== Checking Database State After Failure ===")
	dbSchema := db.Schema()
	if dbSchema.HasTable("users") {
		fmt.Println("⚠️  'users' table still exists (partial migration state)")
	} else {
		fmt.Println("✓ 'users' table does not exist")
	}

	// Clean up for next example
	_ = dbSchema.DropIfExists("users")
	_ = dbSchema.DropIfExists("posts")
	_ = dbSchema.DropIfExists("migration_tracker")

	// Example 2: Migration failure with transactions
	fmt.Println("\n=== Example 2: Migration Failure WITH Transactions ===")
	migratorWithTx := migrator.NewMigrator(db)
	migratorWithTx.SetTransactionsEnabled(true)
	fmt.Println("Transactions enabled: true")

	if err := migratorWithTx.AddMigration(&CreateMigrationTrackerTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}
	if err := migratorWithTx.AddMigration(&CreateUsersTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}
	if err := migratorWithTx.AddMigration(&FailingMigration{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}

	err = migratorWithTx.Up(ctx)
	if err != nil {
		fmt.Printf("Migration failed (expected): %v\n", err)
	}

	// Check what tables exist - should be clean due to rollback
	fmt.Println("\n=== Checking Database State After Failure ===")
	if dbSchema.HasTable("users") {
		fmt.Println("⚠️  'users' table still exists (unexpected - rollback failed)")
	} else {
		fmt.Println("✓ 'users' table does not exist (transaction rolled back successfully)")
	}

	if dbSchema.HasTable("migration_tracker") {
		fmt.Println("✓ 'migration_tracker' table exists (created before failure)")
	} else {
		fmt.Println("⚠️  'migration_tracker' table does not exist")
	}

	// Check migration status
	fmt.Println("\n=== Migration Status ===")
	status, err := migratorWithTx.Status()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	if len(status) == 0 {
		fmt.Println("✓ No migrations recorded (transaction rolled back)")
	} else {
		fmt.Printf("⚠️  %d migrations recorded (unexpected)\n", len(status))
		for _, s := range status {
			fmt.Printf("  - %s: %s\n", s.ID, s.State)
		}
	}

	// Example 3: Successful migration with transactions
	fmt.Println("\n=== Example 3: Successful Migration WITH Transactions ===")
	migratorSuccess := migrator.NewMigrator(db)
	migratorSuccess.SetTransactionsEnabled(true)
	fmt.Println("Transactions enabled: true")

	if err := migratorSuccess.AddMigration(&CreateUsersTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}
	if err := migratorSuccess.AddMigration(&CreatePostsTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}

	err = migratorSuccess.Up(ctx)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("✓ All migrations completed successfully")

	// Verify tables exist
	fmt.Println("\n=== Verifying Database State ===")
	if dbSchema.HasTable("users") {
		fmt.Println("✓ 'users' table exists")
	}
	if dbSchema.HasTable("posts") {
		fmt.Println("✓ 'posts' table exists")
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("✓ Transactions prevent partial migration states")
	fmt.Println("✓ Failed migrations automatically roll back")
	fmt.Println("✓ Database remains in consistent state")
	fmt.Println("✓ Migration tracker reflects actual state")

	return nil
}

// CreateMigrationTrackerTable creates the migration tracking table
type CreateMigrationTrackerTable struct {
	migrator.BaseMigration
}

func (m *CreateMigrationTrackerTable) Signature() string {
	return "2024_06_15_000000_create_migration_tracker_table"
}

func (m *CreateMigrationTrackerTable) Description() string {
	return "Creates the migration_tracker table"
}

func (m *CreateMigrationTrackerTable) Up() error {
	if m.GetSchema().HasTable("migration_tracker") {
		return nil
	}

	return m.GetSchema().Create("migration_tracker", func(blueprint contractsschema.Blueprint) {
		blueprint.String("id")
		blueprint.Primary("id")
		blueprint.Integer("batch")
		blueprint.Text("description")
		blueprint.Timestamp("started_at")
		blueprint.Timestamp("completed_at")
	})
}

func (m *CreateMigrationTrackerTable) Down() error {
	return m.GetSchema().DropIfExists("migration_tracker")
}

// CreateUsersTable creates the users table
type CreateUsersTable struct {
	migrator.BaseMigration
}

func (m *CreateUsersTable) Signature() string {
	return "2024_06_15_120000_create_users_table"
}

func (m *CreateUsersTable) Description() string {
	return "Creates users table"
}

func (m *CreateUsersTable) Up() error {
	if m.GetSchema().HasTable("users") {
		return nil
	}

	return m.GetSchema().Create("users", func(blueprint contractsschema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Unique("email")
		blueprint.Timestamps()
	})
}

func (m *CreateUsersTable) Down() error {
	return m.GetSchema().DropIfExists("users")
}

// CreatePostsTable creates the posts table
type CreatePostsTable struct {
	migrator.BaseMigration
}

func (m *CreatePostsTable) Signature() string {
	return "2024_06_15_120100_create_posts_table"
}

func (m *CreatePostsTable) Description() string {
	return "Creates posts table"
}

func (m *CreatePostsTable) Up() error {
	if m.GetSchema().HasTable("posts") {
		return nil
	}

	return m.GetSchema().Create("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.ID()
		blueprint.Integer("user_id")
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.Timestamps()
	})
}

func (m *CreatePostsTable) Down() error {
	return m.GetSchema().DropIfExists("posts")
}

// FailingMigration is a migration that intentionally fails
type FailingMigration struct {
	migrator.BaseMigration
}

func (m *FailingMigration) Signature() string {
	return "2024_06_15_120200_failing_migration"
}

func (m *FailingMigration) Description() string {
	return "This migration intentionally fails to demonstrate rollback"
}

func (m *FailingMigration) Up() error {
	return fmt.Errorf("intentional migration failure for demonstration")
}

func (m *FailingMigration) Down() error {
	return nil
}
