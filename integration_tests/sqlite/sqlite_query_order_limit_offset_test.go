package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationOrderLimitOffset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Seed data
	users := []models.User{
		{Name: "user_b", Avatar: "avatar2"},
		{Name: "user_a", Avatar: "avatar1"},
		{Name: "user_c", Avatar: "avatar3"},
		{Name: "user_d", Avatar: "avatar4"},
		{Name: "user_e", Avatar: "avatar5"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	t.Run("OrderBy ascending", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Get(&results)
		if err != nil {
			t.Errorf("OrderBy ascending failed: %v", err)
		}
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}
		if results[0].Name != "user_a" {
			t.Errorf("Expected 'user_a', got '%s'", results[0].Name)
		}
		if results[1].Name != "user_b" {
			t.Errorf("Expected 'user_b', got '%s'", results[1].Name)
		}
		if results[2].Name != "user_c" {
			t.Errorf("Expected 'user_c', got '%s'", results[2].Name)
		}
	})

	t.Run("OrderBy descending", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).OrderBy("name", "desc").Get(&results)
		if err != nil {
			t.Errorf("OrderBy descending failed: %v", err)
		}
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}
		if results[0].Name != "user_e" {
			t.Errorf("Expected 'user_e', got '%s'", results[0].Name)
		}
		if results[1].Name != "user_d" {
			t.Errorf("Expected 'user_d', got '%s'", results[1].Name)
		}
		if results[2].Name != "user_c" {
			t.Errorf("Expected 'user_c', got '%s'", results[2].Name)
		}
	})

	t.Run("OrderByDesc method", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).OrderByDesc("name").Get(&results)
		if err != nil {
			t.Errorf("OrderByDesc method failed: %v", err)
		}
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}
		if results[0].Name != "user_e" {
			t.Errorf("Expected 'user_e', got '%s'", results[0].Name)
		}
	})

	t.Run("Multiple OrderBy clauses", func(t *testing.T) {
		// Add another user with same name but different avatar
		if err := query.Model(&models.User{}).Create(&models.User{Name: "user_a", Avatar: "avatar0"}); err != nil {
			t.Fatalf("Failed to create additional user: %v", err)
		}

		var results []models.User
		err := db.Query().Model(&models.User{}).OrderBy("name", "asc").OrderBy("avatar", "asc").Get(&results)
		if err != nil {
			t.Errorf("Multiple OrderBy clauses failed: %v", err)
		}

		// Find "user_a" entries
		var userAEntries []models.User
		for _, u := range results {
			if u.Name == "user_a" {
				userAEntries = append(userAEntries, u)
			}
		}
		if len(userAEntries) != 2 {
			t.Errorf("Expected 2 user_a entries, got %d", len(userAEntries))
		}
		if userAEntries[0].Avatar != "avatar0" {
			t.Errorf("Expected 'avatar0', got '%s'", userAEntries[0].Avatar)
		}
		if userAEntries[1].Avatar != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", userAEntries[1].Avatar)
		}
	})

	t.Run("OrderBy with expressions", func(t *testing.T) {
		var results []models.User
		// SQLite expression
		err := db.Query().Model(&models.User{}).Order("LENGTH(name) DESC").OrderBy("name", "asc").Get(&results)
		if err != nil {
			t.Errorf("OrderBy with expressions failed: %v", err)
		}
		if len(results) == 0 {
			t.Error("Expected non-empty results")
		}
	})

	t.Run("Limit clause", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).Limit(2).Get(&results)
		if err != nil {
			t.Errorf("Limit clause failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Limit with OrderBy", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Limit(2).Get(&results)
		if err != nil {
			t.Errorf("Limit with OrderBy failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].Name != "user_a" {
			t.Errorf("Expected 'user_a', got '%s'", results[0].Name)
		}
		if results[1].Name != "user_a" { // From multiple orderby test
			t.Errorf("Expected 'user_a', got '%s'", results[1].Name)
		}
	})

	t.Run("Limit edge cases (zero)", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).Limit(0).Get(&results)
		if err != nil {
			t.Errorf("Limit edge case (zero) failed: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("Limit edge cases (negative)", func(t *testing.T) {
		var results []models.User
		// GORM/Eloquent usually handles negative limit by ignoring it or applying it literally.
		// In most SQL dialects, negative limit is an error or ignored.
		// Eloquent seems to just pass it to GORM.
		err := db.Query().Model(&models.User{}).Limit(-1).Get(&results)
		if err != nil {
			t.Errorf("Limit edge case (negative) failed: %v", err)
		}
		// Should return all if negative limit is ignored or treated as "no limit"
		if len(results) < 5 {
			t.Errorf("Expected at least 5 results, got %d", len(results))
		}
	})

	t.Run("Offset clause", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Offset(2).Get(&results)
		if err != nil {
			t.Errorf("Offset clause failed: %v", err)
		}
		// Total was 5 + 1 added in "Multiple OrderBy clauses" = 6.
		if len(results) != 4 {
			t.Errorf("Expected 4 results, got %d", len(results))
		}
	})

	t.Run("Offset with Limit", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Offset(2).Limit(2).Get(&results)
		if err != nil {
			t.Errorf("Offset with Limit failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		// Sorted: user_a (avatar0), user_a (avatar1), user_b, user_c, user_d, user_e
		// Offset 2 skip user_a, user_a -> should start with user_b
		if results[0].Name != "user_b" {
			t.Errorf("Expected 'user_b', got '%s'", results[0].Name)
		}
		if results[1].Name != "user_c" {
			t.Errorf("Expected 'user_c', got '%s'", results[1].Name)
		}
	})

	t.Run("InRandomOrder method", func(t *testing.T) {
		var results1 []models.User
		err := db.Query().Model(&models.User{}).InRandomOrder().Get(&results1)
		if err != nil {
			t.Errorf("InRandomOrder method failed: %v", err)
		}

		var results2 []models.User
		err = db.Query().Model(&models.User{}).InRandomOrder().Get(&results2)
		if err != nil {
			t.Errorf("InRandomOrder method (second) failed: %v", err)
		}

		// It's technically possible for two random orders to be the same,
		// but with 6 items it's unlikely (1/720).
		// We just check it returns the correct count.
		if len(results1) != 6 {
			t.Errorf("Expected 6 results, got %d", len(results1))
		}
		if len(results2) != 6 {
			t.Errorf("Expected 6 results, got %d", len(results2))
		}
	})

	t.Run("Random ordering with Limit", func(t *testing.T) {
		var results []models.User
		err := db.Query().Model(&models.User{}).InRandomOrder().Limit(3).Get(&results)
		if err != nil {
			t.Errorf("Random ordering with Limit failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})
}
