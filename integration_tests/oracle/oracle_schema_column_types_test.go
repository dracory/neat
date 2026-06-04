package oracle_test

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
)

func testColumnType(t *testing.T, db *database.Database, name string, setup func(schema.Blueprint), expectedType string, nullable, autoincrement bool, defaultValue string) {
	tableName := fmt.Sprintf("test_types_or_%s", name)
	_ = db.Schema().DropIfExists(tableName)

	// Additional cleanup to handle Oracle case sensitivity
	sqlDB, err := db.DB()
	if err == nil {
		_, _ = sqlDB.Exec(fmt.Sprintf("BEGIN EXECUTE IMMEDIATE 'DROP TABLE %s CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;", tableName))
	}

	err = db.Schema().Create(tableName, func(table schema.Blueprint) {
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

	// Oracle returns column names in uppercase by default
	colName := columns[0].Name
	if colName != "COL" && colName != "col" {
		t.Errorf("Expected column name 'COL' or 'col', got '%s'", colName)
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

func TestOracleSchemaColumnTypeBigIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "big_increments_col", func(t schema.Blueprint) { t.BigIncrements("col") }, "number", false, true, "")
}

func TestOracleSchemaColumnTypeBigInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "big_integer_col", func(t schema.Blueprint) { t.BigInteger("col") }, "number", false, false, "")
}

func TestOracleSchemaColumnTypeBoolean(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "boolean_col", func(t schema.Blueprint) { t.Boolean("col") }, "number", false, false, "")
}

func TestOracleSchemaColumnTypeChar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "char_col", func(t schema.Blueprint) { t.Char("col", 10) }, "char", false, false, "")
}

func TestOracleSchemaColumnTypeDate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "date_col", func(t schema.Blueprint) { t.Date("col") }, "date", false, false, "")
}

func TestOracleSchemaColumnTypeDateTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	// Oracle returns TIMESTAMP(6) by default for TIMESTAMP columns, but we normalize to "timestamp"
	testColumnType(t, db, "datetime_col", func(t schema.Blueprint) { t.DateTime("col") }, "timestamp", false, false, "")
}

func TestOracleSchemaColumnTypeDecimal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "decimal_col", func(t schema.Blueprint) { t.Decimal("col") }, "number", false, false, "")
}

func TestOracleSchemaColumnTypeDouble(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "double_col", func(t schema.Blueprint) { t.Double("col") }, "binary_double", false, false, "")
}

func TestOracleSchemaColumnTypeFloat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "float_col", func(t schema.Blueprint) { t.Float("col") }, "binary_float", false, false, "")
}

func TestOracleSchemaColumnTypeID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "id_col", func(t schema.Blueprint) { t.ID("col") }, "number", false, true, "")
}

func TestOracleSchemaColumnTypeIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "increments_col", func(t schema.Blueprint) { t.Increments("col") }, "number", false, true, "")
}

func TestOracleSchemaColumnTypeInteger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "integer_col", func(t schema.Blueprint) { t.Integer("col") }, "number", false, false, "")
}

func TestOracleSchemaColumnTypeIntegerIncrements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "integer_increments_col", func(t schema.Blueprint) { t.IntegerIncrements("col") }, "number", false, true, "")
}

func TestOracleSchemaColumnTypeString(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "string_col", func(t schema.Blueprint) { t.String("col", 255) }, "varchar2", false, false, "")
}

func TestOracleSchemaColumnTypeText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "text_col", func(t schema.Blueprint) { t.Text("col") }, "clob", false, false, "")
}

func TestOracleSchemaColumnTypeTimestamp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	// Oracle returns TIMESTAMP(6) by default for TIMESTAMP columns, but we normalize to "timestamp"
	testColumnType(t, db, "timestamp_col", func(t schema.Blueprint) { t.Timestamp("col") }, "timestamp", false, false, "")
}

func TestOracleSchemaColumnTypeNullable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "nullable_col", func(t schema.Blueprint) { t.String("col").Nullable() }, "varchar2", true, false, "")
}

func TestOracleSchemaColumnTypeDefault(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	db := SetupOracleTest(t)
	testColumnType(t, db, "default_col", func(t schema.Blueprint) { t.String("col").Default("test") }, "varchar2", false, false, "test")
}
