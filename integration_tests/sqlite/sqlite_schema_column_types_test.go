package sqlite

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/database/schema/grammars"
)

func testColumnType(t *testing.T, db *database.Database, name string, setup func(schema.Blueprint), expectedName, expectedType string, nullable, autoincrement bool, defaultValue string) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tableName := fmt.Sprintf("test_types_%s", name)
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		setup(table)
	})
	if err != nil {
		t.Fatalf("Failed to create table for %s: %v", name, err)
	}

	columns, err := db.Schema().GetColumns(tableName)
	if err != nil {
		t.Fatalf("Failed to get columns for %s: %v", name, err)
	}
	if len(columns) != 1 {
		t.Fatalf("Expected 1 column for %s, got %d", name, len(columns))
	}

	if columns[0].Name != expectedName {
		t.Errorf("Wrong name for %s: expected %s, got %s", name, expectedName, columns[0].Name)
	}
	if columns[0].TypeName != expectedType {
		t.Errorf("Wrong type for %s: expected %s, got %s", name, expectedType, columns[0].TypeName)
	}
	if columns[0].Nullable != nullable {
		t.Errorf("Wrong nullable for %s: expected %v, got %v", name, nullable, columns[0].Nullable)
	}
	if columns[0].Autoincrement != autoincrement {
		t.Errorf("Wrong autoincrement for %s: expected %v, got %v", name, autoincrement, columns[0].Autoincrement)
	}
	if defaultValue != "" && columns[0].Default != defaultValue {
		t.Errorf("Wrong default for %s: expected %s, got %s", name, defaultValue, columns[0].Default)
	}

	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table for %s: %v", name, err)
	}
}

func TestSQLiteSchemaColumnTypeBigIncrements(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "big_increments_col", func(t schema.Blueprint) { t.BigIncrements("col") }, "col", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeBigInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "big_integer_col", func(t schema.Blueprint) { t.BigInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeBoolean(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "boolean_col", func(t schema.Blueprint) { t.Boolean("col") }, "col", "tinyint", false, false, "")
}

func TestSQLiteSchemaColumnTypeChar(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "char_col", func(t schema.Blueprint) { t.Char("col", 10) }, "col", "varchar", false, false, "")
}

func TestSQLiteSchemaColumnTypeDate(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "date_col", func(t schema.Blueprint) { t.Date("col") }, "col", "date", false, false, "")
}

func TestSQLiteSchemaColumnTypeDateTime(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "datetime_col", func(t schema.Blueprint) { t.DateTime("col") }, "col", "datetime", false, false, "")
}

func TestSQLiteSchemaColumnTypeDateTimeTz(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "datetime_tz_col", func(t schema.Blueprint) { t.DateTimeTz("col") }, "col", "datetime", false, false, "")
}

func TestSQLiteSchemaColumnTypeDecimal(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "decimal_col", func(t schema.Blueprint) { t.Decimal("col") }, "col", "numeric", false, false, "")
}

func TestSQLiteSchemaColumnTypeDouble(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "double_col", func(t schema.Blueprint) { t.Double("col") }, "col", "double", false, false, "")
}

func TestSQLiteSchemaColumnTypeEnum(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "enum_col", func(t schema.Blueprint) { t.Enum("col", []any{"a", "b"}) }, "col", "varchar", false, false, "")
}

func TestSQLiteSchemaColumnTypeFloat(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "float_col", func(t schema.Blueprint) { t.Float("col") }, "col", "float", false, false, "")
}

func TestSQLiteSchemaColumnTypeID(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "id_col", func(t schema.Blueprint) { t.ID("col") }, "col", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeIDDefault(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "id_default", func(t schema.Blueprint) { t.ID() }, "id", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeIncrements(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "increments_col", func(t schema.Blueprint) { t.Increments("col") }, "col", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "integer_col", func(t schema.Blueprint) { t.Integer("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeIntegerIncrements(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "integer_increments_col", func(t schema.Blueprint) { t.IntegerIncrements("col") }, "col", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeJson(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "json_col", func(t schema.Blueprint) { t.Json("col") }, "col", "text", false, false, "")
}

func TestSQLiteSchemaColumnTypeJsonb(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "jsonb_col", func(t schema.Blueprint) { t.Jsonb("col") }, "col", "text", false, false, "")
}

func TestSQLiteSchemaColumnTypeLongText(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "long_text_col", func(t schema.Blueprint) { t.LongText("col") }, "col", "text", false, false, "")
}

func TestSQLiteSchemaColumnTypeMediumIncrements(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "medium_increments_col", func(t schema.Blueprint) { t.MediumIncrements("col") }, "col", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeMediumInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "medium_integer_col", func(t schema.Blueprint) { t.MediumInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeMediumText(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "medium_text_col", func(t schema.Blueprint) { t.MediumText("col") }, "col", "text", false, false, "")
}

func TestSQLiteSchemaColumnTypeSmallIncrements(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "small_increments_col", func(t schema.Blueprint) { t.SmallIncrements("col") }, "col", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeSmallInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "small_integer_col", func(t schema.Blueprint) { t.SmallInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeString(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "string_col", func(t schema.Blueprint) { t.String("col", 255) }, "col", "varchar", false, false, "")
}

func TestSQLiteSchemaColumnTypeText(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "text_col", func(t schema.Blueprint) { t.Text("col") }, "col", "text", false, false, "")
}

func TestSQLiteSchemaColumnTypeTime(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "time_col", func(t schema.Blueprint) { t.Time("col") }, "col", "time", false, false, "")
}

func TestSQLiteSchemaColumnTypeTimeTz(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "time_tz_col", func(t schema.Blueprint) { t.TimeTz("col") }, "col", "time", false, false, "")
}

func TestSQLiteSchemaColumnTypeTimestamp(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "timestamp_col", func(t schema.Blueprint) { t.Timestamp("col") }, "col", "datetime", false, false, "")
}

