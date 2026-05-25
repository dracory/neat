//go:build disabled

package mysql

import (
	"testing"
	"github.com/dracory/neat/contracts/database/schema"
)

func TestMySQLSchemaRenameColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)

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
		t.Error("Column 'old_name' should exist")
	}
	if db.Schema().HasColumn(tableName, "new_name") {
		t.Error("Column 'new_name' should not exist")
	}

	// Rename column via Table method
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.RenameColumn("old_name", "new_name")
	})
	if err != nil {
		t.Fatalf("Failed to rename column: %v", err)
	}

	// Verify column was renamed
	if db.Schema().HasColumn(tableName, "old_name") {
		t.Error("Column 'old_name' should not exist after rename")
	}
	if !db.Schema().HasColumn(tableName, "new_name") {
		t.Error("Column 'new_name' should exist after rename")
	}

	// Rename column via Schema method
	err = db.Schema().RenameColumn(tableName, "new_name", "final_name")
	if err != nil {
		t.Fatalf("Failed to rename column via Schema method: %v", err)
	}

	// Verify final rename
	if db.Schema().HasColumn(tableName, "new_name") {
		t.Error("Column 'new_name' should not exist after final rename")
	}
	if !db.Schema().HasColumn(tableName, "final_name") {
		t.Error("Column 'final_name' should exist after final rename")
	}
}
