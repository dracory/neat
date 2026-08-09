package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestCockroachDBIntegrationRawUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupCockroachDBTest(t)
	query := databaseConn.Query()

	user := models.User{Name: "raw_update_user", Avatar: "original"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "raw_update_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	t.Skip("Raw expressions not currently supported in neat")
}

func TestCockroachDBIntegrationRawWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupCockroachDBTest(t)
	query := databaseConn.Query()

	users := []models.User{
		{Name: "raw_where_user1", Avatar: "avatar1"},
		{Name: "raw_where_user2", Avatar: "avatar2"},
		{Name: "other_user", Avatar: "other"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("SUBSTR(name, 1, 14) = ?", "raw_where_user").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to query with raw where: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}
}

func TestCockroachDBIntegrationDatabaseFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := SetupCockroachDBTest(t)
	query := databaseConn.Query()

	user := models.User{Name: "db_functions_user", Avatar: "avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "db_functions_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	t.Skip("Raw expressions not currently supported in neat")
}
