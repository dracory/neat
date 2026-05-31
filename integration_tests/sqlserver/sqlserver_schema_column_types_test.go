package sqlserver

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
)

func testColumnType(t *testing.T, db *database.Database, name string, setup func(schema.Blueprint), expectedType string, nullable, autoincrement bool, defaultValue string) {
	tableName := fmt.Sprintf("test_types_mssql_%s", name)
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

func TestSQLServerSchemaColumnTypeBigIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "big_increments_col", func(t schema.Blueprint) { t.BigIncrements("col") }, "bigint", false, true, "")
}

func TestSQLServerSchemaColumnTypeBigInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "big_integer_col", func(t schema.Blueprint) { t.BigInteger("col") }, "bigint", false, false, "")
}

func TestSQLServerSchemaColumnTypeBoolean(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "boolean_col", func(t schema.Blueprint) { t.Boolean("col") }, "bit", false, false, "")
}

func TestSQLServerSchemaColumnTypeChar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "char_col", func(t schema.Blueprint) { t.Char("col", 10) }, "nchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeDate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "date_col", func(t schema.Blueprint) { t.Date("col") }, "date", false, false, "")
}

func TestSQLServerSchemaColumnTypeDateTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "datetime_col", func(t schema.Blueprint) { t.DateTime("col") }, "datetime2", false, false, "")
}

func TestSQLServerSchemaColumnTypeDateTimeTz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "datetime_tz_col", func(t schema.Blueprint) { t.DateTimeTz("col") }, "datetimeoffset", false, false, "")
}

func TestSQLServerSchemaColumnTypeDecimal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "decimal_col", func(t schema.Blueprint) { t.Decimal("col") }, "decimal", false, false, "")
}

func TestSQLServerSchemaColumnTypeDouble(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "double_col", func(t schema.Blueprint) { t.Double("col") }, "float", false, false, "")
}

func TestSQLServerSchemaTypeEnum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "enum_col", func(t schema.Blueprint) { t.Enum("col", []any{"a", "b"}) }, "nvarchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeFloat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "float_col", func(t schema.Blueprint) { t.Float("col") }, "float", false, false, "")
}

func TestSQLServerSchemaColumnTypeID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "id_col", func(t schema.Blueprint) { t.ID("col") }, "bigint", false, true, "")
}

func TestSQLServerSchemaColumnTypeIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "increments_col", func(t schema.Blueprint) { t.Increments("col") }, "int", false, true, "")
}

func TestSQLServerSchemaColumnTypeInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "integer_col", func(t schema.Blueprint) { t.Integer("col") }, "int", false, false, "")
}

func TestSQLServerSchemaColumnTypeIntegerIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "integer_increments_col", func(t schema.Blueprint) { t.IntegerIncrements("col") }, "int", false, true, "")
}

func TestSQLServerSchemaColumnTypeJson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "json_col", func(t schema.Blueprint) { t.Json("col") }, "nvarchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeJsonb(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "jsonb_col", func(t schema.Blueprint) { t.Jsonb("col") }, "nvarchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeLongText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "long_text_col", func(t schema.Blueprint) { t.LongText("col") }, "nvarchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeMediumIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "medium_increments_col", func(t schema.Blueprint) { t.MediumIncrements("col") }, "int", false, true, "")
}

func TestSQLServerSchemaColumnTypeMediumInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "medium_integer_col", func(t schema.Blueprint) { t.MediumInteger("col") }, "int", false, false, "")
}

func TestSQLServerSchemaColumnTypeMediumText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "medium_text_col", func(t schema.Blueprint) { t.MediumText("col") }, "nvarchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeSmallIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "small_increments_col", func(t schema.Blueprint) { t.SmallIncrements("col") }, "smallint", false, true, "")
}

func TestSQLServerSchemaColumnTypeSmallInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "small_integer_col", func(t schema.Blueprint) { t.SmallInteger("col") }, "smallint", false, false, "")
}

func TestSQLServerSchemaColumnTypeString(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "string_col", func(t schema.Blueprint) { t.String("col", 255) }, "nvarchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "text_col", func(t schema.Blueprint) { t.Text("col") }, "nvarchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "time_col", func(t schema.Blueprint) { t.Time("col") }, "time", false, false, "")
}

func TestSQLServerSchemaColumnTypeTimeTz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "time_tz_col", func(t schema.Blueprint) { t.TimeTz("col") }, "time", false, false, "")
}

func TestSQLServerSchemaColumnTypeTimestamp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "timestamp_col", func(t schema.Blueprint) { t.Timestamp("col") }, "datetime2", false, false, "")
}

func TestSQLServerSchemaColumnTypeTimestampTz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "timestamp_tz_col", func(t schema.Blueprint) { t.TimestampTz("col") }, "datetimeoffset", false, false, "")
}

func TestSQLServerSchemaColumnTypeTinyIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "tiny_increments_col", func(t schema.Blueprint) { t.TinyIncrements("col") }, "tinyint", false, true, "")
}

func TestSQLServerSchemaColumnTypeTinyInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "tiny_integer_col", func(t schema.Blueprint) { t.TinyInteger("col") }, "tinyint", false, false, "")
}

func TestSQLServerSchemaColumnTypeTinyText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "tiny_text_col", func(t schema.Blueprint) { t.TinyText("col") }, "nvarchar", false, false, "")
}

func TestSQLServerSchemaColumnTypeUnsignedBigInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "unsigned_big_integer_col", func(t schema.Blueprint) { t.UnsignedBigInteger("col") }, "bigint", false, false, "")
}

func TestSQLServerSchemaColumnTypeUnsignedInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "unsigned_integer_col", func(t schema.Blueprint) { t.UnsignedInteger("col") }, "int", false, false, "")
}

func TestSQLServerSchemaColumnTypeUnsignedMediumInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "unsigned_medium_integer_col", func(t schema.Blueprint) { t.UnsignedMediumInteger("col") }, "int", false, false, "")
}

func TestSQLServerSchemaColumnTypeUnsignedSmallInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "unsigned_small_integer_col", func(t schema.Blueprint) { t.UnsignedSmallInteger("col") }, "smallint", false, false, "")
}

func TestSQLServerSchemaColumnTypeUnsignedTinyInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "unsigned_tiny_integer_col", func(t schema.Blueprint) { t.UnsignedTinyInteger("col") }, "tinyint", false, false, "")
}

func TestSQLServerSchemaColumnTypeNullable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "nullable_col", func(t schema.Blueprint) { t.String("col").Nullable() }, "nvarchar", true, false, "")
}

func TestSQLServerSchemaColumnTypeDefault(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupSQLServerTest(t)
	testColumnType(t, db, "default_col", func(t schema.Blueprint) { t.String("col").Default("test") }, "nvarchar", false, false, "('test')")
}
