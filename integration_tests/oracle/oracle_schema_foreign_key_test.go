package oracle_test

import (
	"strings"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestOracleSchemaForeignKeyCreateTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("skipped - Oracle foreign key handling needs investigation (ORA-01735)")

	db := SetupOracleTest(t)

	userTable := "fk_users"
	postTable := "fk_posts"
	_ = db.Schema().Drop(postTable)
	_ = db.Schema().Drop(userTable)
	_ = db.Schema().DropIfExists(postTable)
	_ = db.Schema().DropIfExists(userTable)

	err := db.Schema().Create(userTable, func(table schema.Blueprint) {
		table.Integer("id")
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create user table: %v", err)
	}

	err = db.Schema().Create(postTable, func(table schema.Blueprint) {
		table.ID()
		table.Integer("user_id")
		table.String("title")
	})
	if err != nil {
		t.Fatalf("Failed to create post table: %v", err)
	}

	// Add foreign key separately
	err = db.Schema().Table(postTable, func(table schema.Blueprint) {
		table.Foreign("user_id").References("id").On(userTable).CascadeOnDelete()
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

	t.Logf("Foreign key: %+v", foreignKeys[0])

	if len(foreignKeys[0].Columns) != 1 || foreignKeys[0].Columns[0] != "USER_ID" && foreignKeys[0].Columns[0] != "user_id" {
		t.Errorf("Expected columns [user_id], got %v", foreignKeys[0].Columns)
	}
	if !strings.EqualFold(foreignKeys[0].ForeignTable, userTable) {
		t.Errorf("Expected foreign table %s, got %s", userTable, foreignKeys[0].ForeignTable)
	}
	if len(foreignKeys[0].ForeignColumns) != 1 || foreignKeys[0].ForeignColumns[0] != "ID" && foreignKeys[0].ForeignColumns[0] != "id" {
		t.Errorf("Expected foreign columns [id], got %v", foreignKeys[0].ForeignColumns)
	}
	if foreignKeys[0].OnDelete != "cascade" {
		t.Errorf("Expected on_delete 'cascade', got %s", foreignKeys[0].OnDelete)
	}
	// Oracle doesn't support ON UPDATE, so it will be empty
	if foreignKeys[0].OnUpdate != "" {
		t.Errorf("Expected on_update '', got %s", foreignKeys[0].OnUpdate)
	}

	_ = db.Schema().Drop(postTable)
	_ = db.Schema().Drop(userTable)
}
