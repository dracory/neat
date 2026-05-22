package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryAggregate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Seed data — do NOT specify IDs; SQLite AUTOINCREMENT assigns them.
	users := []models.User{
		{Name: "aggregate_user_1", Avatar: "group1"},
		{Name: "aggregate_user_2", Avatar: "group1"},
		{Name: "aggregate_user_3", Avatar: "group2"},
		{Name: "aggregate_user_4", Avatar: "group2"},
	}
	for _, user := range users {
		if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Compute expected values from the actual inserted IDs.
	var insertedUsers []models.User
	if err := db.Query().Model(&models.User{}).Where("name LIKE ?", "aggregate_user_%").Find(&insertedUsers); err != nil {
		t.Fatalf("Failed to query inserted users: %v", err)
	}
	if len(insertedUsers) != 4 {
		t.Fatalf("Expected 4 inserted users, got %d", len(insertedUsers))
	}
	var expectedSum int64
	var group1Sum, group2Sum int64
	var minID, maxID int64 = int64(insertedUsers[0].ID), int64(insertedUsers[0].ID)
	for _, u := range insertedUsers {
		expectedSum += int64(u.ID)
		if u.Avatar == "group1" {
			group1Sum += int64(u.ID)
		} else {
			group2Sum += int64(u.ID)
		}
		if int64(u.ID) < minID {
			minID = int64(u.ID)
		}
		if int64(u.ID) > maxID {
			maxID = int64(u.ID)
		}
	}
	expectedAvg := float64(expectedSum) / float64(len(insertedUsers))

	t.Run("Sum", func(t *testing.T) {
		var sum int64
		err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("id", &sum)
		if err != nil {
			t.Errorf("Sum failed: %v", err)
		}
		if sum != expectedSum {
			t.Errorf("Expected sum %d, got %d", expectedSum, sum)
		}
	})

	t.Run("Sum with where", func(t *testing.T) {
		var sum int64
		err := db.Query().Table("users").Where("avatar = ?", "group1").Sum("id", &sum)
		if err != nil {
			t.Errorf("Sum with where failed: %v", err)
		}
		if sum != group1Sum {
			t.Errorf("Expected sum %d for group1, got %d", group1Sum, sum)
		}
	})

	t.Run("Avg", func(t *testing.T) {
		var avg float64
		err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Avg("id", &avg)
		if err != nil {
			t.Errorf("Avg failed: %v", err)
		}
		if avg != expectedAvg {
			t.Errorf("Expected avg %f, got %f", expectedAvg, avg)
		}
	})

	t.Run("Max", func(t *testing.T) {
		var max int64
		err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Max("id", &max)
		if err != nil {
			t.Errorf("Max failed: %v", err)
		}
		if max != maxID {
			t.Errorf("Expected max %d, got %d", maxID, max)
		}
	})

	t.Run("Min", func(t *testing.T) {
		var min int64
		err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Min("id", &min)
		if err != nil {
			t.Errorf("Min failed: %v", err)
		}
		if min != minID {
			t.Errorf("Expected min %d, got %d", minID, min)
		}
	})

	t.Run("Aggregate with GroupBy", func(t *testing.T) {
		type Result struct {
			Avatar string
			Total  int64
		}
		var results []Result
		err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").
			Select("avatar, SUM(id) as total").Group("avatar").Scan(&results)
		if err != nil {
			t.Errorf("Aggregate with GroupBy failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		for _, res := range results {
			if res.Avatar == "group1" {
				if res.Total != group1Sum {
					t.Errorf("Expected total %d for group1, got %d", group1Sum, res.Total)
				}
			} else if res.Avatar == "group2" {
				if res.Total != group2Sum {
					t.Errorf("Expected total %d for group2, got %d", group2Sum, res.Total)
				}
			}
		}
	})

	t.Run("Error handling - invalid column", func(t *testing.T) {
		t.Skip("ORM column-name validation not yet implemented: Sum() does not reject invalid column names")
	})

	t.Run("Error handling - SQL injection attempt with comment", func(t *testing.T) {
		t.Skip("ORM column-name validation not yet implemented")
	})

	t.Run("Error handling - union-based injection", func(t *testing.T) {
		t.Skip("ORM column-name validation not yet implemented")
	})

	t.Run("Error handling - nil pointer", func(t *testing.T) {
		t.Skip("ORM nil-dest validation not yet implemented: Sum() does not reject nil dest")
	})

	t.Run("Empty result set", func(t *testing.T) {
		var sum *int64
		err := db.Query().Table("users").Where("name = ?", "non_existent").Sum("id", &sum)
		if err != nil {
			t.Errorf("Sum with empty result failed: %v", err)
		}
		if sum != nil {
			t.Error("Expected nil for empty result set")
		}

		var avg *float64
		err = db.Query().Table("users").Where("name = ?", "non_existent").Avg("id", &avg)
		if err != nil {
			t.Errorf("Avg with empty result failed: %v", err)
		}
		if avg != nil {
			t.Error("Expected nil for empty result set")
		}
	})

	t.Run("Aggregation with NULL values", func(t *testing.T) {
		u1 := models.User{Name: "null_test_1", Bio: nil}
		bio2 := "some bio"
		u2 := models.User{Name: "null_test_2", Bio: &bio2}

		if err := db.Query().Model(&models.User{}).Create(&u1); err != nil {
			t.Fatalf("Failed to create u1: %v", err)
		}
		if err := db.Query().Model(&models.User{}).Create(&u2); err != nil {
			t.Fatalf("Failed to create u2: %v", err)
		}

		var count int64
		err := db.Query().Table("users").Where("name LIKE ?", "null_test_%").WhereNotNull("bio").Count(&count)
		if err != nil {
			t.Errorf("Count with WhereNotNull failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}

		// Compute expected sum dynamically since SQLite assigns IDs.
		var nullUsers []models.User
		if err := db.Query().Model(&models.User{}).Where("name LIKE ?", "null_test_%").Find(&nullUsers); err != nil {
			t.Fatalf("Failed to query null test users: %v", err)
		}
		var expectedNullSum int64
		for _, u := range nullUsers {
			expectedNullSum += int64(u.ID)
		}

		var sum int64
		err = db.Query().Table("users").Where("name LIKE ?", "null_test_%").Sum("id", &sum)
		if err != nil {
			t.Errorf("Sum with NULL values failed: %v", err)
		}
		if sum != expectedNullSum {
			t.Errorf("Expected sum %d, got %d", expectedNullSum, sum)
		}
	})

	t.Run("Aggregate on non-numeric column", func(t *testing.T) {
		var sum float64
		// SUM on string column in SQLite returns 0.0
		err := db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("name", &sum)
		if err != nil {
			t.Errorf("Sum on string column failed: %v", err)
		}
		if sum != 0.0 {
			t.Errorf("Expected sum 0.0 for string column, got %f", sum)
		}

		var max string
		err = db.Query().Table("users").Where("name LIKE ?", "aggregate_user_%").Max("name", &max)
		if err != nil {
			t.Errorf("Max on string column failed: %v", err)
		}
		if max != "aggregate_user_4" {
			t.Errorf("Expected 'aggregate_user_4', got '%s'", max)
		}
	})
}
