package schema

import (
	"errors"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

// MockGrammar is a test double for the Grammar interface
type MockGrammar struct {
	compileAddFunc           func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileChangeFunc        func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileCommentFunc       func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileCreateFunc        func(blueprint schema.Blueprint) (string, error)
	compileDropFunc          func(blueprint schema.Blueprint) (string, error)
	compileDropColumnFunc    func(blueprint schema.Blueprint, command *schema.Command) ([]string, error)
	compileDropForeignFunc   func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileDropFullTextFunc  func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileDropIfExistsFunc  func(blueprint schema.Blueprint) (string, error)
	compileDropIndexFunc     func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileDropPrimaryFunc   func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileDropUniqueFunc    func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileForeignFunc       func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileFullTextFunc      func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileIndexFunc         func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compilePrimaryFunc       func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileRenameFunc        func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileRenameColumnFunc  func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	compileRenameIndexFunc   func(schema schema.Schema, blueprint schema.Blueprint, command *schema.Command) ([]string, error)
	compileUniqueFunc        func(blueprint schema.Blueprint, command *schema.Command) (string, error)
	getAttributeCommandsFunc func() []string
}

func (m *MockGrammar) CompileAdd(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileAddFunc != nil {
		return m.compileAddFunc(blueprint, command)
	}
	return "ADD COLUMN " + command.Column.GetName(), nil
}

func (m *MockGrammar) CompileChange(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileChangeFunc != nil {
		return m.compileChangeFunc(blueprint, command)
	}
	return "CHANGE COLUMN " + command.Column.GetName(), nil
}

func (m *MockGrammar) CompileColumns(schema, table string) string {
	return "SELECT columns"
}

func (m *MockGrammar) CompileComment(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileCommentFunc != nil {
		return m.compileCommentFunc(blueprint, command)
	}
	return "COMMENT ON COLUMN " + command.Column.GetName(), nil
}

func (m *MockGrammar) CompileCreate(blueprint schema.Blueprint) (string, error) {
	if m.compileCreateFunc != nil {
		return m.compileCreateFunc(blueprint)
	}
	return "CREATE TABLE " + blueprint.GetTableName(), nil
}

func (m *MockGrammar) CompileDrop(blueprint schema.Blueprint) (string, error) {
	if m.compileDropFunc != nil {
		return m.compileDropFunc(blueprint)
	}
	return "DROP TABLE " + blueprint.GetTableName(), nil
}

func (m *MockGrammar) CompileDropAllDomains(domains []string) (string, error) {
	return "DROP DOMAINS", nil
}

func (m *MockGrammar) CompileDropAllTables(tables []string) (string, error) {
	return "DROP ALL TABLES", nil
}

func (m *MockGrammar) CompileDropAllTypes(types []string) (string, error) {
	return "DROP ALL TYPES", nil
}

func (m *MockGrammar) CompileDropAllViews(views []string) (string, error) {
	return "DROP ALL VIEWS", nil
}

func (m *MockGrammar) CompileDropColumn(blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
	if m.compileDropColumnFunc != nil {
		return m.compileDropColumnFunc(blueprint, command)
	}
	return []string{"DROP COLUMN " + command.Columns[0]}, nil
}

func (m *MockGrammar) CompileDropForeign(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileDropForeignFunc != nil {
		return m.compileDropForeignFunc(blueprint, command)
	}
	return "DROP FOREIGN KEY", nil
}

func (m *MockGrammar) CompileDropFullText(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileDropFullTextFunc != nil {
		return m.compileDropFullTextFunc(blueprint, command)
	}
	return "DROP FULLTEXT INDEX", nil
}

func (m *MockGrammar) CompileDropIfExists(blueprint schema.Blueprint) (string, error) {
	if m.compileDropIfExistsFunc != nil {
		return m.compileDropIfExistsFunc(blueprint)
	}
	return "DROP TABLE IF EXISTS " + blueprint.GetTableName(), nil
}

func (m *MockGrammar) CompileDropIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileDropIndexFunc != nil {
		return m.compileDropIndexFunc(blueprint, command)
	}
	return "DROP INDEX " + command.Index, nil
}

func (m *MockGrammar) CompileDropPrimary(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileDropPrimaryFunc != nil {
		return m.compileDropPrimaryFunc(blueprint, command)
	}
	return "DROP PRIMARY KEY", nil
}

