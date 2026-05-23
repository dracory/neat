package sqlite

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationGroupHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
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

	t.Run("GroupBy multiple columns", func(t *testing.T) {
		type Result struct {
			Name   string
			Avatar string
			Count  int64
		}
		var results []Result
		// In this data set, name+avatar is unique, so count should be 1 for each
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
			Select("name, avatar, COUNT(*) as count").Group("name").Group("avatar").OrderBy("name", "asc").Scan(&results)
		if err != nil {
			t.Errorf("GroupBy multiple columns failed: %v", err)
		}
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
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

	t.Run("Having with aggregate functions", func(t *testing.T) {
		type Result struct {
			Avatar string
			MaxID  uint
		}
		var results []Result
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
			Select("avatar, MAX(id) as max_id").Group("avatar").Having("MAX(id) > 0").OrderBy("avatar", "asc").Scan(&results)
		if err != nil {
			t.Errorf("Having with aggregate functions failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("GroupBy with Having and Where", func(t *testing.T) {
		type Result struct {
			Avatar string
			Count  int64
		}
		var results []Result
		err := db.Query().Model(&models.User{}).
			Where("name LIKE ?", "group_user_%").
			Where("avatar = ?", "avatar2").
			Select("avatar, COUNT(*) as count").
			Group("avatar").
			Having("count > ?", 1).
			Scan(&results)
		if err != nil {
			t.Errorf("GroupBy with Having and Where failed: %v", err)
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
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
			Select("avatar, COUNT(*) as count").Group("avatar").
			Having("COUNT(*) > ?", 1).Having("COUNT(*) < ?", 5).Scan(&results)
		if err != nil {
			t.Errorf("Multiple Having clauses failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Having with subquery callback", func(t *testing.T) {
		type Result struct {
			Avatar string
			Count  int64
		}
		var results []Result
		err := db.Query().Model(&models.User{}).
			Where("name LIKE ?", "group_user_%").
			Group("avatar").
			Select("avatar, COUNT(*) as count").
			Having("COUNT(*) > (?)", func(q contractsorm.Query) contractsorm.Query {
				return q.Model(&models.User{}).Where("avatar = ?", "avatar1").Where("name LIKE ?", "group_user_%").Select("COUNT(*)")
			}).
			Scan(&results)
		if err != nil {
			t.Errorf("Having with subquery callback failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
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
		err := query.Model(&models.User{}).
			Where("name LIKE ?", "group_user_%").
			Group("avatar").
			Select("avatar, COUNT(*) as count").
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
		if results[0].Count != 3 {
			t.Errorf("Expected count 3, got %d", results[0].Count)
		}
	})
}
