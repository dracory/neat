package soft_delete

import (
	"testing"
	"time"
)

func TestSoftDeletesIsDeleted(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.IsDeleted() {
		t.Error("Expected IsDeleted to return false when DeletedAt is nil")
	}

	now := time.Now()
	sd.DeletedAt = &now

	if !sd.IsDeleted() {
		t.Error("Expected IsDeleted to return true when DeletedAt is set")
	}
}

func TestSoftDeletesDelete(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.IsDeleted() {
		t.Error("Expected IsDeleted to return false before Delete")
	}

	sd.Delete()

	if !sd.IsDeleted() {
		t.Error("Expected IsDeleted to return true after Delete")
	}

	if sd.DeletedAt == nil {
		t.Error("Expected DeletedAt to be set after Delete")
	}
}

func TestSoftDeletesRestore(t *testing.T) {
	sd := &SoftDeletes{}

	// First delete
	sd.Delete()
	if !sd.IsDeleted() {
		t.Error("Expected IsDeleted to return true after Delete")
	}

	// Then restore
	sd.Restore()
	if sd.IsDeleted() {
		t.Error("Expected IsDeleted to return false after Restore")
	}

	if sd.DeletedAt != nil {
		t.Error("Expected DeletedAt to be nil after Restore")
	}
}

func TestSoftDeletesGetDeletedAt(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.GetDeletedAt() != nil {
		t.Error("Expected GetDeletedAt to return nil when DeletedAt is nil")
	}

	now := time.Now()
	sd.DeletedAt = &now

	if sd.GetDeletedAt() == nil {
		t.Error("Expected GetDeletedAt to return non-nil when DeletedAt is set")
	}

	if !sd.GetDeletedAt().Equal(now) {
		t.Error("Expected GetDeletedAt to return the correct timestamp")
	}
}

func TestDeletedAtColumn(t *testing.T) {
	if DeletedAtColumn != "deleted_at" {
		t.Errorf("Expected DeletedAtColumn to be 'deleted_at', got '%s'", DeletedAtColumn)
	}
}
