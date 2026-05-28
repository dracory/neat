package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryDeleteByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	user := models.User{Name: "delete_user_model"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	var createdUser models.User
	if err := db.Query().Model(&models.User{}).Where("name = ?", "delete_user_model").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	res, err := db.Query().Model(&models.User{}).Where("id = ?", createdUser.ID).Delete(&models.User{})
	if err != nil {
		t.Errorf("Delete by model failed: %v", err)
	}
	if res == nil {
		t.Error("Expected result to be returned")
		return
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}
}

func TestSQLiteIntegrationQueryDeleteByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	user := models.User{Name: "delete_user_table"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	res, err := db.Query().Table("users").Where("name = ?", "delete_user_table").Delete()
	if err != nil {
		t.Errorf("Delete by table failed: %v", err)
	}
	if res == nil {
		t.Error("Expected result to be returned")
		return
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}
}

func TestSQLiteIntegrationQueryDeleteByModelWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	user1 := models.User{Name: "delete_user_where_1"}
	user2 := models.User{Name: "delete_user_where_2"}
	if err := db.Query().Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	res, err := db.Query().Model(&models.User{}).Where("name = ?", "delete_user_where_1").Delete(&models.User{})
	if err != nil {
		t.Errorf("Delete by model with where failed: %v", err)
	}
	if res == nil {
		t.Error("Expected result to be returned")
		return
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var foundUser2 models.User
	err = db.Query().Model(&models.User{}).Where("name = ?", "delete_user_where_2").First(&foundUser2)
	if err != nil {
		t.Errorf("User 2 should still exist: %v", err)
	}
}
