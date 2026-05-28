package schema

import (
	"testing"
)

func TestColumnDefinition_Change(t *testing.T) {
	column := &ColumnDefinition{}

	result := column.Change()

	if !column.GetChange() {
		t.Error("Change should set change flag to true")
	}
	if result != column {
		t.Error("Change should return the same instance for chaining")
	}
}

func TestColumnDefinition_GetChange(t *testing.T) {
	t.Run("when change is not set", func(t *testing.T) {
		column := &ColumnDefinition{}
		if column.GetChange() {
			t.Error("GetChange should return false when change is not set")
		}
	})

	t.Run("when change is set to true", func(t *testing.T) {
		column := &ColumnDefinition{}
		column.Change()
		if !column.GetChange() {
			t.Error("GetChange should return true after calling Change")
		}
	})
}
