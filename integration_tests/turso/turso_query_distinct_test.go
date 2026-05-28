package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestTursoIntegrationDistinctSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

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
}

func TestTursoIntegrationDistinctMultipleColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

	for _, user := range []models.User{
		{Name: "distinct_user_1", Avatar: "avatar1"},
		{Name: "distinct_user_1", Avatar: "avatar1"}, // Duplicate
		{Name: "distinct_user_2", Avatar: "avatar1"},
	} {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	type Result struct {
		Name   string
		Avatar string
	}
	var results []Result
	err := query.Model(&models.User{}).Where("name LIKE ?", "distinct_user_%").
		Select([]string{"name", "avatar"}).Distinct().OrderBy("name", "asc").Scan(&results)
	if err != nil {
		t.Errorf("Distinct multiple columns failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 distinct pairs, got %d", len(results))
	}
}

func TestTursoIntegrationDistinctWithCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

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
}

func TestTursoIntegrationDistinctWithoutArgsWithCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

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
		Select("avatar").Distinct().Count(&count)
	if err != nil {
		t.Errorf("Distinct without args with Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestTursoIntegrationDistinctWithSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	query := db.Query()

	for _, user := range []models.User{
		{Name: "distinct_user_1", Avatar: "avatar1"},
		{Name: "distinct_user_1", Avatar: "avatar1"}, // Duplicate
		{Name: "distinct_user_2", Avatar: "avatar1"},
	} {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	type Result struct {
		Name   string
		Avatar string
	}
	var results []Result
	err := query.Model(&models.User{}).
		Select([]string{"name", "avatar"}).
		Distinct().
		Where("name LIKE ?", "distinct_user_%").
		OrderBy("name", "asc").
		Scan(&results)

	if err != nil {
		t.Errorf("Distinct with select failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 distinct results, got %d", len(results))
	}
}
