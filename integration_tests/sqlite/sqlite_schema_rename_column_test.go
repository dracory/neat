package sqlite

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestSQLiteSchemaRenameColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_rename_column"

	// Create a table
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("old_name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	defer func() {
		_ = db.Schema().Drop(tableName)
	}()

	// Verify column exists
	if !db.Schema().HasColumn(tableName, "old_name") {
		t.Error("old_name column should exist")
	}
	if db.Schema().HasColumn(tableName, "new_name") {
		t.Error("new_name column should not exist yet")
	}

	// Rename column via Table method
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.RenameColumn("old_name", "new_name")
	})
	if err != nil {
		t.Fatalf("Failed to rename column via Table: %v", err)
	}

	// Verify column was renamed
	if db.Schema().HasColumn(tableName, "old_name") {
		t.Error("old_name column should not exist after rename")
	}
	if !db.Schema().HasColumn(tableName, "new_name") {
		t.Error("new_name column should exist after rename")
	}

	// Rename column via Schema method
	err = db.Schema().RenameColumn(tableName, "new_name", "final_name")
	if err != nil {
		t.Fatalf("Failed to rename column via Schema: %v", err)
	}

	// Verify final rename
	if db.Schema().HasColumn(tableName, "new_name") {
		t.Error("new_name column should not exist after final rename")
	}
	if !db.Schema().HasColumn(tableName, "final_name") {
		t.Error("final_name column should exist after final rename")
	}
}