func (m *MockGrammar) CompileDropUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileDropUniqueFunc != nil {
		return m.compileDropUniqueFunc(blueprint, command)
	}
	return "DROP UNIQUE INDEX", nil
}

func (m *MockGrammar) CompileForeign(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileForeignFunc != nil {
		return m.compileForeignFunc(blueprint, command)
	}
	return "ADD FOREIGN KEY", nil
}

func (m *MockGrammar) CompileForeignKeys(schema, table string) string {
	return "SELECT foreign keys"
}

func (m *MockGrammar) CompileFullText(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileFullTextFunc != nil {
		return m.compileFullTextFunc(blueprint, command)
	}
	return "ADD FULLTEXT INDEX", nil
}

func (m *MockGrammar) CompileIndex(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileIndexFunc != nil {
		return m.compileIndexFunc(blueprint, command)
	}
	return "CREATE INDEX " + command.Index, nil
}

func (m *MockGrammar) CompileIndexes(schema, table string) string {
	return "SELECT indexes"
}

func (m *MockGrammar) CompilePrimary(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compilePrimaryFunc != nil {
		return m.compilePrimaryFunc(blueprint, command)
	}
	return "ADD PRIMARY KEY", nil
}

func (m *MockGrammar) CompileRename(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileRenameFunc != nil {
		return m.compileRenameFunc(blueprint, command)
	}
	return "RENAME TABLE TO " + command.To, nil
}

func (m *MockGrammar) CompileRenameColumn(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileRenameColumnFunc != nil {
		return m.compileRenameColumnFunc(blueprint, command)
	}
	return "RENAME COLUMN " + command.From + " TO " + command.To, nil
}

func (m *MockGrammar) CompileRenameIndex(schema schema.Schema, blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
	if m.compileRenameIndexFunc != nil {
		return m.compileRenameIndexFunc(schema, blueprint, command)
	}
	return []string{"RENAME INDEX " + command.From + " TO " + command.To}, nil
}

func (m *MockGrammar) CompileTables(database string) string {
	return "SELECT tables"
}

func (m *MockGrammar) CompileTypes() string {
	return "SELECT types"
}

func (m *MockGrammar) CompileUnique(blueprint schema.Blueprint, command *schema.Command) (string, error) {
	if m.compileUniqueFunc != nil {
		return m.compileUniqueFunc(blueprint, command)
	}
	return "ADD UNIQUE INDEX " + command.Index, nil
}

func (m *MockGrammar) CompileViews(database string) string {
	return "SELECT views"
}

func (m *MockGrammar) GetAttributeCommands() []string {
	if m.getAttributeCommandsFunc != nil {
		return m.getAttributeCommandsFunc()
	}
	return []string{}
}

