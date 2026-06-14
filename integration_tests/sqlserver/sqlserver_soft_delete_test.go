package sqlserver_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLServerIntegrationSoftDelete verifies that deleting a model with a
// soft_deleted_at column performs a soft delete: the row is hidden from normal
// queries but visible when WithSoftDeleted() is used, and SoftDeletedAt is non-nil.
func TestSQLServerIntegrationSoftDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	user := models.User{Name: "soft_delete_user", Avatar: "avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "soft_delete_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	res, err := query.Model(&models.User{}).Where("name = ?", "soft_delete_user").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var notFoundUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&notFoundUser)
	if err == nil {
		t.Error("Expected error when finding soft deleted user without WithTrashed")
	}

	if notFoundUser.ID != 0 {
		t.Error("User should not be found without WithTrashed")
	}

	var foundUser models.User
	err = query.Model(&models.User{}).WithSoftDeleted().Where("id = ?", createdUser.ID).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find soft deleted user with WithTrashed: %v", err)
	}

	if foundUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, foundUser.ID)
	}

	if foundUser.SoftDeletedAt == nil {
		t.Error("DeletedAt should be set for soft deleted user")
	}
}

// TestSQLServerIntegrationWithTrashed verifies that after soft-deleting one of
// two users, a normal Find() returns only the active user, while WithTrashed()
// returns both the active and the soft-deleted user.
func TestSQLServerIntegrationWithTrashed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "with_trashed_user1", Avatar: "avatar1"},
		{Name: "with_trashed_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "with_trashed_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	res, err := query.Model(&models.User{}).Where("name = ?", "with_trashed_user1").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var activeUsers []models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "with_trashed_user%").Find(&activeUsers)
	if err != nil {
		t.Fatalf("Failed to find active users: %v", err)
	}

	if len(activeUsers) != 1 {
		t.Errorf("Expected 1 active user, got %d", len(activeUsers))
	}

	var allUsers []models.User
	err = query.Model(&models.User{}).WithSoftDeleted().Where("name LIKE ?", "with_trashed_user%").Find(&allUsers)
	if err != nil {
		t.Fatalf("Failed to find all users with WithTrashed: %v", err)
	}

	if len(allUsers) != 2 {
		t.Errorf("Expected 2 users with WithTrashed, got %d", len(allUsers))
	}
}
