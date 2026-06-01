package turso

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/common"
	"github.com/dracory/neat/integration_tests/models"
)

func TestTursoIntegrationGroupBySingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestGroupBySingleColumn(t, db)
}

func TestTursoIntegrationGroupByMultipleColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedGroupHavingTestData(t, db)

	type Result struct {
		Name   string
		Avatar string
		Count  int64
	}
	results := make([]Result, 0)
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
		Select("name, avatar, COUNT(*) as count").Group("name").Group("avatar").OrderBy("name", "asc").Scan(&results)
	if err != nil {
		t.Errorf("GroupBy multiple columns failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

func TestTursoIntegrationHavingClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestHavingClause(t, db)
}

func TestTursoIntegrationHavingWithAggregateFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedGroupHavingTestData(t, db)

	type Result struct {
		Avatar string
		MaxID  uint
	}
	results := make([]Result, 0)
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
		Select("avatar, MAX(id) as max_id").Group("avatar").Having("MAX(id) > 0").OrderBy("avatar", "asc").Scan(&results)
	if err != nil {
		t.Errorf("Having with aggregate functions failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestTursoIntegrationGroupByWithHavingAndWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedGroupHavingTestData(t, db)

	type Result struct {
		Avatar string
		Count  int64
	}
	results := make([]Result, 0)
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
}

func TestTursoIntegrationMultipleHavingClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedGroupHavingTestData(t, db)

	type Result struct {
		Avatar string
		Count  int64
	}
	results := make([]Result, 0)
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "group_user_%").
		Select("avatar, COUNT(*) as count").Group("avatar").
		Having("COUNT(*) > ?", 1).Having("COUNT(*) < ?", 5).Scan(&results)
	if err != nil {
		t.Errorf("Multiple Having clauses failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestTursoIntegrationHavingWithSubqueryCallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedGroupHavingTestData(t, db)

	type Result struct {
		Avatar string
		Count  int64
	}
	results := make([]Result, 0)
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
}

func TestTursoIntegrationHavingWithSubqueryInArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedGroupHavingTestData(t, db)
	query := db.Query()

	type Result struct {
		Avatar string
		Count  int64
	}
	results := make([]Result, 0)
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
}