func (m *MockGrammar) TypeBigInteger(column schema.ColumnDefinition) string    { return "BIGINT" }
func (m *MockGrammar) TypeBoolean(column schema.ColumnDefinition) string       { return "BOOLEAN" }
func (m *MockGrammar) TypeChar(column schema.ColumnDefinition) string          { return "CHAR" }
func (m *MockGrammar) TypeDate(column schema.ColumnDefinition) string          { return "DATE" }
func (m *MockGrammar) TypeDateTime(column schema.ColumnDefinition) string      { return "DATETIME" }
func (m *MockGrammar) TypeDateTimeTz(column schema.ColumnDefinition) string    { return "DATETIMETZ" }
func (m *MockGrammar) TypeDecimal(column schema.ColumnDefinition) string       { return "DECIMAL" }
func (m *MockGrammar) TypeDouble(column schema.ColumnDefinition) string        { return "DOUBLE" }
func (m *MockGrammar) TypeEnum(column schema.ColumnDefinition) string          { return "ENUM" }
func (m *MockGrammar) TypeFloat(column schema.ColumnDefinition) string         { return "FLOAT" }
func (m *MockGrammar) TypeInteger(column schema.ColumnDefinition) string       { return "INTEGER" }
func (m *MockGrammar) TypeJson(column schema.ColumnDefinition) string          { return "JSON" }
func (m *MockGrammar) TypeJsonb(column schema.ColumnDefinition) string         { return "JSONB" }
func (m *MockGrammar) TypeLongText(column schema.ColumnDefinition) string      { return "LONGTEXT" }
func (m *MockGrammar) TypeMediumInteger(column schema.ColumnDefinition) string { return "MEDIUMINT" }
func (m *MockGrammar) TypeMediumText(column schema.ColumnDefinition) string    { return "MEDIUMTEXT" }
func (m *MockGrammar) TypeSmallInteger(column schema.ColumnDefinition) string  { return "SMALLINT" }
func (m *MockGrammar) TypeString(column schema.ColumnDefinition) string        { return "VARCHAR" }
func (m *MockGrammar) TypeText(column schema.ColumnDefinition) string          { return "TEXT" }
func (m *MockGrammar) TypeTime(column schema.ColumnDefinition) string          { return "TIME" }
func (m *MockGrammar) TypeTimeTz(column schema.ColumnDefinition) string        { return "TIMETZ" }
func (m *MockGrammar) TypeTimestamp(column schema.ColumnDefinition) string     { return "TIMESTAMP" }
func (m *MockGrammar) TypeTimestampTz(column schema.ColumnDefinition) string   { return "TIMESTAMPTZ" }
func (m *MockGrammar) TypeTinyInteger(column schema.ColumnDefinition) string   { return "TINYINT" }
func (m *MockGrammar) TypeTinyText(column schema.ColumnDefinition) string      { return "TINYTEXT" }

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
	t.Run("Change modifies last added column", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.String("name", 100)
		result := bp.Change()

		if result == nil {
			t.Fatal("expected non-nil result")
		}
		// After Change(), the ADD command is removed and replaced with CHANGE
		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command (change), got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "change" {
			t.Errorf("expected change command, got %s", bp.commands[0].Name)
		}
	})

	t.Run("Change returns nil when no columns exist", func(t *testing.T) {
		bp2 := &Blueprint{table: "test_table"}
		result := bp2.Change()

		if result != nil {
			t.Fatal("expected nil when no columns exist")
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

func TestBlueprint_SkipTransaction(t *testing.T) {
	t.Run("ShouldSkipTransaction returns false by default", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		if bp.ShouldSkipTransaction() {
			t.Error("expected ShouldSkipTransaction to be false by default")
		}
	})

	t.Run("SkipTransaction sets skipTransaction to true", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.SkipTransaction()
		if !bp.ShouldSkipTransaction() {
			t.Error("expected ShouldSkipTransaction to be true after SkipTransaction")
		}
	})

	t.Run("RenameIndex calls SkipTransaction automatically", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.RenameIndex("old", "new")
		if !bp.ShouldSkipTransaction() {
			t.Error("expected ShouldSkipTransaction to be true after RenameIndex")
		}
	})
}

func TestBlueprint_hasChangeCommand(t *testing.T) {
	t.Run("returns false when no CHANGE command exists for column", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		if bp.hasChangeCommand("name") {
			t.Error("expected false when no change command exists")
		}
	})

	t.Run("returns true when CHANGE command exists for column", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.String("name")
		bp.Change()
		if !bp.hasChangeCommand("name") {
			t.Error("expected true when change command exists")
		}
	})

	t.Run("returns false for different column name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.String("name")
		bp.Change()
		if bp.hasChangeCommand("other") {
			t.Error("expected false for different column name")
		}
	})
}

func TestBlueprint_modifyColumn(t *testing.T) {
	t.Run("returns nil when no columns exist", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		result := bp.modifyColumn()
		if result != nil {
			t.Error("expected nil when no columns exist")
		}
	})

	t.Run("marks last column as changed", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.String("name")
		result := bp.modifyColumn()

		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if !result.GetChange() {
			t.Error("expected column to be marked as changed")
		}
	})

	t.Run("removes ADD command for the changed column", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.String("name")

		// Before modifyColumn, there should be an ADD command
		hasAddBefore := false
		for _, cmd := range bp.commands {
			if cmd.Name == "add" && cmd.Column.GetName() == "name" {
				hasAddBefore = true
				break
			}
		}
		if !hasAddBefore {
			t.Error("expected ADD command to exist before modifyColumn")
		}

		bp.Change()

		// After modifyColumn, there should be a CHANGE command instead
		hasChange := false
		for _, cmd := range bp.commands {
			if cmd.Name == "change" && cmd.Column.GetName() == "name" {
				hasChange = true
				break
			}
		}
		if !hasChange {
			t.Error("expected CHANGE command to exist after modifyColumn")
		}
	})
}

