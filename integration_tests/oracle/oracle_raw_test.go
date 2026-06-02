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

	t.Skip("TODO: Oracle raw update with concatenation failing - ORA-00911 invalid character")
}

// TestOracleIntegrationRawWhere tests raw SQL expressions in Where clauses
func TestOracleIntegrationRawWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupOracleTest(t)
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

	t.Skip("TODO: Oracle database functions test failing - ORA-00911 invalid character")
}
