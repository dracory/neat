package schema

import (
	"testing"
)

func TestNewBlueprint(t *testing.T) {
	prefix := "test_prefix"
	table := "test_table"

	bp := NewBlueprint(nil, prefix, table)

	if bp == nil {
		t.Fatal("expected non-nil blueprint")
	}
	if bp.prefix != prefix {
		t.Errorf("expected %s, got %s", prefix, bp.prefix)
	}
	if bp.table != table {
		t.Errorf("expected %s, got %s", table, bp.table)
	}
	if len(bp.columns) != 0 {
		t.Errorf("expected empty columns, got %v", bp.columns)
	}
	if len(bp.commands) != 0 {
		t.Errorf("expected empty commands, got %v", bp.commands)
	}
}

func TestBlueprint_CreateAndAddColumn(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("creates column with correct type and name", func(t *testing.T) {
		column := bp.createAndAddColumn("string", "name")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "string" {
			t.Errorf("expected string, got %s", column.GetType())
		}
		if column.GetName() != "name" {
			t.Errorf("expected name, got %s", column.GetName())
		}
		if len(bp.columns) != 1 {
			t.Errorf("expected 1 column, got %d", len(bp.columns))
		}
	})

	t.Run("adds Add command when not in Create mode", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		bp2.createAndAddColumn("integer", "age")

		if len(bp2.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp2.commands))
		}
		if bp2.commands[0].Name != "add" {
			t.Errorf("expected add command, got %s", bp2.commands[0].Name)
		}
	})

	t.Run("does not add Add command when in Create mode", func(t *testing.T) {
		bp3 := &Blueprint{table: "test_table"}
		bp3.Create()
		bp3.createAndAddColumn("integer", "age")

		if len(bp3.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp3.commands))
		}
		if bp3.commands[0].Name != "create" {
			t.Errorf("expected create command, got %s", bp3.commands[0].Name)
		}
	})
}

func TestBlueprint_ColumnMethods(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("String method", func(t *testing.T) {
		column := bp.String("name", 100)

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "string" {
			t.Errorf("expected string, got %s", column.GetType())
		}
		if column.GetName() != "name" {
			t.Errorf("expected name, got %s", column.GetName())
		}
		if column.GetLength() != 100 {
			t.Errorf("expected 100, got %d", column.GetLength())
		}
	})

	t.Run("String method with default length", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		column := bp2.String("email")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "string" {
			t.Errorf("expected string, got %s", column.GetType())
		}
		if column.GetName() != "email" {
			t.Errorf("expected email, got %s", column.GetName())
		}
		if column.GetLength() != 255 {
			t.Errorf("expected 255, got %d", column.GetLength())
		}
	})

	t.Run("Integer method", func(t *testing.T) {
		bp3 := &Blueprint{table: "test_table"}
		column := bp3.Integer("age")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "integer" {
			t.Errorf("expected integer, got %s", column.GetType())
		}
		if column.GetName() != "age" {
			t.Errorf("expected age, got %s", column.GetName())
		}
	})

	t.Run("Text method", func(t *testing.T) {
		bp4 := &Blueprint{table: "test_table"}
		column := bp4.Text("description")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "text" {
			t.Errorf("expected text, got %s", column.GetType())
		}
		if column.GetName() != "description" {
			t.Errorf("expected description, got %s", column.GetName())
		}
	})

	t.Run("Boolean method", func(t *testing.T) {
		bp5 := &Blueprint{table: "test_table"}
		column := bp5.Boolean("is_active")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "boolean" {
			t.Errorf("expected boolean, got %s", column.GetType())
		}
		if column.GetName() != "is_active" {
			t.Errorf("expected is_active, got %s", column.GetName())
		}
	})
}

func TestBlueprint_ID(t *testing.T) {
	t.Run("ID with default name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.ID()

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetName() != "id" {
			t.Errorf("expected id, got %s", column.GetName())
		}
		if !column.GetAutoIncrement() {
			t.Error("expected auto increment to be true")
		}
	})

	t.Run("ID with custom name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.ID("custom_id")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetName() != "custom_id" {
			t.Errorf("expected custom_id, got %s", column.GetName())
		}
		if !column.GetAutoIncrement() {
			t.Error("expected auto increment to be true")
		}
	})
}

func TestBlueprint_Timestamps(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	bp.Timestamps()

	if len(bp.columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(bp.columns))
	}
	if bp.columns[0].GetName() != "created_at" {
		t.Errorf("expected created_at, got %s", bp.columns[0].GetName())
	}
	if bp.columns[1].GetName() != "updated_at" {
		t.Errorf("expected updated_at, got %s", bp.columns[1].GetName())
	}
	if !bp.columns[0].GetNullable() {
		t.Error("expected created_at to be nullable")
	}
	if !bp.columns[1].GetNullable() {
		t.Error("expected updated_at to be nullable")
	}
}

