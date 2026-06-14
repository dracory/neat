package soft_delete

import (
	"testing"
	"time"
)

// ── SoftDeletes tests (default: soft_deleted_at) ─────────────────────────────

func TestSoftDeletesIsDeleted(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false when SoftDeletedAt is nil")
	}

	now := time.Now()
	sd.SoftDeletedAt = &now

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true when SoftDeletedAt is set")
	}
}

func TestSoftDeletesDelete(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false before Delete")
	}

	sd.Delete()

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true after Delete")
	}

	if sd.SoftDeletedAt == nil {
		t.Error("Expected SoftDeletedAt to be set after Delete")
	}
}

func TestSoftDeletesRestore(t *testing.T) {
	sd := &SoftDeletes{}

	sd.Delete()
	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true after Delete")
	}

	sd.RestoreSoftDeleted()
	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false after RestoreSoftDeleted")
	}

	if sd.SoftDeletedAt != nil {
		t.Error("Expected SoftDeletedAt to be nil after RestoreSoftDeleted")
	}
}

func TestSoftDeletesGetDeletedAt(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.GetSoftDeletedAt() != nil {
		t.Error("Expected GetSoftDeletedAt to return nil when SoftDeletedAt is nil")
	}

	now := time.Now()
	sd.SoftDeletedAt = &now

	if sd.GetSoftDeletedAt() == nil {
		t.Error("Expected GetSoftDeletedAt to return non-nil when SoftDeletedAt is set")
	}

	if !sd.GetSoftDeletedAt().Equal(now) {
		t.Error("Expected GetSoftDeletedAt to return the correct timestamp")
	}
}

func TestSoftDeleteAtColumn(t *testing.T) {
	if SoftDeleteAtColumn != "soft_deleted_at" {
		t.Errorf("Expected SoftDeleteAtColumn to be 'soft_deleted_at', got '%s'", SoftDeleteAtColumn)
	}
}

func TestDeletedAtColumnName(t *testing.T) {
	if DeletedAtColumnName != "deleted_at" {
		t.Errorf("Expected DeletedAtColumnName to be 'deleted_at', got '%s'", DeletedAtColumnName)
	}
}

// TestSoftDeletesSoftDelete verifies SoftDelete() sets SoftDeletedAt and IsSoftDeleted() returns true.
func TestSoftDeletesSoftDelete(t *testing.T) {
	sd := &SoftDeletes{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false before SoftDelete")
	}

	sd.SoftDelete()

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true after SoftDelete")
	}
	if sd.SoftDeletedAt == nil {
		t.Error("Expected SoftDeletedAt to be set after SoftDelete")
	}
}

// TestSoftDeletesDeleteIsAlias verifies Delete() delegates to SoftDelete().
func TestSoftDeletesDeleteIsAlias(t *testing.T) {
	sd1 := &SoftDeletes{}
	sd1.SoftDelete()

	sd2 := &SoftDeletes{}
	sd2.Delete()

	if sd1.IsSoftDeleted() != sd2.IsSoftDeleted() {
		t.Error("Expected Delete() and SoftDelete() to produce the same IsSoftDeleted() result")
	}
	if (sd1.SoftDeletedAt == nil) != (sd2.SoftDeletedAt == nil) {
		t.Error("Expected Delete() and SoftDelete() to both set SoftDeletedAt")
	}
}

func TestSoftDeletesSoftDeletedAtColumn(t *testing.T) {
	sd := &SoftDeletes{}
	if sd.SoftDeletedAtColumn() != "soft_deleted_at" {
		t.Errorf("Expected 'soft_deleted_at', got %q", sd.SoftDeletedAtColumn())
	}
}

// ── SoftDeletedAt tests (explicit soft_deleted_at embed) ─────────────────────

func TestSoftDeletedAtIsDeleted(t *testing.T) {
	sd := &SoftDeletedAt{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false when SoftDeletedAt is nil")
	}

	now := time.Now()
	sd.SoftDeletedAt = &now

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true when SoftDeletedAt is set")
	}
}

