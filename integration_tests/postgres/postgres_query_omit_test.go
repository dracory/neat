//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryOmit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	t.Run("Omit during select", func(t *testing.T) {
		user := models.User{Name: "omit_user", Avatar: "omit_avatar"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		var foundUser models.User
		err := query.Model(&models.User{}).Omit("avatar").Find(&foundUser, user.ID)
		if err != nil {
			t.Errorf("Find with Omit failed: %v", err)
		}
		if foundUser.Name != "omit_user" {
			t.Errorf("Expected 'omit_user', got '%s'", foundUser.Name)
		}
		if foundUser.Avatar != "" {
			t.Error("Expected Avatar to be empty")
		}
	})

	t.Run("Omit during update", func(t *testing.T) {
		user := models.User{Name: "update_omit_user", Avatar: "original_avatar"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		user.Name = "updated_name"
		user.Avatar = "updated_avatar"

		err := query.Model(&user).Omit("Avatar").Save(&user)
		if err != nil {
			t.Errorf("Save with Omit failed: %v", err)
		}

		var foundUser models.User
		err = query.Model(&models.User{}).Find(&foundUser, user.ID)
		if err != nil {
			t.Errorf("Find failed: %v", err)
		}
		if foundUser.Name != "updated_name" {
			t.Errorf("Expected 'updated_name', got '%s'", foundUser.Name)
		}
		if foundUser.Avatar != "original_avatar" {
			t.Errorf("Expected 'original_avatar', got '%s'", foundUser.Avatar)
		}
	})
}
