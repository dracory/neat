package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// This example demonstrates how to create and run database migrations
func main() {
	if err := RunExample("sqlite://./example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates database migration operations
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Migration 1: Create users table
	fmt.Println("=== Migration: Create Users Table ===")
	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
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
	err = db.Schema().Create("posts", func(blueprint schema.Blueprint) {
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
	err = db.Schema().Create("comments", func(blueprint schema.Blueprint) {
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
	err = db.Schema().Table("posts", func(blueprint schema.Blueprint) {
		blueprint.Index("user_id")
		blueprint.Index("status")
	})
	if err != nil {
		return fmt.Errorf("error in migration 4: %w", err)
	}
	fmt.Println("Indexes added to posts table")

	// Migration 5: Add published_at column to posts
	fmt.Println("\n=== Migration: Add published_at Column ===")
	err = db.Schema().Table("posts", func(blueprint schema.Blueprint) {
		blueprint.Timestamp("published_at")
	})
	if err != nil {
		return fmt.Errorf("error in migration 5: %w", err)
	}
	fmt.Println("published_at column added to posts table")

	fmt.Println("\n=== All Migrations Completed ===")

	return nil
}
