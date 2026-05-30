//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryOmitDuringSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "omit_user", Avatar: "omit_avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "omit_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	var foundUser models.User
	err := query.Model(&models.User{}).Omit("avatar").Where("id = ?", createdUser.ID).First(&foundUser)
	if err != nil {
		t.Errorf("Omit during select failed: %v", err)
	}
	if foundUser.Name != "omit_user" {
		t.Errorf("Expected 'omit_user', got '%s'", foundUser.Name)
	}
	if foundUser.Avatar != "" {
		t.Errorf("Expected empty avatar, got '%s'", foundUser.Avatar)
	}
}

func TestPostgresIntegrationQueryOmitDuringUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "update_omit_user", Avatar: "original_avatar"}
	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "update_omit_user").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	createdUser.Name = "updated_name"
	createdUser.Avatar = "updated_avatar"

	_, err := query.Model(&models.User{}).Omit("Avatar").Where("id = ?", createdUser.ID).Update("name", "updated_name")
	if err != nil {
		t.Errorf("Omit during update failed: %v", err)
	}

	var foundUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&foundUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if foundUser.Name != "updated_name" {
		t.Errorf("Expected 'updated_name', got '%s'", foundUser.Name)
	}
	if foundUser.Avatar != "original_avatar" {
		t.Errorf("Expected 'original_avatar', got '%s'", foundUser.Avatar)
	}
}
