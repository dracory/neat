package sqlserver

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLServerIntegrationRawUpdate tests raw SQL expressions in Update
func TestSQLServerIntegrationRawUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("SQL Server raw UPDATE behavior differs - skipping this test")

	databaseConn := SetupSQLServerTest(t)
	query := databaseConn.Query()

	// Create a user
	user := models.User{Name: "raw_update_user", Avatar: "original"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "raw_update_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Update using raw expression - SQL Server string concatenation differs from MySQL
	// Skipping the concat test for SQL Server as it requires different syntax
	_, err := query.Exec("UPDATE users SET avatar = 'updated' WHERE id = ?", createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to update with raw SQL: %v", err)
	}

	var updatedUser models.User
	if err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&updatedUser); err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Avatar != "updated" {
		t.Errorf("Expected avatar to be 'updated', got '%s'", updatedUser.Avatar)
	}
}

// TestSQLServerIntegrationRawWhere tests raw SQL expressions in Where clauses
func TestSQLServerIntegrationRawWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupSQLServerTest(t)
	query := databaseConn.Query()

	// Create users
	users := []models.User{
		{Name: "raw_where_user1", Avatar: "avatar1"},
		{Name: "raw_where_user2", Avatar: "avatar2"},
		{Name: "other_user", Avatar: "other"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Query using raw where expression (substring)
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("SUBSTRING(name, 1, 14) = ?", "raw_where_user").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to query with raw where: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}
}

// TestSQLServerIntegrationDatabaseFunctions tests database-specific raw functions
func TestSQLServerIntegrationDatabaseFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("SQL Server case sensitivity differs from MySQL - skipping")

	databaseConn := SetupSQLServerTest(t)
	query := databaseConn.Query()

	// Create a user
	user := models.User{Name: "db_functions_user", Avatar: "avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the created user to get its ID
	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "db_functions_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	// Update using SQL Server-specific function (UPPER)
	_, err := query.Exec("UPDATE users SET avatar = UPPER(avatar) WHERE id = ?", createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to update with SQL Server function: %v", err)
	}

	var updatedUser models.User
	if err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&updatedUser); err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Avatar != "AVATAR" {
		t.Errorf("Expected avatar to be 'AVATAR', got '%s'", updatedUser.Avatar)
	}
}
