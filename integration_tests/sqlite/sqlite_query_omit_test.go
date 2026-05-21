//go:build integration

package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryOmit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	t.Run("Omit during select", func(t *testing.T) {
		user := models.User{Name: "omit_user", Avatar: "omit_avatar"}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		var foundUser models.User
		err := query.Model(&models.User{}).Omit("avatar").Find(&foundUser, user.ID)
		if err != nil {
			t.Errorf("Omit during select failed: %v", err)
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

		// Update using model, but omit avatar
		err := query.Model(&user).Omit("Avatar").Save(&user)
		if err != nil {
			t.Errorf("Omit during update failed: %v", err)
		}

		var foundUser models.User
		if err := query.Find(&foundUser, user.ID); err != nil {
			t.Errorf("Failed to find user: %v", err)
		}
		if foundUser.Name != "updated_name" {
			t.Errorf("Expected 'updated_name', got '%s'", foundUser.Name)
		}
		if foundUser.Avatar != "original_avatar" {
			t.Errorf("Expected 'original_avatar', got '%s'", foundUser.Avatar)
		}
	})
}
