package sqlite

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestSQLiteSchemaIndexCreateHasGetListing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_index_table"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
		table.String("email")
		table.Index("name")
		table.Unique("email")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if !db.Schema().HasIndex(tableName, "test_index_table_name_index") {
		t.Error("Expected name_index to exist")
	}
	if !db.Schema().HasIndex(tableName, "test_index_table_email_unique") {
		t.Error("Expected email_unique to exist")
	}

	indexes := db.Schema().GetIndexListing(tableName)
	nameIndexFound := false
	emailUniqueFound := false
	for _, idx := range indexes {
		if idx == "test_index_table_name_index" {
			nameIndexFound = true
		}
		if idx == "test_index_table_email_unique" {
			emailUniqueFound = true
		}
	}
	if !nameIndexFound {
		t.Error("name_index not found in listing")
	}
	if !emailUniqueFound {
		t.Error("email_unique not found in listing")
	}

	_ = db.Schema().Drop(tableName)
}

func TestSQLiteSchemaIndexMultiColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_multi_index"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("first_name")
		table.String("last_name")
		table.Index("first_name", "last_name").Name("name_idx")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if !db.Schema().HasIndex(tableName, "name_idx") {
		t.Error("Expected name_idx to exist")
	}

	indexes, err := db.Schema().GetIndexes(tableName)
	if err != nil {
		t.Fatalf("Failed to get indexes: %v", err)
	}

	found := false
	for _, idx := range indexes {
		if idx.Name == "name_idx" {
			if len(idx.Columns) != 2 || idx.Columns[0] != "first_name" || idx.Columns[1] != "last_name" {
				t.Errorf("Expected columns [first_name, last_name], got %v", idx.Columns)
			}
			if idx.Unique {
				t.Error("Expected non-unique index")
			}
			if idx.Primary {
				t.Error("Expected non-primary index")
			}
			found = true
		}
	}
	if !found {
		t.Error("name_idx not found in indexes")
	}

	_ = db.Schema().Drop(tableName)
}

func TestSQLiteSchemaIndexGetIndexesVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_get_indexes"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.Integer("id")
		table.String("name")
		table.String("email")
		table.Primary("id")
		table.Unique("name")
		table.Index("email")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	indexes, err := db.Schema().GetIndexes(tableName)
	if err != nil {
		t.Fatalf("Failed to get indexes: %v", err)
	}

	primaryFound := false
	for _, idx := range indexes {
		if idx.Primary {
			if len(idx.Columns) != 1 || idx.Columns[0] != "id" {
				t.Errorf("Expected primary key on [id], got %v", idx.Columns)
			}
			primaryFound = true
		}
	}
	if !primaryFound {
		t.Error("Primary key not found")
	}

	uniqueFound := false
	for _, idx := range indexes {
		if idx.Name == "test_get_indexes_name_unique" {
			if len(idx.Columns) != 1 || idx.Columns[0] != "name" {
				t.Errorf("Expected unique index on [name], got %v", idx.Columns)
			}
			uniqueFound = true
		}
	}
	if !uniqueFound {
		t.Error("Unique index not found")
	}

	normalFound := false
	for _, idx := range indexes {
		if idx.Name == "test_get_indexes_email_index" {
			if len(idx.Columns) != 1 || idx.Columns[0] != "email" {
				t.Errorf("Expected index on [email], got %v", idx.Columns)
			}
			if idx.Unique {
				t.Error("Expected non-unique index")
			}
			normalFound = true
		}
	}
	if !normalFound {
		t.Error("Normal index not found")
	}

	_ = db.Schema().Drop(tableName)
}

func TestSQLiteSchemaIndexDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_drop_index"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
		table.String("email")
		table.Index("name")
		table.Index("email").Name("email_idx")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if !db.Schema().HasIndex(tableName, "test_drop_index_name_index") {
		t.Error("Expected name_index to exist")
	}
	if !db.Schema().HasIndex(tableName, "email_idx") {
		t.Error("Expected email_idx to exist")
	}

	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.DropIndex("name")
	})
	if err != nil {
		t.Fatalf("Failed to drop index: %v", err)
	}
	if db.Schema().HasIndex(tableName, "test_drop_index_name_index") {
		t.Error("name_index should not exist after drop")
	}

	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.DropIndexByName("email_idx")
	})
	if err != nil {
		t.Fatalf("Failed to drop index by name: %v", err)
	}
	if db.Schema().HasIndex(tableName, "email_idx") {
		t.Error("email_idx should not exist after drop")
	}

	_ = db.Schema().Drop(tableName)
}

func TestSQLiteSchemaIndexRename(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("RenameIndex is currently problematic in SQLite with savepoints")

	db := SetupSQLiteTest(t)

	tableName := "test_rename_index"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
		table.Index("name").Name("old_idx")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	if !db.Schema().HasIndex(tableName, "old_idx") {
		t.Error("Expected old_idx to exist")
	}

	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.RenameIndex("old_idx", "new_idx")
	})
	if err != nil {
		t.Fatalf("Failed to rename index: %v", err)
	}

	if db.Schema().HasIndex(tableName, "old_idx") {
		t.Error("old_idx should not exist after rename")
	}
	if !db.Schema().HasIndex(tableName, "new_idx") {
		t.Error("new_idx should exist after rename")
	}

	_ = db.Schema().Drop(tableName)
}
