package query

import (
	"testing"
	"time"
)

func TestHasSoftDeleteCapability(t *testing.T) {
	type ModelWithSoftDelete struct {
		DeletedAt *time.Time
	}

	type ModelWithoutSoftDelete struct {
		Name string
	}

	// Test model with soft delete capability
	modelWithDelete := &ModelWithSoftDelete{}
	if !hasSoftDeleteCapability(modelWithDelete) {
		t.Error("Expected model with DeletedAt field to have soft delete capability")
	}

	// Test model without soft delete capability
	modelWithoutDelete := &ModelWithoutSoftDelete{}
	if hasSoftDeleteCapability(modelWithoutDelete) {
		t.Error("Expected model without DeletedAt field to not have soft delete capability")
	}

	// Test nil model
	if hasSoftDeleteCapability(nil) {
		t.Error("Expected nil model to not have soft delete capability")
	}
}