func TestSQLiteSchemaColumnTypeTimestampTz(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "timestamp_tz_col", func(t schema.Blueprint) { t.TimestampTz("col") }, "col", "datetime", false, false, "")
}

func TestSQLiteSchemaColumnTypeTinyIncrements(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "tiny_increments_col", func(t schema.Blueprint) { t.TinyIncrements("col") }, "col", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeTinyInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "tiny_integer_col", func(t schema.Blueprint) { t.TinyInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeTinyText(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "tiny_text_col", func(t schema.Blueprint) { t.TinyText("col") }, "col", "text", false, false, "")
}

func TestSQLiteSchemaColumnTypeUnsignedBigInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "unsigned_big_integer_col", func(t schema.Blueprint) { t.UnsignedBigInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeUnsignedInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "unsigned_integer_col", func(t schema.Blueprint) { t.UnsignedInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeUnsignedMediumInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "unsigned_medium_integer_col", func(t schema.Blueprint) { t.UnsignedMediumInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeUnsignedSmallInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "unsigned_small_integer_col", func(t schema.Blueprint) { t.UnsignedSmallInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeUnsignedTinyInteger(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "unsigned_tiny_integer_col", func(t schema.Blueprint) { t.UnsignedTinyInteger("col") }, "col", "integer", false, false, "")
}

func TestSQLiteSchemaColumnTypeNullable(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "nullable_col", func(t schema.Blueprint) { t.String("col").Nullable() }, "col", "varchar", true, false, "")
}

func TestSQLiteSchemaColumnTypeDefault(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "default_col", func(t schema.Blueprint) { t.String("col").Default("test") }, "col", "varchar", false, false, "'test'")
}

func TestSQLiteSchemaColumnTypeIDCustom(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "id_custom", func(t schema.Blueprint) { t.ID("custom") }, "custom", "integer", false, true, "")
}

func TestSQLiteSchemaColumnTypeTimestampUseCurrent(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "timestamp_use_current", func(t schema.Blueprint) { t.Timestamp("col").UseCurrent() }, "col", "datetime", false, false, "CURRENT_TIMESTAMP")
}

func TestSQLiteSchemaColumnTypeTimestampTzUseCurrent(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "timestamp_tz_use_current", func(t schema.Blueprint) { t.TimestampTz("col").UseCurrent() }, "col", "datetime", false, false, "CURRENT_TIMESTAMP")
}

func TestSQLiteSchemaColumnTypeDateTimeUseCurrent(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "datetime_use_current", func(t schema.Blueprint) { t.DateTime("col").UseCurrent() }, "col", "datetime", false, false, "CURRENT_TIMESTAMP")
}

func TestSQLiteSchemaColumnTypeDateTimeTzUseCurrent(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "datetime_tz_use_current", func(t schema.Blueprint) { t.DateTimeTz("col").UseCurrent() }, "col", "datetime", false, false, "CURRENT_TIMESTAMP")
}

func TestSQLiteSchemaColumnTypeDefaultBoolTrue(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "default_bool_true", func(t schema.Blueprint) { t.Boolean("col").Default(true) }, "col", "tinyint", false, false, "'1'")
}

func TestSQLiteSchemaColumnTypeDefaultBoolFalse(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "default_bool_false", func(t schema.Blueprint) { t.Boolean("col").Default(false) }, "col", "tinyint", false, false, "'0'")
}

func TestSQLiteSchemaColumnTypeDefaultInt(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "default_int", func(t schema.Blueprint) { t.Integer("col").Default(123) }, "col", "integer", false, false, "'123'")
}

func TestSQLiteSchemaColumnTypeDefaultExpression(t *testing.T) {
	db := SetupSQLiteTest(t)
	testColumnType(t, db, "default_expression", func(t schema.Blueprint) { t.Integer("col").Default(grammars.Expression("(1 + 1)")) }, "col", "integer", false, false, "1 + 1")
}

func TestSQLiteEnumConstraint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_enum_constraint"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.Enum("status", []any{"active", "inactive"})
	})
	if err != nil {
		t.Fatalf("Failed to create table with enum: %v", err)
	}

	// Test valid value
	err = db.Query().Table(tableName).Create(map[string]any{
		"status": "active",
	})
	if err != nil {
		t.Errorf("Failed to insert valid enum value: %v", err)
	}

	// Test invalid value
	err = db.Query().Table(tableName).Create(map[string]any{
		"status": "pending",
	})
	if err == nil {
		t.Error("Should fail due to CHECK constraint")
	}

	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}
