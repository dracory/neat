package sqlite

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryIncrementDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Seed data
	user := models.User{Name: "increment_user", ID: 1, Avatar: "group1"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("Increment", func(t *testing.T) {
		res, err := db.Query().Table("users").Where("id = ?", 1).Increment("id")
		if err != nil {
			t.Errorf("Increment failed: %v", err)
		}
		if res.RowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
		}

		var updatedUser models.User
		err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
		if err != nil {
			t.Errorf("Failed to find updated user: %v", err)
		}
		if updatedUser.ID != 2 {
			t.Errorf("Expected ID 2, got %d", updatedUser.ID)
		}
	})

	t.Run("Increment by amount", func(t *testing.T) {
		res, err := db.Query().Table("users").Where("name = ?", "increment_user").Increment("id", 5)
		if err != nil {
			t.Errorf("Increment by amount failed: %v", err)
		}
		if res.RowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
		}

		var updatedUser models.User
		err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
		if err != nil {
			t.Errorf("Failed to find updated user: %v", err)
		}
		if updatedUser.ID != 7 {
			t.Errorf("Expected ID 7, got %d", updatedUser.ID)
		}
	})

	t.Run("Decrement", func(t *testing.T) {
		res, err := db.Query().Table("users").Where("name = ?", "increment_user").Decrement("id")
		if err != nil {
			t.Errorf("Decrement failed: %v", err)
		}
		if res.RowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
		}

		var updatedUser models.User
		err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
		if err != nil {
			t.Errorf("Failed to find updated user: %v", err)
		}
		if updatedUser.ID != 6 {
			t.Errorf("Expected ID 6, got %d", updatedUser.ID)
		}
	})

	t.Run("Decrement by amount", func(t *testing.T) {
		res, err := db.Query().Table("users").Where("name = ?", "increment_user").Decrement("id", 3)
		if err != nil {
			t.Errorf("Decrement by amount failed: %v", err)
		}
		if res.RowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
		}

		var updatedUser models.User
		err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
		if err != nil {
			t.Errorf("Failed to find updated user: %v", err)
		}
		if updatedUser.ID != 3 {
			t.Errorf("Expected ID 3, got %d", updatedUser.ID)
		}
	})

	t.Run("With where conditions", func(t *testing.T) {
		// Create another user
		user2 := models.User{Name: "other_user", ID: 10, Avatar: "group2"}
		if err := db.Query().Model(&models.User{}).Create(&user2); err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}

		res, err := db.Query().Table("users").Where("avatar = ?", "group2").Increment("id", 10)
		if err != nil {
			t.Errorf("Increment with where failed: %v", err)
		}
		if res.RowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
		}

		var updatedUser models.User
		err = db.Query().Table("users").Where("name = ?", "other_user").First(&updatedUser)
		if err != nil {
			t.Errorf("Failed to find updated user: %v", err)
		}
		if updatedUser.ID != 20 {
			t.Errorf("Expected ID 20, got %d", updatedUser.ID)
		}

		// Verify first user was NOT updated
		var firstUser models.User
		err = db.Query().Table("users").Where("name = ?", "increment_user").First(&firstUser)
		if err != nil {
			t.Errorf("Failed to find first user: %v", err)
		}
		if firstUser.ID != 3 {
			t.Errorf("Expected ID 3, got %d", firstUser.ID)
		}
	})

	t.Run("Increment with extra columns", func(t *testing.T) {
		res, err := db.Query().Table("users").Where("name = ?", "increment_user").Increment("id", 1, map[string]any{"avatar": "updated_group"})
		if err != nil {
			t.Errorf("Increment with extra columns failed: %v", err)
		}
		if res.RowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
		}

		var updatedUser models.User
		err = db.Query().Table("users").Where("name = ?", "increment_user").First(&updatedUser)
		if err != nil {
			t.Errorf("Failed to find updated user: %v", err)
		}
		if updatedUser.ID != 4 { // Previous was 3
			t.Errorf("Expected ID 4, got %d", updatedUser.ID)
		}
		if updatedUser.Avatar != "updated_group" {
			t.Errorf("Expected avatar 'updated_group', got '%s'", updatedUser.Avatar)
		}
	})

	t.Run("Invalid column", func(t *testing.T) {
		_, err := db.Query().Table("users").Increment("invalid; column")
		if err == nil {
			t.Error("Expected error for invalid column")
		}
		if !strings.Contains(err.Error(), "invalid column name") {
			t.Errorf("Expected error to contain 'invalid column name', got '%s'", err.Error())
		}
	})
}
