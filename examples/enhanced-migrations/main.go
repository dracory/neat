package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/migration"
)

// This example demonstrates the enhanced migration system features
func main() {
	// Example 1: Default datetime format with descriptions
	if err := RunDateTimeFormatExample("sqlite://./example_datetime.db"); err != nil {
		log.Fatalf("DateTime format example failed: %v", err)
	}

	// Example 2: Date sequence format
	if err := RunDateFormatExample("sqlite://./example_date.db"); err != nil {
		log.Fatalf("Date format example failed: %v", err)
	}

	// Example 3: Unix timestamp format (legacy)
	if err := RunUnixFormatExample("sqlite://./example_unix.db"); err != nil {
		log.Fatalf("Unix format example failed: %v", err)
	}

	// Example 4: Custom format with validation disabled
	if err := RunCustomFormatExample("sqlite://./example_custom.db"); err != nil {
		log.Fatalf("Custom format example failed: %v", err)
	}

	// Example 5: Performance tracking and history
	if err := RunPerformanceTrackingExample("sqlite://./example_performance.db"); err != nil {
		log.Fatalf("Performance tracking example failed: %v", err)
	}
}

// RunDateTimeFormatExample demonstrates the default datetime format (YYYY_MM_DD_HHMM_description)
func RunDateTimeFormatExample(dsn string) error {
	fmt.Println("=== DateTime Format Example ===")
	fmt.Println("Format: YYYY_MM_DD_HHMM_description")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Register migrations with datetime format IDs
	migration.RegisterMigration("2026_06_15_1200_create_users_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("users", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.String("email")
				blueprint.Unique("email")
				blueprint.String("password")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("users")
		},
	})

	migration.RegisterMigration("2026_06_15_1300_create_posts_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("posts", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.Integer("user_id")
				blueprint.String("title")
				blueprint.Text("content")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("posts")
		},
	})

	// Run migrations
	fmt.Println("Running migrations with datetime format...")
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	fmt.Println("Migrations completed successfully")

	// Check status
	status, err := db.MigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Println("\nMigration Status:")
	for _, s := range status {
		fmt.Printf("  %s: Batch %d, Ran: %v\n", s.Name, s.Batch, s.Ran)
	}

	fmt.Println()
	return nil
}

// RunDateFormatExample demonstrates the date sequence format (YYYY_MM_DD_NNN_description)
func RunDateFormatExample(dsn string) error {
	fmt.Println("=== Date Sequence Format Example ===")
	fmt.Println("Format: YYYY_MM_DD_NNN_description (NNN = sequence number)")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Register migrations with date format IDs
	migration.RegisterMigration("2026_06_15_001_create_products_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("products", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.Decimal("price")
				blueprint.Integer("stock")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("products")
		},
	})

	migration.RegisterMigration("2026_06_15_002_create_orders_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("orders", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.Integer("product_id")
				blueprint.Integer("quantity")
				blueprint.Decimal("total")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("orders")
		},
	})

	// Run migrations
	fmt.Println("Running migrations with date sequence format...")
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	fmt.Println("Migrations completed successfully")

	// Check status
	status, err := db.MigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Println("\nMigration Status:")
	for _, s := range status {
		fmt.Printf("  %s: Batch %d, Ran: %v\n", s.Name, s.Batch, s.Ran)
	}

	fmt.Println()
	return nil
}

// RunUnixFormatExample demonstrates the unix timestamp format (legacy)
func RunUnixFormatExample(dsn string) error {
	fmt.Println("=== Unix Timestamp Format Example ===")
	fmt.Println("Format: unix_timestamp_description (legacy format)")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Register migrations with unix timestamp IDs
	migration.RegisterMigration("1717080000_create_categories_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("categories", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.String("slug")
				blueprint.Text("description")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("categories")
		},
	})

	// Run migrations
	fmt.Println("Running migrations with unix timestamp format...")
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	fmt.Println("Migrations completed successfully")

	// Check status
	status, err := db.MigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Println("\nMigration Status:")
	for _, s := range status {
		fmt.Printf("  %s: Batch %d, Ran: %v\n", s.Name, s.Batch, s.Ran)
	}

	fmt.Println()
	return nil
}

