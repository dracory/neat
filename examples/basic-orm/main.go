package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// This example demonstrates basic ORM usage with the query builder
func main() {
	if err := RunExample("sqlite://./example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates basic ORM usage with the query builder
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create users table for the example
	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("age")
		blueprint.String("status")
		blueprint.Timestamp("created_at")
	})
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Get all records
	fmt.Println("=== Get All Records ===")
	var results []map[string]any
	err = db.Query().Table("users").Get(&results)
	if err != nil {
		return fmt.Errorf("Error getting records: %v", err)
	} else {
		fmt.Printf("Found %d records\n", len(results))
	}

	// Find by ID
	fmt.Println("\n=== Find by ID ===")
	var user map[string]any
	err = db.Query().Table("users").Where("id = ?", 1).Get(&user)
	if err != nil {
		return fmt.Errorf("Error finding user: %v", err)
	} else {
		fmt.Printf("Found user: %v\n", user)
	}

	// Create a new record
	fmt.Println("\n=== Create Record ===")
	newUser := map[string]any{
		"name":       "John Doe",
		"email":      "john@example.com",
		"age":        30,
		"status":     "active",
		"created_at": "2026-05-12 18:00:00",
	}
	err = db.Query().Table("users").Create(newUser)
	if err != nil {
		return fmt.Errorf("Error creating user: %v", err)
	} else {
		fmt.Println("User created successfully")
	}

	// Update a record
	fmt.Println("\n=== Update Record ===")
	_, err = db.Query().Table("users").Where("id = ?", 1).Update(map[string]any{
		"name": "Jane Doe",
	})
	if err != nil {
		return fmt.Errorf("Error updating user: %v", err)
	} else {
		fmt.Println("User updated successfully")
	}

	// Delete a record
	fmt.Println("\n=== Delete Record ===")
	_, err = db.Query().Table("users").Where("id = ?", 1).Delete()
	if err != nil {
		return fmt.Errorf("Error deleting user: %v", err)
	} else {
		fmt.Println("User deleted successfully")
	}

	// Advanced query with multiple conditions
	fmt.Println("\n=== Advanced Query ===")
	var advancedResults []map[string]any
	err = db.Query().Table("users").
		Where("age > ?", 18).
		Where("status = ?", "active").
		OrderBy("created_at", "desc").
		Limit(10).
		Get(&advancedResults)
	if err != nil {
		return fmt.Errorf("Error in advanced query: %v", err)
	} else {
		fmt.Printf("Found %d active users over 18\n", len(advancedResults))
	}

	return nil
}