func TestBlueprint_addAttributeCommands(t *testing.T) {
	t.Run("adds comment commands for columns with comments", func(t *testing.T) {
		grammar := &MockGrammar{
			getAttributeCommandsFunc: func() []string {
				return []string{"comment"}
			},
		}
		bp := &Blueprint{table: "test_table"}
		bp.String("name").Comment("User name")
		bp.addAttributeCommands(grammar)

		found := false
		for _, cmd := range bp.commands {
			if cmd.Name == "comment" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected comment command to be added")
		}
	})

	t.Run("does not add comment commands when no columns have comments", func(t *testing.T) {
		grammar := &MockGrammar{
			getAttributeCommandsFunc: func() []string {
				return []string{"comment"}
			},
		}
		bp := &Blueprint{table: "test_table"}
		bp.String("name")
		bp.addAttributeCommands(grammar)

		for _, cmd := range bp.commands {
			if cmd.Name == "comment" {
				t.Error("expected no comment command when no column has comment")
			}
		}
	})

	t.Run("handles multiple attribute commands", func(t *testing.T) {
		grammar := &MockGrammar{
			getAttributeCommandsFunc: func() []string {
				return []string{"comment", "virtual"}
			},
		}
		bp := &Blueprint{table: "test_table"}
		bp.String("name").Comment("User name")
		bp.addAttributeCommands(grammar)

		// Only comment should be added since only comment is supported
		found := false
		for _, cmd := range bp.commands {
			if cmd.Name == "comment" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected comment command to be added")
		}
	})
}

func TestBlueprint_addImpliedCommands(t *testing.T) {
	t.Run("calls addAttributeCommands", func(t *testing.T) {
		grammar := &MockGrammar{
			getAttributeCommandsFunc: func() []string {
				return []string{"comment"}
			},
		}
		bp := &Blueprint{table: "test_table"}
		bp.String("name").Comment("User name")
		bp.addImpliedCommands(grammar)

		found := false
		for _, cmd := range bp.commands {
			if cmd.Name == "comment" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected comment command to be added via addImpliedCommands")
		}
	})
}

