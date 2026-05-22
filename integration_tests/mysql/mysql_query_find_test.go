//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestMySQLIntegrationQueryFind tests Find operations
func TestMySQLIntegrationQueryFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	t.Run("find by ID", func(t *testing.T) {
		user := models.User{Name: "find_user_id"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		// Get the created user to get its ID
		var createdUser models.User
		if err := query.Model(&models.User{}).Where("name = ?", "find_user_id").First(&createdUser); err != nil {
			t.Fatalf("Failed to get created user: %v", err)
		}

		var foundUser models.User
		err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&foundUser)
		if err != nil {
			t.Errorf("Find by ID failed: %v", err)
		}
		if foundUser.ID != createdUser.ID {
			t.Errorf("Expected ID %d, got %d", createdUser.ID, foundUser.ID)
		}
		if foundUser.Name != "find_user_id" {
			t.Errorf("Expected name 'find_user_id', got '%s'", foundUser.Name)
		}
	})

	t.Run("find with where", func(t *testing.T) {
		user := models.User{Name: "find_user_where"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		var foundUser models.User
		err := query.Model(&models.User{}).Where("name = ?", "find_user_where").First(&foundUser)
		if err != nil {
			t.Errorf("Find with where failed: %v", err)
		}
		if foundUser.Name != "find_user_where" {
			t.Errorf("Expected name 'find_user_where', got '%s'", foundUser.Name)
		}
	})

	t.Run("find with conditions", func(t *testing.T) {
		user1 := models.User{Name: "find_user_cond_1", Avatar: "avatar1"}
		user2 := models.User{Name: "find_user_cond_2", Avatar: "avatar2"}
		if err := query.Model(&models.User{}).Create(&user1); err != nil {
			t.Fatalf("Failed to create test user 1: %v", err)
		}
		if err := query.Model(&models.User{}).Create(&user2); err != nil {
			t.Fatalf("Failed to create test user 2: %v", err)
		}

		var users []models.User
		err := query.Model(&models.User{}).Where("avatar = ?", "avatar1").Find(&users)
		if err != nil {
			t.Errorf("Find with conditions failed: %v", err)
		}
		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}
	})
}
