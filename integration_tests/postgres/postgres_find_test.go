//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

// setupFindTest creates a database connection and sets up test tables
func setupFindTest(t *testing.T) *database.Database {
	return SetupPostgresConnection(t)
}

// TestPostgresIntegrationFirst tests First operation
func TestPostgresIntegrationFirst(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Create users
	users := []models.User{
		{Name: "first_user1", Avatar: "avatar1"},
		{Name: "first_user2", Avatar: "avatar2"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Test First - should get the first record
	var user models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "first_user%").First(&user)
	if err != nil {
		t.Fatalf("Failed to get first user: %v", err)
	}

	if user.Name != "first_user1" && user.Name != "first_user2" {
		t.Errorf("Expected first_user1 or first_user2, got '%s'", user.Name)
	}
}

// TestPostgresIntegrationFind tests Find operation
func TestPostgresIntegrationFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Create users
	users := []models.User{
		{Name: "find_user1", Avatar: "avatar1"},
		{Name: "find_user2", Avatar: "avatar2"},
		{Name: "find_user3", Avatar: "avatar3"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Test Find - should get all matching records
	var foundUsers []models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "find_user%").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users: %v", err)
	}

	if len(foundUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(foundUsers))
	}
}

// TestPostgresIntegrationCreate tests Create operation
func TestPostgresIntegrationCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Test Create single record
	user := models.User{Name: "create_user", Avatar: "avatar"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify the record was created by querying it
	var createdUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "create_user").First(&createdUser)
	if err != nil {
		t.Fatalf("Failed to query created user: %v", err)
	}

	if createdUser.Name != "create_user" {
		t.Errorf("Expected 'create_user', got '%s'", createdUser.Name)
	}

	// Test Create multiple records
	users := []models.User{
		{Name: "create_user1", Avatar: "avatar1"},
		{Name: "create_user2", Avatar: "avatar2"},
	}
	err = query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Verify the records were created
	var foundUsers []models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "create_user%").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to query created users: %v", err)
	}

	if len(foundUsers) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(foundUsers))
	}
}

// TestPostgresIntegrationUpdate tests Update operation
func TestPostgresIntegrationUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Create a user
	user := models.User{Name: "update_user", Avatar: "old_avatar"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "update_user").First(&createdUser)
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Test Update
	_, err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).Update("avatar", "new_avatar")
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Avatar != "new_avatar" {
		t.Errorf("Expected 'new_avatar', got '%s'", updatedUser.Avatar)
	}
}

// TestPostgresIntegrationDelete tests Delete operation
func TestPostgresIntegrationDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Create a user
	user := models.User{Name: "delete_user", Avatar: "avatar"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "delete_user").First(&createdUser)
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Test Delete
	_, err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify deletion
	var deletedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&deletedUser)
	if err == nil {
		t.Error("Expected error for deleted user, got nil")
	}
}