func TestBlueprint_ToSql(t *testing.T) {
	t.Run("compiles CREATE command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Create()
		bp.String("name")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d", len(statements))
		}
		if statements[0] != "CREATE TABLE test_table" {
			t.Errorf("expected 'CREATE TABLE test_table', got '%s'", statements[0])
		}
	})

	t.Run("compiles DROP command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Drop()

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d", len(statements))
		}
		if statements[0] != "DROP TABLE test_table" {
			t.Errorf("expected 'DROP TABLE test_table', got '%s'", statements[0])
		}
	})

	t.Run("compiles DROP IF EXISTS command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropIfExists()

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d", len(statements))
		}
		if statements[0] != "DROP TABLE IF EXISTS test_table" {
			t.Errorf("expected 'DROP TABLE IF EXISTS test_table', got '%s'", statements[0])
		}
	})

	t.Run("compiles ADD command for columns", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.String("name")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d", len(statements))
		}
		if statements[0] != "ADD COLUMN name" {
			t.Errorf("expected 'ADD COLUMN name', got '%s'", statements[0])
		}
	})

	t.Run("skips ADD command for columns marked for change", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.String("name").Change()

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have CHANGE command, not ADD
		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d: %v", len(statements), statements)
		}
		if statements[0] != "CHANGE COLUMN name" {
			t.Errorf("expected 'CHANGE COLUMN name', got '%s'", statements[0])
		}
	})

	t.Run("compiles CHANGE command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		col := bp.String("name")
		col.Change()

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "CHANGE COLUMN name" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected CHANGE COLUMN name in statements, got %v", statements)
		}
	})

	t.Run("compiles DROP COLUMN command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropColumn("name")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d", len(statements))
		}
		if statements[0] != "DROP COLUMN name" {
			t.Errorf("expected 'DROP COLUMN name', got '%s'", statements[0])
		}
	})

	t.Run("compiles RENAME command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Rename("new_table")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d", len(statements))
		}
		if statements[0] != "RENAME TABLE TO new_table" {
			t.Errorf("expected 'RENAME TABLE TO new_table', got '%s'", statements[0])
		}
	})

	t.Run("compiles RENAME COLUMN command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.RenameColumn("old_name", "new_name")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d", len(statements))
		}
		if statements[0] != "RENAME COLUMN old_name TO new_name" {
			t.Errorf("expected 'RENAME COLUMN old_name TO new_name', got '%s'", statements[0])
		}
	})

	t.Run("compiles RENAME INDEX command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.RenameIndex("old_idx", "new_idx")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement, got %d", len(statements))
		}
		if statements[0] != "RENAME INDEX old_idx TO new_idx" {
			t.Errorf("expected 'RENAME INDEX old_idx TO new_idx', got '%s'", statements[0])
		}
	})

	t.Run("compiles INDEX command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Index("name")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "CREATE INDEX test_table_name_index" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected CREATE INDEX statement, got %v", statements)
		}
	})

	t.Run("compiles UNIQUE command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Unique("email")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "ADD UNIQUE INDEX test_table_email_unique" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected ADD UNIQUE INDEX statement, got %v", statements)
		}
	})

	t.Run("compiles PRIMARY command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Primary("id")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "ADD PRIMARY KEY" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected ADD PRIMARY KEY statement, got %v", statements)
		}
	})

	t.Run("compiles DROP INDEX command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropIndex("name")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "DROP INDEX test_table_name_index" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected DROP INDEX statement, got %v", statements)
		}
	})

	t.Run("compiles DROP UNIQUE command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropUnique("email")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "DROP UNIQUE INDEX" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected DROP UNIQUE INDEX statement, got %v", statements)
		}
	})

	t.Run("compiles DROP PRIMARY command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropPrimary("id")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "DROP PRIMARY KEY" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected DROP PRIMARY KEY statement, got %v", statements)
		}
	})

	t.Run("compiles FOREIGN command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Foreign("user_id")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "ADD FOREIGN KEY" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected ADD FOREIGN KEY statement, got %v", statements)
		}
	})

	t.Run("compiles DROP FOREIGN command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropForeign("user_id")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "DROP FOREIGN KEY" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected DROP FOREIGN KEY statement, got %v", statements)
		}
	})

	t.Run("compiles FULLTEXT command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.FullText("content")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "ADD FULLTEXT INDEX" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected ADD FULLTEXT INDEX statement, got %v", statements)
		}
	})

	t.Run("compiles DROP FULLTEXT command", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropFullText("content")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "DROP FULLTEXT INDEX" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected DROP FULLTEXT INDEX statement, got %v", statements)
		}
	})

	t.Run("compiles COMMENT command", func(t *testing.T) {
		grammar := &MockGrammar{
			getAttributeCommandsFunc: func() []string {
				return []string{"comment"}
			},
		}
		bp := NewBlueprint(nil, "", "test_table")
		bp.String("name").Comment("User name")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, stmt := range statements {
			if stmt == "COMMENT ON COLUMN name" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected COMMENT ON COLUMN statement, got %v", statements)
		}
	})

	t.Run("skips commands marked as ShouldBeSkipped", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Create()
		bp.commands[0].ShouldBeSkipped = true

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 0 {
			t.Errorf("expected 0 statements when skipped, got %d: %v", len(statements), statements)
		}
	})

	t.Run("returns error when grammar CompileAdd fails", func(t *testing.T) {
		grammar := &MockGrammar{
			compileAddFunc: func(blueprint schema.Blueprint, command *schema.Command) (string, error) {
				return "", errors.New("compile error")
			},
		}
		bp := NewBlueprint(nil, "", "test_table")
		bp.String("name")

		_, err := bp.ToSql(grammar)
		if err == nil {
			t.Error("expected error when grammar fails")
		}
	})

	t.Run("returns error when grammar CompileChange fails", func(t *testing.T) {
		grammar := &MockGrammar{
			compileChangeFunc: func(blueprint schema.Blueprint, command *schema.Command) (string, error) {
				return "", errors.New("compile change error")
			},
		}
		bp := NewBlueprint(nil, "", "test_table")
		bp.String("name").Change()

		_, err := bp.ToSql(grammar)
		if err == nil {
			t.Error("expected error when grammar fails")
		}
	})

	t.Run("returns error when grammar CompileCreate fails", func(t *testing.T) {
		grammar := &MockGrammar{
			compileCreateFunc: func(blueprint schema.Blueprint) (string, error) {
				return "", errors.New("compile create error")
			},
		}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Create()

		_, err := bp.ToSql(grammar)
		if err == nil {
			t.Error("expected error when grammar fails")
		}
	})

	t.Run("returns error when grammar CompileDrop fails", func(t *testing.T) {
		grammar := &MockGrammar{
			compileDropFunc: func(blueprint schema.Blueprint) (string, error) {
				return "", errors.New("compile drop error")
			},
		}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Drop()

		_, err := bp.ToSql(grammar)
		if err == nil {
			t.Error("expected error when grammar fails")
		}
	})

	t.Run("returns error when grammar CompileDropColumn fails", func(t *testing.T) {
		grammar := &MockGrammar{
			compileDropColumnFunc: func(blueprint schema.Blueprint, command *schema.Command) ([]string, error) {
				return nil, errors.New("compile drop column error")
			},
		}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropColumn("name")

		_, err := bp.ToSql(grammar)
		if err == nil {
			t.Error("expected error when grammar fails")
		}
	})

	t.Run("returns error when grammar CompileDropIfExists fails", func(t *testing.T) {
		grammar := &MockGrammar{
			compileDropIfExistsFunc: func(blueprint schema.Blueprint) (string, error) {
				return "", errors.New("compile drop if exists error")
			},
		}
		bp := NewBlueprint(nil, "", "test_table")
		bp.DropIfExists()

		_, err := bp.ToSql(grammar)
		if err == nil {
			t.Error("expected error when grammar fails")
		}
	})

	t.Run("handles multiple statements", func(t *testing.T) {
		grammar := &MockGrammar{}
		bp := NewBlueprint(nil, "", "test_table")
		bp.Create()
		bp.String("name")
		bp.String("email")

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statements) != 1 {
			t.Errorf("expected 1 statement for create, got %d", len(statements))
		}
	})

	t.Run("filters out empty statements", func(t *testing.T) {
		grammar := &MockGrammar{
			compileChangeFunc: func(blueprint schema.Blueprint, command *schema.Command) (string, error) {
				return "", nil
			},
		}
		bp := NewBlueprint(nil, "", "test_table")
		bp.String("name")
		bp.Change()

		statements, err := bp.ToSql(grammar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The empty change statement should be filtered out
		for _, stmt := range statements {
			if stmt == "" {
				t.Error("expected empty statements to be filtered out")
			}
		}
	})
}

