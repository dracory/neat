//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationQueryLogEnableAndCapture verifies that query logging
// captures executed SQL statements.
func TestAztablesIntegrationQueryLogEnableAndCapture(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	db.EnableQueryLog()

	var users []AzUser
	err := db.Query().Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Find(&users)
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}

	logs := db.GetQueryLog()
	if len(logs) == 0 {
		t.Error("Expected logs to be captured")
	}
	if len(logs) > 0 {
		if logs[0].Query == "" {
			t.Error("Log query should not be empty")
		}
	}
}

// TestAztablesIntegrationQueryLogFlush verifies that FlushQueryLog clears
// captured logs.
func TestAztablesIntegrationQueryLogFlush(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	db.EnableQueryLog()

	var users []AzUser
	_ = db.Query().Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Find(&users)

	if len(db.GetQueryLog()) == 0 {
		t.Error("Expected logs before flush")
	}

	db.FlushQueryLog()
	if len(db.GetQueryLog()) != 0 {
		t.Error("Expected no logs after flush")
	}
}
