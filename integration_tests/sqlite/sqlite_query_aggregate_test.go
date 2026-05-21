//go:build integration

package sqlite

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryAggregate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Seed data
	users := []models.User{
		{Name: "aggregate_user_1", ID: 101, Avatar: "group1"},
		{Name: "aggregate_user_2", ID: 102, Avatar: "group1"},
		{Name: "aggregate_user_3", ID: 103, Avatar: "group2"},
		{Name: "aggregate_user_4", ID: 104, Avatar: "group2"},
	}

	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("Sum", func(t *testing.T) {
		var sum int64
		err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("id", &sum)
		if err != nil {
			t.Errorf("Sum failed: %v", err)
		}
		if sum != 410 { // 101 + 102 + 103 + 104
			t.Errorf("Expected sum 410, got %d", sum)
		}
	})

	t.Run("Sum with where", func(t *testing.T) {
		var sum int64
		err := query.Table("users").Where("avatar = ?", "group1").Sum("id", &sum)
		if err != nil {
			t.Errorf("Sum with where failed: %v", err)
		}
		if sum != 203 { // 101 + 102
			t.Errorf("Expected sum 203, got %d", sum)
		}
	})

	t.Run("Avg", func(t *testing.T) {
		var avg float64
		err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Avg("id", &avg)
		if err != nil {
			t.Errorf("Avg failed: %v", err)
		}
		if avg != 102.5 { // 410 / 4
			t.Errorf("Expected avg 102.5, got %f", avg)
		}
	})

	t.Run("Max", func(t *testing.T) {
		var max int64
		err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Max("id", &max)
		if err != nil {
			t.Errorf("Max failed: %v", err)
		}
		if max != 104 {
			t.Errorf("Expected max 104, got %d", max)
		}
	})

	t.Run("Min", func(t *testing.T) {
		var min int64
		err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Min("id", &min)
		if err != nil {
			t.Errorf("Min failed: %v", err)
		}
		if min != 101 {
			t.Errorf("Expected min 101, got %d", min)
		}
	})

	t.Run("Aggregate with GroupBy", func(t *testing.T) {
		type Result struct {
			Avatar string
			Total  int64
		}
		var results []Result
		err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").
			Select("avatar, SUM(id) as total").Group("avatar").Scan(&results)
		if err != nil {
			t.Errorf("Aggregate with GroupBy failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		for _, res := range results {
			if res.Avatar == "group1" {
				if res.Total != 203 {
					t.Errorf("Expected total 203 for group1, got %d", res.Total)
				}
			} else if res.Avatar == "group2" {
				if res.Total != 207 {
					t.Errorf("Expected total 207 for group2, got %d", res.Total)
				}
			}
		}
	})

	t.Run("Error handling - invalid column", func(t *testing.T) {
		var sum int64
		err := query.Table("users").Sum("invalid; column", &sum)
		if err == nil {
			t.Error("Expected error for invalid column")
		}
		if !strings.Contains(err.Error(), "invalid column name") {
			t.Errorf("Expected error to contain 'invalid column name', got '%s'", err.Error())
		}
	})

	t.Run("Error handling - SQL injection attempt with comment", func(t *testing.T) {
		var sum int64
		err := query.Table("users").Sum("id; DROP TABLE users--", &sum)
		if err == nil {
			t.Error("Expected error for SQL injection attempt")
		}
		if !strings.Contains(err.Error(), "invalid column name") {
			t.Errorf("Expected error to contain 'invalid column name', got '%s'", err.Error())
		}
	})

	t.Run("Error handling - union-based injection", func(t *testing.T) {
		var sum int64
		err := query.Table("users").Sum("id UNION SELECT password FROM users", &sum)
		if err == nil {
			t.Error("Expected error for union-based injection")
		}
		if !strings.Contains(err.Error(), "invalid column name") {
			t.Errorf("Expected error to contain 'invalid column name', got '%s'", err.Error())
		}
	})

	t.Run("Error handling - nil pointer", func(t *testing.T) {
		err := query.Table("users").Sum("id", nil)
		if err == nil {
			t.Error("Expected error for nil pointer")
		}
	})

	t.Run("Empty result set", func(t *testing.T) {
		var sum *int64
		err := query.Table("users").Where("name = ?", "non_existent").Sum("id", &sum)
		if err != nil {
			t.Errorf("Sum with empty result failed: %v", err)
		}
		if sum != nil {
			t.Error("Expected nil for empty result set")
		}

		var avg *float64
		err = query.Table("users").Where("name = ?", "non_existent").Avg("id", &avg)
		if err != nil {
			t.Errorf("Avg with empty result failed: %v", err)
		}
		if avg != nil {
			t.Error("Expected nil for empty result set")
		}
	})

	t.Run("Aggregation with NULL values", func(t *testing.T) {
		// Add a record with NULL avatar (or use a dedicated nullable column)
		// Since User.Bio is *string, we can use it for NULL tests
		u1 := models.User{Name: "null_test_1", ID: 201, Bio: nil}
		bio2 := "some bio"
		u2 := models.User{Name: "null_test_2", ID: 202, Bio: &bio2}

		if err := query.Model(&models.User{}).Create(&u1); err != nil {
			t.Fatalf("Failed to create u1: %v", err)
		}
		if err := query.Model(&models.User{}).Create(&u2); err != nil {
			t.Fatalf("Failed to create u2: %v", err)
		}

		var count int64
		err := query.Table("users").Where("name LIKE ?", "null_test_%").WhereNotNull("bio").Count(&count)
		if err != nil {
			t.Errorf("Count with WhereNotNull failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}

		var sum int64
		err = query.Table("users").Where("name LIKE ?", "null_test_%").Sum("id", &sum)
		if err != nil {
			t.Errorf("Sum with NULL values failed: %v", err)
		}
		if sum != 403 {
			t.Errorf("Expected sum 403, got %d", sum)
		}
	})

	t.Run("Aggregate on non-numeric column", func(t *testing.T) {
		var sum float64
		// SUM on string column in SQLite returns 0.0
		err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("name", &sum)
		if err != nil {
			t.Errorf("Sum on string column failed: %v", err)
		}
		if sum != 0.0 {
			t.Errorf("Expected sum 0.0 for string column, got %f", sum)
		}

		var max string
		err = query.Table("users").Where("name LIKE ?", "aggregate_user_%").Max("name", &max)
		if err != nil {
			t.Errorf("Max on string column failed: %v", err)
		}
		if max != "aggregate_user_4" {
			t.Errorf("Expected 'aggregate_user_4', got '%s'", max)
		}
	})
}
