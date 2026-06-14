package mysql_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestMySQLIntegrationSoftDelete tests soft delete behavior
func TestMySQLIntegrationSoftDelete(t *testing.T) {

	db := SetupMySQLTest(t)
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

	if foundUser.DeletedAt == nil {
		t.Error("DeletedAt should be set for soft deleted user")
	}
}

// TestMySQLIntegrationWithTrashed tests WithTrashed method
func TestMySQLIntegrationWithTrashed(t *testing.T) {

	db := SetupMySQLTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "with_trashed_user1", Avatar: "avatar1"},
		{Name: "with_trashed_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "with_trashed_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
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
		return
	}

	if activeUsers[0].Name != "with_trashed_user2" {
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

// TestMySQLIntegrationForceDelete tests ForceDelete method
func TestMySQLIntegrationForceDelete(t *testing.T) {

	db := SetupMySQLTest(t)
	query := db.Query()

	// Create a user
	user := models.User{Name: "force_delete_user", Avatar: "avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "force_delete_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Soft delete the user first
	res, err := query.Model(&models.User{}).Where("name = ?", "force_delete_user").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to soft delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify user is soft deleted
	var softDeletedUser models.User
	err = query.Model(&models.User{}).WithTrashed().Where("id = ?", createdUser.ID).First(&softDeletedUser)
	if err != nil {
		t.Fatalf("Failed to find soft deleted user: %v", err)
	}

	if softDeletedUser.DeletedAt == nil {
		t.Error("User should be soft deleted")
	}

	// Force delete the user (permanent deletion)
	res, err = query.Model(&models.User{}).Where("name = ?", "force_delete_user").ForceDelete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to force delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify user is permanently deleted (not found even with WithTrashed)
	var permanentlyDeletedUser models.User
	err = query.Model(&models.User{}).WithTrashed().Where("id = ?", createdUser.ID).First(&permanentlyDeletedUser)
	if err == nil {
		t.Error("Expected error when finding permanently deleted user")
	}

	if permanentlyDeletedUser.ID != 0 {
		t.Error("User should be permanently deleted")
	}
}

// TestMySQLIntegrationRestore tests Restore method
func TestMySQLIntegrationRestore(t *testing.T) {

	db := SetupMySQLTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "restore_user1", Avatar: "avatar1"},
		{Name: "restore_user2", Avatar: "avatar2"},
		{Name: "restore_user3", Avatar: "avatar3"},
		{Name: "restore_user4", Avatar: "avatar4"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Fetch users from database to get actual IDs
	var fetchedUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "restore_user%").Find(&fetchedUsers); err != nil {
		t.Fatalf("Failed to fetch users: %v", err)
	}
	if len(fetchedUsers) != 4 {
		t.Fatalf("Expected 4 users, got %d", len(fetchedUsers))
	}
	users = fetchedUsers

	// Soft delete all users
	res, err := query.Model(&models.User{}).Where("avatar = ?", "avatar1").OrWhere("avatar = ?", "avatar2").OrWhere("avatar = ?", "avatar3").OrWhere("avatar = ?", "avatar4").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete users: %v", err)
	}

	if res.RowsAffected != 4 {
		t.Errorf("Expected 4 rows affected, got %d", res.RowsAffected)
	}

	// Restore user1 with WithTrashed
	res, err = query.Model(&models.User{}).WithTrashed().Where("name = ?", "restore_user1").RestoreSoftDeleted(&models.User{})
	if err != nil {
		t.Fatalf("Failed to restore user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Restore user2 using Model method
	res, err = query.Model(&models.User{}).WithTrashed().Where("name = ?", "restore_user2").RestoreSoftDeleted()
	if err != nil {
		t.Fatalf("Failed to restore user with Model: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Restore user3 using model instance
	res, err = query.Model(&models.User{}).WithTrashed().RestoreSoftDeleted(&users[2])
	if err != nil {
		t.Fatalf("Failed to restore user instance: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Restore user4
	res, err = query.Model(&models.User{}).WithTrashed().RestoreSoftDeleted(&users[3])
	if err != nil {
		t.Fatalf("Failed to restore user4: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify all users are restored (can be found without WithTrashed)
	var count int64
	err = query.Model(&models.User{}).Where("avatar = ?", "avatar1").OrWhere("avatar = ?", "avatar2").OrWhere("avatar = ?", "avatar3").OrWhere("avatar = ?", "avatar4").Count(&count)
	if err != nil {
		t.Fatalf("Failed to count restored users: %v", err)
	}

	if count != 4 {
		t.Errorf("Expected 4 restored users, got %d", count)
	}
}

// TestMySQLIntegrationSoftDeleteWithConditions tests soft delete with where conditions
func TestMySQLIntegrationSoftDeleteWithConditions(t *testing.T) {

	db := SetupMySQLTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "soft_delete_cond_user1", Avatar: "avatar1"},
		{Name: "soft_delete_cond_user2", Avatar: "avatar1"},
		{Name: "soft_delete_cond_user3", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete users with avatar1
	res, err := query.Model(&models.User{}).Where("avatar = ?", "avatar1").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete users: %v", err)
	}

	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	// Verify only users with avatar1 are soft deleted
	var remainingUsers []models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "soft_delete_cond_user%").Find(&remainingUsers)
	if err != nil {
		t.Fatalf("Failed to find remaining users: %v", err)
	}

	if len(remainingUsers) != 1 {
		t.Errorf("Expected 1 remaining user, got %d", len(remainingUsers))
		return
	}

	if remainingUsers[0].Name != "soft_delete_cond_user3" {
		t.Errorf("Expected 'soft_delete_cond_user3', got '%s'", remainingUsers[0].Name)
	}
}

// TestMySQLIntegrationRestoreWithConditions tests restore with conditions
func TestMySQLIntegrationRestoreWithConditions(t *testing.T) {

	db := SetupMySQLTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "restore_cond_user1", Avatar: "avatar1"},
		{Name: "restore_cond_user2", Avatar: "avatar1"},
		{Name: "restore_cond_user3", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete all users
	res, err := query.Model(&models.User{}).Where("name LIKE ?", "restore_cond_user%").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete users: %v", err)
	}

	if res.RowsAffected != 3 {
		t.Errorf("Expected 3 rows affected, got %d", res.RowsAffected)
	}

	// Restore only users with avatar1
	res, err = query.Model(&models.User{}).WithTrashed().Where("avatar = ?", "avatar1").RestoreSoftDeleted(&models.User{})
	if err != nil {
		t.Fatalf("Failed to restore users: %v", err)
	}

	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	// Verify only users with avatar1 are restored
	var restoredUsers []models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "restore_cond_user%").Find(&restoredUsers)
	if err != nil {
		t.Fatalf("Failed to find restored users: %v", err)
	}

	if len(restoredUsers) != 2 {
		t.Errorf("Expected 2 restored users, got %d", len(restoredUsers))
	}
}

// TestMySQLIntegrationOnlyTrashed tests OnlyTrashed method
func TestMySQLIntegrationOnlyTrashed(t *testing.T) {

	db := SetupMySQLTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "only_trashed_user1", Avatar: "avatar1"},
		{Name: "only_trashed_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete one user
	res, err := query.Model(&models.User{}).Where("name = ?", "only_trashed_user1").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Test OnlyTrashed - should only find the deleted user
	var deletedUsers []models.User
	err = query.Model(&models.User{}).OnlyTrashed().Where("name LIKE ?", "only_trashed_user%").Find(&deletedUsers)
	if err != nil {
		t.Fatalf("Failed to find users with OnlyTrashed: %v", err)
	}

	if len(deletedUsers) != 1 {
		t.Errorf("Expected 1 deleted user, got %d", len(deletedUsers))
		return
	}

	if deletedUsers[0].Name != "only_trashed_user1" {
		t.Errorf("Expected 'only_trashed_user1', got '%s'", deletedUsers[0].Name)
	}
}

// TestMySQLIntegrationWithoutTrashed tests WithoutTrashed method
func TestMySQLIntegrationWithoutTrashed(t *testing.T) {

	db := SetupMySQLTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "without_trashed_user1", Avatar: "avatar1"},
		{Name: "without_trashed_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Soft delete one user
	res, err := query.Model(&models.User{}).Where("name = ?", "without_trashed_user1").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Test WithoutTrashed after WithTrashed
	var activeUsers []models.User
	err = query.Model(&models.User{}).WithTrashed().WithoutTrashed().Where("name LIKE ?", "without_trashed_user%").Find(&activeUsers)
	if err != nil {
		t.Fatalf("Failed to find users with WithoutTrashed: %v", err)
	}

	if len(activeUsers) != 1 {
		t.Errorf("Expected 1 active user, got %d", len(activeUsers))
		return
	}

	if activeUsers[0].Name != "without_trashed_user2" {
		t.Errorf("Expected 'without_trashed_user2', got '%s'", activeUsers[0].Name)
	}
}