func TestBlueprint_DropByNameMethods(t *testing.T) {
	t.Run("DropForeignByName", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropForeignByName("fk_user_id")

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "dropForeign" {
			t.Errorf("expected dropForeign command, got %s", bp.commands[0].Name)
		}
		if bp.commands[0].Index != "fk_user_id" {
			t.Errorf("expected Index to be fk_user_id, got %s", bp.commands[0].Index)
		}
	})

	t.Run("DropIndexByName", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropIndexByName("idx_email")

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "dropIndex" {
			t.Errorf("expected dropIndex command, got %s", bp.commands[0].Name)
		}
		if bp.commands[0].Index != "idx_email" {
			t.Errorf("expected Index to be idx_email, got %s", bp.commands[0].Index)
		}
	})

	t.Run("DropFullTextByName", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropFullTextByName("ft_content")

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "dropFullText" {
			t.Errorf("expected dropFullText command, got %s", bp.commands[0].Name)
		}
		if bp.commands[0].Index != "ft_content" {
			t.Errorf("expected Index to be ft_content, got %s", bp.commands[0].Index)
		}
	})

	t.Run("DropUniqueByName", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropUniqueByName("uniq_email")

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "dropUnique" {
			t.Errorf("expected dropUnique command, got %s", bp.commands[0].Name)
		}
		if bp.commands[0].Index != "uniq_email" {
			t.Errorf("expected Index to be uniq_email, got %s", bp.commands[0].Index)
		}
	})
}

