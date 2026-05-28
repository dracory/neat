//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestMySQLSchemaForeignKeyCreateTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	userTable := "fk_users"
	postTable := "fk_posts"
	db.Schema().DropIfExists(postTable)
	db.Schema().DropIfExists(userTable)

	err := db.Schema().Create(userTable, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create user table: %v", err)
	}

	err = db.Schema().Create(postTable, func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.String("title")
		table.Foreign("user_id").References("id").On(userTable).CascadeOnDelete().RestrictOnUpdate()
	})
	if err != nil {
		t.Fatalf("Failed to create post table: %v", err)
	}

	foreignKeys, err := db.Schema().GetForeignKeys(postTable)
	if err != nil {
		t.Fatalf("Failed to get foreign keys: %v", err)
	}
	if len(foreignKeys) != 1 {
		t.Fatalf("Expected 1 foreign key, got %d", len(foreignKeys))
	}

	if len(foreignKeys[0].Columns) != 1 || foreignKeys[0].Columns[0] != "user_id" {
		t.Errorf("Expected columns [user_id], got %v", foreignKeys[0].Columns)
	}
	if foreignKeys[0].ForeignTable != userTable {
		t.Errorf("Expected foreign table %s, got %s", userTable, foreignKeys[0].ForeignTable)
	}
	if len(foreignKeys[0].ForeignColumns) != 1 || foreignKeys[0].ForeignColumns[0] != "id" {
		t.Errorf("Expected foreign columns [id], got %v", foreignKeys[0].ForeignColumns)
	}
	if foreignKeys[0].OnDelete != "cascade" {
		t.Errorf("Expected on_delete 'cascade', got %s", foreignKeys[0].OnDelete)
	}
	if foreignKeys[0].OnUpdate != "restrict" {
		t.Errorf("Expected on_update 'restrict', got %s", foreignKeys[0].OnUpdate)
	}

	db.Schema().Drop(postTable)
	db.Schema().Drop(userTable)
}

func TestMySQLSchemaForeignKeyAddToExistingTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	userTable := "fk_users_alter"
	postTable := "fk_posts_alter"
	db.Schema().DropIfExists(postTable)
	db.Schema().DropIfExists(userTable)

	err := db.Schema().Create(userTable, func(table schema.Blueprint) {
		table.ID()
	})
	if err != nil {
		t.Fatalf("Failed to create user table: %v", err)
	}

	err = db.Schema().Create(postTable, func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
	})
	if err != nil {
		t.Fatalf("Failed to create post table: %v", err)
	}

	err = db.Schema().Table(postTable, func(table schema.Blueprint) {
		table.Foreign("user_id").References("id").On(userTable).CascadeOnDelete().CascadeOnUpdate()
	})
	if err != nil {
		t.Fatalf("Failed to add foreign key: %v", err)
	}

	foreignKeys, err := db.Schema().GetForeignKeys(postTable)
	if err != nil {
		t.Fatalf("Failed to get foreign keys: %v", err)
	}
	if len(foreignKeys) != 1 {
		t.Fatalf("Expected 1 foreign key, got %d", len(foreignKeys))
	}

	if foreignKeys[0].OnDelete != "cascade" {
		t.Errorf("Expected on_delete 'cascade', got %s", foreignKeys[0].OnDelete)
	}
	if foreignKeys[0].OnUpdate != "cascade" {
		t.Errorf("Expected on_update 'cascade', got %s", foreignKeys[0].OnUpdate)
	}

	db.Schema().Drop(postTable)
	db.Schema().Drop(userTable)
}

func TestMySQLSchemaForeignKeyCustomNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	userTable := "fk_users_named"
	postTable := "fk_posts_named"
	db.Schema().DropIfExists(postTable)
	db.Schema().DropIfExists(userTable)

	err := db.Schema().Create(userTable, func(table schema.Blueprint) {
		table.ID()
	})
	if err != nil {
		t.Fatalf("Failed to create user table: %v", err)
	}

	err = db.Schema().Create(postTable, func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.Foreign("user_id").References("id").On(userTable).Name("custom_fk_name")
	})
	if err != nil {
		t.Fatalf("Failed to create post table: %v", err)
	}

	foreignKeys, err := db.Schema().GetForeignKeys(postTable)
	if err != nil {
		t.Fatalf("Failed to get foreign keys: %v", err)
	}
	if len(foreignKeys) != 1 {
		t.Fatalf("Expected 1 foreign key, got %d", len(foreignKeys))
	}

	if foreignKeys[0].Name != "custom_fk_name" {
		t.Errorf("Expected foreign key name 'custom_fk_name', got %s", foreignKeys[0].Name)
	}

	db.Schema().Drop(postTable)
	db.Schema().Drop(userTable)
}

func TestMySQLSchemaForeignKeyDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	userTable := "fk_users_drop"
	postTable := "fk_posts_drop"
	db.Schema().DropIfExists(postTable)
	db.Schema().DropIfExists(userTable)

	err := db.Schema().Create(userTable, func(table schema.Blueprint) {
		table.ID()
	})
	if err != nil {
		t.Fatalf("Failed to create user table: %v", err)
	}

	err = db.Schema().Create(postTable, func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.Foreign("user_id").References("id").On(userTable)
	})
	if err != nil {
		t.Fatalf("Failed to create post table: %v", err)
	}

	if !db.Schema().HasTable(postTable) {
		t.Error("Post table should exist")
	}
	foreignKeys, _ := db.Schema().GetForeignKeys(postTable)
	if len(foreignKeys) != 1 {
		t.Fatalf("Expected 1 foreign key, got %d", len(foreignKeys))
	}

	err = db.Schema().Table(postTable, func(table schema.Blueprint) {
		table.DropForeign("user_id")
	})
	if err != nil {
		t.Fatalf("Failed to drop foreign key: %v", err)
	}

	foreignKeys, err = db.Schema().GetForeignKeys(postTable)
	if err != nil {
		t.Fatalf("Failed to get foreign keys: %v", err)
	}
	if len(foreignKeys) != 0 {
		t.Errorf("Expected 0 foreign keys, got %d", len(foreignKeys))
	}

	err = db.Schema().Table(postTable, func(table schema.Blueprint) {
		table.Foreign("user_id").References("id").On(userTable).Name("another_fk")
	})
	if err != nil {
		t.Fatalf("Failed to add foreign key: %v", err)
	}

	err = db.Schema().Table(postTable, func(table schema.Blueprint) {
		table.DropForeignByName("another_fk")
	})
	if err != nil {
		t.Fatalf("Failed to drop foreign key by name: %v", err)
	}

	foreignKeys, err = db.Schema().GetForeignKeys(postTable)
	if err != nil {
		t.Fatalf("Failed to get foreign keys: %v", err)
	}
	if len(foreignKeys) != 0 {
		t.Errorf("Expected 0 foreign keys, got %d", len(foreignKeys))
	}

	db.Schema().Drop(postTable)
	db.Schema().Drop(userTable)
}
