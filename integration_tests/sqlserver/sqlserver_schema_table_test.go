package sqlserver

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

// TestSQLServerSchemaTableCreateHasDrop verifies the full lifecycle of a table:
// DropIfExists (no-op when absent), Create, HasTable, Drop, and a second
// DropIfExists (no-op when absent again).
func TestSQLServerSchemaTableCreateHasDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	tableName := "test_table"

	db.Schema().DropIfExists(tableName)
	if db.Schema().HasTable(tableName) {
		t.Error("Table should not exist after DropIfExists")
	}

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	if !db.Schema().HasTable(tableName) {
		t.Error("Table should exist after Create")
	}

	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
	if db.Schema().HasTable(tableName) {
		t.Error("Table should not exist after Drop")
	}

	err = db.Schema().DropIfExists(tableName)
	if err != nil {
		t.Fatalf("Failed to drop if exists: %v", err)
	}
}

// TestSQLServerSchemaTableRename verifies that Rename() moves a table to the
// new name, that the old name no longer exists, and that the new name does.
func TestSQLServerSchemaTableRename(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	oldName := "old_table"
	newName := "new_table"

	db.Schema().DropIfExists(oldName)
	db.Schema().DropIfExists(newName)

	err := db.Schema().Create(oldName, func(table schema.Blueprint) {
		table.ID()
	})
	if err != nil {
		t.Fatalf("Failed to create old table: %v", err)
	}

	err = db.Schema().Rename(oldName, newName)
	if err != nil {
		t.Fatalf("Failed to rename table: %v", err)
	}

	if db.Schema().HasTable(oldName) {
		t.Error("Old table should not exist after rename")
	}
	if !db.Schema().HasTable(newName) {
		t.Error("New table should exist after rename")
	}

	db.Schema().Drop(newName)
}

// TestSQLServerSchemaTableGetTables verifies that GetTableListing() and
// GetTables() both reflect newly created tables in their results.
func TestSQLServerSchemaTableGetTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	table1 := "table1"
	table2 := "table2"

	db.Schema().DropIfExists(table1)
	db.Schema().DropIfExists(table2)

	if err := db.Schema().Create(table1, func(table schema.Blueprint) { table.ID() }); err != nil {
		t.Fatalf("Failed to create table1: %v", err)
	}
	if err := db.Schema().Create(table2, func(table schema.Blueprint) { table.ID() }); err != nil {
		t.Fatalf("Failed to create table2: %v", err)
	}

	tables := db.Schema().GetTableListing()
	table1Found := false
	table2Found := false
	for _, t := range tables {
		if t == table1 {
			table1Found = true
		}
		if t == table2 {
			table2Found = true
		}
	}
	if !table1Found {
		t.Error("table1 should be in table listing")
	}
	if !table2Found {
		t.Error("table2 should be in table listing")
	}

	tableInfos, err := db.Schema().GetTables()
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	found1 := false
	found2 := false
	for _, ti := range tableInfos {
		if ti.Name == table1 {
			found1 = true
		}
		if ti.Name == table2 {
			found2 = true
		}
	}
	if !found1 {
		t.Error("table1 should be in GetTables result")
	}
	if !found2 {
		t.Error("table2 should be in GetTables result")
	}

	db.Schema().Drop(table1)
	db.Schema().Drop(table2)
}

// TestSQLServerSchemaTableModify verifies that Table() can alter an existing
// table by adding a new column without error.
func TestSQLServerSchemaTableModify(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	tableName := "modify_table"
	db.Schema().DropIfExists(tableName)

	if err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
	}); err != nil {
		t.Fatalf("Failed to create modify table: %v", err)
	}

	if err := db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.String("name")
	}); err != nil {
		t.Fatalf("Failed to modify table: %v", err)
	}

	db.Schema().Drop(tableName)
}
