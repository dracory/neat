package sqlserver

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLServerIntegrationQueryDeleteByModel verifies that Delete() via a model
// removes exactly one row and reports RowsAffected = 1.
func TestSQLServerIntegrationQueryDeleteByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	user := models.User{Name: "delete_user_model"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "delete_user_model").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	res, err := query.Model(&models.User{}).Where("id = ?", createdUser.ID).Delete(&models.User{})
	if err != nil {
		t.Errorf("Delete by model failed: %v", err)
	}
	if res != nil && res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}
}

// TestSQLServerIntegrationQueryDeleteByTable verifies that Delete() via a raw
// table name with a WHERE clause removes exactly one matching row.
func TestSQLServerIntegrationQueryDeleteByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	user := models.User{Name: "delete_user_table"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	res, err := query.Table("users").Where("name = ?", "delete_user_table").Delete()
	if err != nil {
		t.Errorf("Delete by table failed: %v", err)
	}
	if res != nil && res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}
}

// TestSQLServerIntegrationQueryDeleteByModelWithWhere verifies that a targeted
// DELETE only removes the row matching the WHERE clause, leaving other rows intact.
func TestSQLServerIntegrationQueryDeleteByModelWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	user1 := models.User{Name: "delete_user_where_1"}
	user2 := models.User{Name: "delete_user_where_2"}
	if err := query.Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}
	if err := query.Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	res, err := query.Model(&models.User{}).Where("name = ?", "delete_user_where_1").Delete(&models.User{})
	if err != nil {
		t.Errorf("Delete by model with where failed: %v", err)
	}
	if res != nil && res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var foundUser2 models.User
	err = query.Model(&models.User{}).Where("name = ?", "delete_user_where_2").First(&foundUser2)
	if err != nil {
		t.Errorf("User 2 should still exist: %v", err)
	}
}
