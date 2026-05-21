package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlueprint(t *testing.T) {
	prefix := "test_prefix"
	table := "test_table"

	bp := NewBlueprint(nil, prefix, table)

	assert.NotNil(t, bp)
	assert.Equal(t, prefix, bp.prefix)
	assert.Equal(t, table, bp.table)
	assert.Empty(t, bp.columns)
	assert.Empty(t, bp.commands)
}

func TestBlueprint_CreateAndAddColumn(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("creates column with correct type and name", func(t *testing.T) {
		column := bp.createAndAddColumn("string", "name")

		assert.NotNil(t, column)
		assert.Equal(t, "string", column.GetType())
		assert.Equal(t, "name", column.GetName())
		assert.Len(t, bp.columns, 1)
	})

	t.Run("adds Add command when not in Create mode", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		bp2.createAndAddColumn("integer", "age")

		assert.Len(t, bp2.commands, 1)
		assert.Equal(t, "add", bp2.commands[0].Name)
	})

	t.Run("does not add Add command when in Create mode", func(t *testing.T) {
		bp3 := &Blueprint{table: "test_table"}
		bp3.Create()
		bp3.createAndAddColumn("integer", "age")

		assert.Len(t, bp3.commands, 1)
		assert.Equal(t, "create", bp3.commands[0].Name)
	})
}

func TestBlueprint_ColumnMethods(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("String method", func(t *testing.T) {
		column := bp.String("name", 100)

		assert.NotNil(t, column)
		assert.Equal(t, "string", column.GetType())
		assert.Equal(t, "name", column.GetName())
		assert.Equal(t, 100, column.GetLength())
	})

	t.Run("String method with default length", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		column := bp2.String("email")

		assert.NotNil(t, column)
		assert.Equal(t, "string", column.GetType())
		assert.Equal(t, "email", column.GetName())
		assert.Equal(t, 255, column.GetLength()) // DefaultStringLength
	})

	t.Run("Integer method", func(t *testing.T) {
		bp3 := &Blueprint{table: "test_table"}
		column := bp3.Integer("age")

		assert.NotNil(t, column)
		assert.Equal(t, "integer", column.GetType())
		assert.Equal(t, "age", column.GetName())
	})

	t.Run("Text method", func(t *testing.T) {
		bp4 := &Blueprint{table: "test_table"}
		column := bp4.Text("description")

		assert.NotNil(t, column)
		assert.Equal(t, "text", column.GetType())
		assert.Equal(t, "description", column.GetName())
	})

	t.Run("Boolean method", func(t *testing.T) {
		bp5 := &Blueprint{table: "test_table"}
		column := bp5.Boolean("is_active")

		assert.NotNil(t, column)
		assert.Equal(t, "boolean", column.GetType())
		assert.Equal(t, "is_active", column.GetName())
	})
}

func TestBlueprint_ID(t *testing.T) {
	t.Run("ID with default name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.ID()

		assert.NotNil(t, column)
		assert.Equal(t, "id", column.GetName())
		assert.True(t, column.GetAutoIncrement())
	})

	t.Run("ID with custom name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.ID("custom_id")

		assert.NotNil(t, column)
		assert.Equal(t, "custom_id", column.GetName())
		assert.True(t, column.GetAutoIncrement())
	})
}

func TestBlueprint_Timestamps(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	bp.Timestamps()

	assert.Len(t, bp.columns, 2)
	assert.Equal(t, "created_at", bp.columns[0].GetName())
	assert.Equal(t, "updated_at", bp.columns[1].GetName())
	assert.True(t, bp.columns[0].GetNullable())
	assert.True(t, bp.columns[1].GetNullable())
}

func TestBlueprint_SoftDeletes(t *testing.T) {
	t.Run("SoftDeletes with default name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.SoftDeletes()

		assert.NotNil(t, column)
		assert.Equal(t, "deleted_at", column.GetName())
		assert.True(t, column.GetNullable())
	})

	t.Run("SoftDeletes with custom name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.SoftDeletes("removed_at")

		assert.NotNil(t, column)
		assert.Equal(t, "removed_at", column.GetName())
		assert.True(t, column.GetNullable())
	})
}

func TestBlueprint_CommandMethods(t *testing.T) {
	t.Run("Create method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Create()

		assert.Len(t, bp.commands, 1)
		assert.Equal(t, "create", bp.commands[0].Name)
	})

	t.Run("Drop method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Drop()

		assert.Len(t, bp.commands, 1)
		assert.Equal(t, "drop", bp.commands[0].Name)
	})

	t.Run("DropIfExists method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropIfExists()

		assert.Len(t, bp.commands, 1)
		assert.Equal(t, "dropIfExists", bp.commands[0].Name)
	})

	t.Run("DropColumn method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropColumn("name", "email")

		assert.Len(t, bp.commands, 1)
		assert.Equal(t, "dropColumn", bp.commands[0].Name)
		assert.Equal(t, []string{"name", "email"}, bp.commands[0].Columns)
	})

	t.Run("Rename method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Rename("new_table")

		assert.Len(t, bp.commands, 1)
		assert.Equal(t, "rename", bp.commands[0].Name)
		assert.Equal(t, "new_table", bp.commands[0].To)
	})
}

