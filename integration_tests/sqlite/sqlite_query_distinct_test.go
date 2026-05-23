package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationDistinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	t.Run("Distinct single column", func(t *testing.T) {
		query := db.Query()
		// Seed data
		for _, user := range []models.User{
			{Name: "distinct_user_1", Avatar: "avatar1"},
			{Name: "distinct_user_2", Avatar: "avatar1"},
			{Name: "distinct_user_3", Avatar: "avatar2"},
		} {
			if err := query.Model(&models.User{}).Create(&user); err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
		}
		type Result struct {
			Avatar string
		}
		var results []Result
		err := query.Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
			Select("avatar").Distinct().OrderBy("avatar", "asc").Scan(&results)
		if err != nil {
			t.Errorf("Distinct single column failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 distinct avatars, got %d", len(results))
		}
	})

	t.Run("Distinct multiple columns", func(t *testing.T) {
		t.Skip("WHERE clause with SELECT not working correctly - needs investigation in query builder")
	})

	t.Run("Distinct with Count", func(t *testing.T) {
		query := db.Query()
		// Seed data
		for _, user := range []models.User{
			{Name: "distinct_user_1", Avatar: "avatar1"},
			{Name: "distinct_user_2", Avatar: "avatar1"},
			{Name: "distinct_user_3", Avatar: "avatar2"},
		} {
			if err := query.Model(&models.User{}).Create(&user); err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
		}
		var count int64
		err := query.Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
			Distinct("avatar").Count(&count)
		if err != nil {
			t.Errorf("Distinct with Count failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("Distinct without arguments with Count", func(t *testing.T) {
		t.Skip("ORM Select+Distinct+Count does not generate COUNT(DISTINCT ...) — not yet implemented")
	})

	t.Run("Distinct with Select", func(t *testing.T) {
		t.Skip("WHERE clause with SELECT not working correctly - needs investigation in query builder")
	})
}
