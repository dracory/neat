package sqlite

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
)

func TestSQLiteSchemaColumnTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	types := []struct {
		name          string
		setup         func(table schema.Blueprint)
		expectedName  string
		expectedType  string
		nullable      bool
		autoincrement bool
		defaultValue  string
	}{
		{name: "big_increments_col", setup: func(t schema.Blueprint) { t.BigIncrements("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "big_integer_col", setup: func(t schema.Blueprint) { t.BigInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "boolean_col", setup: func(t schema.Blueprint) { t.Boolean("col") }, expectedName: "col", expectedType: "tinyint", nullable: false, autoincrement: false},
		{name: "char_col", setup: func(t schema.Blueprint) { t.Char("col", 10) }, expectedName: "col", expectedType: "varchar", nullable: false, autoincrement: false},
		{name: "date_col", setup: func(t schema.Blueprint) { t.Date("col") }, expectedName: "col", expectedType: "date", nullable: false, autoincrement: false},
		{name: "datetime_col", setup: func(t schema.Blueprint) { t.DateTime("col") }, expectedName: "col", expectedType: "datetime", nullable: false, autoincrement: false},
		{name: "datetime_tz_col", setup: func(t schema.Blueprint) { t.DateTimeTz("col") }, expectedName: "col", expectedType: "datetime", nullable: false, autoincrement: false},
		{name: "decimal_col", setup: func(t schema.Blueprint) { t.Decimal("col") }, expectedName: "col", expectedType: "numeric", nullable: false, autoincrement: false},
		{name: "double_col", setup: func(t schema.Blueprint) { t.Double("col") }, expectedName: "col", expectedType: "double", nullable: false, autoincrement: false},
		{name: "enum_col", setup: func(t schema.Blueprint) { t.Enum("col", []any{"a", "b"}) }, expectedName: "col", expectedType: "varchar", nullable: false, autoincrement: false},
		{name: "float_col", setup: func(t schema.Blueprint) { t.Float("col") }, expectedName: "col", expectedType: "float", nullable: false, autoincrement: false},
		{name: "id_col", setup: func(t schema.Blueprint) { t.ID("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "id_default", setup: func(t schema.Blueprint) { t.ID() }, expectedName: "id", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "increments_col", setup: func(t schema.Blueprint) { t.Increments("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "integer_col", setup: func(t schema.Blueprint) { t.Integer("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "integer_increments_col", setup: func(t schema.Blueprint) { t.IntegerIncrements("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "json_col", setup: func(t schema.Blueprint) { t.Json("col") }, expectedName: "col", expectedType: "text", nullable: false, autoincrement: false},
		{name: "jsonb_col", setup: func(t schema.Blueprint) { t.Jsonb("col") }, expectedName: "col", expectedType: "text", nullable: false, autoincrement: false},
		{name: "long_text_col", setup: func(t schema.Blueprint) { t.LongText("col") }, expectedName: "col", expectedType: "text", nullable: false, autoincrement: false},
		{name: "medium_increments_col", setup: func(t schema.Blueprint) { t.MediumIncrements("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "medium_integer_col", setup: func(t schema.Blueprint) { t.MediumInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "medium_text_col", setup: func(t schema.Blueprint) { t.MediumText("col") }, expectedName: "col", expectedType: "text", nullable: false, autoincrement: false},
		{name: "small_increments_col", setup: func(t schema.Blueprint) { t.SmallIncrements("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "small_integer_col", setup: func(t schema.Blueprint) { t.SmallInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "string_col", setup: func(t schema.Blueprint) { t.String("col", 255) }, expectedName: "col", expectedType: "varchar", nullable: false, autoincrement: false},
		{name: "text_col", setup: func(t schema.Blueprint) { t.Text("col") }, expectedName: "col", expectedType: "text", nullable: false, autoincrement: false},
		{name: "time_col", setup: func(t schema.Blueprint) { t.Time("col") }, expectedName: "col", expectedType: "time", nullable: false, autoincrement: false},
		{name: "time_tz_col", setup: func(t schema.Blueprint) { t.TimeTz("col") }, expectedName: "col", expectedType: "time", nullable: false, autoincrement: false},
		{name: "timestamp_col", setup: func(t schema.Blueprint) { t.Timestamp("col") }, expectedName: "col", expectedType: "datetime", nullable: false, autoincrement: false},
		{name: "timestamp_tz_col", setup: func(t schema.Blueprint) { t.TimestampTz("col") }, expectedName: "col", expectedType: "datetime", nullable: false, autoincrement: false},
		{name: "tiny_increments_col", setup: func(t schema.Blueprint) { t.TinyIncrements("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "tiny_integer_col", setup: func(t schema.Blueprint) { t.TinyInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "tiny_text_col", setup: func(t schema.Blueprint) { t.TinyText("col") }, expectedName: "col", expectedType: "text", nullable: false, autoincrement: false},
		{name: "unsigned_big_integer_col", setup: func(t schema.Blueprint) { t.UnsignedBigInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "unsigned_integer_col", setup: func(t schema.Blueprint) { t.UnsignedInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "unsigned_medium_integer_col", setup: func(t schema.Blueprint) { t.UnsignedMediumInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "unsigned_small_integer_col", setup: func(t schema.Blueprint) { t.UnsignedSmallInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "unsigned_tiny_integer_col", setup: func(t schema.Blueprint) { t.UnsignedTinyInteger("col") }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false},
		{name: "nullable_col", setup: func(t schema.Blueprint) { t.String("col").Nullable() }, expectedName: "col", expectedType: "varchar", nullable: true, autoincrement: false},
		{name: "default_col", setup: func(t schema.Blueprint) { t.String("col").Default("test") }, expectedName: "col", expectedType: "varchar", nullable: false, autoincrement: false, defaultValue: "'test'"},
		{name: "id_custom", setup: func(t schema.Blueprint) { t.ID("custom") }, expectedName: "custom", expectedType: "integer", nullable: false, autoincrement: true},
		{name: "timestamp_use_current", setup: func(t schema.Blueprint) { t.Timestamp("col").UseCurrent() }, expectedName: "col", expectedType: "datetime", nullable: false, autoincrement: false, defaultValue: "CURRENT_TIMESTAMP"},
		{name: "timestamp_tz_use_current", setup: func(t schema.Blueprint) { t.TimestampTz("col").UseCurrent() }, expectedName: "col", expectedType: "datetime", nullable: false, autoincrement: false, defaultValue: "CURRENT_TIMESTAMP"},
		{name: "datetime_use_current", setup: func(t schema.Blueprint) { t.DateTime("col").UseCurrent() }, expectedName: "col", expectedType: "datetime", nullable: false, autoincrement: false, defaultValue: "CURRENT_TIMESTAMP"},
		{name: "datetime_tz_use_current", setup: func(t schema.Blueprint) { t.DateTimeTz("col").UseCurrent() }, expectedName: "col", expectedType: "datetime", nullable: false, autoincrement: false, defaultValue: "CURRENT_TIMESTAMP"},
		{name: "default_bool_true", setup: func(t schema.Blueprint) { t.Boolean("col").Default(true) }, expectedName: "col", expectedType: "tinyint", nullable: false, autoincrement: false, defaultValue: "'1'"},
		{name: "default_bool_false", setup: func(t schema.Blueprint) { t.Boolean("col").Default(false) }, expectedName: "col", expectedType: "tinyint", nullable: false, autoincrement: false, defaultValue: "'0'"},
		{name: "default_int", setup: func(t schema.Blueprint) { t.Integer("col").Default(123) }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false, defaultValue: "'123'"},
		{name: "default_expression", setup: func(t schema.Blueprint) { t.Integer("col").Default(grammars.Expression("(1 + 1)")) }, expectedName: "col", expectedType: "integer", nullable: false, autoincrement: false, defaultValue: "1 + 1"},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			tableName := fmt.Sprintf("test_types_%s", tt.name)
			_ = db.Schema().DropIfExists(tableName)

			err := db.Schema().Create(tableName, func(table schema.Blueprint) {
				tt.setup(table)
			})
			if err != nil {
				t.Fatalf("Failed to create table for %s: %v", tt.name, err)
			}

			columns, err := db.Schema().GetColumns(tableName)
			if err != nil {
				t.Fatalf("Failed to get columns for %s: %v", tt.name, err)
			}
			if len(columns) != 1 {
				t.Fatalf("Expected 1 column for %s, got %d", tt.name, len(columns))
			}

			if columns[0].Name != tt.expectedName {
				t.Errorf("Wrong name for %s: expected %s, got %s", tt.name, tt.expectedName, columns[0].Name)
			}
			if columns[0].TypeName != tt.expectedType {
				t.Errorf("Wrong type for %s: expected %s, got %s", tt.name, tt.expectedType, columns[0].TypeName)
			}
			if columns[0].Nullable != tt.nullable {
				t.Errorf("Wrong nullable for %s: expected %v, got %v", tt.name, tt.nullable, columns[0].Nullable)
			}
			if columns[0].Autoincrement != tt.autoincrement {
				t.Errorf("Wrong autoincrement for %s: expected %v, got %v", tt.name, tt.autoincrement, columns[0].Autoincrement)
			}
			if tt.defaultValue != "" && columns[0].Default != tt.defaultValue {
				t.Errorf("Wrong default for %s: expected %s, got %s", tt.name, tt.defaultValue, columns[0].Default)
			}

			err = db.Schema().Drop(tableName)
			if err != nil {
				t.Fatalf("Failed to drop table for %s: %v", tt.name, err)
			}
		})
	}
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