func TestSoftDeletedAtSoftDelete(t *testing.T) {
	sd := &SoftDeletedAt{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false before SoftDelete")
	}

	sd.SoftDelete()

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true after SoftDelete")
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
		t.Error("Expected Delete() and SoftDelete() to produce the same IsSoftDeleted() result")
	}
	if (sd1.SoftDeletedAt == nil) != (sd2.SoftDeletedAt == nil) {
		t.Error("Expected Delete() and SoftDelete() to both set SoftDeletedAt")
	}
}

func TestSoftDeletedAtRestore(t *testing.T) {
	sd := &SoftDeletedAt{}

	sd.SoftDelete()
	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true after SoftDelete")
	}

	sd.RestoreSoftDeleted()
	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false after RestoreSoftDeleted")
	}
	if sd.SoftDeletedAt != nil {
		t.Error("Expected SoftDeletedAt to be nil after RestoreSoftDeleted")
	}
}

func TestSoftDeletedAtGetDeletedAt(t *testing.T) {
	sd := &SoftDeletedAt{}

	if sd.GetSoftDeletedAt() != nil {
		t.Error("Expected GetSoftDeletedAt to return nil initially")
	}

	sd.SoftDelete()

	if sd.GetSoftDeletedAt() == nil {
		t.Error("Expected GetSoftDeletedAt to return non-nil after SoftDelete")
	}
}

func TestSoftDeletedAtSoftDeletedAtColumn(t *testing.T) {
	sd := &SoftDeletedAt{}
	if sd.SoftDeletedAtColumn() != "soft_deleted_at" {
		t.Errorf("Expected 'soft_deleted_at', got %q", sd.SoftDeletedAtColumn())
	}
}

// ── DeletedAt tests (Laravel-compatible: deleted_at) ─────────────────────────

func TestDeletedAtIsDeleted(t *testing.T) {
	sd := &DeletedAt{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false when DeletedAt is nil")
	}

	now := time.Now()
	sd.DeletedAt = &now

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true when DeletedAt is set")
	}
}

func TestDeletedAtSoftDelete(t *testing.T) {
	sd := &DeletedAt{}

	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false before SoftDelete")
	}

	sd.SoftDelete()

	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true after SoftDelete")
	}
	if sd.DeletedAt == nil {
		t.Error("Expected DeletedAt to be set after SoftDelete")
	}
}

func TestDeletedAtDeleteIsAlias(t *testing.T) {
	sd1 := &DeletedAt{}
	sd1.SoftDelete()

	sd2 := &DeletedAt{}
	sd2.Delete()

	if sd1.IsSoftDeleted() != sd2.IsSoftDeleted() {
		t.Error("Expected Delete() and SoftDelete() to produce the same IsSoftDeleted() result")
	}
	if (sd1.DeletedAt == nil) != (sd2.DeletedAt == nil) {
		t.Error("Expected Delete() and SoftDelete() to both set DeletedAt")
	}
}

func TestDeletedAtRestore(t *testing.T) {
	sd := &DeletedAt{}

	sd.SoftDelete()
	if !sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return true after SoftDelete")
	}

	sd.RestoreSoftDeleted()
	if sd.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to return false after RestoreSoftDeleted")
	}
	if sd.DeletedAt != nil {
		t.Error("Expected DeletedAt to be nil after RestoreSoftDeleted")
	}
}

func TestDeletedAtGetSoftDeletedAt(t *testing.T) {
	sd := &DeletedAt{}

	if sd.GetSoftDeletedAt() != nil {
		t.Error("Expected GetSoftDeletedAt to return nil initially")
	}

	sd.SoftDelete()

	if sd.GetSoftDeletedAt() == nil {
		t.Error("Expected GetSoftDeletedAt to return non-nil after SoftDelete")
	}
}

func TestDeletedAtSoftDeletedAtColumn(t *testing.T) {
	sd := &DeletedAt{}
	if sd.SoftDeletedAtColumn() != "deleted_at" {
		t.Errorf("Expected 'deleted_at', got %q", sd.SoftDeletedAtColumn())
	}
}
