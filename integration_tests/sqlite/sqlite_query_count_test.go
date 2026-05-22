package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLiteIntegrationQueryCount tests Count operations
func TestSQLiteIntegrationQueryCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	t.Run("count basic", func(t *testing.T) {
		user1 := models.User{Name: "count_user_1"}
		user2 := models.User{Name: "count_user_1"}
		if err := query.Model(&models.User{}).Create(&user1); err != nil {
			t.Fatalf("Failed to create user 1: %v", err)
		}
		if err := query.Model(&models.User{}).Create(&user2); err != nil {
			t.Fatalf("Failed to create user 2: %v", err)
		}

		var count int64
		if err := query.Model(&models.User{}).Where("name = ?", "count_user_1").Count(&count); err != nil {
			t.Errorf("Count failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("count with table", func(t *testing.T) {
		user1 := models.User{Name: "count_table_user_1", Avatar: "avatar1"}
		user2 := models.User{Name: "count_table_user_2", Avatar: "avatar2"}
		if err := query.Model(&models.User{}).Create(&user1); err != nil {
			t.Fatalf("Failed to create user 1: %v", err)
		}
		if err := query.Model(&models.User{}).Create(&user2); err != nil {
			t.Fatalf("Failed to create user 2: %v", err)
		}

		var count int64
		if err := query.Table("users").Where("avatar = ?", "avatar1").Count(&count); err != nil {
			t.Errorf("Count failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})
}
