package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationDistinctSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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

	type Result struct {
		Avatar string
	}
	var results []Result
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
		Select("avatar").Distinct().OrderBy("avatar", "asc").Scan(&results)
	if err != nil {
		t.Errorf("Distinct failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if len(results) >= 2 {
		if results[0].Avatar != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", results[0].Avatar)
		}
		if results[1].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[1].Avatar)
		}
	}
}

func TestPostgresIntegrationDistinctWithCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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

	var count int64
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
		Distinct("avatar").Count(&count)
	if err != nil {
		t.Errorf("Distinct with Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}
