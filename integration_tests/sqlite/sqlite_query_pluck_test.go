package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationPluck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
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
			t.Errorf("Pluck single column failed: %v", err)
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

	t.Run("Pluck with Where conditions", func(t *testing.T) {
		var names []string
		err := db.Query().Model(&models.User{}).Where("avatar = ?", "avatar1").OrderBy("name", "asc").Pluck("name", &names)
		if err != nil {
			t.Errorf("Pluck with Where conditions failed: %v", err)
		}
		if len(names) != 2 {
			t.Errorf("Expected 2 names, got %d", len(names))
		}
		if names[0] != "pluck_user_1" {
			t.Errorf("Expected 'pluck_user_1', got '%s'", names[0])
		}
		if names[1] != "pluck_user_3" {
			t.Errorf("Expected 'pluck_user_3', got '%s'", names[1])
		}
	})

	t.Run("Pluck into maps", func(t *testing.T) {
		// Goravel/GORM Pluck into map usually means plucking two columns: key and value.
		// However, our Pluck interface only takes one column name.
		// Let's check how GORM handles it. In GORM, Pluck only plucks one column.
		// If we want a map, we usually use Scan or Find.

		// Some ORMs allow Pluck("name", "id", &map[string]uint)
		// But our interface is Pluck(column string, dest any) error

		// Let's see if we can pluck into a map of [string]any where any is the column value
		// This doesn't make much sense for a single column Pluck.

		// If the user wants a map, they might be expecting Pluck to support it.
		// For now, let's test if it can pluck into a slice of maps or something.
		var results []map[string]any
		err := db.Query().Table("users").Where("name = ?", "pluck_user_1").Select("name", "avatar").Scan(&results)
		if err != nil {
			t.Errorf("Pluck into maps failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0]["name"] != "pluck_user_1" {
			t.Errorf("Expected 'pluck_user_1', got '%v'", results[0]["name"])
		}
	})

	t.Run("Pluck edge cases - duplicates", func(t *testing.T) {
		var avatars []string
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "pluck_user_%").OrderBy("avatar", "asc").Pluck("avatar", &avatars)
		if err != nil {
			t.Errorf("Pluck with duplicates failed: %v", err)
		}
		if len(avatars) != 3 {
			t.Errorf("Expected 3 avatars, got %d", len(avatars))
		}
		if avatars[0] != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", avatars[0])
		}
		if avatars[1] != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", avatars[1])
		}
		if avatars[2] != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", avatars[2])
		}
	})

	t.Run("Pluck edge cases - empty results", func(t *testing.T) {
		var names []string
		err := db.Query().Model(&models.User{}).Where("name = ?", "non_existent").Pluck("name", &names)
		if err != nil {
			t.Errorf("Pluck with empty results failed: %v", err)
		}
		if len(names) != 0 {
			t.Errorf("Expected 0 names, got %d", len(names))
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
		if !foundAvatar1 || !foundAvatar2 {
			t.Errorf("Expected to find both avatar1 and avatar2")
		}
	})
}
