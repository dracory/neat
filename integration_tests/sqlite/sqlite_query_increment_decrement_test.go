package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryIncrementDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Create a test user
	user := models.User{Name: "increment_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("Increment", func(t *testing.T) {
		// For now, skip since User model doesn't have a numeric column to increment
		t.Skip("User model needs a numeric column for increment testing")
	})
}
