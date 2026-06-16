package main

import (
	"context"
	"fmt"
	"log"

	_ "modernc.org/sqlite"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema"
	"github.com/dracory/neat/database/migrator"
)

// This example demonstrates transaction control in the schemer package
// for safe migration execution with automatic rollback on failure
func main() {
	if err := RunTransactionExample("sqlite://./example_transactions.db"); err != nil {
		log.Fatalf("Transaction example failed: %v", err)
	}
}

// RunTransactionExample demonstrates transaction control in migrations
func RunTransactionExample(dsn string) error {
	fmt.Println("=== Schemer Transaction Control Example ===")
	fmt.Println("This demonstrates transaction settings for migration safety")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create schemer instance
	schemerInstance := migrator.NewMigrator(db)

	// Configure transaction settings
	fmt.Println("=== Configuring Transaction Settings ===")
	schemerInstance.SetTransactionsEnabled(true)
	fmt.Println("Transactions enabled: true")

	schemerInstance.SetTransactionIsolationLevel("SERIALIZABLE")
	fmt.Println("Transaction isolation level: SERIALIZABLE")

	// Add migrations
	if err := schemerInstance.AddMigration(&CreateMigrationTrackerTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}
	if err := schemerInstance.AddMigration(&CreateUsersTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}
	if err := schemerInstance.AddMigration(&CreatePostsTable{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}

	// Run migrations with transaction wrapping
	fmt.Println("\n=== Running Migrations with Transaction Wrapping ===")
	ctx := context.Background()
	if err := schemerInstance.Up(ctx); err != nil {
		return fmt.Errorf("migration up failed: %w", err)
	}

	fmt.Println("\n=== Migrations Completed Successfully ===")

	// Check migration status
	fmt.Println("\n=== Migration Status ===")
	status, err := schemerInstance.Status()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	for _, s := range status {
		fmt.Printf("Migration: %s - State: %s\n", s.ID, s.State)
	}

	// Demonstrate disabling transactions for large migrations
	fmt.Println("\n=== Disabling Transactions for Large Migrations ===")
	schemerInstance.SetTransactionsEnabled(false)
	fmt.Println("Transactions enabled: false")

	// Add a large migration
	if err := schemerInstance.AddMigration(&AddPostsIndexes{}); err != nil {
		return fmt.Errorf("failed to add migration: %w", err)
	}

	// Run without transaction wrapping
	if err := schemerInstance.Up(ctx); err != nil {
		return fmt.Errorf("migration up failed: %w", err)
	}

	fmt.Println("\n=== Large Migration Completed ===")

	return nil
}

// CreateMigrationTrackerTable creates the migration tracking table
type CreateMigrationTrackerTable struct {
	schema.BaseMigration
}

func (m *CreateMigrationTrackerTable) Signature() string {
	return "2024_06_15_000000_create_migration_tracker_table"
}

func (m *CreateMigrationTrackerTable) Description() string {
	return "Creates the migration_tracker table for tracking migration execution"
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
	schema.BaseMigration
}

func (m *CreateUsersTable) Signature() string {
	return "2024_06_15_120000_create_users_table"
}

func (m *CreateUsersTable) Description() string {
	return "Creates users table with authentication fields"
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
		blueprint.String("password")
		blueprint.String("status")
		blueprint.Timestamps()
	})
}

func (m *CreateUsersTable) Down() error {
	return m.GetSchema().DropIfExists("users")
}

// CreatePostsTable creates the posts table
type CreatePostsTable struct {
	schema.BaseMigration
}

func (m *CreatePostsTable) Signature() string {
	return "2024_06_15_120100_create_posts_table"
}

func (m *CreatePostsTable) Description() string {
	return "Creates posts table with user relationships"
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
		blueprint.String("status")
		blueprint.Timestamps()
	})
}

func (m *CreatePostsTable) Down() error {
	return m.GetSchema().DropIfExists("posts")
}

// AddPostsIndexes adds indexes to the posts table
type AddPostsIndexes struct {
	schema.BaseMigration
}

func (m *AddPostsIndexes) Signature() string {
	return "2024_06_15_120200_add_posts_indexes"
}

func (m *AddPostsIndexes) Description() string {
	return "Adds performance indexes to posts table"
}

func (m *AddPostsIndexes) Up() error {
	return m.GetSchema().Table("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.Index("user_id")
		blueprint.Index("status")
	})
}

func (m *AddPostsIndexes) Down() error {
	// Note: Dropping indexes is database-specific
	return nil
}
