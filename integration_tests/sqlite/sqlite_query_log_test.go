package sqlite

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	t.Run("EnableQueryLog and capture queries", func(t *testing.T) {
		db.EnableQueryLog()

		// Run some queries
		var users []models.User
		err := db.Query().Get(&users)
		if err != nil {
			t.Errorf("Get query failed: %v", err)
		}

		err = db.Query().Table("users").Where("id = ?", 1).First(&models.User{})
		if err != nil {
			t.Errorf("First query failed: %v", err)
		}

		logs := db.GetQueryLog()
		if len(logs) != 2 {
			t.Errorf("Expected 2 log entries, got %d", len(logs))
		}
		if !strings.Contains(logs[0].Query, "SELECT") {
			t.Error("Expected SELECT in query log")
		}
		if !strings.Contains(logs[0].Query, "users") {
			t.Error("Expected 'users' in query log")
		}
		if logs[0].Time <= 0 {
			t.Error("Expected positive time in log")
		}

		if !strings.Contains(logs[1].Query, "SELECT") {
			t.Error("Expected SELECT in second query log")
		}
		if !strings.Contains(logs[1].Query, "users") {
			t.Error("Expected 'users' in second query log")
		}
		if logs[1].Time <= 0 {
			t.Error("Expected positive time in second log")
		}
	})

	t.Run("FlushQueryLog", func(t *testing.T) {
		db.EnableQueryLog()

		var users []models.User
		_ = db.Query().Get(&users)

		if len(db.GetQueryLog()) == 0 {
			t.Error("Expected non-empty query log")
		}

		db.FlushQueryLog()
		if len(db.GetQueryLog()) != 0 {
			t.Error("Expected empty query log after flush")
		}
	})

	t.Run("DisableQueryLog", func(t *testing.T) {
		db.FlushQueryLog()
		db.DisableQueryLog()

		var users []models.User
		_ = db.Query().Get(&users)

		if len(db.GetQueryLog()) != 0 {
			t.Error("Expected empty query log when disabled")
		}
	})

	t.Run("Query Log with bindings", func(t *testing.T) {
		db.EnableQueryLog()
		db.FlushQueryLog()

		_ = db.Query().Table("users").Where("name = ?", "test").First(&models.User{})

		logs := db.GetQueryLog()
		if len(logs) != 1 {
			t.Errorf("Expected 1 log entry, got %d", len(logs))
		}
		// GORM's Trace might return the SQL with values already interpolated or as a raw string
		// depending on how it's called.
		if !strings.Contains(logs[0].Query, "test") {
			t.Error("Expected 'test' in query log")
		}
	})

	t.Run("Query Log on specific query builder", func(t *testing.T) {
		// Fresh query builder
		query := db.Query()
		query.EnableQueryLog()

		var users []models.User
		_ = query.Get(&users)

		if len(query.GetQueryLog()) == 0 {
			t.Error("Expected non-empty query log on query builder")
		}

		// db instance still has its own context/state
		// In our implementation, db.Query() might share the same base context if not using WithContext
		// Let's check if the main db instance captured it.
		// If db.ormInstance.query and the one returned by db.Query() share the same r.ctx pointer,
		// they will share the same log slice.
	})
}
