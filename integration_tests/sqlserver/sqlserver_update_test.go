package sqlserver_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLServerIntegrationUpdateByModel verifies that Update() via a model with
// a WHERE clause updates all matching rows and reports the correct RowsAffected.
func TestSQLServerIntegrationUpdateByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "update_by_model_name", Avatar: "original_avatar"},
		{Name: "update_by_model_name", Avatar: "original_avatar1"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "update_by_model_name").Find(&createdUsers)
	if err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 1 {
		t.Fatalf("Expected at least 1 created user, got %d", len(createdUsers))
	}

	res, err := query.Model(&models.User{}).Where("name = ?", "update_by_model_name").Update("avatar", "updated_avatar")
	if err != nil {
		t.Fatalf("Failed to update users: %v", err)
	}

	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUsers[0].ID).First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	if updatedUser.Avatar != "updated_avatar" {
		t.Errorf("Expected avatar to be 'updated_avatar', got '%s'", updatedUser.Avatar)
	}
}

// TestSQLServerIntegrationUpdateByTable verifies that Update() via a raw table
// name with a WHERE clause updates all matching rows and reports the correct
// RowsAffected.
func TestSQLServerIntegrationUpdateByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "update_by_table_name", Avatar: "original_avatar"},
		{Name: "update_by_table_name", Avatar: "original_avatar1"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "update_by_table_name").Find(&createdUsers)
	if err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 1 {
		t.Fatalf("Expected at least 1 created user, got %d", len(createdUsers))
	}

	res, err := query.Table("users").Where("name = ?", "update_by_table_name").Update("avatar", "updated_avatar")
	if err != nil {
		t.Fatalf("Failed to update users: %v", err)
	}

	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUsers[0].ID).First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	if updatedUser.Avatar != "updated_avatar" {
		t.Errorf("Expected avatar to be 'updated_avatar', got '%s'", updatedUser.Avatar)
	}
}
