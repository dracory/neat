package arraysource

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rows := []map[string]any{
			{"id": 1, "name": "Alice"},
		}
		model := New(rows)
		if model.TableName() == "" {
			t.Error("expected table name to be auto-generated")
		}
		gotRows, _ := model.Rows()
		if len(gotRows) != 1 || gotRows[0]["name"] != "Alice" {
			t.Errorf("expected rows to be copy, got %v", gotRows)
		}
	})

	t.Run("shallow copy", func(t *testing.T) {
		rows := []map[string]any{
			{"id": 1, "name": "Alice"},
		}
		model := New(rows)
		rows[0]["name"] = "Bob"
		gotRows, _ := model.Rows()
		if gotRows[0]["name"] != "Alice" {
			t.Errorf("expected snapshot semantics, got modified value: %v", gotRows[0]["name"])
		}
	})

	t.Run("panic on empty", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic on empty rows")
			}
		}()
		New(nil)
	})
}

func TestNewWithSchema(t *testing.T) {
	rows := []map[string]any{
		{"id": 1, "name": "Alice"},
	}
	schema := map[string]string{
		"id":   "int",
		"name": "string",
	}
	model := NewWithSchema(rows, schema)
	if model.Schema()["id"] != "int" {
		t.Errorf("expected schema to be set, got %v", model.Schema())
	}
}

func TestTable(t *testing.T) {
	model := New([]map[string]any{{"id": 1}})
	model.Table("custom_table")
	if model.TableName() != "custom_table" {
		t.Errorf("expected custom_table, got %s", model.TableName())
	}
}
