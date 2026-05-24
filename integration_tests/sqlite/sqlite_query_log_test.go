//go:build integration

package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryLogEnableQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	db.EnableQueryLog()

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
}

func TestSQLiteIntegrationQueryLogFlushQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Query logging not yet implemented in ORM instance")
}

func TestSQLiteIntegrationQueryLogDisableQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Query logging not yet implemented in ORM instance")
}

func TestSQLiteIntegrationQueryLogOnSpecificQueryBuilder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Query logging not yet implemented in ORM instance")
}
