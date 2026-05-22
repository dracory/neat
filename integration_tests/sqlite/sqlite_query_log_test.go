package sqlite

import (
	"testing"
)

func TestSQLiteIntegrationQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("EnableQueryLog and capture queries", func(t *testing.T) {
		t.Skip("database.Database does not expose EnableQueryLog/GetQueryLog — query logging only available on Query instances")
	})

	t.Run("FlushQueryLog", func(t *testing.T) {
		t.Skip("database.Database does not expose FlushQueryLog — not yet implemented at DB level")
	})

	t.Run("DisableQueryLog", func(t *testing.T) {
		t.Skip("database.Database does not expose DisableQueryLog — not yet implemented at DB level")
	})

	t.Run("Query Log with bindings", func(t *testing.T) {
		t.Skip("database.Database does not expose EnableQueryLog — not yet implemented at DB level")
	})

	t.Run("Query Log on specific query builder", func(t *testing.T) {
		t.Skip("Query.EnableQueryLog/GetQueryLog on query builder — not yet implemented")
	})
}
