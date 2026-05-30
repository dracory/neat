package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/migration"
)

// This example demonstrates both the schema builder approach and the migration system
func main() {
	if err := RunSchemaBuilderExample("sqlite://./example_schema.db"); err != nil {
		log.Fatalf("Schema builder example failed: %v", err)
	}

	if err := RunMigrationSystemExample("sqlite://./example_migration.db"); err != nil {
		log.Fatalf("Migration system example failed: %v", err)
	}
}

// RunSchemaBuilderExample demonstrates using the schema builder directly
// This is useful for simple scripts or when you don't need version control
func RunSchemaBuilderExample(dsn string) error {
	fmt.Println("=== Schema Builder Approach ===")
	fmt.Println("This approach uses the schema builder directly without version control")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Migration 1: Create users table
	fmt.Println("=== Migration: Create Users Table ===")
	err = db.Schema().Create("users", func(blueprint contractsschema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Unique("email")
		blueprint.String("password")
		blueprint.String("status")
		blueprint.Timestamps()
		blueprint.SoftDeletes()
	})
	if err != nil {
		return fmt.Errorf("error in migration 1: %w", err)
	}
	fmt.Println("Users table created")

	// Migration 2: Create posts table with foreign key
	fmt.Println("\n=== Migration: Create Posts Table ===")
	err = db.Schema().Create("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.ID()
		blueprint.Integer("user_id")
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.String("status")
		blueprint.Timestamps()

		// Note: Foreign key constraints skipped for SQLite compatibility
		// blueprint.Foreign("user_id")
	})
	if err != nil {
		return fmt.Errorf("error in migration 2: %w", err)
	}
	fmt.Println("Posts table created with foreign key")

	// Migration 3: Create comments table
	fmt.Println("\n=== Migration: Create Comments Table ===")
	err = db.Schema().Create("comments", func(blueprint contractsschema.Blueprint) {
		blueprint.ID()
		blueprint.Integer("post_id")
		blueprint.Integer("user_id")
		blueprint.Text("comment")
		blueprint.Timestamps()

		// Note: Foreign key constraints skipped for SQLite compatibility
		// blueprint.Foreign("post_id")
		// blueprint.Foreign("user_id")
	})
	if err != nil {
		return fmt.Errorf("error in migration 3: %w", err)
	}
	fmt.Println("Comments table created with foreign keys")

	// Migration 4: Add indexes to posts table
	fmt.Println("\n=== Migration: Add Indexes to Posts ===")
	err = db.Schema().Table("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.Index("user_id")
		blueprint.Index("status")
	})
	if err != nil {
		return fmt.Errorf("error in migration 4: %w", err)
	}
	fmt.Println("Indexes added to posts table")

	// Migration 5: Add published_at column to posts
	fmt.Println("\n=== Migration: Add published_at Column ===")
	err = db.Schema().Table("posts", func(blueprint contractsschema.Blueprint) {
		blueprint.Timestamp("published_at")
	})
	if err != nil {
		return fmt.Errorf("error in migration 5: %w", err)
	}
	fmt.Println("published_at column added to posts table")

	fmt.Println()
	fmt.Println("=== Schema Builder Example Completed ===")
	fmt.Println()

	return nil
}

// Register migrations in init()
func init() {
	// Register migration 1: Create users table
	migration.RegisterMigration("001_create_users_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("users", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.String("email")
				blueprint.Unique("email")
				blueprint.String("password")
				blueprint.String("status")
				blueprint.Timestamps()
				blueprint.SoftDeletes()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("users")
		},
	})

	// Register migration 2: Create posts table
	migration.RegisterMigration("002_create_posts_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("posts", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.Integer("user_id")
				blueprint.String("title")
				blueprint.Text("content")
				blueprint.String("status")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("posts")
		},
	})

	// Register migration 3: Create comments table
	migration.RegisterMigration("003_create_comments_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("comments", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.Integer("post_id")
				blueprint.Integer("user_id")
				blueprint.Text("comment")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("comments")
		},
	})

	// Register migration 4: Add indexes to posts table
	migration.RegisterMigration("004_add_posts_indexes", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Table("posts", func(blueprint contractsschema.Blueprint) {
				blueprint.Index("user_id")
				blueprint.Index("status")
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.Table("posts", func(blueprint contractsschema.Blueprint) {
				// Note: Dropping indexes is database-specific
				// This is a simplified example
			})
		},
	})

	// Register migration 5: Add published_at column to posts
	migration.RegisterMigration("005_add_published_to_posts", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Table("posts", func(blueprint contractsschema.Blueprint) {
				blueprint.Timestamp("published_at")
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.Table("posts", func(blueprint contractsschema.Blueprint) {
				blueprint.DropColumn("published_at")
			})
		},
	})
}

// RunMigrationSystemExample demonstrates using the migration system
// This approach provides version control and rollback capabilities
func RunMigrationSystemExample(dsn string) error {
	fmt.Println("=== Migration System Approach ===")
	fmt.Println("This approach uses the migration system with version control and rollback")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Run all pending migrations
	fmt.Println("=== Running Migrations ===")
	err = db.Migrate()
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	fmt.Println("Migrations completed successfully")

	// Check migration status
	fmt.Println("\n=== Migration Status ===")
	status, err := db.MigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	for _, s := range status {
		fmt.Printf("%s: Batch %d, Ran: %v\n", s.Name, s.Batch, s.Ran)
	}

	// Demonstrate rollback
	fmt.Println("\n=== Rolling Back Last Migration ===")
	err = db.MigrateDown(1)
	if err != nil {
		return fmt.Errorf("failed to rollback: %w", err)
	}
	fmt.Println("Rollback completed")

	// Check status again
	fmt.Println("\n=== Migration Status After Rollback ===")
	status, err = db.MigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	for _, s := range status {
		fmt.Printf("%s: Batch %d, Ran: %v\n", s.Name, s.Batch, s.Ran)
	}

	return nil
}
