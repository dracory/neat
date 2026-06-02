package postgres_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestPostgreSQLIntegrationSoftDelete tests soft delete behavior
func TestPostgreSQLIntegrationSoftDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create a user
	user := models.User{Name: "soft_delete_user", Avatar: "avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "soft_delete_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Soft delete the user
	res, err := query.Model(&models.User{}).Where("name = ?", "soft_delete_user").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify the user is not found without WithTrashed
	var notFoundUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&notFoundUser)
	if err == nil {
		t.Error("Expected error when finding soft deleted user without WithTrashed")
	}

	if notFoundUser.ID != 0 {
		t.Error("User should not be found without WithTrashed")
	}

	// Verify the user is found with WithTrashed
	var foundUser models.User
	err = query.Model(&models.User{}).WithTrashed().Where("id = ?", createdUser.ID).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find soft deleted user with WithTrashed: %v", err)
	}

	if foundUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, foundUser.ID)
	}

	if foundUser.DeletedAt.IsZero() {
		t.Error("DeletedAt should be set for soft deleted user")
	}
}

// TestPostgreSQLIntegrationWithTrashed tests WithTrashed method
func TestPostgreSQLIntegrationWithTrashed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "with_trashed_user1", Avatar: "avatar1"},
		{Name: "with_trashed_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete one user
	res, err := query.Model(&models.User{}).Where("name = ?", "with_trashed_user1").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Without WithTrashed, should only find non-deleted users
	var activeUsers []models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "with_trashed_user%").Find(&activeUsers)
	if err != nil {
		t.Fatalf("Failed to find active users: %v", err)
	}

	if len(activeUsers) != 1 {
		t.Errorf("Expected 1 active user, got %d", len(activeUsers))
	}

	if len(activeUsers) >= 1 && activeUsers[0].Name != "with_trashed_user2" {
		t.Errorf("Expected 'with_trashed_user2', got '%s'", activeUsers[0].Name)
	}

	// With WithTrashed, should find all users including deleted
	var allUsers []models.User
	err = query.Model(&models.User{}).WithTrashed().Where("name LIKE ?", "with_trashed_user%").Find(&allUsers)
	if err != nil {
		t.Fatalf("Failed to find all users with WithTrashed: %v", err)
	}

	if len(allUsers) != 2 {
		t.Errorf("Expected 2 users with WithTrashed, got %d", len(allUsers))
	}
}
