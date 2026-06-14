package soft_delete

import (
	"testing"
	"time"
)

func TestSoftDeletesIsDeleted(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return false when DeletedAt is nil")
	}

	now := time.Now()
	sd.DeletedAt = &now

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return true when DeletedAt is set")
	}
}

func TestSoftDeletesDelete(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return false before Delete")
	}

	sd.Delete()

	if !sd.IsSoftDeleted() {
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
	if !sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return true after Delete")
	}

	// Then restore
	sd.Restore()
	if sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return false after Restore")
	}

	if sd.DeletedAt != nil {
		t.Error("Expected DeletedAt to be nil after Restore")
	}
}

func TestSoftDeletesGetDeletedAt(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.GetSoftDeletedAt() != nil {
		t.Error("Expected GetDeletedAt to return nil when DeletedAt is nil")
	}

	now := time.Now()
	sd.DeletedAt = &now

	if sd.GetSoftDeletedAt() == nil {
		t.Error("Expected GetDeletedAt to return non-nil when DeletedAt is set")
	}

	if !sd.GetSoftDeletedAt().Equal(now) {
		t.Error("Expected GetDeletedAt to return the correct timestamp")
	}
}

func TestDeletedAtColumn(t *testing.T) {
	if SoftDeleteAtColumn != "deleted_at" {
		t.Errorf("Expected SoftDeleteAtColumn to be 'deleted_at', got '%s'", SoftDeleteAtColumn)
	}
}

// TestSoftDeletesSoftDelete verifies SoftDelete() sets DeletedAt and IsDeleted() returns true.
func TestSoftDeletesSoftDelete(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return false before SoftDelete")
	}

	sd.SoftDelete()

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return true after SoftDelete")
	}
	if sd.DeletedAt == nil {
		t.Error("Expected DeletedAt to be set after SoftDelete")
	}
}

// TestSoftDeletesDeleteIsAlias verifies Delete() delegates to SoftDelete() and produces
// the same result — confirming the alias contract.
func TestSoftDeletesDeleteIsAlias(t *testing.T) {
	sd1 := &SoftDeletes{}
	sd1.SoftDelete()

	sd2 := &SoftDeletes{}
	sd2.Delete()

	if sd1.IsSoftDeleted() != sd2.IsSoftDeleted() {
		t.Error("Expected Delete() and SoftDelete() to produce the same IsDeleted() result")
	}
	if (sd1.DeletedAt == nil) != (sd2.DeletedAt == nil) {
		t.Error("Expected Delete() and SoftDelete() to both set DeletedAt")
	}
}

// ── SoftDeletedAt tests ───────────────────────────────────────────────────────

func TestSoftDeletedAtIsDeleted(t *testing.T) {
	sd := &SoftDeletedAt{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return false when SoftDeletedAt is nil")
	}

	now := time.Now()
	sd.SoftDeletedAt = &now

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return true when SoftDeletedAt is set")
	}
}

func TestSoftDeletedAtSoftDelete(t *testing.T) {
	sd := &SoftDeletedAt{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return false before SoftDelete")
	}

	sd.SoftDelete()

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return true after SoftDelete")
	}
	if sd.SoftDeletedAt == nil {
		t.Error("Expected SoftDeletedAt to be set after SoftDelete")
	}
}

// TestSoftDeletedAtDeleteIsAlias verifies Delete() delegates to SoftDelete().
func TestSoftDeletedAtDeleteIsAlias(t *testing.T) {
	sd1 := &SoftDeletedAt{}
	sd1.SoftDelete()

	sd2 := &SoftDeletedAt{}
	sd2.Delete()

	if sd1.IsSoftDeleted() != sd2.IsSoftDeleted() {
		t.Error("Expected Delete() and SoftDelete() to produce the same IsDeleted() result")
	}
	if (sd1.SoftDeletedAt == nil) != (sd2.SoftDeletedAt == nil) {
		t.Error("Expected Delete() and SoftDelete() to both set SoftDeletedAt")
	}
}

func TestSoftDeletedAtRestore(t *testing.T) {
	sd := &SoftDeletedAt{}

	sd.SoftDelete()
	if !sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return true after SoftDelete")
	}

	sd.Restore()
	if sd.IsSoftDeleted() {
		t.Error("Expected IsDeleted to return false after Restore")
	}
	if sd.SoftDeletedAt != nil {
		t.Error("Expected SoftDeletedAt to be nil after Restore")
	}
}

func TestSoftDeletedAtGetDeletedAt(t *testing.T) {
	sd := &SoftDeletedAt{}

	if sd.GetSoftDeletedAt() != nil {
		t.Error("Expected GetDeletedAt to return nil initially")
	}

	sd.SoftDelete()

	if sd.GetSoftDeletedAt() == nil {
		t.Error("Expected GetDeletedAt to return non-nil after SoftDelete")
	}
}

func TestSoftDeletedAtDeletedAtColumn(t *testing.T) {
	sd := &SoftDeletedAt{}
	if sd.SoftDeletedAtColumn() != "soft_deleted_at" {
		t.Errorf("Expected 'soft_deleted_at', got %q", sd.SoftDeletedAtColumn())
	}
}
