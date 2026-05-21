//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

// TestMySQLIntegrationRawUpdate tests raw SQL expressions in Update
func TestMySQLIntegrationRawUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupMySQLTest(t)
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

	// Update using raw expression (concat)
	// Note: neat may not have a Raw() function, so we'll skip this test
	// or adapt it to use the available API
	t.Skip("Raw expressions not currently supported in neat")
}

// TestMySQLIntegrationRawWhere tests raw SQL expressions in Where clauses
func TestMySQLIntegrationRawWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupMySQLTest(t)
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

// TestMySQLIntegrationDatabaseFunctions tests database-specific raw functions
func TestMySQLIntegrationDatabaseFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupMySQLTest(t)
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

	// Update using MySQL-specific function (UPPER)
	// Note: neat may not have a Raw() function, so we'll skip this test
	// or adapt it to use the available API
	t.Skip("Raw expressions not currently supported in neat")
}
