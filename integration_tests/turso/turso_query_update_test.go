package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestTursoIntegrationQueryUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

	// Create test data
	user := models.User{Name: "update_user"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update user
	_, err = query.Model(&models.User{}).Where("id = ?", user.ID).Update(map[string]any{
		"name": "updated_user",
	})
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Verify update
	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", user.ID).First(&updatedUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.Name != "updated_user" {
		t.Errorf("Expected name 'updated_user', got '%s'", updatedUser.Name)
	}
}

func TestTursoIntegrationQueryDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

	// Create test data
	user := models.User{Name: "delete_user"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Delete user
	_, err = query.Model(&models.User{}).Where("id = ?", user.ID).Delete()
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify deletion
	var deletedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", user.ID).First(&deletedUser)
	if err == nil {
		t.Error("Should have error when finding deleted user")
	}
}