// RunCustomFormatExample demonstrates custom format with validation disabled
func RunCustomFormatExample(dsn string) error {
	fmt.Println("=== Custom Format Example ===")
	fmt.Println("Format: No prefix validation (custom names)")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Register migrations with custom IDs
	migration.RegisterMigration("initial_setup", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("settings", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("key")
				blueprint.Text("value")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("settings")
		},
	})

	migration.RegisterMigration("add_indexes", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Table("settings", func(blueprint contractsschema.Blueprint) {
				blueprint.Index("key")
			})
		},
		Down: func(schema contractsschema.Schema) error {
			// Note: Dropping indexes is database-specific
			return nil
		},
	})

	// Run migrations
	fmt.Println("Running migrations with custom format...")
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	fmt.Println("Migrations completed successfully")

	// Check status
	status, err := db.MigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Println("\nMigration Status:")
	for _, s := range status {
		fmt.Printf("  %s: Batch %d, Ran: %v\n", s.Name, s.Batch, s.Ran)
	}

	fmt.Println()
	return nil
}

// RunPerformanceTrackingExample demonstrates performance tracking and migration history
func RunPerformanceTrackingExample(dsn string) error {
	fmt.Println("=== Performance Tracking Example ===")
	fmt.Println("Tracking migration execution time and history")
	fmt.Println()

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Register migrations with descriptions
	migration.RegisterMigration("2026_06_15_1400_create_auth_tables", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			// Simulate a slow migration
			time.Sleep(100 * time.Millisecond)
			return schema.Create("auth_users", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("username")
				blueprint.String("email")
				blueprint.String("password_hash")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("auth_users")
		},
	})

	migration.RegisterMigration("2026_06_15_1500_create_permissions_table", migration.Migration{
		Up: func(schema contractsschema.Schema) error {
			// Simulate a faster migration
			time.Sleep(50 * time.Millisecond)
			return schema.Create("permissions", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.String("slug")
				blueprint.Text("description")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("permissions")
		},
	})

	// Run migrations
	fmt.Println("Running migrations with performance tracking...")
	startTime := time.Now()
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	totalDuration := time.Since(startTime)
	fmt.Printf("All migrations completed in %v\n", totalDuration)

	// Get detailed migration status
	status, err := db.MigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Println("\nMigration Status:")
	for _, s := range status {
		fmt.Printf("  %s: Batch %d, Ran: %v\n", s.Name, s.Batch, s.Ran)
	}

	// Query migration table directly to show performance data
	fmt.Println("\nMigration History (from database):")
	var history []map[string]any
	err = db.Query().Table("migrations").
		OrderBy("id", "asc").
		Get(&history)
	if err != nil {
		return fmt.Errorf("failed to query migration history: %w", err)
	}

	for _, record := range history {
		fmt.Printf("  Migration: %v\n", record["migration"])
		fmt.Printf("  Batch: %v\n", record["batch"])
		if record["description"] != nil {
			fmt.Printf("  Description: %v\n", record["description"])
		}
		if record["started_at"] != nil {
			fmt.Printf("  Started: %v\n", record["started_at"])
		}
		if record["completed_at"] != nil {
			fmt.Printf("  Completed: %v\n", record["completed_at"])
			// Calculate duration
			if started, ok := record["started_at"].(string); ok {
				if completed, ok := record["completed_at"].(string); ok {
					startTime, _ := time.Parse(time.RFC3339, started)
					completedTime, _ := time.Parse(time.RFC3339, completed)
					duration := completedTime.Sub(startTime)
					fmt.Printf("  Duration: %v\n", duration)
				}
			}
		}
		fmt.Println()
	}

	fmt.Println()
	return nil
}
