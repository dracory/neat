package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestTursoIntegrationQueryLogEnableQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
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

func TestTursoIntegrationQueryLogFlushQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	db.EnableQueryLog()

	var users []models.User
	_ = db.Query().Model(&models.User{}).Find(&users)

	if len(db.GetQueryLog()) == 0 {
		t.Error("Expected logs to be captured before flush")
	}

	db.FlushQueryLog()
	if len(db.GetQueryLog()) != 0 {
		t.Error("Expected logs to be cleared after flush")
	}
}

func TestTursoIntegrationQueryLogDisableQueryLog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	db.EnableQueryLog()

	var users []models.User
	_ = db.Query().Model(&models.User{}).Find(&users)

	if len(db.GetQueryLog()) == 0 {
		t.Error("Expected logs to be captured when enabled")
	}

	db.DisableQueryLog()
	db.FlushQueryLog()
	_ = db.Query().Model(&models.User{}).Find(&users)

	if len(db.GetQueryLog()) != 0 {
		t.Error("Expected no logs to be captured when disabled")
	}
}

func TestTursoIntegrationQueryLogOnSpecificQueryBuilder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	db.DisableQueryLog() // Global logging disabled

	query := db.Query()
	query.EnableQueryLog() // Local logging enabled

	var users []models.User
	_ = query.Model(&models.User{}).Find(&users)

	// Check if the query captured its own logs
	logs := query.GetQueryLog()
	if len(logs) == 0 {
		t.Error("Expected logs to be captured on the specific query builder")
	}

	// Global log should still be empty (if they were separate, but here they might be shared)
	// Actually, in our implementation, query.EnableQueryLog() only affects that query instance and its clones.
	// But they share the SAME queryLog pointer if they come from the same ORM instance.
	// Wait, db.Query() returns a clone.

	globalLogs := db.GetQueryLog()
	if len(globalLogs) == 0 {
		t.Error("Expected global logs to ALSO see the logs because they share the pointer")
	}
}
