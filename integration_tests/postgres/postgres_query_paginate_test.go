//go:build disabled

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationPaginate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Seed data
	for i := 1; i <= 15; i++ {
		user := models.User{
			Name:   "paginate_user_" + string(rune(64+i)), // A, B, C...
			Avatar: "avatar",
		}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("Paginate method - first page", func(t *testing.T) {
		var users []models.User
		var total int64
		err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Paginate(1, 5, &users, &total)
		if err != nil {
			t.Errorf("Paginate failed: %v", err)
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
	})

	t.Run("Pagination with conditions", func(t *testing.T) {
		var users []models.User
		var total int64
		err := db.Query().Model(&models.User{}).Where("name <= ?", "paginate_user_E").OrderBy("name", "asc").Paginate(1, 3, &users, &total)
		if err != nil {
			t.Errorf("Paginate with conditions failed: %v", err)
		}
		if total != 5 {
			t.Errorf("Expected total 5, got %d", total)
		}
		if len(users) != 3 {
			t.Errorf("Expected 3 users, got %d", len(users))
		}
	})

	t.Run("Pagination with Select aliases", func(t *testing.T) {
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
	})
}
