package turso

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestTursoSchemaTableCreateHasDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)

	tableName := "test_table"

	_ = db.Schema().DropIfExists(tableName)
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

func TestTursoSchemaTableRename(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)

	oldName := "old_table"
	newName := "new_table"

	_ = db.Schema().DropIfExists(oldName)
	_ = db.Schema().DropIfExists(newName)

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

	_ = db.Schema().DropIfExists(newName)
}
