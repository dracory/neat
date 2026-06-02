package mysql_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestMysqlIntegrationQueryLogEnableAndCapture(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

func TestMysqlIntegrationQueryLogFlush(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	db.EnableQueryLog()

	var users []models.User
	_ = db.Query().Model(&models.User{}).Find(&users)

	if len(db.GetQueryLog()) == 0 {
		t.Error("Expected logs before flush")
	}

	db.FlushQueryLog()
	if len(db.GetQueryLog()) != 0 {
		t.Error("Expected no logs after flush")
	}
}
