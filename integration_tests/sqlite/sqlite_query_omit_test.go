package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryOmit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Create a test user
	user := models.User{Name: "omit_user", Avatar: "omit_avatar"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("Omit during select", func(t *testing.T) {
		t.Skip("GAP-13: Omit during select needs more work - column mapping issue")
	})

	t.Run("Omit during update", func(t *testing.T) {
		t.Skip("ORM Omit().Save() generates invalid SQL (near SET: syntax error) — not yet fixed")
	})
}
