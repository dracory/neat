package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// This example demonstrates schema builder usage for creating and modifying tables
func main() {
	if err := RunExample("sqlite://./example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates schema builder operations
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create a new table
	fmt.Println("=== Create Table ===")
	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Unique("email")
		blueprint.String("phone")
		blueprint.Integer("age")
		blueprint.String("status")
		blueprint.Timestamps()
		blueprint.SoftDeletes()
	})
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}
	fmt.Println("Table created successfully")

	// Modify an existing table
	fmt.Println("\n=== Modify Table ===")
	err = db.Schema().Table("users", func(blueprint schema.Blueprint) {
		blueprint.String("address")
		blueprint.String("city")
		blueprint.DropColumn("phone")
	})
	if err != nil {
		return fmt.Errorf("error modifying table: %w", err)
	}
	fmt.Println("Table modified successfully")

	// Add indexes
	fmt.Println("\n=== Add Index ===")
	err = db.Schema().Table("users", func(blueprint schema.Blueprint) {
		blueprint.Index("email")
		blueprint.Unique("name")
	})
	if err != nil {
		return fmt.Errorf("error adding indexes: %w", err)
	}
	fmt.Println("Indexes added successfully")

	// Drop a table
	fmt.Println("\n=== Drop Table ===")
	err = db.Schema().Drop("users")
	if err != nil {
		return fmt.Errorf("error dropping table: %w", err)
	}
	fmt.Println("Table dropped successfully")

	return nil
}
