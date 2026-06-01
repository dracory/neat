package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	contractsseeder "github.com/dracory/neat/contracts/database/seeder"
)

// This example demonstrates how to use seeders to populate your database with initial data
func main() {
	if err := RunExample("sqlite://./example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates seeder usage
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create tables for the example
	err = createTables(db)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Example 1: Run seeders using SeedOnce (only runs once)
	fmt.Println("=== Example 1: Using SeedOnce() ===")
	userSeeder1 := &UserSeeder{db: db}
	roleSeeder1 := &RoleSeeder{db: db}
	seeders1 := []contractsseeder.Seeder{roleSeeder1, userSeeder1}
	err = db.SeedOnce(seeders1)
	if err != nil {
		return fmt.Errorf("failed to run seeders once: %w", err)
	}
	fmt.Println("Seeders executed once")

	// Call SeedOnce again to demonstrate it skips
	err = db.SeedOnce(seeders1)
	if err != nil {
		return fmt.Errorf("failed to run seeders once (second call): %w", err)
	}
	fmt.Println("Seeders skipped on second call (already run)")

	// Example 2: Run seeders using Seed method (runs every time)
	fmt.Println("\n=== Example 2: Using Seed() ===")
	userSeeder2 := &UserSeeder{db: db}
	roleSeeder2 := &RoleSeeder{db: db}
	seeders2 := []contractsseeder.Seeder{roleSeeder2, userSeeder2}
	err = db.Seed(seeders2)
	if err != nil {
		return fmt.Errorf("failed to run seeders: %w", err)
	}
	fmt.Println("Seeders executed successfully (Seed runs every time)")

	// Example 3: Using the Seeder facade for advanced operations
	fmt.Println("\n=== Example 3: Using Seeder Facade ===")
	facade := db.Seeder()
	facade.Register(seeders1)

	// Get a specific seeder
	s := facade.GetSeeder("user_seeder")
	if s != nil {
		fmt.Printf("Found seeder: %s\n", s.Signature())
	}

	// Get all seeders
	allSeeders := facade.GetSeeders()
	fmt.Printf("Total registered seeders: %d\n", len(allSeeders))

	// Example 4: Verify seeded data
	fmt.Println("\n=== Example 4: Verify Seeded Data ===")
	var users []map[string]any
	err = db.Query().Table("users").Get(&users)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}
	fmt.Printf("Total users seeded: %d\n", len(users))

	var roles []map[string]any
	err = db.Query().Table("roles").Get(&roles)
	if err != nil {
		return fmt.Errorf("failed to get roles: %w", err)
	}
	fmt.Printf("Total roles seeded: %d\n", len(roles))

	return nil
}

// RunExampleForTest runs the example and returns the database for assertions
func RunExampleForTest(dsn string) (*neat.Database, error) {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create tables for the example
	err = createTables(db)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Example 1: Run seeders using SeedOnce (only runs once)
	userSeeder1 := &UserSeeder{db: db}
	roleSeeder1 := &RoleSeeder{db: db}
	seeders1 := []contractsseeder.Seeder{roleSeeder1, userSeeder1}
	err = db.SeedOnce(seeders1)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run seeders once: %w", err)
	}

	// Call SeedOnce again to demonstrate it skips
	err = db.SeedOnce(seeders1)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run seeders once (second call): %w", err)
	}

	// Example 2: Run seeders using Seed method (runs every time)
	userSeeder2 := &UserSeeder{db: db}
	roleSeeder2 := &RoleSeeder{db: db}
	seeders2 := []contractsseeder.Seeder{roleSeeder2, userSeeder2}
	err = db.Seed(seeders2)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run seeders: %w", err)
	}

	// Example 3: Using the Seeder facade for advanced operations
	facade := db.Seeder()
	facade.Register(seeders1)

	return db, nil
}

// createTables creates the necessary tables for the example
func createTables(db *neat.Database) error {
	// Create roles table
	err := db.Schema().Create("roles", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("description")
		blueprint.Timestamp("created_at")
	})
	if err != nil {
		return fmt.Errorf("failed to create roles table: %w", err)
	}

	// Create users table
	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("role_id")
		blueprint.Timestamp("created_at")
	})
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}

// UserSeeder seeds the users table with sample data
type UserSeeder struct {
	db *neat.Database
}

func (s *UserSeeder) Signature() string {
	return "user_seeder"
}

func (s *UserSeeder) Run() error {
	users := []map[string]any{
		{"name": "John Doe", "email": "john@example.com", "role_id": 1, "created_at": "2026-05-30 00:00:00"},
		{"name": "Jane Smith", "email": "jane@example.com", "role_id": 2, "created_at": "2026-05-30 00:00:00"},
		{"name": "Bob Johnson", "email": "bob@example.com", "role_id": 1, "created_at": "2026-05-30 00:00:00"},
	}

	for _, user := range users {
		err := s.db.Query().Table("users").Create(user)
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
	}

	fmt.Printf("Seeded %d users\n", len(users))
	return nil
}

// RoleSeeder seeds the roles table with sample data
type RoleSeeder struct {
	db *neat.Database
}

func (s *RoleSeeder) Signature() string {
	return "role_seeder"
}

func (s *RoleSeeder) Run() error {
	roles := []map[string]any{
		{"name": "Admin", "description": "Administrator with full access", "created_at": "2026-05-30 00:00:00"},
		{"name": "User", "description": "Standard user with limited access", "created_at": "2026-05-30 00:00:00"},
		{"name": "Guest", "description": "Guest user with read-only access", "created_at": "2026-05-30 00:00:00"},
	}

	for _, role := range roles {
		err := s.db.Query().Table("roles").Create(role)
		if err != nil {
			return fmt.Errorf("failed to insert role: %w", err)
		}
	}

	fmt.Printf("Seeded %d roles\n", len(roles))
	return nil
}
