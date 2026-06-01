package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
	"github.com/dracory/neat/integration_tests/models"
)

func TestTursoIntegrationPluckSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedPluckTestData(t, db)

	names := make([]string, 0)
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
}

func TestTursoIntegrationPluckWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedPluckTestData(t, db)

	names := make([]string, 0)
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
}

func TestTursoIntegrationPluckIntoMaps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedPluckTestData(t, db)

	var results []map[string]any
	err := db.Query().Model(&models.User{}).Where("name = ?", "pluck_user_1").Select("name, avatar").Scan(&results)
	if err != nil {
		t.Errorf("Pluck into maps failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "pluck_user_1" {
		t.Errorf("Expected 'pluck_user_1', got '%v'", results[0]["name"])
	}
}

func TestTursoIntegrationPluckDuplicates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedPluckTestData(t, db)

	avatars := make([]string, 0)
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
}

func TestTursoIntegrationPluckEmptyResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedPluckTestData(t, db)

	var names []string
	err := db.Query().Model(&models.User{}).Where("name = ?", "non_existent").Pluck("name", &names)
	if err != nil {
		t.Errorf("Pluck with empty results failed: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("Expected 0 names, got %d", len(names))
	}
}

func TestTursoIntegrationPluckDistinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestPluckWithDistinct(t, db)
}
