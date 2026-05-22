package sqlite

import (
	"testing"
)

func TestSQLiteLockForUpdate(t *testing.T) {
	t.Skip("ORM LockForUpdate() generates 'FOR UPDATE' syntax not supported by SQLite")
}

func TestSQLiteSharedLock(t *testing.T) {
	t.Skip("ORM SharedLock() generates 'LOCK IN SHARE MODE' syntax not supported by SQLite")
}
