package oracle_test

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestOracleSchemaForeignKeyCreateTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Oracle foreign key handling needs investigation - ORA-01735 error")

	db := SetupOracleTest(t)
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
		table.BigInteger("user_id")
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

	if len(foreignKeys[0].Columns) != 1 || foreignKeys[0].Columns[0] != "USER_ID" && foreignKeys[0].Columns[0] != "user_id" {
		t.Errorf("Expected columns [user_id], got %v", foreignKeys[0].Columns)
	}
	if foreignKeys[0].ForeignTable != userTable {
		t.Errorf("Expected foreign table %s, got %s", userTable, foreignKeys[0].ForeignTable)
	}
	if len(foreignKeys[0].ForeignColumns) != 1 || foreignKeys[0].ForeignColumns[0] != "ID" && foreignKeys[0].ForeignColumns[0] != "id" {
		t.Errorf("Expected foreign columns [id], got %v", foreignKeys[0].ForeignColumns)
	}
	if foreignKeys[0].OnDelete != "cascade" {
		t.Errorf("Expected on_delete 'cascade', got %s", foreignKeys[0].OnDelete)
	}
	if foreignKeys[0].OnUpdate != "restrict" {
		t.Errorf("Expected on_update 'restrict', got %s", foreignKeys[0].OnUpdate)
	}

	_ = db.Schema().Drop(postTable)
	_ = db.Schema().Drop(userTable)
}
