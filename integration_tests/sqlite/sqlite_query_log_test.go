//go:build integration

package sqlite

import (
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
		err := db.Query().Model(&models.User{}).Find(&users)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		logs := db.GetQueryLog()
		if len(logs) == 0 {
			t.Error("Expected logs to be captured")
		}
		if len(logs) > 0 {
			logContent := logs[0].Query
			if len(logContent) == 0 {
				t.Error("Log query should not be empty")
			}
		}
	})

	t.Run("FlushQueryLog", func(t *testing.T) {
		t.Skip("Query logging not yet implemented in ORM instance")
	})

	t.Run("DisableQueryLog", func(t *testing.T) {
		t.Skip("Query logging not yet implemented in ORM instance")
	})

	t.Run("Query Log on specific query builder", func(t *testing.T) {
		t.Skip("Query logging not yet implemented in ORM instance")
	})
}
