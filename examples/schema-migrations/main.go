package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema"
)

// This example demonstrates the interface-based migration system
// which provides a cleaner, more structured approach to schema migrations
func main() {
	if err := RunInterfaceBasedMigrations("sqlite://./example_schema_migrations.db"); err != nil {
		log.Fatalf("Interface-based migration example failed: %v", err)
	}
}

// RunInterfaceBasedMigrations demonstrates the interface-based migration system
func RunInterfaceBasedMigrations(dsn string) error {
	fmt.Println("=== Interface-Based Migration System ===")
	fmt.Println("This approach uses structured migration objects with clean interfaces")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration instances
	migrations := []contractsschema.MigrationInterface{
		&CreateMigrationTrackerTable{},
		&CreateUsersTable{},
		&CreatePostsTable{},
		&CreateCommentsTable{},
		&AddPostsIndexes{},
		&AddPublishedToPosts{},
	}

	// Register migrations with schema (automatic schema injection via SchemaSetter)
	schema := db.Schema()
	schema.Register(migrations)

	// Run all migrations
	fmt.Println("=== Running Migrations ===")
	for _, migration := range migrations {
		fmt.Printf("Running migration: %s\n", migration.Signature())
		fmt.Printf("Description: %s\n", migration.Description())
		if err := migration.Up(); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Signature(), err)
		}
		fmt.Printf("Migration %s completed successfully\n", migration.Signature())
	}

	fmt.Println("\n=== All Migrations Completed ===")

	// Demonstrate rollback (last migration only)
	fmt.Println("\n=== Rolling Back Last Migration ===")
	lastMigration := migrations[len(migrations)-1]
	fmt.Printf("Rolling back migration: %s\n", lastMigration.Signature())
	if err := lastMigration.Down(); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	fmt.Printf("Migration %s rolled back successfully\n", lastMigration.Signature())

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
	return "Creates users table with authentication fields and soft deletes"
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
		blueprint.SoftDeletes()
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
	return "Creates posts table with user relationships and status tracking"
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

		// Note: Foreign key constraints skipped for SQLite compatibility
		// blueprint.Foreign("user_id")
	})
}

func (m *CreatePostsTable) Down() error {
	return m.GetSchema().DropIfExists("posts")
}

// CreateCommentsTable creates the comments table
type CreateCommentsTable struct {
	schema.BaseMigration
}

func (m *CreateCommentsTable) Signature() string {
	return "2024_06_15_120200_create_comments_table"
}

func (m *CreateCommentsTable) Description() string {
	return "Creates comments table for post discussions with user relationships"
}

func (m *CreateCommentsTable) Up() error {
	if m.GetSchema().HasTable("comments") {
		return nil
	}

	return m.GetSchema().Create("comments", func(blueprint contractsschema.Blueprint) {
		blueprint.ID()
		blueprint.Integer("post_id")
		blueprint.Integer("user_id")
		blueprint.Text("comment")
		blueprint.Timestamps()

		// Note: Foreign key constraints skipped for SQLite compatibility
		// blueprint.Foreign("post_id")
		// blueprint.Foreign("user_id")
	})
}

func (m *CreateCommentsTable) Down() error {
	return m.GetSchema().DropIfExists("comments")
}

// AddPostsIndexes adds indexes to the posts table
type AddPostsIndexes struct {
	schema.BaseMigration
}

func (m *AddPostsIndexes) Signature() string {
	return "2024_06_15_120300_add_posts_indexes"
}

func (m *AddPostsIndexes) Description() string {
	return "Adds performance indexes to posts table for user_id and status columns"
}

func (m *AddPostsIndexes) Up() error {
	return m.GetSchema().Table("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.Index("user_id")
		blueprint.Index("status")
	})
}

func (m *AddPostsIndexes) Down() error {
	// Note: Dropping indexes is database-specific
	// This is a simplified example
	return nil
}

// AddPublishedToPosts adds published_at column to posts table
type AddPublishedToPosts struct {
	schema.BaseMigration
}

func (m *AddPublishedToPosts) Signature() string {
	return "2024_06_15_120400_add_published_to_posts"
}

func (m *AddPublishedToPosts) Description() string {
	return "Adds published_at timestamp column to posts table for scheduling"
}

func (m *AddPublishedToPosts) Up() error {
	return m.GetSchema().Table("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.Timestamp("published_at")
	})
}

func (m *AddPublishedToPosts) Down() error {
	return m.GetSchema().Table("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.DropColumn("published_at")
	})
}
