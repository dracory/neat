package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// This example demonstrates the schema Migration interface approach
// This is an alternative design pattern that uses interface-based migrations
// instead of the current function-based registration approach
func main() {
	if err := RunExample("sqlite://./example_schema_migrations.db"); err != nil {
		log.Fatalf("Schema migration example failed: %v", err)
	}
}

// RunExample demonstrates using the schema Migration interface
func RunExample(dsn string) error {
	fmt.Println("=== Schema Migration Interface Approach ===")
	fmt.Println("This approach uses the schema Migration interface with Signature(), Up(), and Down() methods")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create migrations implementing the schema Migration interface
	// Note: We need to pass the schema to each migration since the interface
	// doesn't provide a way to access it
	schema := db.Schema()
	migrations := []contractsschema.Migration{
		NewCreateUsersTable(schema),
		NewCreatePostsTable(schema),
		NewAddPostsIndexes(schema),
	}

	// Register migrations with the schema
	db.Schema().Register(migrations)

	// Run migrations by calling Up() on each
	fmt.Println("=== Running Migrations ===")
	for _, migration := range db.Schema().Migrations() {
		fmt.Printf("Running migration: %s\n", migration.Signature())
		if err := migration.Up(); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Signature(), err)
		}
		fmt.Printf("Migration %s completed successfully\n", migration.Signature())
	}

	// Verify tables were created
	fmt.Println("\n=== Verification ===")
	tables := db.Schema().GetTableListing()
	fmt.Printf("Tables created: %v\n", tables)

	// Demonstrate rollback
	fmt.Println("\n=== Rolling Back Migrations ===")
	// Rollback in reverse order
	migrationsList := db.Schema().Migrations()
	for i := len(migrationsList) - 1; i >= 0; i-- {
		migration := migrationsList[i]
		fmt.Printf("Rolling back migration: %s\n", migration.Signature())
		if err := migration.Down(); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Signature(), err)
		}
		fmt.Printf("Migration %s rolled back successfully\n", migration.Signature())
	}

	// Verify tables were dropped
	fmt.Println("\n=== Verification After Rollback ===")
	tables = db.Schema().GetTableListing()
	fmt.Printf("Tables remaining: %v\n", tables)

	return nil
}

// CreateUsersTable creates the users table
type CreateUsersTable struct {
	schema contractsschema.Schema
}

func NewCreateUsersTable(schema contractsschema.Schema) *CreateUsersTable {
	return &CreateUsersTable{schema: schema}
}

func (m *CreateUsersTable) Signature() string {
	return "create_users_table"
}

func (m *CreateUsersTable) Up() error {
	fmt.Println("  Creating users table...")
	return m.schema.Create("users", func(blueprint contractsschema.Blueprint) {
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
	fmt.Println("  Dropping users table...")
	return m.schema.DropIfExists("users")
}

// CreatePostsTable creates the posts table
type CreatePostsTable struct {
	schema contractsschema.Schema
}

func NewCreatePostsTable(schema contractsschema.Schema) *CreatePostsTable {
	return &CreatePostsTable{schema: schema}
}

func (m *CreatePostsTable) Signature() string {
	return "create_posts_table"
}

func (m *CreatePostsTable) Up() error {
	fmt.Println("  Creating posts table...")
	return m.schema.Create("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.ID()
		blueprint.Integer("user_id")
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.String("status")
		blueprint.Timestamps()
	})
}

func (m *CreatePostsTable) Down() error {
	fmt.Println("  Dropping posts table...")
	return m.schema.DropIfExists("posts")
}

// AddPostsIndexes adds indexes to the posts table
type AddPostsIndexes struct {
	schema contractsschema.Schema
}

func NewAddPostsIndexes(schema contractsschema.Schema) *AddPostsIndexes {
	return &AddPostsIndexes{schema: schema}
}

func (m *AddPostsIndexes) Signature() string {
	return "add_posts_indexes"
}

func (m *AddPostsIndexes) Up() error {
	fmt.Println("  Adding indexes to posts table...")
	return m.schema.Table("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.Index("user_id")
		blueprint.Index("status")
	})
}

func (m *AddPostsIndexes) Down() error {
	fmt.Println("  Removing indexes from posts table...")
	// Note: Dropping indexes is database-specific
	// This is a simplified example
	return nil
}
