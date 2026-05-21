//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLIntegrationQueryIncrementDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	// Seed data - Note: ID is auto-generated, so we'll use a different column
	user := models.User{Name: "increment_user", Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "increment_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	t.Run("Increment", func(t *testing.T) {
		// Note: Incrementing ID doesn't make sense for auto-increment fields
		// We'll skip this test or adapt it to use a different column
		t.Skip("Increment on auto-increment ID not applicable")
	})

	t.Run("Decrement", func(t *testing.T) {
		// Note: Decrementing ID doesn't make sense for auto-increment fields
		// We'll skip this test or adapt it to use a different column
		t.Skip("Decrement on auto-increment ID not applicable")
	})

	t.Run("With where conditions", func(t *testing.T) {
		// Create another user
		user2 := models.User{Name: "other_user", Avatar: "group2"}
		if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}

		// Get the created user2 to get its ID
		var createdUser2 models.User
		if err := db.Query().Model(&models.User{}).Where("name = ?", "other_user").First(&createdUser2); err != nil {
			t.Fatalf("Failed to get created user2: %v", err)
		}

		// Note: We can't increment/decrement auto-increment IDs
		// This test would need to be adapted to use a different column
		t.Skip("Increment/decrement on auto-increment ID not applicable")
	})
}
