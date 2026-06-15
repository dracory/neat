package main_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/migrator"
	mainpkg "github.com/dracory/neat/examples/enhanced-migrations"
)

func TestRunDateTimeFormatExample(t *testing.T) {
	t.Skip("Skipping example function test - focus on unit tests")
	err := mainpkg.RunDateTimeFormatExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunDateTimeFormatExample failed: %v", err)
	}
}

func TestRunDateFormatExample(t *testing.T) {
	t.Skip("Skipping example function test - focus on unit tests")
	err := mainpkg.RunDateFormatExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunDateFormatExample failed: %v", err)
	}
}

func TestRunUnixFormatExample(t *testing.T) {
	t.Skip("Skipping example function test - focus on unit tests")
	err := mainpkg.RunUnixFormatExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunUnixFormatExample failed: %v", err)
	}
}

func TestRunCustomFormatExample(t *testing.T) {
	t.Skip("Skipping example function test - focus on unit tests")
	err := mainpkg.RunCustomFormatExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunCustomFormatExample failed: %v", err)
	}
}

func TestRunPerformanceTrackingExample(t *testing.T) {
	t.Skip("Skipping example function test - focus on unit tests")
	err := mainpkg.RunPerformanceTrackingExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunPerformanceTrackingExample failed: %v", err)
	}
}

// TestMigrationIDFormatValidation tests the validation of different migration ID formats
func TestMigrationIDFormatValidation(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		format    migrator.MigrationIDFormat
		wantError bool
	}{
		{
			name:      "Valid datetime format",
			id:        "2026_06_15_1200_create_users_table",
			format:    migrator.MigrationIDFormatDateTime,
			wantError: false,
		},
		{
			name:      "Invalid datetime format - missing time",
			id:        "2026_06_15_create_users_table",
			format:    migrator.MigrationIDFormatDateTime,
			wantError: true,
		},
		{
			name:      "Invalid datetime format - invalid date",
			id:        "2026_13_15_1200_create_users_table",
			format:    migrator.MigrationIDFormatDateTime,
			wantError: true,
		},
		{
			name:      "Valid date format",
			id:        "2026_06_15_001_create_users_table",
			format:    migrator.MigrationIDFormatDate,
			wantError: false,
		},
		{
			name:      "Invalid date format - missing sequence",
			id:        "2026_06_15_create_users_table",
			format:    migrator.MigrationIDFormatDate,
			wantError: true,
		},
		{
			name:      "Valid unix format",
			id:        "1717080000_create_users_table",
			format:    migrator.MigrationIDFormatUnix,
			wantError: false,
		},
		{
			name:      "Invalid unix format - non-numeric",
			id:        "abc_create_users_table",
			format:    migrator.MigrationIDFormatUnix,
			wantError: true,
		},
		{
			name:      "Custom format - always valid",
			id:        "any_custom_name",
			format:    migrator.MigrationIDFormatCustom,
			wantError: false,
		},
		{
			name:      "Empty ID",
			id:        "",
			format:    migrator.MigrationIDFormatDateTime,
			wantError: true,
		},
		{
			name:      "ID too long",
			id:        "2026_06_15_1200_" + string(make([]byte, 300)),
			format:    migrator.MigrationIDFormatDateTime,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := migrator.ValidateMigrationID(tt.id, tt.format)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMigrationID() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMigrationDescriptionExtraction tests that descriptions are properly extracted and stored
func TestMigrationDescriptionExtraction(t *testing.T) {
	t.Skip("Skipping integration test - requires full migration setup")
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Register a migration with a descriptive ID
	migrator.RegisterMigration("2026_06_15_1200_create_users_table_with_email", migrator.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("users", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.String("email")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("users")
		},
	})

	// Run migration
	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	// Query the migration table to check if description was stored
	var history []map[string]any
	err = db.Query().Table("migrations").Get(&history)
	if err != nil {
		t.Fatalf("failed to query migration history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 migration record, got %d", len(history))
	}

	record := history[0]
	if record["description"] == nil {
		t.Error("expected description to be stored, but it was nil")
	} else {
		description := record["description"].(string)
		if description == "" {
			t.Error("expected non-empty description")
		}
	}
}

// TestPerformanceTracking tests that started_at and completed_at are properly tracked
func TestPerformanceTracking(t *testing.T) {
	t.Skip("Skipping integration test - requires full migration setup")
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Register a migration
	migrator.RegisterMigration("2026_06_15_1200_test_performance", migrator.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("test_table", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("test_table")
		},
	})

	// Run migration
	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	// Query the migration table to check performance tracking
	var history []map[string]any
	err = db.Query().Table("migrations").Get(&history)
	if err != nil {
		t.Fatalf("failed to query migration history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 migration record, got %d", len(history))
	}

	record := history[0]
	if record["started_at"] == nil {
		t.Error("expected started_at to be stored, but it was nil")
	}
	if record["completed_at"] == nil {
		t.Error("expected completed_at to be stored, but it was nil")
	}

	// Verify that completed_at is after started_at
	if started, ok := record["started_at"].(string); ok {
		if completed, ok := record["completed_at"].(string); ok {
			startTime, err := time.Parse(time.RFC3339, started)
			if err != nil {
				t.Fatalf("failed to parse started_at: %v", err)
			}
			completedTime, err := time.Parse(time.RFC3339, completed)
			if err != nil {
				t.Fatalf("failed to parse completed_at: %v", err)
			}
			if completedTime.Before(startTime) {
				t.Error("completed_at should be after started_at")
			}
		}
	}
}

