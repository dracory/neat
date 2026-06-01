package common

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

// SeedGroupHavingTestData creates test data for group/having tests
func SeedGroupHavingTestData(t *testing.T, db *database.Database) {
	query := db.Query()
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
}

// TestGroupBySingleColumn tests grouping by a single column
func TestGroupBySingleColumn(t *testing.T, db *database.Database) {
	SeedGroupHavingTestData(t, db)

	type Result struct {
		Avatar string
		Count  int64
	}
	var results []Result
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
		Select("avatar, COUNT(*) as count").Group("avatar").OrderBy("avatar", "asc").Scan(&results)
	if err != nil {
		t.Errorf("GroupBy failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if len(results) >= 2 {
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
	}
}

// TestHavingClause tests HAVING clause functionality
func TestHavingClause(t *testing.T, db *database.Database) {
	SeedGroupHavingTestData(t, db)

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
	if len(results) >= 1 {
		if results[0].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[0].Avatar)
		}
		if results[0].Count != 3 {
			t.Errorf("Expected count 3, got %d", results[0].Count)
		}
	}
}

// TestMultipleHavingClauses tests multiple HAVING clauses
func TestMultipleHavingClauses(t *testing.T, db *database.Database) {
	SeedGroupHavingTestData(t, db)

	type Result struct {
		Avatar string
		Count  int64
	}
	var results []Result
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
	if len(results) >= 1 {
		if results[0].Avatar != "avatar1" {
			t.Errorf("Expected 'avatar1', got '%s'", results[0].Avatar)
		}
		if results[0].Count != 2 {
			t.Errorf("Expected count 2, got %d", results[0].Count)
		}
	}
}

// TestHavingWithSubqueryCallback tests HAVING with subquery callback
func TestHavingWithSubqueryCallback(t *testing.T, db *database.Database) {
	SeedGroupHavingTestData(t, db)

	type Result struct {
		Avatar string
		Count  int64
	}
	var results []Result
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
	if len(results) >= 1 {
		if results[0].Avatar != "avatar2" {
			t.Errorf("Expected 'avatar2', got '%s'", results[0].Avatar)
		}
		if results[0].Count != 3 {
			t.Errorf("Expected count 3, got %d", results[0].Count)
		}
	}
}

// TestHavingWithSubqueryInArgs tests HAVING with subquery in arguments
func TestHavingWithSubqueryInArgs(t *testing.T, db *database.Database) {
	SeedGroupHavingTestData(t, db)

	type Result struct {
		Avatar string
		Count  int64
	}
	var results []Result
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
	if len(results) >= 1 && results[0].Avatar != "avatar2" {
		t.Errorf("Expected 'avatar2', got '%s'", results[0].Avatar)
	}
}
