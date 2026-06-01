package mysql

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestMySQLSchemaColumnModifiers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)

	tableName := "test_column_modifiers"
	_ = db.Schema().DropIfExists(tableName)

	// Create a table with various modifiers
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("nullable_col").Nullable()
		table.String("default_string").Default("hello")
		table.Integer("default_int").Default(123)
		table.Boolean("default_bool").Default(true)
		table.Timestamp("use_current").UseCurrent()
		table.Timestamp("use_current_on_update").UseCurrent().UseCurrentOnUpdate()
		table.Integer("unsigned_col").Unsigned()
		table.String("collation_col").Collation("utf8mb4_unicode_ci")
		table.String("comment_col").Comment("this is a comment")
	})
	if err != nil {
		t.Fatalf("Failed to create table with modifiers: %v", err)
	}

	columns, err := db.Schema().GetColumns(tableName)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	colMap := make(map[string]schema.Column)
	for _, col := range columns {
		colMap[col.Name] = col
	}

	if !colMap["nullable_col"].Nullable {
		t.Error("nullable_col should be nullable")
	}
	if colMap["default_string"].Default != "hello" {
		t.Errorf("Expected default 'hello', got '%s'", colMap["default_string"].Default)
	}
	if colMap["default_int"].Default != "123" {
		t.Errorf("Expected default '123', got '%s'", colMap["default_int"].Default)
	}
	// MySQL boolean is tinyint(1), default 1
	if colMap["default_bool"].Default != "1" {
		t.Errorf("Expected default '1', got '%s'", colMap["default_bool"].Default)
	}
	if colMap["use_current"].Default == "" {
		t.Error("use_current should have CURRENT_TIMESTAMP default")
	}
	if colMap["use_current_on_update"].Default == "" {
		t.Error("use_current_on_update should have CURRENT_TIMESTAMP default")
	}

	if colMap["collation_col"].Collation != "utf8mb4_unicode_ci" {
		t.Errorf("Expected collation 'utf8mb4_unicode_ci', got '%s'", colMap["collation_col"].Collation)
	}
	if colMap["comment_col"].Comment != "this is a comment" {
		t.Errorf("Expected comment 'this is a comment', got '%s'", colMap["comment_col"].Comment)
	}

	// Test After and First
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.String("first_col").First()
		table.String("after_col").After("id")
	})
	if err != nil {
		t.Fatalf("Failed to add first and after columns: %v", err)
	}

	columns, err = db.Schema().GetColumns(tableName)
	if err != nil {
		t.Fatalf("Failed to get columns after adding first/after: %v", err)
	}

	if len(columns) == 0 {
		t.Fatal("GetColumns returned empty slice")
	}

	// Verify order: first_col, id, after_col, ...
	if columns[0].Name != "first_col" {
		t.Errorf("Expected first column 'first_col', got '%s'", columns[0].Name)
	}
	if columns[1].Name != "id" {
		t.Errorf("Expected second column 'id', got '%s'", columns[1].Name)
	}
	if columns[2].Name != "after_col" {
		t.Errorf("Expected third column 'after_col', got '%s'", columns[2].Name)
	}

	// Test Change modifier
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.Text("default_string").Change()
	})
	if err != nil {
		t.Fatalf("Failed to change column type: %v", err)
	}

	columns, err = db.Schema().GetColumns(tableName)
	if err != nil {
		t.Fatalf("Failed to get columns after change: %v", err)
	}
	for _, col := range columns {
		if col.Name == "default_string" {
			if len(col.Type) == 0 {
				t.Error("Changed column should have type")
			}
		}
	}

	// Advanced Change cases
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.String("comment_col").Comment("new comment").Change()
		table.String("nullable_col").Default("new default").Change()
		table.String("after_col").First().Change()
	})
	if err != nil {
		t.Fatalf("Failed to change advanced columns: %v", err)
	}

	columns, err = db.Schema().GetColumns(tableName)
	if err != nil {
		t.Fatalf("Failed to get columns after advanced change: %v", err)
	}

	colMap = make(map[string]schema.Column)
	for _, col := range columns {
		colMap[col.Name] = col
	}

	if colMap["comment_col"].Comment != "new comment" {
		t.Errorf("Expected comment 'new comment', got '%s'", colMap["comment_col"].Comment)
	}
	if colMap["nullable_col"].Default != "new default" {
		t.Errorf("Expected default 'new default', got '%s'", colMap["nullable_col"].Default)
	}
	if len(columns) > 0 && columns[0].Name != "after_col" {
		t.Errorf("Expected first column 'after_col' after change, got '%s'", columns[0].Name)
	}

	_ = db.Schema().Drop(tableName)
}
