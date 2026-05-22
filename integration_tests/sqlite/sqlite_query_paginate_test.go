package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationPaginate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
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
	})

	t.Run("Paginate method - second page", func(t *testing.T) {
		var users []models.User
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
	})

	t.Run("Paginate method - last page", func(t *testing.T) {
		var users []models.User
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
	})

	t.Run("Pagination with conditions", func(t *testing.T) {
		var users []models.User
		var total int64
		// Only users A-E
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
	})

	t.Run("Pagination edge cases - page beyond bounds", func(t *testing.T) {
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
	})

	t.Run("Pagination edge cases - empty results", func(t *testing.T) {
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
	})

	t.Run("Pagination with Select aliases", func(t *testing.T) {
		var results []struct {
			UserName string
		}
		var total int64
		// GORM's Count doesn't handle aliases in Select well if they are passed as a single string.
		// We might need to fix Paginate to handle this.
		err := db.Query().Table("users").Select("name as user_name").OrderBy("name", "asc").Paginate(1, 5, &results, &total)
		if err != nil {
			t.Errorf("Pagination with Select aliases failed: %v", err)
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

	t.Run("Count with Select alias (sanity check)", func(t *testing.T) {
		var total int64
		err := db.Query().Table("users").Select("name as user_name").Count(&total)
		// If this fails, then it's a general Count issue with Select aliases in our implementation
		if err != nil {
			t.Errorf("Count with Select alias failed: %v", err)
		}
		if total != 15 {
			t.Errorf("Expected total 15, got %d", total)
		}
	})
}
