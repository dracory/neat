//go:build integration

package sqlite

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestSQLiteSchemaForeignKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	t.Run("Create Table with Foreign Keys", func(t *testing.T) {
		userTable := "fk_users"
		postTable := "fk_posts"
		_ = db.Schema().DropIfExists(postTable)
		_ = db.Schema().DropIfExists(userTable)

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
			t.Errorf("Expected on delete cascade, got %s", foreignKeys[0].OnDelete)
		}
		if foreignKeys[0].OnUpdate != "restrict" {
			t.Errorf("Expected on update restrict, got %s", foreignKeys[0].OnUpdate)
		}

		_ = db.Schema().Drop(postTable)
		_ = db.Schema().Drop(userTable)
	})

	t.Run("Foreign Key with Cascade Options", func(t *testing.T) {
		userTable := "fk_users_cascade"
		postTable := "fk_posts_cascade"
		_ = db.Schema().DropIfExists(postTable)
		_ = db.Schema().DropIfExists(userTable)

		err := db.Schema().Create(userTable, func(table schema.Blueprint) {
			table.ID()
		})
		if err != nil {
			t.Fatalf("Failed to create user table: %v", err)
		}

		err = db.Schema().Create(postTable, func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.Foreign("user_id").References("id").On(userTable).CascadeOnDelete().CascadeOnUpdate()
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

		if foreignKeys[0].OnDelete != "cascade" {
			t.Errorf("Expected on delete cascade, got %s", foreignKeys[0].OnDelete)
		}
		if foreignKeys[0].OnUpdate != "cascade" {
			t.Errorf("Expected on update cascade, got %s", foreignKeys[0].OnUpdate)
		}

		_ = db.Schema().Drop(postTable)
		_ = db.Schema().Drop(userTable)
	})

	t.Run("Custom Foreign Key Names (SQLite Limitation)", func(t *testing.T) {
		userTable := "fk_users_named"
		postTable := "fk_posts_named"
		_ = db.Schema().DropIfExists(postTable)
		_ = db.Schema().DropIfExists(userTable)

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

		// SQLite does not preserve foreign key names in pragma_foreign_key_list
		// Our current processor doesn't seem to set the name for SQLite
		if foreignKeys[0].Name != "" {
			t.Errorf("Expected empty name for SQLite, got %s", foreignKeys[0].Name)
		}

		_ = db.Schema().Drop(postTable)
		_ = db.Schema().Drop(userTable)
	})

	t.Run("SQLite doesn't support dropping foreign keys", func(t *testing.T) {
		userTable := "fk_users_drop"
		postTable := "fk_posts_drop"
		_ = db.Schema().DropIfExists(postTable)
		_ = db.Schema().DropIfExists(userTable)

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

		// SQLite's CompileDropForeign returns an empty string, so this should do nothing and not error
		err = db.Schema().Table(postTable, func(table schema.Blueprint) {
			table.DropForeign("user_id")
		})
		if err != nil {
			t.Fatalf("Drop foreign should not error: %v", err)
		}

		// Verify it's still there
		foreignKeys, err := db.Schema().GetForeignKeys(postTable)
		if err != nil {
			t.Fatalf("Failed to get foreign keys: %v", err)
		}
		if len(foreignKeys) != 1 {
			t.Errorf("Expected 1 foreign key, got %d", len(foreignKeys))
		}

		_ = db.Schema().Drop(postTable)
		_ = db.Schema().Drop(userTable)
	})
}