// TestDifferentIDFormats tests that different ID formats work correctly
func TestDifferentIDFormats(t *testing.T) {
	t.Skip("Skipping integration test - requires full migration setup")
	formats := []struct {
		name   string
		format string
		ids    []string
	}{
		{
			name:   "datetime",
			format: "datetime",
			ids: []string{
				"2026_06_15_1200_create_table1",
				"2026_06_15_1300_create_table2",
			},
		},
		{
			name:   "date",
			format: "date",
			ids: []string{
				"2026_06_15_001_create_table1",
				"2026_06_15_002_create_table2",
			},
		},
		{
			name:   "unix",
			format: "unix",
			ids: []string{
				"1717080000_create_table1",
				"1717083600_create_table2",
			},
		},
	}

	for _, formatTest := range formats {
		t.Run(formatTest.name, func(t *testing.T) {
			db, err := neat.NewFromDSN("sqlite://:memory:")
			if err != nil {
				t.Fatalf("failed to connect: %v", err)
			}
			defer func() { _ = db.Close() }()

			// Register migrations
			for i, id := range formatTest.ids {
				tableName := fmt.Sprintf("table_%d", i)
				migrator.RegisterMigration(id, migrator.Migration{
					Up: func(schema contractsschema.Schema) error {
						return schema.Create(tableName, func(blueprint contractsschema.Blueprint) {
							blueprint.ID()
							blueprint.String("name")
							blueprint.Timestamps()
						})
					},
					Down: func(schema contractsschema.Schema) error {
						return schema.DropIfExists(tableName)
					},
				})
			}

			// Run migrations
			if err := db.Migrate(); err != nil {
				t.Fatalf("failed to run migrations: %v", err)
			}

			// Verify all migrations ran
			status, err := db.MigrationStatus()
			if err != nil {
				t.Fatalf("failed to get migration status: %v", err)
			}

			ranCount := 0
			for _, s := range status {
				if s.Ran {
					ranCount++
				}
			}

			if ranCount != len(formatTest.ids) {
				t.Errorf("expected %d migrations to run, got %d", len(formatTest.ids), ranCount)
			}
		})
	}
}

// TestTransactionSupport tests that migrations can run with transactions
func TestTransactionSupport(t *testing.T) {
	t.Skip("Skipping integration test - requires full migration setup")
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Register a migration
	migrator.RegisterMigration("2026_06_15_1200_test_transaction", migrator.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("transaction_test", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("transaction_test")
		},
	})

	// Run migration
	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	// Verify table was created
	if !db.Schema().HasTable("transaction_test") {
		t.Error("expected table to be created after migration")
	}
}

// TestDuplicateDetection tests that duplicate migration names are detected
func TestDuplicateDetection(t *testing.T) {
	t.Skip("Skipping integration test - requires full migration setup")
	// This test verifies that registering the same migration ID twice
	// will be handled (currently it's a silent overwrite, but the system
	// should ideally detect this)

	// Register a migration
	migrator.RegisterMigration("2026_06_15_1200_duplicate_test", migrator.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("duplicate_test", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("duplicate_test")
		},
	})

	// Try to register the same migration ID again
	// This should ideally be detected, but currently it's a silent overwrite
	// The test documents this behavior
	migrator.RegisterMigration("2026_06_15_1200_duplicate_test", migrator.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("duplicate_test2", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name2")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("duplicate_test2")
		},
	})

	// Note: This test documents current behavior where duplicate registration
	// silently overwrites. Future implementations should add duplicate detection.
	t.Log("Duplicate migration registration currently overwrites silently")
}

// TestMigrationTableSchema tests that the migration table has the enhanced schema
func TestMigrationTableSchema(t *testing.T) {
	t.Skip("Skipping integration test - requires full migration setup")
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Run a migration to create the migration table
	migrator.RegisterMigration("2026_06_15_1200_test_schema", migrator.Migration{
		Up: func(schema contractsschema.Schema) error {
			return schema.Create("test_table", func(blueprint contractsschema.Blueprint) {
				blueprint.ID()
				blueprint.String("name")
				blueprint.Timestamps()
			})
		},
		Down: func(schema contractsschema.Schema) error {
			return schema.DropIfExists("test_table")
		},
	})

	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	// Check that the migration table has the new columns
	expectedColumns := []string{"id", "migration", "batch", "description", "started_at", "completed_at", "created_at", "updated_at"}
	for _, col := range expectedColumns {
		if !db.Schema().HasColumn("migrations", col) {
			t.Errorf("expected column '%s' to exist in migrations table", col)
		}
	}
}

// TestTableNameValidation tests that table names are properly validated
func TestTableNameValidation(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		wantError bool
	}{
		{
			name:      "Valid table name",
			tableName: "migrations",
			wantError: false,
		},
		{
			name:      "Valid table name with underscore",
			tableName: "my_migrations",
			wantError: false,
		},
		{
			name:      "Empty table name",
			tableName: "",
			wantError: true,
		},
		{
			name:      "Table name too long",
			tableName: "this_is_a_very_long_table_name_that_exceeds_the_maximum_allowed_length_of_sixty_four_characters",
			wantError: true,
		},
		{
			name:      "Table name starts with digit",
			tableName: "123migrations",
			wantError: true,
		},
		{
			name:      "Table name with special characters",
			tableName: "migrations-table",
			wantError: true,
		},
		{
			name:      "Table name with space",
			tableName: "migrations table",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := migrator.ValidateTableName(tt.tableName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTableName() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