func TestBlueprint_MoreColumnMethods(t *testing.T) {
	t.Run("Char method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Char("code", 10)

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "char" {
			t.Errorf("expected char, got %s", column.GetType())
		}
		if column.GetLength() != 10 {
			t.Errorf("expected length 10, got %d", column.GetLength())
		}
	})

	t.Run("Char method with default length", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Char("code")

		if column.GetLength() != 255 {
			t.Errorf("expected default length 255, got %d", column.GetLength())
		}
	})

	t.Run("Decimal method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Decimal("amount")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "decimal" {
			t.Errorf("expected decimal, got %s", column.GetType())
		}
	})

	t.Run("Date method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Date("birthday")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "date" {
			t.Errorf("expected date, got %s", column.GetType())
		}
	})

	t.Run("DateTime method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.DateTime("created_at")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "dateTime" {
			t.Errorf("expected dateTime, got %s", column.GetType())
		}
	})

	t.Run("DateTime method with precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.DateTime("created_at", 6)

		if column.GetPrecision() != 6 {
			t.Errorf("expected precision 6, got %d", column.GetPrecision())
		}
	})

	t.Run("DateTimeTz method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.DateTimeTz("created_at")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "dateTimeTz" {
			t.Errorf("expected dateTimeTz, got %s", column.GetType())
		}
	})

	t.Run("Double method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Double("value")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "double" {
			t.Errorf("expected double, got %s", column.GetType())
		}
	})

	t.Run("Json method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Json("data")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "json" {
			t.Errorf("expected json, got %s", column.GetType())
		}
	})

	t.Run("Jsonb method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Jsonb("data")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "jsonb" {
			t.Errorf("expected jsonb, got %s", column.GetType())
		}
	})

	t.Run("LongText method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.LongText("content")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "longText" {
			t.Errorf("expected longText, got %s", column.GetType())
		}
	})

	t.Run("MediumInteger method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.MediumInteger("count")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "mediumInteger" {
			t.Errorf("expected mediumInteger, got %s", column.GetType())
		}
	})

	t.Run("MediumText method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.MediumText("content")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "mediumText" {
			t.Errorf("expected mediumText, got %s", column.GetType())
		}
	})

	t.Run("Time method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Time("start_time")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "time" {
			t.Errorf("expected time, got %s", column.GetType())
		}
	})

	t.Run("Time method with precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Time("start_time", 3)

		if column.GetPrecision() != 3 {
			t.Errorf("expected precision 3, got %d", column.GetPrecision())
		}
	})

	t.Run("TimeTz method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.TimeTz("start_time")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "timeTz" {
			t.Errorf("expected timeTz, got %s", column.GetType())
		}
	})

	t.Run("Timestamp method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Timestamp("created_at")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "timestamp" {
			t.Errorf("expected timestamp, got %s", column.GetType())
		}
	})

	t.Run("Timestamp method with precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Timestamp("created_at", 6)

		if column.GetPrecision() != 6 {
			t.Errorf("expected precision 6, got %d", column.GetPrecision())
		}
	})

	t.Run("TimestampTz method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.TimestampTz("created_at")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "timestampTz" {
			t.Errorf("expected timestampTz, got %s", column.GetType())
		}
	})

	t.Run("TinyInteger method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.TinyInteger("tiny_val")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "tinyInteger" {
			t.Errorf("expected tinyInteger, got %s", column.GetType())
		}
	})

	t.Run("TinyText method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.TinyText("tiny_content")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "tinyText" {
			t.Errorf("expected tinyText, got %s", column.GetType())
		}
	})

	t.Run("SmallInteger method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.SmallInteger("small_val")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "smallInteger" {
			t.Errorf("expected smallInteger, got %s", column.GetType())
		}
	})

	t.Run("BigInteger method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.BigInteger("big_val")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "bigInteger" {
			t.Errorf("expected bigInteger, got %s", column.GetType())
		}
	})

	t.Run("Column method", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Column("custom", "customType")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "customType" {
			t.Errorf("expected customType, got %s", column.GetType())
		}
		if column.GetName() != "custom" {
			t.Errorf("expected custom, got %s", column.GetName())
		}
	})
}

func TestBlueprint_SoftDeletesVariants(t *testing.T) {
	t.Run("SoftDeletesTz", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.SoftDeletesTz()

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "timestampTz" {
			t.Errorf("expected timestampTz, got %s", column.GetType())
		}
		if column.GetName() != "deleted_at" {
			t.Errorf("expected deleted_at, got %s", column.GetName())
		}
	})

	t.Run("SoftDeletesTz with custom name", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.SoftDeletesTz("removed_at")

		if column.GetName() != "removed_at" {
			t.Errorf("expected removed_at, got %s", column.GetName())
		}
	})
}

func TestBlueprint_DropSoftDeletesVariants(t *testing.T) {
	t.Run("DropSoftDeletes with default", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropSoftDeletes()

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "dropColumn" {
			t.Errorf("expected dropColumn command, got %s", bp.commands[0].Name)
		}
		if len(bp.commands[0].Columns) != 1 || bp.commands[0].Columns[0] != "deleted_at" {
			t.Errorf("expected [deleted_at], got %v", bp.commands[0].Columns)
		}
	})

	t.Run("DropSoftDeletes with custom column", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropSoftDeletes("removed_at")

		if len(bp.commands[0].Columns) != 1 || bp.commands[0].Columns[0] != "removed_at" {
			t.Errorf("expected [removed_at], got %v", bp.commands[0].Columns)
		}
	})

	t.Run("DropSoftDeletesTz with default", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropSoftDeletesTz()

		if len(bp.commands[0].Columns) != 1 || bp.commands[0].Columns[0] != "deleted_at" {
			t.Errorf("expected [deleted_at], got %v", bp.commands[0].Columns)
		}
	})

	t.Run("DropSoftDeletesTz with custom column", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropSoftDeletesTz("removed_at")

		if len(bp.commands[0].Columns) != 1 || bp.commands[0].Columns[0] != "removed_at" {
			t.Errorf("expected [removed_at], got %v", bp.commands[0].Columns)
		}
	})
}

