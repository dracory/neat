package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryCountBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	user1 := models.User{Name: "count_user_1"}
	user2 := models.User{Name: "count_user_1"}
	if err := db.Query().Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create user 1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user 2: %v", err)
	}

	var count int64
	if err := db.Query().Model(&models.User{}).Where("name = ?", "count_user_1").Count(&count); err != nil {
		t.Errorf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestSQLiteIntegrationQueryCountWithTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	user1 := models.User{Name: "count_table_user_1", Avatar: "avatar1_cnt"}
	user2 := models.User{Name: "count_table_user_2", Avatar: "avatar2_cnt"}
	if err := db.Query().Model(&models.User{}).Create(&user1); err != nil {
		t.Fatalf("Failed to create user 1: %v", err)
	}
	if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
		t.Fatalf("Failed to create user 2: %v", err)
	}

	var count int64
	if err := db.Query().Table("users").Where("avatar = ?", "avatar1_cnt").Count(&count); err != nil {
		t.Errorf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}
