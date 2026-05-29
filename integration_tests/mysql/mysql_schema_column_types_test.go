//go:build integration

package mysql

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
)

func testColumnType(t *testing.T, db *database.Database, name string, setup func(schema.Blueprint), expectedType string, nullable, autoincrement bool, defaultValue string) {
	tableName := fmt.Sprintf("test_types_my_%s", name)
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

	if columns[0].Name != "col" {
		t.Errorf("Expected column name 'col', got '%s'", columns[0].Name)
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

func TestMySQLSchemaColumnTypeBigIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "big_increments_col", func(t schema.Blueprint) { t.BigIncrements("col") }, "bigint", false, true, "")
}

func TestMySQLSchemaColumnTypeBigInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "big_integer_col", func(t schema.Blueprint) { t.BigInteger("col") }, "bigint", false, false, "")
}

func TestMySQLSchemaColumnTypeBoolean(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "boolean_col", func(t schema.Blueprint) { t.Boolean("col") }, "tinyint", false, false, "")
}

func TestMySQLSchemaColumnTypeChar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "char_col", func(t schema.Blueprint) { t.Char("col", 10) }, "char", false, false, "")
}

func TestMySQLSchemaColumnTypeDate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "date_col", func(t schema.Blueprint) { t.Date("col") }, "date", false, false, "")
}

func TestMySQLSchemaColumnTypeDateTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "datetime_col", func(t schema.Blueprint) { t.DateTime("col") }, "datetime", false, false, "")
}

func TestMySQLSchemaColumnTypeDateTimeTz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "datetime_tz_col", func(t schema.Blueprint) { t.DateTimeTz("col") }, "datetime", false, false, "")
}

func TestMySQLSchemaColumnTypeDecimal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "decimal_col", func(t schema.Blueprint) { t.Decimal("col") }, "decimal", false, false, "")
}

func TestMySQLSchemaColumnTypeDouble(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "double_col", func(t schema.Blueprint) { t.Double("col") }, "double", false, false, "")
}

func TestMySQLSchemaColumnTypeEnum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "enum_col", func(t schema.Blueprint) { t.Enum("col", []any{"a", "b"}) }, "enum", false, false, "")
}

func TestMySQLSchemaColumnTypeFloat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "float_col", func(t schema.Blueprint) { t.Float("col") }, "double", false, false, "")
}

func TestMySQLSchemaColumnTypeID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "id_col", func(t schema.Blueprint) { t.ID("col") }, "bigint", false, true, "")
}

func TestMySQLSchemaColumnTypeIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "increments_col", func(t schema.Blueprint) { t.Increments("col") }, "int", false, true, "")
}

func TestMySQLSchemaColumnTypeInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "integer_col", func(t schema.Blueprint) { t.Integer("col") }, "int", false, false, "")
}

func TestMySQLSchemaColumnTypeIntegerIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "integer_increments_col", func(t schema.Blueprint) { t.IntegerIncrements("col") }, "int", false, true, "")
}

func TestMySQLSchemaColumnTypeJson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "json_col", func(t schema.Blueprint) { t.Json("col") }, "json", false, false, "")
}

func TestMySQLSchemaColumnTypeJsonb(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "jsonb_col", func(t schema.Blueprint) { t.Jsonb("col") }, "json", false, false, "")
}

func TestMySQLSchemaColumnTypeLongText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "long_text_col", func(t schema.Blueprint) { t.LongText("col") }, "longtext", false, false, "")
}

func TestMySQLSchemaColumnTypeMediumIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "medium_increments_col", func(t schema.Blueprint) { t.MediumIncrements("col") }, "mediumint", false, true, "")
}

func TestMySQLSchemaColumnTypeMediumInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "medium_integer_col", func(t schema.Blueprint) { t.MediumInteger("col") }, "mediumint", false, false, "")
}

func TestMySQLSchemaColumnTypeMediumText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "medium_text_col", func(t schema.Blueprint) { t.MediumText("col") }, "mediumtext", false, false, "")
}

func TestMySQLSchemaColumnTypeSmallIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "small_increments_col", func(t schema.Blueprint) { t.SmallIncrements("col") }, "smallint", false, true, "")
}

func TestMySQLSchemaColumnTypeSmallInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "small_integer_col", func(t schema.Blueprint) { t.SmallInteger("col") }, "smallint", false, false, "")
}

func TestMySQLSchemaColumnTypeString(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "string_col", func(t schema.Blueprint) { t.String("col", 255) }, "varchar", false, false, "")
}

func TestMySQLSchemaColumnTypeText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "text_col", func(t schema.Blueprint) { t.Text("col") }, "text", false, false, "")
}

func TestMySQLSchemaColumnTypeTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "time_col", func(t schema.Blueprint) { t.Time("col") }, "time", false, false, "")
}

func TestMySQLSchemaColumnTypeTimeTz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "time_tz_col", func(t schema.Blueprint) { t.TimeTz("col") }, "time", false, false, "")
}

func TestMySQLSchemaColumnTypeTimestamp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "timestamp_col", func(t schema.Blueprint) { t.Timestamp("col") }, "timestamp", false, false, "")
}

func TestMySQLSchemaColumnTypeTimestampTz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "timestamp_tz_col", func(t schema.Blueprint) { t.TimestampTz("col") }, "timestamp", false, false, "")
}

func TestMySQLSchemaColumnTypeTinyIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "tiny_increments_col", func(t schema.Blueprint) { t.TinyIncrements("col") }, "tinyint", false, true, "")
}

func TestMySQLSchemaColumnTypeTinyInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "tiny_integer_col", func(t schema.Blueprint) { t.TinyInteger("col") }, "tinyint", false, false, "")
}

func TestMySQLSchemaColumnTypeTinyText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "tiny_text_col", func(t schema.Blueprint) { t.TinyText("col") }, "tinytext", false, false, "")
}

func TestMySQLSchemaColumnTypeUnsignedBigInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "unsigned_big_integer_col", func(t schema.Blueprint) { t.UnsignedBigInteger("col") }, "bigint", false, false, "")
}

func TestMySQLSchemaColumnTypeUnsignedInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "unsigned_integer_col", func(t schema.Blueprint) { t.UnsignedInteger("col") }, "int", false, false, "")
}

func TestMySQLSchemaColumnTypeUnsignedMediumInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "unsigned_medium_integer_col", func(t schema.Blueprint) { t.UnsignedMediumInteger("col") }, "mediumint", false, false, "")
}

func TestMySQLSchemaColumnTypeUnsignedSmallInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "unsigned_small_integer_col", func(t schema.Blueprint) { t.UnsignedSmallInteger("col") }, "smallint", false, false, "")
}

func TestMySQLSchemaColumnTypeUnsignedTinyInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "unsigned_tiny_integer_col", func(t schema.Blueprint) { t.UnsignedTinyInteger("col") }, "tinyint", false, false, "")
}

func TestMySQLSchemaColumnTypeNullable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "nullable_col", func(t schema.Blueprint) { t.String("col").Nullable() }, "varchar", true, false, "")
}

func TestMySQLSchemaColumnTypeDefault(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupMySQLTest(t)
	testColumnType(t, db, "default_col", func(t schema.Blueprint) { t.String("col").Default("test") }, "varchar", false, false, "test")
}
