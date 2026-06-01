package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// User represents a user model for the factory example
type User struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Email    string `db:"email"`
	Age      int    `db:"age"`
	Status   string `db:"status"`
	IsActive bool   `db:"is_active"`
}

// This example demonstrates the Factory pattern for creating test data
func main() {
	if err := RunExample("sqlite://./factory_example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates Factory pattern usage
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
		blueprint.Boolean("is_active")
	})
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Example 1: Create a single user with Factory
	fmt.Println("=== Example 1: Create Single User ===")
	user := &User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Status:   "active",
		IsActive: true,
	}
	_, err = db.Factory().Table("users").Create(user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Println("User created successfully")

	// Example 2: Create multiple users with Count()
	fmt.Println("\n=== Example 2: Bulk Create with Count() ===")
	templateUser := &User{
		Name:     "Template User",
		Email:    "template@example.com",
		Age:      25,
		Status:   "pending",
		IsActive: false,
	}
	_, err = db.Factory().Table("users").Count(3).Create(templateUser)
	if err != nil {
		return fmt.Errorf("failed to create bulk users: %w", err)
	}
	fmt.Println("3 users created successfully")

	// Example 3: Create another user
	fmt.Println("\n=== Example 3: Create Another User ===")
	anotherUser := &User{
		Name:     "Custom User",
		Email:    "custom@example.com",
		Age:      35,
		Status:   "active",
		IsActive: true,
	}
	_, err = db.Factory().Table("users").Create(anotherUser)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Println("User created successfully")

	// Example 4: Create without firing model events (CreateQuietly)
	fmt.Println("\n=== Example 4: Create Quietly (No Events) ===")
	quietUser := &User{
		Name:     "Quiet User",
		Email:    "quiet@example.com",
		Age:      28,
		Status:   "active",
		IsActive: true,
	}
	_, err = db.Factory().Table("users").CreateQuietly(quietUser)
	if err != nil {
		return fmt.Errorf("failed to create user quietly: %w", err)
	}
	fmt.Println("User created quietly (without events)")

	// Example 5: Make instances without persisting to database
	fmt.Println("\n=== Example 5: Make (In-Memory Only) ===")
	makeUser := &User{
		Name:  "Make User",
		Email: "make@example.com",
		Age:   22,
	}
	_, err = db.Factory().Make(makeUser)
	if err != nil {
		return fmt.Errorf("failed to make user: %w", err)
	}
	fmt.Printf("User created in memory: %+v\n", makeUser)
	fmt.Println("(Note: This user was not persisted to the database)")

	// Example 6: Bulk make with Count()
	fmt.Println("\n=== Example 6: Bulk Make with Count() ===")
	bulkTemplate := &User{
		Name:  "Bulk Template",
		Email: "bulk@example.com",
		Age:   30,
	}
	bulkUsers, err := db.Factory().Count(2).Make(bulkTemplate)
	if err != nil {
		return fmt.Errorf("failed to bulk make users: %w", err)
	}
	users, ok := bulkUsers.([]*User)
	if !ok {
		return fmt.Errorf("unexpected type returned from Make: %T", bulkUsers)
	}
	fmt.Printf("%d users created in memory (not persisted)\n", len(users))

	// Example 7: Create another user
	fmt.Println("\n=== Example 7: Create Final User ===")
	finalUser := &User{
		Name:     "Final User",
		Email:    "final@example.com",
		Age:      25,
		Status:   "active",
		IsActive: true,
	}
	_, err = db.Factory().Table("users").Create(finalUser)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Println("User created successfully")

	// Verify created records
	fmt.Println("\n=== Verification: Count All Users ===")
	var count int64
	err = db.Query().Table("users").Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	fmt.Printf("Total users in database: %d\n", count)

	return nil
}
