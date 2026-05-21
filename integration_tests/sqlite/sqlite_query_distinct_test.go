//go:build integration

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
	query := db.Query()

	// Seed data
	users := []models.User{
		{Name: "distinct_user_1", Avatar: "avatar1"},
		{Name: "distinct_user_2", Avatar: "avatar1"},
		{Name: "distinct_user_3", Avatar: "avatar2"},
	}

	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("Distinct single column", func(t *testing.T) {
		type Result struct {
			Avatar string
		}
		var results []Result
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
			Select("avatar").Distinct().OrderBy("avatar", "asc").Scan(&results)
		if err != nil {
			t.Errorf("Distinct single column failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].Avatar != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", results[0].Avatar)
		}
		if results[1].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[1].Avatar)
		}
	})

	t.Run("Distinct multiple columns", func(t *testing.T) {
		type Result struct {
			Name   string
			Avatar string
		}
		var results []Result
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
			Select("name", "avatar").Distinct().OrderBy("name", "asc").Scan(&results)
		if err != nil {
			t.Errorf("Distinct multiple columns failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})

	t.Run("Distinct with Count", func(t *testing.T) {
		var count int64
		// GORM Distinct + Count might need special handling
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
			Distinct("avatar").Count(&count)
		if err != nil {
			t.Errorf("Distinct with Count failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("Distinct without arguments with Count", func(t *testing.T) {
		var count int64
		// This should count unique rows based on the selection
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
			Select("avatar").Distinct().Count(&count)
		if err != nil {
			t.Errorf("Distinct without arguments with Count failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("Distinct with Select", func(t *testing.T) {
		type Result struct {
			Avatar string
		}
		var results []Result
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
			Distinct("avatar").OrderBy("avatar", "asc").Scan(&results)
		if err != nil {
			t.Errorf("Distinct with Select failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].Avatar != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", results[0].Avatar)
		}
		if results[1].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[1].Avatar)
		}
	})
}