func TestBlueprint_DropTimestampsVariants(t *testing.T) {
	t.Run("DropTimestamps", func(t *testing.T) {
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
	})

	t.Run("DropTimestampsTz", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.DropTimestampsTz()

		if len(bp.commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(bp.commands))
		}
		if bp.commands[0].Name != "dropColumn" {
			t.Errorf("expected dropColumn command, got %s", bp.commands[0].Name)
		}
		if len(bp.commands[0].Columns) != 2 || bp.commands[0].Columns[0] != "created_at" || bp.commands[0].Columns[1] != "updated_at" {
			t.Errorf("expected [created_at, updated_at], got %v", bp.commands[0].Columns)
		}
	})
}

func TestBlueprint_TimestampsVariants(t *testing.T) {
	t.Run("Timestamps with precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.Timestamps(6)

		if len(bp.columns) != 2 {
			t.Errorf("expected 2 columns, got %d", len(bp.columns))
		}
		if bp.columns[0].GetPrecision() != 6 {
			t.Errorf("expected precision 6 for created_at, got %d", bp.columns[0].GetPrecision())
		}
		if bp.columns[1].GetPrecision() != 6 {
			t.Errorf("expected precision 6 for updated_at, got %d", bp.columns[1].GetPrecision())
		}
	})

	t.Run("TimestampsTz", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.TimestampsTz()

		if len(bp.columns) != 2 {
			t.Errorf("expected 2 columns, got %d", len(bp.columns))
		}
		if bp.columns[0].GetType() != "timestampTz" {
			t.Errorf("expected timestampTz for created_at, got %s", bp.columns[0].GetType())
		}
	})

	t.Run("TimestampsTz with precision", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		bp.TimestampsTz(3)

		if bp.columns[0].GetPrecision() != 3 {
			t.Errorf("expected precision 3 for created_at, got %d", bp.columns[0].GetPrecision())
		}
	})
}

func TestBlueprint_IncrementVariants(t *testing.T) {
	t.Run("Increments", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.Increments("id")

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

	t.Run("BigIncrements", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.BigIncrements("id")

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

	t.Run("MediumIncrements", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.MediumIncrements("id")

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

	t.Run("SmallIncrements", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.SmallIncrements("id")

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

	t.Run("TinyIncrements", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.TinyIncrements("id")

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

	t.Run("IntegerIncrements", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.IntegerIncrements("id")

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
}

func TestBlueprint_UnsignedVariants(t *testing.T) {
	t.Run("UnsignedInteger", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.UnsignedInteger("count")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "integer" {
			t.Errorf("expected integer, got %s", column.GetType())
		}
		if !column.GetUnsigned() {
			t.Error("expected unsigned to be true")
		}
	})

	t.Run("UnsignedBigInteger", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.UnsignedBigInteger("big_count")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "bigInteger" {
			t.Errorf("expected bigInteger, got %s", column.GetType())
		}
		if !column.GetUnsigned() {
			t.Error("expected unsigned to be true")
		}
	})

	t.Run("UnsignedMediumInteger", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.UnsignedMediumInteger("medium_count")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "mediumInteger" {
			t.Errorf("expected mediumInteger, got %s", column.GetType())
		}
		if !column.GetUnsigned() {
			t.Error("expected unsigned to be true")
		}
	})

	t.Run("UnsignedSmallInteger", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.UnsignedSmallInteger("small_count")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "smallInteger" {
			t.Errorf("expected smallInteger, got %s", column.GetType())
		}
		if !column.GetUnsigned() {
			t.Error("expected unsigned to be true")
		}
	})

	t.Run("UnsignedTinyInteger", func(t *testing.T) {
		bp := &Blueprint{table: "test_table"}
		column := bp.UnsignedTinyInteger("tiny_count")

		if column == nil {
			t.Fatal("expected non-nil column")
		}
		if column.GetType() != "tinyInteger" {
			t.Errorf("expected tinyInteger, got %s", column.GetType())
		}
		if !column.GetUnsigned() {
			t.Error("expected unsigned to be true")
		}
	})
}
