package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLiteIntegrationQueryFind tests Find operations
func TestSQLiteIntegrationQueryFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Create test user
	user := models.User{Name: "find_user", Avatar: "find_avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "find_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	t.Run("find by ID", func(t *testing.T) {
		var foundUser models.User
		err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&foundUser)
		if err != nil {
			t.Errorf("Find by ID failed: %v", err)
		}
		if foundUser.ID != createdUser.ID {
			t.Errorf("Expected ID %d, got %d", createdUser.ID, foundUser.ID)
		}
		if foundUser.Name != "find_user" {
			t.Errorf("Expected name 'find_user', got '%s'", foundUser.Name)
		}
	})

	t.Run("find with where", func(t *testing.T) {
		var foundUser models.User
		err := query.Model(&models.User{}).Where("name = ?", "find_user").First(&foundUser)
		if err != nil {
			t.Errorf("Find with where failed: %v", err)
		}
		if foundUser.Name != "find_user" {
			t.Errorf("Expected name 'find_user', got '%s'", foundUser.Name)
		}
	})

	t.Run("find with conditions", func(t *testing.T) {
		var foundUser models.User
		err := query.Model(&models.User{}).Where("name = ?", "find_user").Where("avatar = ?", "find_avatar").First(&foundUser)
		if err != nil {
			t.Errorf("Find with conditions failed: %v", err)
		}
		if foundUser.Name != "find_user" {
			t.Errorf("Expected name 'find_user', got '%s'", foundUser.Name)
		}
	})
}
