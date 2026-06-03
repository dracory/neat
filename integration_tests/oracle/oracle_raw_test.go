package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestOracleIntegrationRawUpdate tests raw SQL expressions in Update
func TestOracleIntegrationRawUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupOracleTest(t)
	query := databaseConn.Query()

	// Cleanup test data
	t.Cleanup(func() {
		sqlDB, err := databaseConn.DB()
		if err == nil {
			sqlDB.Exec(`DELETE FROM USERS WHERE NAME LIKE 'raw_update_user'`)
		}
	})

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
	// Use Exec with raw SQL for raw expression updates
	_, err := query.Exec("UPDATE users SET avatar = avatar || '_updated' WHERE id = ?", createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to update with raw SQL: %v", err)
	}

	var updatedUser models.User
	if err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&updatedUser); err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Avatar != "original_updated" {
		t.Errorf("Expected avatar to be 'original_updated', got '%s'", updatedUser.Avatar)
	}
}

// TestOracleIntegrationRawWhere tests raw SQL expressions in Where clauses
func TestOracleIntegrationRawWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupOracleTest(t)
	query := databaseConn.Query()

	// Cleanup test data
	t.Cleanup(func() {
		sqlDB, err := databaseConn.DB()
		if err == nil {
			sqlDB.Exec(`DELETE FROM USERS WHERE NAME LIKE 'raw_where_user%' OR NAME = 'other_user'`)
		}
	})

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
	err := query.Model(&models.User{}).Where("SUBSTR(name, 1, 14) = ?", "raw_where_user").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to query with raw where: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}
}

// TestOracleIntegrationDatabaseFunctions tests database-specific raw functions
func TestOracleIntegrationDatabaseFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupOracleTest(t)
	query := databaseConn.Query()

	// Cleanup test data
	t.Cleanup(func() {
		sqlDB, err := databaseConn.DB()
		if err == nil {
			sqlDB.Exec(`DELETE FROM USERS WHERE NAME LIKE 'db_functions_user'`)
		}
	})

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

	// Update using Oracle-specific function (UPPER)
	_, err := query.Exec("UPDATE users SET avatar = UPPER(avatar) WHERE id = ?", createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to update with Oracle function: %v", err)
	}

	var updatedUser models.User
	if err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&updatedUser); err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Avatar != "AVATAR" {
		t.Errorf("Expected avatar to be 'AVATAR', got '%s'", updatedUser.Avatar)
	}
}
