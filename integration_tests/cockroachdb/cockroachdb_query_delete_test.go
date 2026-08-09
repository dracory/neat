package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestCockroachDBIntegrationQueryDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{Name: "delete_query_user", Avatar: "avatar"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "delete_query_user").First(&createdUser)
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	_, err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	var deletedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&deletedUser)
	if err == nil {
		t.Error("Expected error for deleted user, got nil")
	}
}

func TestCockroachDBIntegrationQueryDeleteWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "delete_where_user1", Avatar: "avatar1"},
		{Name: "delete_where_user2", Avatar: "avatar2"},
		{Name: "keep_user", Avatar: "avatar3"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	res, err := query.Model(&models.User{}).Where("name LIKE ?", "delete_where_user%").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete users: %v", err)
	}
	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	var count int64
	err = query.Model(&models.User{}).Where("name LIKE ?", "delete_where_user%").Count(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 remaining, got %d", count)
	}

	err = query.Model(&models.User{}).Where("name = ?", "keep_user").Count(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 remaining, got %d", count)
	}
}

func TestCockroachDBIntegrationQueryDeleteByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{Name: "delete_table_user", Avatar: "avatar"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "delete_table_user").First(&createdUser)
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	_, err = query.Table("users").Where("id = ?", createdUser.ID).Delete()
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	var deletedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&deletedUser)
	if err == nil {
		t.Error("Expected error for deleted user, got nil")
	}
}

func TestCockroachDBIntegrationQueryDeleteAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "delete_all_user1"},
		{Name: "delete_all_user2"},
		{Name: "delete_all_user3"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	res, err := query.Model(&models.User{}).Where("name LIKE ?", "delete_all_user%").Delete(&models.User{})
	if err != nil {
		t.Fatalf("Failed to delete all users: %v", err)
	}
	if res.RowsAffected != 3 {
		t.Errorf("Expected 3 rows affected, got %d", res.RowsAffected)
	}

	var count int64
	err = query.Model(&models.User{}).Where("name LIKE ?", "delete_all_user%").Count(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 remaining, got %d", count)
	}
}
