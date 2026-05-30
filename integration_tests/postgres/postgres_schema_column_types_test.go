//go:build integration

package postgres

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestPostgreSQLSchemaColumnTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	types := []struct {
		name          string
		setup         func(table schema.Blueprint)
		expectedType  string
		nullable      bool
		autoincrement bool
		defaultValue  string
	}{
		{"big_increments_col", func(t schema.Blueprint) { t.BigIncrements("col") }, "int8", false, true, ""},
		{"big_integer_col", func(t schema.Blueprint) { t.BigInteger("col") }, "int8", false, false, ""},
		{"boolean_col", func(t schema.Blueprint) { t.Boolean("col") }, "bool", false, false, ""},
		{"char_col", func(t schema.Blueprint) { t.Char("col", 10) }, "bpchar", false, false, ""},
		{"date_col", func(t schema.Blueprint) { t.Date("col") }, "date", false, false, ""},
		{"datetime_col", func(t schema.Blueprint) { t.DateTime("col") }, "timestamp", false, false, ""},
		{"datetime_tz_col", func(t schema.Blueprint) { t.DateTimeTz("col") }, "timestamptz", false, false, ""},
		{"decimal_col", func(t schema.Blueprint) { t.Decimal("col") }, "numeric", false, false, ""},
		{"double_col", func(t schema.Blueprint) { t.Double("col") }, "float8", false, false, ""},
		{"enum_col", func(t schema.Blueprint) { t.Enum("col", []any{"a", "b"}) }, "varchar", false, false, ""},
		{"enum_int_col", func(t schema.Blueprint) { t.Enum("col", []any{1, 2}) }, "varchar", false, false, ""},
		{"float_col", func(t schema.Blueprint) { t.Float("col") }, "float8", false, false, ""},
		{"id_col", func(t schema.Blueprint) { t.ID("col") }, "int8", false, true, ""},
		{"increments_col", func(t schema.Blueprint) { t.Increments("col") }, "int4", false, true, ""},
		{"integer_col", func(t schema.Blueprint) { t.Integer("col") }, "int4", false, false, ""},
		{"integer_increments_col", func(t schema.Blueprint) { t.IntegerIncrements("col") }, "int4", false, true, ""},
		{"json_col", func(t schema.Blueprint) { t.Json("col") }, "json", false, false, ""},
		{"jsonb_col", func(t schema.Blueprint) { t.Jsonb("col") }, "jsonb", false, false, ""},
		{"long_text_col", func(t schema.Blueprint) { t.LongText("col") }, "text", false, false, ""},
		{"medium_increments_col", func(t schema.Blueprint) { t.MediumIncrements("col") }, "int4", false, true, ""},
		{"medium_integer_col", func(t schema.Blueprint) { t.MediumInteger("col") }, "int4", false, false, ""},
		{"medium_text_col", func(t schema.Blueprint) { t.MediumText("col") }, "text", false, false, ""},
		{"small_increments_col", func(t schema.Blueprint) { t.SmallIncrements("col") }, "int2", false, true, ""},
		{"small_integer_col", func(t schema.Blueprint) { t.SmallInteger("col") }, "int2", false, false, ""},
		{"string_col", func(t schema.Blueprint) { t.String("col", 255) }, "varchar", false, false, ""},
		{"text_col", func(t schema.Blueprint) { t.Text("col") }, "text", false, false, ""},
		{"time_col", func(t schema.Blueprint) { t.Time("col") }, "time", false, false, ""},
		{"time_tz_col", func(t schema.Blueprint) { t.TimeTz("col") }, "timetz", false, false, ""},
		{"timestamp_col", func(t schema.Blueprint) { t.Timestamp("col") }, "timestamp", false, false, ""},
		{"timestamp_tz_col", func(t schema.Blueprint) { t.TimestampTz("col") }, "timestamptz", false, false, ""},
		{"tiny_increments_col", func(t schema.Blueprint) { t.TinyIncrements("col") }, "int2", false, true, ""},
		{"tiny_integer_col", func(t schema.Blueprint) { t.TinyInteger("col") }, "int2", false, false, ""},
		{"tiny_text_col", func(t schema.Blueprint) { t.TinyText("col") }, "varchar", false, false, ""},
		{"unsigned_big_integer_col", func(t schema.Blueprint) { t.UnsignedBigInteger("col") }, "int8", false, false, ""},
		{"unsigned_integer_col", func(t schema.Blueprint) { t.UnsignedInteger("col") }, "int4", false, false, ""},
		{"unsigned_medium_integer_col", func(t schema.Blueprint) { t.UnsignedMediumInteger("col") }, "int4", false, false, ""},
		{"unsigned_small_integer_col", func(t schema.Blueprint) { t.UnsignedSmallInteger("col") }, "int2", false, false, ""},
		{"unsigned_tiny_integer_col", func(t schema.Blueprint) { t.UnsignedTinyInteger("col") }, "int2", false, false, ""},
		{"nullable_col", func(t schema.Blueprint) { t.String("col").Nullable() }, "varchar", true, false, ""},
		{"default_col", func(t schema.Blueprint) { t.String("col").Default("test") }, "varchar", false, false, "'test'::character varying"},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			tableName := fmt.Sprintf("test_types_pg_%s", tt.name)
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

			if columns[0].Name != "col" {
				t.Errorf("Expected column name 'col', got '%s'", columns[0].Name)
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

func TestPostgreSQLEnumConstraint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	tableName := "test_enum_constraint"
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.Enum("status", []any{"active", "inactive"})
	})
	if err != nil {
		t.Fatalf("Failed to create table with enum: %v", err)
	}

	_ = db.Schema().Drop(tableName)
}