func TestBlueprint_GetterMethods(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	bp.String("name")
	bp.Integer("age")

	t.Run("GetTableName", func(t *testing.T) {
		assert.Equal(t, "test_table", bp.GetTableName())
	})

	t.Run("GetAddedColumns", func(t *testing.T) {
		columns := bp.GetAddedColumns()

		assert.Len(t, columns, 2)
		assert.Equal(t, "name", columns[0].GetName())
		assert.Equal(t, "age", columns[1].GetName())
	})

	t.Run("GetCommands", func(t *testing.T) {
		commands := bp.GetCommands()

		assert.NotNil(t, commands)
		assert.Len(t, commands, 2) // Two add commands
	})
}

func TestBlueprint_HasCommand(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("returns false when command doesn't exist", func(t *testing.T) {
		assert.False(t, bp.HasCommand("create"))
	})

	t.Run("returns true when command exists", func(t *testing.T) {
		bp.Create()
		assert.True(t, bp.HasCommand("create"))
	})

	t.Run("returns false for different command", func(t *testing.T) {
		assert.False(t, bp.HasCommand("drop"))
	})
}

func TestBlueprint_SetTable(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	bp.SetTable("new_table")

	assert.Equal(t, "new_table", bp.GetTableName())
}

func TestBlueprint_Change(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("Change modifies last added column", func(t *testing.T) {
		bp.String("name", 100)
		result := bp.Change()

		assert.NotNil(t, result)
		assert.Len(t, bp.commands, 2) // add + change
		assert.Equal(t, "change", bp.commands[1].Name)
	})

	t.Run("Change returns nil when no columns exist", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		result := bp2.Change()

		assert.Nil(t, result)
	})
}

func TestBlueprint_IndexMethods(t *testing.T) {
	bp := &Blueprint{table: "test_table"}

	t.Run("Index method", func(t *testing.T) {
		indexDef := bp.Index("name", "email")

		assert.NotNil(t, indexDef)
		assert.Len(t, bp.commands, 1)
		assert.Equal(t, "index", bp.commands[0].Name)
		assert.Equal(t, []string{"name", "email"}, bp.commands[0].Columns)
	})

	t.Run("Unique method", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		uniqueDef := bp2.Unique("email")

		assert.NotNil(t, uniqueDef)
		assert.Len(t, bp2.commands, 1)
		assert.Equal(t, "unique", bp2.commands[0].Name)
	})

	t.Run("Primary method", func(t *testing.T) {
		bp3 := &Blueprint{table: "test_table"}
		bp3.Primary("id")

		assert.Len(t, bp3.commands, 1)
		assert.Equal(t, "primary", bp3.commands[0].Name)
	})
}

func TestBlueprint_isCreate(t *testing.T) {
	t.Run("returns false when no Create command", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		assert.False(t, bp.isCreate())
	})

	t.Run("returns true when Create command exists", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Create()
		assert.True(t, bp.isCreate())
	})

	t.Run("returns false when only Drop command exists", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Drop()
		assert.False(t, bp.isCreate())
	})
}

func TestBlueprint_createIndexName(t *testing.T) {
	t.Run("creates index name without prefix", func(t *testing.T) {
		bp := &Blueprint{table: "users", prefix: ""}
		indexName := bp.createIndexName("index", []string{"name", "email"})

		assert.Equal(t, "users_name_email_index", indexName)
	})

	t.Run("creates index name with prefix", func(t *testing.T) {
		bp := &Blueprint{table: "users", prefix: "test_"}
		indexName := bp.createIndexName("index", []string{"name"})

		assert.Equal(t, "test_users_name_index", indexName)
	})

	t.Run("handles table with schema", func(t *testing.T) {
		bp := &Blueprint{table: "myschema.users", prefix: "test_"}
		indexName := bp.createIndexName("index", []string{"id"})

		assert.Equal(t, "myschema_test_users_id_index", indexName)
	})

	t.Run("replaces special characters", func(t *testing.T) {
		bp := &Blueprint{table: "my-table", prefix: ""}
		indexName := bp.createIndexName("index", []string{"name"})

		assert.Equal(t, "my_table_name_index", indexName)
	})
}

func TestBlueprint_Enum(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	allowed := []any{"active", "inactive", "pending"}
	column := bp.Enum("status", allowed)

	assert.NotNil(t, column)
	assert.Equal(t, "enum", column.GetType())
	assert.Equal(t, "status", column.GetName())
	assert.Equal(t, allowed, column.GetAllowed())
}

func TestBlueprint_Float(t *testing.T) {
	t.Run("Float with default precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Float("price")

		assert.NotNil(t, column)
		assert.Equal(t, "float", column.GetType())
		assert.Equal(t, "price", column.GetName())
		assert.Equal(t, 53, column.GetPrecision())
	})

	t.Run("Float with custom precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Float("price", 32)

		assert.NotNil(t, column)
		assert.Equal(t, 32, column.GetPrecision())
	})
}

func TestBlueprint_DropTimestamps(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	bp.DropTimestamps()

	assert.Len(t, bp.commands, 1)
	assert.Equal(t, "dropColumn", bp.commands[0].Name)
	assert.Equal(t, []string{"created_at", "updated_at"}, bp.commands[0].Columns)
}

func TestBlueprint_RenameIndex(t *testing.T) {
	bp := &Blueprint{table: "test_table"}
	bp.RenameIndex("old_index", "new_index")

	assert.Len(t, bp.commands, 1)
	assert.Equal(t, "renameIndex", bp.commands[0].Name)
	assert.Equal(t, "old_index", bp.commands[0].From)
	assert.Equal(t, "new_index", bp.commands[0].To)
}
