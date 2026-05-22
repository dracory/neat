package sqlite

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Seed data
	users := []models.User{
		{Name: "value_user_1", Avatar: "avatar_1"},
		{Name: "value_user_2", Avatar: "avatar_2"},
	}
	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("Value basic", func(t *testing.T) {
		var name string
		err := db.Query().Model(&models.User{}).OrderBy("id", "asc").Value("name", &name)
		if err != nil {
			t.Errorf("Value basic failed: %v", err)
		}
		if name != "value_user_1" {
			t.Errorf("Expected 'value_user_1', got '%s'", name)
		}
	})

	t.Run("Value with where", func(t *testing.T) {
		var name string
		err := db.Query().Model(&models.User{}).Where("avatar = ?", "avatar_2").Value("name", &name)
		if err != nil {
			t.Errorf("Value with where failed: %v", err)
		}
		if name != "value_user_2" {
			t.Errorf("Expected 'value_user_2', got '%s'", name)
		}
	})

	t.Run("Value with order", func(t *testing.T) {
		var name string
		err := db.Query().Model(&models.User{}).OrderBy("id", "desc").Value("name", &name)
		if err != nil {
			t.Errorf("Value with order failed: %v", err)
		}
		if name != "value_user_2" {
			t.Errorf("Expected 'value_user_2', got '%s'", name)
		}
	})

	t.Run("Value not found", func(t *testing.T) {
		var name string
		err := db.Query().Model(&models.User{}).Where("name = ?", "non_existent").Value("name", &name)
		if err != nil {
			t.Errorf("Value not found failed: %v", err)
		}
		if name != "" {
			t.Error("Expected empty name")
		}
	})

	t.Run("ToSql Value", func(t *testing.T) {
		var name string
		sql := db.Query().Model(&models.User{}).Where("id = ?", 1).ToSql().Value("name", &name)
		if !strings.Contains(sql, "SELECT `name` FROM `users`") {
			t.Error("Expected SELECT `name` FROM `users`")
		}
		if !strings.Contains(sql, "WHERE `id` = ?") {
			t.Error("Expected WHERE `id` = ?")
		}
		if !strings.Contains(sql, "LIMIT 1") {
			t.Error("Expected LIMIT 1")
		}
	})
}
