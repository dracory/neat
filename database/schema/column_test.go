package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumnDefinition_Change(t *testing.T) {
	column := &ColumnDefinition{}

	result := column.Change()

	assert.True(t, column.GetChange(), "Change should set change flag to true")
	assert.Same(t, column, result, "Change should return the same instance for chaining")
}

func TestColumnDefinition_GetChange(t *testing.T) {
	t.Run("when change is not set", func(t *testing.T) {
		column := &ColumnDefinition{}
		assert.False(t, column.GetChange(), "GetChange should return false when change is not set")
	})

	t.Run("when change is set to true", func(t *testing.T) {
		column := &ColumnDefinition{}
		column.Change()
		assert.True(t, column.GetChange(), "GetChange should return true after calling Change")
	})
}
