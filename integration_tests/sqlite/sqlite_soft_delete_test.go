package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLiteIntegrationSoftDelete tests soft delete behavior
func TestSQLiteIntegrationSoftDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
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

	if foundUser.DeletedAt == nil || foundUser.DeletedAt.IsZero() {
		t.Error("DeletedAt should be set for soft deleted user")
	}
}

// TestSQLiteIntegrationWithTrashed tests WithTrashed method
func TestSQLiteIntegrationWithTrashed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Create users
	users := []models.User{
		{Name: "with_trashed_user1", Avatar: "avatar1"},
		{Name: "with_trashed_user2", Avatar: "avatar2"},
	}
	if err := db.Query().Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete one user
	res, err := db.Query().Model(&models.User{}).Where("name = ?", "with_trashed_user1").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Without WithTrashed, should only find non-deleted users
	var activeUsers []models.User
	if err := db.Query().Model(&models.User{}).Where("name LIKE ?", "with_trashed_user%").Find(&activeUsers); err != nil {
		t.Fatalf("Failed to find active users: %v", err)
	}

	if len(activeUsers) != 1 {
		t.Errorf("Expected 1 active user, got %d", len(activeUsers))
	}

	if activeUsers[0].Name != "with_trashed_user2" {
		t.Errorf("Expected 'with_trashed_user2', got '%s'", activeUsers[0].Name)
	}

	// With WithTrashed, should find all users including deleted
	var allUsers []models.User
	if err := db.Query().Model(&models.User{}).WithTrashed().Where("name LIKE ?", "with_trashed_user%").Find(&allUsers); err != nil {
		t.Fatalf("Failed to find all users with WithTrashed: %v", err)
	}

	if len(allUsers) != 2 {
		t.Errorf("Expected 2 users with WithTrashed, got %d", len(allUsers))
	}
}

// TestSQLiteIntegrationForceDelete tests ForceDelete method
func TestSQLiteIntegrationForceDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Create a user
	user := models.User{Name: "force_delete_user", Avatar: "avatar"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	if err := db.Query().Model(&models.User{}).Where("name = ?", "force_delete_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Soft delete the user first
	res, err := db.Query().Model(&models.User{}).Where("name = ?", "force_delete_user").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to soft delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify user is soft deleted
	var softDeletedUser models.User
	if err := db.Query().Model(&models.User{}).WithTrashed().Where("id = ?", createdUser.ID).First(&softDeletedUser); err != nil {
		t.Fatalf("Failed to find soft deleted user: %v", err)
	}

	if softDeletedUser.DeletedAt == nil || softDeletedUser.DeletedAt.IsZero() {
		t.Error("User should be soft deleted")
	}

	// Force delete the user (permanent deletion)
	res, err = db.Query().Model(&models.User{}).Where("name = ?", "force_delete_user").ForceDelete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to force delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify user is permanently deleted (not found even with WithTrashed)
	var permanentlyDeletedUser models.User
	err = db.Query().Model(&models.User{}).WithTrashed().Where("id = ?", createdUser.ID).First(&permanentlyDeletedUser)
	if err == nil {
		t.Error("Expected error when finding permanently deleted user")
	}

	if permanentlyDeletedUser.ID != 0 {
		t.Error("User should be permanently deleted")
	}
}

// TestSQLiteIntegrationRestore tests Restore method
func TestSQLiteIntegrationRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Create users
	users := []models.User{
		{Name: "restore_user1", Avatar: "avatar1"},
		{Name: "restore_user2", Avatar: "avatar2"},
		{Name: "restore_user3", Avatar: "avatar3"},
		{Name: "restore_user4", Avatar: "avatar4"},
	}
	if err := db.Query().Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete all users
	res, err := db.Query().Model(&models.User{}).Where("avatar = ?", "avatar1").OrWhere("avatar = ?", "avatar2").OrWhere("avatar = ?", "avatar3").OrWhere("avatar = ?", "avatar4").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete users: %v", err)
	}

	if res.RowsAffected != 4 {
		t.Errorf("Expected 4 rows affected, got %d", res.RowsAffected)
	}

	// Restore user1 with WithTrashed
	res, err = db.Query().Model(&models.User{}).WithTrashed().Where("name = ?", "restore_user1").Restore(&models.User{})
	if err != nil {
		t.Fatalf("Failed to restore user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Restore user2 using Model method
	res, err = db.Query().Model(&models.User{}).WithTrashed().Where("name = ?", "restore_user2").Restore()
	if err != nil {
		t.Fatalf("Failed to restore user with Model: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Restore user3 using model instance
	res, err = db.Query().Model(&models.User{}).WithTrashed().Restore(&users[2])
	if err != nil {
		t.Fatalf("Failed to restore user instance: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Restore user4
	res, err = db.Query().Model(&models.User{}).WithTrashed().Restore(&users[3])
	if err != nil {
		t.Fatalf("Failed to restore user4: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify all users are restored (can be found without WithTrashed)
	var count int64
	if err := db.Query().Model(&models.User{}).Where("avatar = ?", "avatar1").OrWhere("avatar = ?", "avatar2").OrWhere("avatar = ?", "avatar3").OrWhere("avatar = ?", "avatar4").Count(&count); err != nil {
		t.Fatalf("Failed to count restored users: %v", err)
	}

	if count != 4 {
		t.Errorf("Expected 4 restored users, got %d", count)
	}
}

// TestSQLiteIntegrationOnlyTrashed tests OnlyTrashed method
func TestSQLiteIntegrationOnlyTrashed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Create users
	users := []models.User{
		{Name: "only_trashed_user1", Avatar: "avatar1"},
		{Name: "only_trashed_user2", Avatar: "avatar2"},
	}
	if err := db.Query().Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete one user
	res, err := db.Query().Model(&models.User{}).Where("name = ?", "only_trashed_user1").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Test OnlyTrashed - should only find the deleted user
	var deletedUsers []models.User
	if err := db.Query().Model(&models.User{}).OnlyTrashed().Where("name LIKE ?", "only_trashed_user%").Find(&deletedUsers); err != nil {
		t.Fatalf("Failed to find users with OnlyTrashed: %v", err)
	}

	if len(deletedUsers) != 1 {
		t.Errorf("Expected 1 deleted user, got %d", len(deletedUsers))
	}

	if deletedUsers[0].Name != "only_trashed_user1" {
		t.Errorf("Expected 'only_trashed_user1', got '%s'", deletedUsers[0].Name)
	}
}

// TestSQLiteIntegrationWithoutTrashed tests WithoutTrashed method
func TestSQLiteIntegrationWithoutTrashed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Create users
	users := []models.User{
		{Name: "without_trashed_user1", Avatar: "avatar1"},
		{Name: "without_trashed_user2", Avatar: "avatar2"},
	}
	if err := db.Query().Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete one user
	res, err := db.Query().Model(&models.User{}).Where("name = ?", "without_trashed_user1").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Test WithoutTrashed after WithTrashed
	var activeUsers []models.User
	if err := db.Query().Model(&models.User{}).WithTrashed().WithoutTrashed().Where("name LIKE ?", "without_trashed_user%").Find(&activeUsers); err != nil {
		t.Fatalf("Failed to find users with WithoutTrashed: %v", err)
	}

	if len(activeUsers) != 1 {
		t.Errorf("Expected 1 active user, got %d", len(activeUsers))
	}

	if activeUsers[0].Name != "without_trashed_user2" {
		t.Errorf("Expected 'without_trashed_user2', got '%s'", activeUsers[0].Name)
	}
}
