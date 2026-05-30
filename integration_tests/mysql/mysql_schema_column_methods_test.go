
package mysql

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
)

func setupColumnMethodsTable(t *testing.T, db *database.Database, tableName string) {
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
		table.String("age")
		table.String("weight")
		table.String("height")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if !db.Schema().HasTable(tableName) {
		t.Error("Table should exist")
	}
}

func TestMySQLSchemaColumnMethodsHasColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	tableName := "test_column_methods"
	setupColumnMethodsTable(t, db, tableName)

	if !db.Schema().HasColumn(tableName, "name") {
		t.Error("Column 'name' should exist")
	}
	if !db.Schema().HasColumn(tableName, "age") {
		t.Error("Column 'age' should exist")
	}
	if !db.Schema().HasColumn(tableName, "weight") {
		t.Error("Column 'weight' should exist")
	}
	if !db.Schema().HasColumn(tableName, "height") {
		t.Error("Column 'height' should exist")
	}
	if db.Schema().HasColumn(tableName, "nonexistent") {
		t.Error("Nonexistent column should not exist")
	}

	db.Schema().Drop(tableName)
}

func TestMySQLSchemaColumnMethodsHasColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	tableName := "test_column_methods"
	setupColumnMethodsTable(t, db, tableName)

	if !db.Schema().HasColumns(tableName, []string{"name", "age", "weight"}) {
		t.Error("All columns should exist")
	}
	if !db.Schema().HasColumns(tableName, []string{"id", "name"}) {
		t.Error("ID and name should exist")
	}
	if db.Schema().HasColumns(tableName, []string{"name", "nonexistent"}) {
		t.Error("Should fail if one column doesn't exist")
	}
	if db.Schema().HasColumns(tableName, []string{"nonexistent1", "nonexistent2"}) {
		t.Error("Should fail if all columns don't exist")
	}

	db.Schema().Drop(tableName)
}

func TestMySQLSchemaColumnMethodsDropColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	tableName := "test_column_methods"
	setupColumnMethodsTable(t, db, tableName)

	err := db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.DropColumn("name", "age")
	})
	if err != nil {
		t.Fatalf("Failed to drop columns: %v", err)
	}

	if db.Schema().HasColumn(tableName, "name") {
		t.Error("Column 'name' should not exist after drop")
	}
	if db.Schema().HasColumn(tableName, "age") {
		t.Error("Column 'age' should not exist after drop")
	}
	if !db.Schema().HasColumn(tableName, "weight") {
		t.Error("Column 'weight' should still exist")
	}
	if !db.Schema().HasColumn(tableName, "height") {
		t.Error("Column 'height' should still exist")
	}

	db.Schema().Drop(tableName)
}

func TestMySQLSchemaColumnMethodsDropColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	tableName := "test_column_methods"
	setupColumnMethodsTable(t, db, tableName)

	err := db.Schema().DropColumns(tableName, []string{"weight"})
	if err != nil {
		t.Fatalf("Failed to drop columns via DropColumns: %v", err)
	}

	if db.Schema().HasColumn(tableName, "weight") {
		t.Error("Column 'weight' should not exist after DropColumns")
	}
	if !db.Schema().HasColumn(tableName, "height") {
		t.Error("Column 'height' should still exist")
	}

	db.Schema().Drop(tableName)
}

func TestMySQLSchemaColumnMethodsHasColumnsAfterDrops(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	tableName := "test_column_methods"
	setupColumnMethodsTable(t, db, tableName)

	db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.DropColumn("name", "age")
	})
	db.Schema().DropColumns(tableName, []string{"weight"})

	if db.Schema().HasColumns(tableName, []string{"name", "age", "weight"}) {
		t.Error("All dropped columns should not exist")
	}
	if !db.Schema().HasColumns(tableName, []string{"id", "height"}) {
		t.Error("Remaining columns should exist")
	}

	db.Schema().Drop(tableName)
}

func TestMySQLSchemaColumnMethodsSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	tableName := "test_column_methods_single"

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("single_column")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if !db.Schema().HasColumn(tableName, "single_column") {
		t.Error("single_column should exist")
	}
	if !db.Schema().HasColumns(tableName, []string{"single_column"}) {
		t.Error("single_column should exist in HasColumns")
	}

	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.DropColumn("single_column")
	})
	if err != nil {
		t.Fatalf("Failed to drop single column: %v", err)
	}

	if db.Schema().HasColumn(tableName, "single_column") {
		t.Error("single_column should not exist after drop")
	}
	if db.Schema().HasColumns(tableName, []string{"single_column"}) {
		t.Error("single_column should not exist in HasColumns after drop")
	}

	db.Schema().Drop(tableName)
}

func TestMySQLSchemaColumnMethodsEmptyColumnList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	tableName := "test_column_methods_empty"

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("test_column")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if !db.Schema().HasColumns(tableName, []string{}) {
		t.Error("HasColumns with empty list should return true")
	}

	db.Schema().Drop(tableName)
}
