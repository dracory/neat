package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryOmitDuringSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	user := models.User{Name: "omit_user", Avatar: "omit_avatar"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var result models.User
	err := db.Query().Model(&models.User{}).Omit("avatar").Where("name = ?", "omit_user").First(&result)
	if err != nil {
		t.Errorf("Omit during select failed: %v", err)
	}
	if result.Avatar != "" {
		t.Errorf("Expected empty avatar, got '%s'", result.Avatar)
	}
	if result.Name != "omit_user" {
		t.Errorf("Expected 'omit_user', got '%s'", result.Name)
	}
}

func TestSQLiteIntegrationQueryOmitDuringUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	user := models.User{Name: "omit_user", Avatar: "omit_avatar"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user.Name = "omit_user_updated"
	user.Avatar = "should_not_update"
	err := db.Query().Model(&models.User{}).Omit("avatar").Where("name = ?", "omit_user").Save(&user)
	if err != nil {
		t.Errorf("Omit during update failed: %v", err)
	}

	var result models.User
	err = db.Query().Model(&models.User{}).Where("name = ?", "omit_user_updated").First(&result)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if result.Avatar != "omit_avatar" {
		t.Errorf("Expected avatar to remain 'omit_avatar', got '%s'", result.Avatar)
	}
	if result.Name != "omit_user_updated" {
		t.Errorf("Expected 'omit_user_updated', got '%s'", result.Name)
	}
}
