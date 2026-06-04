package oracle_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

// TestOracleSchemaColumnChange tests column change operations
func TestOracleSchemaColumnChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)

	tableName := "test_column_change"

	// Clean up table if it exists from previous test run
	if db.Schema().HasTable(tableName) {
		_ = db.Schema().Drop(tableName)
	}

	// Create a table with initial columns
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("change_length")
		table.String("change_type")
		table.String("change_nullable")
		table.String("change_not_nullable").Nullable() // Start as nullable
		table.String("change_default")
		table.String("change_remove_default").Default("test_value")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Verify table was created
	if !db.Schema().HasTable(tableName) {
		t.Error("Table should exist")
	}
	if !db.Schema().HasColumn(tableName, "change_length") {
		t.Error("Column change_length should exist")
	}
	if !db.Schema().HasColumn(tableName, "change_type") {
		t.Error("Column change_type should exist")
	}

	// Modify columns using Change()
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		// Change column length
		table.String("change_length", 100).Change()

		// Change column type - use String instead of Text to avoid LONG to LOB conversion error
		table.String("change_type", 2000).Change()

		// Change to nullable
		table.String("change_nullable").Nullable().Change()

		// Change to not nullable (now works since it starts as nullable)
		table.String("change_not_nullable").Change()

		// Add default value
		table.String("change_default").Default("new_default").Change()

		// Remove default value
		table.String("change_remove_default").Change()
	})
	if err != nil {
		t.Fatalf("Failed to modify table: %v", err)
	}

	// Verify modifications
	if !db.Schema().HasColumn(tableName, "change_length") {
		t.Error("Column change_length should still exist")
	}
	if !db.Schema().HasColumn(tableName, "change_type") {
		t.Error("Column change_type should still exist")
	}
	if !db.Schema().HasColumn(tableName, "change_not_nullable") {
		t.Error("Column change_not_nullable should still exist")
	}

	// Clean up
	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}

// TestOracleSchemaColumnChangeType tests changing column types
func TestOracleSchemaColumnChangeType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)

	tableName := "test_change_type"

	// Clean up table if it exists from previous test run
	_ = db.Schema().DropIfExists(tableName)
	sqlDB, err := db.DB()
	if err == nil {
		_, _ = sqlDB.Exec(fmt.Sprintf("BEGIN EXECUTE IMMEDIATE 'DROP TABLE %s CASCADE CONSTRAINTS'; EXCEPTION WHEN OTHERS THEN NULL; END;", strings.ToUpper(tableName)))
	}

	// Create table
	err = db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("varchar_col")
		table.Integer("int_col").Nullable() // Start as nullable
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Change column types
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		// Use String instead of Text to avoid LONG to LOB conversion error
		table.String("varchar_col", 2000).Change()
		// Now we can change to BigInteger since it starts as nullable
		table.BigInteger("int_col").Change()
	})
	if err != nil {
		t.Fatalf("Failed to change column types: %v", err)
	}

	// Clean up
	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}

// TestOracleSchemaColumnChangeNullable tests changing nullable status
func TestOracleSchemaColumnChangeNullable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)

	tableName := "test_change_nullable"

	// Clean up table if it exists from previous test run
	if db.Schema().HasTable(tableName) {
		_ = db.Schema().Drop(tableName)
	}

	// Create table
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("not_null_col")
		table.String("null_col").Nullable()
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Change nullable status
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.String("not_null_col").Nullable().Change()
		table.String("null_col").Change()
	})
	if err != nil {
		t.Fatalf("Failed to change nullable status: %v", err)
	}

	// Clean up
	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}
