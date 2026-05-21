//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/contracts/database/schema"
)

func TestPostgreSQLSchemaTimestamp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	t.Run("Timestamps and TimestampsTz", func(t *testing.T) {
		tableName := "test_timestamps"
		_ = db.Schema().DropIfExists(tableName)

		err := db.Schema().Create(tableName, func(table schema.Blueprint) {
			table.ID()
			table.Timestamps()
		})
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		columns, err := db.Schema().GetColumns(tableName)
		if err != nil {
			t.Fatalf("Failed to get columns: %v", err)
		}

		hasCreatedAt := false
		hasUpdatedAt := false
		for _, col := range columns {
			if col.Name == "created_at" {
				hasCreatedAt = true
				if col.TypeName != "timestamp" {
					t.Errorf("Expected type 'timestamp', got '%s'", col.TypeName)
				}
				if !col.Nullable {
					t.Error("created_at should be nullable")
				}
			}
			if col.Name == "updated_at" {
				hasUpdatedAt = true
				if col.TypeName != "timestamp" {
					t.Errorf("Expected type 'timestamp', got '%s'", col.TypeName)
				}
				if !col.Nullable {
					t.Error("updated_at should be nullable")
				}
			}
		}
		if !hasCreatedAt {
			t.Error("created_at column should exist")
		}
		if !hasUpdatedAt {
			t.Error("updated_at column should exist")
		}

		_ = db.Schema().Drop(tableName)

		tableNameTz := "test_timestamps_tz"
		_ = db.Schema().DropIfExists(tableNameTz)
		err = db.Schema().Create(tableNameTz, func(table schema.Blueprint) {
			table.ID()
			table.TimestampsTz()
		})
		if err != nil {
			t.Fatalf("Failed to create table with TimestampsTz: %v", err)
		}

		columns, err = db.Schema().GetColumns(tableNameTz)
		if err != nil {
			t.Fatalf("Failed to get columns: %v", err)
		}

		hasCreatedAt = false
		hasUpdatedAt = false
		for _, col := range columns {
			if col.Name == "created_at" {
				hasCreatedAt = true
				if col.TypeName != "timestamptz" {
					t.Errorf("Expected type 'timestamptz', got '%s'", col.TypeName)
				}
			}
			if col.Name == "updated_at" {
				hasUpdatedAt = true
				if col.TypeName != "timestamptz" {
					t.Errorf("Expected type 'timestamptz', got '%s'", col.TypeName)
				}
			}
		}
		if !hasCreatedAt {
			t.Error("created_at column should exist")
		}
		if !hasUpdatedAt {
			t.Error("updated_at column should exist")
		}

		_ = db.Schema().Drop(tableNameTz)
	})

	t.Run("SoftDeletes and SoftDeletesTz", func(t *testing.T) {
		tableName := "test_soft_deletes"
		_ = db.Schema().DropIfExists(tableName)

		err := db.Schema().Create(tableName, func(table schema.Blueprint) {
			table.ID()
			table.SoftDeletes()
			table.SoftDeletesTz("deleted_at_tz")
		})
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		columns, err := db.Schema().GetColumns(tableName)
		if err != nil {
			t.Fatalf("Failed to get columns: %v", err)
		}

		hasDeletedAt := false
		hasDeletedAtTz := false
		for _, col := range columns {
			if col.Name == "deleted_at" {
				hasDeletedAt = true
				if col.TypeName != "timestamp" {
					t.Errorf("Expected type 'timestamp', got '%s'", col.TypeName)
				}
				if !col.Nullable {
					t.Error("deleted_at should be nullable")
				}
			}
			if col.Name == "deleted_at_tz" {
				hasDeletedAtTz = true
				if col.TypeName != "timestamptz" {
					t.Errorf("Expected type 'timestamptz', got '%s'", col.TypeName)
				}
				if !col.Nullable {
					t.Error("deleted_at_tz should be nullable")
				}
			}
		}
		if !hasDeletedAt {
			t.Error("deleted_at column should exist")
		}
		if !hasDeletedAtTz {
			t.Error("deleted_at_tz column should exist")
		}

		_ = db.Schema().Drop(tableName)
	})

	t.Run("DropTimestamps and DropSoftDeletes", func(t *testing.T) {
		tableName := "test_drop_timestamps"
		_ = db.Schema().DropIfExists(tableName)

		err := db.Schema().Create(tableName, func(table schema.Blueprint) {
			table.ID()
			table.Timestamps()
			table.SoftDeletes()
		})
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		err = db.Schema().Table(tableName, func(table schema.Blueprint) {
			table.DropTimestamps()
			table.DropSoftDeletes()
		})
		if err != nil {
			t.Fatalf("Failed to drop timestamps/soft deletes: %v", err)
		}

		if db.Schema().HasColumn(tableName, "created_at") {
			t.Error("created_at should not exist after drop")
		}
		if db.Schema().HasColumn(tableName, "updated_at") {
			t.Error("updated_at should not exist after drop")
		}
		if db.Schema().HasColumn(tableName, "deleted_at") {
			t.Error("deleted_at should not exist after drop")
		}

		_ = db.Schema().Drop(tableName)

		// Test Tz variants
		tableNameTz := "test_drop_timestamps_tz"
		_ = db.Schema().DropIfExists(tableNameTz)
		err = db.Schema().Create(tableNameTz, func(table schema.Blueprint) {
			table.ID()
			table.TimestampsTz()
			table.SoftDeletesTz("deleted_at_tz")
		})
		if err != nil {
			t.Fatalf("Failed to create table with Tz variants: %v", err)
		}

		err = db.Schema().Table(tableNameTz, func(table schema.Blueprint) {
			table.DropTimestampsTz()
			table.DropSoftDeletesTz("deleted_at_tz")
		})
		if err != nil {
			t.Fatalf("Failed to drop Tz variants: %v", err)
		}

		if db.Schema().HasColumn(tableNameTz, "created_at") {
			t.Error("created_at should not exist after drop")
		}
		if db.Schema().HasColumn(tableNameTz, "updated_at") {
			t.Error("updated_at should not exist after drop")
		}
		if db.Schema().HasColumn(tableNameTz, "deleted_at_tz") {
			t.Error("deleted_at_tz should not exist after drop")
		}

		_ = db.Schema().Drop(tableNameTz)
	})

	t.Run("Precision", func(t *testing.T) {
		tableName := "test_precision"
		_ = db.Schema().DropIfExists(tableName)

		err := db.Schema().Create(tableName, func(table schema.Blueprint) {
			table.ID()
			table.Timestamps(3)
			table.Timestamp("deleted_at", 3).Nullable()
		})
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		columns, err := db.Schema().GetColumns(tableName)
		if err != nil {
			t.Fatalf("Failed to get columns: %v", err)
		}

		for _, col := range columns {
			if col.Name == "created_at" || col.Name == "updated_at" || col.Name == "deleted_at" {
				if col.TypeName != "timestamp" {
					t.Errorf("Expected type 'timestamp', got '%s'", col.TypeName)
				}
			}
		}

		_ = db.Schema().Drop(tableName)
	})
}
