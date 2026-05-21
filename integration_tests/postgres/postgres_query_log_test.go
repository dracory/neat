//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	t.Run("EnableQueryLog and capture queries", func(t *testing.T) {
		db.EnableQueryLog()

		// Run some queries
		var users []models.User
		err := db.Query().Model(&models.User{}).Get(&users)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}

		logs := db.GetQueryLog()
		if len(logs) == 0 {
			t.Error("Expected non-empty logs")
		}
		if len(logs) > 0 {
			// Check if it contains SELECT
			selectFound := false
			for _, log := range logs {
				if len(log.Query) > 0 && log.Query[0:6] == "SELECT" {
					selectFound = true
					break
				}
			}
			if !selectFound {
				t.Error("Expected logs to contain SELECT query")
			}
		}
	})

	t.Run("FlushQueryLog", func(t *testing.T) {
		db.EnableQueryLog()

		var users []models.User
		_ = db.Query().Model(&models.User{}).Get(&users)

		if len(db.GetQueryLog()) == 0 {
			t.Error("Expected non-empty logs before flush")
		}

		db.FlushQueryLog()
		if len(db.GetQueryLog()) != 0 {
			t.Error("Expected empty logs after flush")
		}
	})
}