func TestBlueprint_SoftDeletes(t *testing.T) {
	t.Run("SoftDeletes with default name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.SoftDeletes()

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetName() != "deleted_at" {
			t.Errorf("expected deleted_at, got %s", column.GetName())
		}
		if !column.GetNullable() {
			t.Error("expected deleted_at to be nullable")
		}
	})

	t.Run("SoftDeletes with custom name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.SoftDeletes("removed_at")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetName() != "removed_at" {
			t.Errorf("expected removed_at, got %s", column.GetName())
		}
		if !column.GetNullable() {
			t.Error("expected removed_at to be nullable")
		}
	})
}

func TestBlueprint_CommandMethods(t *testing.T) {
	t.Run("Create method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Create()

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "create" {
			t.Errorf("expected create command, got %s", bp.commands[0].Name)
		}
	})

	t.Run("Drop method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Drop()

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "drop" {
			t.Errorf("expected drop command, got %s", bp.commands[0].Name)
		}
	})

	t.Run("DropIfExists method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropIfExists()

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "dropIfExists" {
			t.Errorf("expected dropIfExists command, got %s", bp.commands[0].Name)
		}
	})

	t.Run("DropColumn method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropColumn("name", "email")

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "dropColumn" {
			t.Errorf("expected dropColumn command, got %s", bp.commands[0].Name)
		}
		if len(bp.commands[0].Columns) != 2 || bp.commands[0].Columns[0] != "name" || bp.commands[0].Columns[1] != "email" {
			t.Errorf("expected [name, email], got %v", bp.commands[0].Columns)
		}
	})

	t.Run("Rename method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Rename("new_table")

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "rename" {
			t.Errorf("expected rename command, got %s", bp.commands[0].Name)
		}
		if bp.commands[0].To != "new_table" {
			t.Errorf("expected new_table, got %s", bp.commands[0].To)
		}
	})
}

func TestBlueprint_GetterMethods(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	bp.String("name")
	bp.Integer("age")

	t.Run("GetTableName", func(t *testing.T) {
		if bp.GetTableName() != "test_table" {
			t.Errorf("expected test_table, got %s", bp.GetTableName())
		}
	})

	t.Run("GetAddedColumns", func(t *testing.T) {
		columns := bp.GetAddedColumns()

		if len(columns) != 2 {
			t.Errorf("expected 2 columns, got %d", len(columns))
		}
		if columns[0].GetName() != "name" {
			t.Errorf("expected name, got %s", columns[0].GetName())
		}
		if columns[1].GetName() != "age" {
			t.Errorf("expected age, got %s", columns[1].GetName())
		}
	})

	t.Run("GetCommands", func(t *testing.T) {
		commands := bp.GetCommands()

		if commands == nil {
			t.Fatal("expected non-nil commands")
		}
		if len(commands) != 2 {
			t.Errorf("expected 2 commands, got %d", len(commands))
		}
	})
}

func TestBlueprint_HasCommand(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("returns false when command doesn't exist", func(t *testing.T) {
		if bp.HasCommand("create") {
			t.Error("expected false for non-existent command")
		}
	})

	t.Run("returns true when command exists", func(t *testing.T) {
		bp.Create()
		if !bp.HasCommand("create") {
			t.Error("expected true for existing command")
		}
	})

	t.Run("returns false for different command", func(t *testing.T) {
		if bp.HasCommand("drop") {
			t.Error("expected false for different command")
		}
	})
}

func TestBlueprint_SetTable(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	bp.SetTable("new_table")

	if bp.GetTableName() != "new_table" {
		t.Errorf("expected new_table, got %s", bp.GetTableName())
	}
}

func TestBlueprint_Change(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("Change modifies last added column", func(t *testing.T) {
		bp.String("name", 100)
		result := bp.Change()

		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if len(bp.commands) != 2 {
			t.Errorf("expected 2 commands, got %d", len(bp.commands))
		}
		if bp.commands[1].Name != "change" {
			t.Errorf("expected change command, got %s", bp.commands[1].Name)
		}
	})

	t.Run("Change returns blueprint for chaining", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		result := bp2.Change()

		if result == nil {
			t.Fatal("expected non-nil result for chaining")
		}
	})
}

