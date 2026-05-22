//go:build integration

package mysql

import (
	"testing"
	"github.com/dracory/neat/contracts/database/schema"
)

func TestMySQLSchemaIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)

	t.Run("Create Index, HasIndex, GetIndexListing", func(t *testing.T) {
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
			t.Error("Index 'test_index_table_name_index' should exist")
		}
		if !db.Schema().HasIndex(tableName, "test_index_table_email_unique") {
			t.Error("Index 'test_index_table_email_unique' should exist")
		}

		indexes := db.Schema().GetIndexListing(tableName)
		nameIndexFound := false
		emailIndexFound := false
		for _, idx := range indexes {
			if idx == "test_index_table_name_index" {
				nameIndexFound = true
			}
			if idx == "test_index_table_email_unique" {
				emailIndexFound = true
			}
		}
		if !nameIndexFound {
			t.Error("Index 'test_index_table_name_index' should be in listing")
		}
		if !emailIndexFound {
			t.Error("Index 'test_index_table_email_unique' should be in listing")
		}

		_ = db.Schema().Drop(tableName)
	})

	t.Run("Multi-column Index", func(t *testing.T) {
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
			t.Error("Index 'name_idx' should exist")
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
					t.Error("Index should not be unique")
				}
				if idx.Primary {
					t.Error("Index should not be primary")
				}
				found = true
			}
		}
		if !found {
			t.Error("Index 'name_idx' not found")
		}

		_ = db.Schema().Drop(tableName)
	})

	t.Run("GetIndexes Verification", func(t *testing.T) {
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

		// Primary key
		primaryFound := false
		for _, idx := range indexes {
			if idx.Primary {
				if len(idx.Columns) != 1 || idx.Columns[0] != "id" {
					t.Errorf("Expected primary column 'id', got %v", idx.Columns)
				}
				primaryFound = true
			}
		}
		if !primaryFound {
			t.Error("Primary key not found")
		}

		// Unique index
		uniqueFound := false
		for _, idx := range indexes {
			if idx.Name == "test_get_indexes_name_unique" {
				if len(idx.Columns) != 1 || idx.Columns[0] != "name" {
					t.Errorf("Expected column 'name', got %v", idx.Columns)
				}
				if !idx.Unique {
					t.Error("Index should be unique")
				}
				uniqueFound = true
			}
		}
		if !uniqueFound {
			t.Error("Unique index not found")
		}

		// Normal index
		normalFound := false
		for _, idx := range indexes {
			if idx.Name == "test_get_indexes_email_index" {
				if len(idx.Columns) != 1 || idx.Columns[0] != "email" {
					t.Errorf("Expected column 'email', got %v", idx.Columns)
				}
				if idx.Unique {
					t.Error("Index should not be unique")
				}
				normalFound = true
			}
		}
		if !normalFound {
			t.Error("Normal index not found")
		}

		_ = db.Schema().Drop(tableName)
	})

	t.Run("DropIndex and DropIndexByName", func(t *testing.T) {
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
			t.Error("Index 'test_drop_index_name_index' should exist")
		}
		if !db.Schema().HasIndex(tableName, "email_idx") {
			t.Error("Index 'email_idx' should exist")
		}

		// DropIndex (by column)
		err = db.Schema().Table(tableName, func(table schema.Blueprint) {
			table.DropIndex("name")
		})
		if err != nil {
			t.Fatalf("Failed to drop index: %v", err)
		}
		if db.Schema().HasIndex(tableName, "test_drop_index_name_index") {
			t.Error("Index 'test_drop_index_name_index' should not exist after drop")
		}

		// DropIndexByName
		err = db.Schema().Table(tableName, func(table schema.Blueprint) {
			table.DropIndexByName("email_idx")
		})
		if err != nil {
			t.Fatalf("Failed to drop index by name: %v", err)
		}
		if db.Schema().HasIndex(tableName, "email_idx") {
			t.Error("Index 'email_idx' should not exist after drop by name")
		}

		_ = db.Schema().Drop(tableName)
	})

	t.Run("RenameIndex", func(t *testing.T) {
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
			t.Error("Index 'old_idx' should exist")
		}

		err = db.Schema().Table(tableName, func(table schema.Blueprint) {
			table.RenameIndex("old_idx", "new_idx")
		})
		if err != nil {
			t.Fatalf("Failed to rename index: %v", err)
		}

		if db.Schema().HasIndex(tableName, "old_idx") {
			t.Error("Index 'old_idx' should not exist after rename")
		}
		if !db.Schema().HasIndex(tableName, "new_idx") {
			t.Error("Index 'new_idx' should exist after rename")
		}

		_ = db.Schema().Drop(tableName)
	})
}
