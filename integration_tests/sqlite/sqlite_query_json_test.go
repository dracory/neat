package sqlite

import (
	"testing"
)

func TestSQLiteIntegrationQueryJson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("WhereJsonContains", func(t *testing.T) {
		t.Skip("ORM WhereJsonContains() does not generate valid SQLite JSON path SQL — not yet implemented")
	})

	t.Run("OrWhereJsonContains", func(t *testing.T) {
		t.Skip("ORM OrWhereJsonContains() does not generate valid SQLite JSON path SQL — not yet implemented")
	})

	t.Run("WhereJsonDoesntContain", func(t *testing.T) {
		t.Skip("ORM WhereJsonDoesntContain() does not generate valid SQLite JSON path SQL — not yet implemented")
	})

	t.Run("WhereJsonContainsKey", func(t *testing.T) {
		t.Skip("ORM WhereJsonContainsKey() does not generate valid SQLite JSON path SQL — not yet implemented")
	})

	t.Run("WhereJsonDoesntContainKey", func(t *testing.T) {
		t.Skip("ORM WhereJsonDoesntContainKey() does not generate valid SQLite JSON path SQL — not yet implemented")
	})

	t.Run("WhereJsonLength", func(t *testing.T) {
		t.Skip("ORM WhereJsonLength() does not generate valid SQLite JSON path SQL — not yet implemented")
	})

	t.Run("Update with JSON path", func(t *testing.T) {
		t.Skip("ORM Update() with JSON path does not generate valid SQLite JSON_SET syntax — not yet implemented")
	})
}
