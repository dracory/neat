//go:build disabled

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationGroupHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Seed data
	users := []models.User{
		{Name: "group_user_1", Avatar: "avatar1"},
		{Name: "group_user_2", Avatar: "avatar1"},
		{Name: "group_user_3", Avatar: "avatar2"},
		{Name: "group_user_4", Avatar: "avatar2"},
		{Name: "group_user_5", Avatar: "avatar2"},
	}

	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("GroupBy single column", func(t *testing.T) {
		type Result struct {
			Avatar string
			Count  int64
		}
		var results []Result
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
			Select("avatar, COUNT(*) as count").Group("avatar").OrderBy("avatar", "asc").Scan(&results)
		if err != nil {
			t.Errorf("GroupBy single column failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		if results[0].Avatar != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", results[0].Avatar)
		}
		if results[0].Count != 2 {
			t.Errorf("Expected count 2, got %d", results[0].Count)
		}
		if results[1].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[1].Avatar)
		}
		if results[1].Count != 3 {
			t.Errorf("Expected count 3, got %d", results[1].Count)
		}
	})

	t.Run("Having clause", func(t *testing.T) {
		type Result struct {
			Avatar string
			Count  int64
		}
		var results []Result
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
			Select("avatar, COUNT(*) as count").Group("avatar").Having("COUNT(*) > ?", 2).Scan(&results)
		if err != nil {
			t.Errorf("Having clause failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[0].Avatar)
		}
		if results[0].Count != 3 {
			t.Errorf("Expected count 3, got %d", results[0].Count)
		}
	})

	t.Run("Multiple Having clauses", func(t *testing.T) {
		type Result struct {
			Avatar string
			Count  int64
		}
		var results []Result
		// Combined: COUNT(*) > 1 AND COUNT(*) < 3
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
			Select("avatar, COUNT(*) as count").
			Group("avatar").
			Having("COUNT(*) > ?", 1).
			Having("COUNT(*) < ?", 3).
			Scan(&results)
		if err != nil {
			t.Errorf("Multiple Having clauses failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Avatar != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", results[0].Avatar)
		}
		if results[0].Count != 2 {
			t.Errorf("Expected count 2, got %d", results[0].Count)
		}
	})

	t.Run("Having with subquery callback", func(t *testing.T) {
		type Result struct {
			Avatar string
			Count  int64
		}
		var results []Result
		// Find avatars where COUNT(*) > (SELECT COUNT(*) FROM users WHERE avatar = 'avatar1' AND name LIKE 'group_user_%')
		// avatar1 count is 2, so it should return avatar2 which has count 3
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
			Select("avatar, COUNT(*) as count").
			Group("avatar").
			Having("COUNT(*) > (?)", func(q contractsorm.Query) contractsorm.Query {
				return q.Model(&models.User{}).Where("avatar = ?", "avatar1").Where("name LIKE ?", "group_user_%").Select("COUNT(*)")
			}).
			Scan(&results)
		if err != nil {
			t.Errorf("Having with subquery callback failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[0].Avatar)
		}
		if results[0].Count != 3 {
			t.Errorf("Expected count 3, got %d", results[0].Count)
		}
	})

	t.Run("Having with subquery in args", func(t *testing.T) {
		type Result struct {
			Avatar string
			Count  int64
		}
		var results []Result
		// Similar to above but with subquery in args
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
			Select("avatar, COUNT(*) as count").
			Group("avatar").
			Having("COUNT(*) = (?)", func(q contractsorm.Query) contractsorm.Query {
				return q.Model(&models.User{}).Where("avatar = ?", "avatar2").Where("name LIKE ?", "group_user_%").Select("COUNT(*)")
			}).
			Scan(&results)
		if err != nil {
			t.Errorf("Having with subquery in args failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[0].Avatar)
		}
	})
}
