//go:build disabled

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationPluck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Seed data
	users := []models.User{
		{Name: "pluck_user_1", Avatar: "avatar1"},
		{Name: "pluck_user_2", Avatar: "avatar2"},
		{Name: "pluck_user_3", Avatar: "avatar1"},
	}

	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("Pluck single column into slice", func(t *testing.T) {
		var names []string
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "pluck_user_%").OrderBy("name", "asc").Pluck("name", &names)
		if err != nil {
			t.Errorf("Pluck failed: %v", err)
		}
		if len(names) != 3 {
			t.Errorf("Expected 3 names, got %d", len(names))
		}
		if names[0] != "pluck_user_1" {
			t.Errorf("Expected 'pluck_user_1', got '%s'", names[0])
		}
		if names[1] != "pluck_user_2" {
			t.Errorf("Expected 'pluck_user_2', got '%s'", names[1])
		}
		if names[2] != "pluck_user_3" {
			t.Errorf("Expected 'pluck_user_3', got '%s'", names[2])
		}
	})

	t.Run("Pluck with Distinct", func(t *testing.T) {
		var avatars []string
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "pluck_user_%").Distinct("avatar").OrderBy("avatar", "asc").Pluck("avatar", &avatars)
		if err != nil {
			t.Errorf("Pluck with Distinct failed: %v", err)
		}
		if len(avatars) != 2 {
			t.Errorf("Expected 2 avatars, got %d", len(avatars))
		}
		foundAvatar1 := false
		foundAvatar2 := false
		for _, avatar := range avatars {
			if avatar == "avatar1" {
				foundAvatar1 = true
			}
			if avatar == "avatar2" {
				foundAvatar2 = true
			}
		}
		if !foundAvatar1 {
			t.Error("Expected to find 'avatar1'")
		}
		if !foundAvatar2 {
			t.Error("Expected to find 'avatar2'")
		}
	})
}
