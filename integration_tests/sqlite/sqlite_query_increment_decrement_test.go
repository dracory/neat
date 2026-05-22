package sqlite

import (
	"testing"
)

func TestSQLiteIntegrationQueryIncrementDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Increment", func(t *testing.T) {
		t.Skip("ORM Increment() has SQL build bug (missing argument) and tests require explicit IDs incompatible with SQLite AUTOINCREMENT")
	})

	t.Run("Increment by amount", func(t *testing.T) {
		t.Skip("ORM Increment() not yet working correctly for SQLite")
	})

	t.Run("Decrement", func(t *testing.T) {
		t.Skip("ORM Decrement() not yet working correctly for SQLite")
	})

	t.Run("Decrement by amount", func(t *testing.T) {
		t.Skip("ORM Decrement() not yet working correctly for SQLite")
	})

	t.Run("With where conditions", func(t *testing.T) {
		t.Skip("ORM Increment() not yet working correctly for SQLite")
	})

	t.Run("Increment with extra columns", func(t *testing.T) {
		t.Skip("ORM Increment() not yet working correctly for SQLite")
	})

	t.Run("Invalid column", func(t *testing.T) {
		t.Skip("ORM column-name validation not yet implemented for Increment()")
	})
}
