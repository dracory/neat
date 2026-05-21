//go:build integration

package sqlite

import (
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQuerySelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	t.Run("Select specific columns", func(t *testing.T) {
		user := models.User{Name: "select_user", Avatar: "select_avatar"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		var foundUser models.User
		err := query.Model(&models.User{}).Select("name").Find(&foundUser, user.ID)
		if err != nil {
			t.Errorf("Select specific columns failed: %v", err)
		}
		if foundUser.Name != "select_user" {
			t.Errorf("Expected 'select_user', got '%s'", foundUser.Name)
		}
		if foundUser.Avatar != "" {
			t.Error("Expected Avatar to be empty")
		}
	})

	t.Run("Select with aliases", func(t *testing.T) {
		user := models.User{Name: "alias_user"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		var result struct {
			UserName string `gorm:"column:user_name"`
		}
		err := query.Model(&models.User{}).Select("name as user_name").Where("id = ?", user.ID).Scan(&result)
		if err != nil {
			t.Errorf("Select with aliases failed: %v", err)
		}
		if result.UserName != "alias_user" {
			t.Errorf("Expected 'alias_user', got '%s'", result.UserName)
		}
	})

	t.Run("Select with raw subqueries", func(t *testing.T) {
		user := models.User{Name: "subquery_user"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		var result struct {
			SubName string `gorm:"column:sub_name"`
		}
		// In SQLite, we can use a subquery in SELECT
		err := query.Model(&models.User{}).
			Select("(SELECT name FROM users WHERE id = ?) as sub_name", user.ID).
			Where("id = ?", user.ID).
			Scan(&result)
		if err != nil {
			t.Errorf("Select with raw subqueries failed: %v", err)
		}
		if result.SubName != "subquery_user" {
			t.Errorf("Expected 'subquery_user', got '%s'", result.SubName)
		}
	})

	t.Run("Select with subquery callbacks", func(t *testing.T) {
		user := models.User{Name: "callback_user"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		var result struct {
			SubName string `gorm:"column:sub_name"`
		}
		err := query.Model(&models.User{}).
			Select(func(q orm.Query) orm.Query {
				return q.Table("users").Select("name").Where("id = ?", user.ID)
			}, "sub_name").
			Where("id = ?", user.ID).
			Scan(&result)

		if err != nil {
			t.Errorf("Select with subquery callbacks failed: %v", err)
		}
		if result.SubName != "callback_user" {
			t.Errorf("Expected 'callback_user', got '%s'", result.SubName)
		}
	})
}
