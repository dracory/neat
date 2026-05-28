package turso

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func seedPaginateTestData(t *testing.T, db *database.Database) {
	query := db.Query()
	for i := 1; i <= 15; i++ {
		user := models.User{
			Name:   "paginate_user_" + string(rune(64+i)),
			Avatar: "avatar",
		}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}
}

func TestTursoIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedPaginateTestData(t, db)

	users := make([]models.User, 0)
	var total int64
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Paginate(1, 5, &users, &total)
	if err != nil {
		t.Errorf("Paginate first page failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
	if users[0].Name != "paginate_user_A" {
		t.Errorf("Expected 'paginate_user_A', got '%s'", users[0].Name)
	}
	if users[4].Name != "paginate_user_E" {
		t.Errorf("Expected 'paginate_user_E', got '%s'", users[4].Name)
	}
}

func TestTursoIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedPaginateTestData(t, db)

	users := make([]models.User, 0)
	var total int64
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Paginate(2, 5, &users, &total)
	if err != nil {
		t.Errorf("Paginate second page failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
	if users[0].Name != "paginate_user_F" {
		t.Errorf("Expected 'paginate_user_F', got '%s'", users[0].Name)
	}
	if users[4].Name != "paginate_user_J" {
		t.Errorf("Expected 'paginate_user_J', got '%s'", users[4].Name)
	}
}

func TestTursoIntegrationPaginateLastPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedPaginateTestData(t, db)

	users := make([]models.User, 0)
	var total int64
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Paginate(3, 5, &users, &total)
	if err != nil {
		t.Errorf("Paginate last page failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
	if users[0].Name != "paginate_user_K" {
		t.Errorf("Expected 'paginate_user_K', got '%s'", users[0].Name)
	}
	if users[4].Name != "paginate_user_O" {
		t.Errorf("Expected 'paginate_user_O', got '%s'", users[4].Name)
	}
}

func TestTursoIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedPaginateTestData(t, db)

	users := make([]models.User, 0)
	var total int64
	err := db.Query().Model(&models.User{}).Where("name <= ?", "paginate_user_E").OrderBy("name", "asc").Paginate(1, 3, &users, &total)
	if err != nil {
		t.Errorf("Pagination with conditions failed: %v", err)
	}
	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
	if users[0].Name != "paginate_user_A" {
		t.Errorf("Expected 'paginate_user_A', got '%s'", users[0].Name)
	}
	if users[2].Name != "paginate_user_C" {
		t.Errorf("Expected 'paginate_user_C', got '%s'", users[2].Name)
	}
}

func TestTursoIntegrationPaginatePageBeyondBounds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedPaginateTestData(t, db)

	var users []models.User
	var total int64
	err := db.Query().Model(&models.User{}).Paginate(10, 5, &users, &total)
	if err != nil {
		t.Errorf("Pagination page beyond bounds failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

func TestTursoIntegrationPaginateEmptyResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedPaginateTestData(t, db)

	var users []models.User
	var total int64
	err := db.Query().Model(&models.User{}).Where("name = ?", "non_existent").Paginate(1, 5, &users, &total)
	if err != nil {
		t.Errorf("Pagination empty results failed: %v", err)
	}
	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

func TestTursoIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedPaginateTestData(t, db)

	var results []struct {
		UserName string
	}
	var total int64
	err := db.Query().Table("users").Select("name as user_name").OrderBy("name", "asc").Paginate(1, 5, &results, &total)
	if err != nil {
		t.Errorf("Paginate with Select aliases failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if len(results) > 0 && results[0].UserName != "paginate_user_A" {
		t.Errorf("Expected 'paginate_user_A', got '%s'", results[0].UserName)
	}
}

func TestTursoIntegrationCountWithSelectAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedPaginateTestData(t, db)

	var count int64
	err := db.Query().Table("users").Select("name as user_name").Count(&count)
	if err != nil {
		t.Errorf("Count with Select alias failed: %v", err)
	}
	if count != 15 {
		t.Errorf("Expected count 15, got %d", count)
	}
}