func TestBlueprint_IndexMethods(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("Index method", func(t *testing.T) {
		indexDef := bp.Index("name", "email")

		if indexDef == nil {
			t.Fatal("expected non-nil indexDef")
		}
		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "index" {
			t.Errorf("expected index command, got %s", bp.commands[0].Name)
		}
		if len(bp.commands[0].Columns) != 2 || bp.commands[0].Columns[0] != "name" || bp.commands[0].Columns[1] != "email" {
			t.Errorf("expected [name, email], got %v", bp.commands[0].Columns)
		}
	})

	t.Run("Unique method", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		uniqueDef := bp2.Unique("email")

		if uniqueDef == nil {
			t.Fatal("expected non-nil uniqueDef")
		}
		if len(bp2.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp2.commands))
		}
		if bp2.commands[0].Name != "unique" {
			t.Errorf("expected unique command, got %s", bp2.commands[0].Name)
		}
	})

	t.Run("Primary method", func(t *testing.T) {
		bp3 := &Blueprint{table: "test_table"}
		bp3.Primary("id")

		if len(bp3.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp3.commands))
		}
		if bp3.commands[0].Name != "primary" {
			t.Errorf("expected primary command, got %s", bp3.commands[0].Name)
		}
	})
}

func TestBlueprint_isCreate(t *testing.T) {
	t.Run("returns false when no Create command", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		if bp.isCreate() {
			t.Error("expected false when no Create command")
		}
	})

	t.Run("returns true when Create command exists", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Create()
		if !bp.isCreate() {
			t.Error("expected true when Create command exists")
		}
	})

	t.Run("returns false when only Drop command exists", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Drop()
		if bp.isCreate() {
			t.Error("expected false when only Drop command exists")
		}
	})
}

func TestBlueprint_createIndexName(t *testing.T) {
	t.Run("creates index name without prefix", func(t *testing.T) {
		bp := &Blueprint{table: "users", prefix: ""}
		indexName := bp.createIndexName("index", []string{"name", "email"})

		if indexName != "users_name_email_index" {
			t.Errorf("expected users_name_email_index, got %s", indexName)
		}
	})

	t.Run("creates index name with prefix", func(t *testing.T) {
		bp := &Blueprint{table: "users", prefix: "test_"}
		indexName := bp.createIndexName("index", []string{"name"})

		if indexName != "test_users_name_index" {
			t.Errorf("expected test_users_name_index, got %s", indexName)
		}
	})

	t.Run("handles table with schema", func(t *testing.T) {
		bp := &Blueprint{table: "myschema.users", prefix: "test_"}
		indexName := bp.createIndexName("index", []string{"id"})

		if indexName != "myschema_test_users_id_index" {
			t.Errorf("expected myschema_test_users_id_index, got %s", indexName)
		}
	})

	t.Run("replaces special characters", func(t *testing.T) {
		bp := &Blueprint{table: "my-table", prefix: ""}
		indexName := bp.createIndexName("index", []string{"name"})

		if indexName != "my_table_name_index" {
			t.Errorf("expected my_table_name_index, got %s", indexName)
		}
	})
}

func TestBlueprint_Enum(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	allowed := []any{"active", "inactive", "pending"}
	column := bp.Enum("status", allowed)

	if column == nil {
		t.Fatal("expected non-nil column")
	}
	if column.GetType() != "enum" {
		t.Errorf("expected enum, got %s", column.GetType())
	}
	if column.GetName() != "status" {
		t.Errorf("expected status, got %s", column.GetName())
	}
	if len(column.GetAllowed()) != 3 {
		t.Errorf("expected 3 allowed values, got %d", len(column.GetAllowed()))
	}
}

func TestBlueprint_Float(t *testing.T) {
	t.Run("Float with default precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Float("price")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "float" {
			t.Errorf("expected float, got %s", column.GetType())
		}
		if column.GetName() != "price" {
			t.Errorf("expected price, got %s", column.GetName())
		}
		if column.GetPrecision() != 53 {
			t.Errorf("expected 53, got %d", column.GetPrecision())
		}
	})

	t.Run("Float with custom precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Float("price", 32)

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetPrecision() != 32 {
			t.Errorf("expected 32, got %d", column.GetPrecision())
		}
	})
}

func TestBlueprint_DropTimestamps(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	bp.DropTimestamps()

	if len(bp.commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(bp.commands))
	}
	if bp.commands[0].Name != "dropColumn" {
		t.Errorf("expected dropColumn command, got %s", bp.commands[0].Name)
	}
	if len(bp.commands[0].Columns) != 2 || bp.commands[0].Columns[0] != "created_at" || bp.commands[0].Columns[1] != "updated_at" {
		t.Errorf("expected [created_at, updated_at], got %v", bp.commands[0].Columns)
	}
}

func TestBlueprint_RenameIndex(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	bp.RenameIndex("old_index", "new_index")

	if len(bp.commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(bp.commands))
	}
	if bp.commands[0].Name != "renameIndex" {
		t.Errorf("expected renameIndex command, got %s", bp.commands[0].Name)
	}
	if bp.commands[0].From != "old_index" {
		t.Errorf("expected old_index, got %s", bp.commands[0].From)
	}
	if bp.commands[0].To != "new_index" {
		t.Errorf("expected new_index, got %s", bp.commands[0].To)
	}
}
