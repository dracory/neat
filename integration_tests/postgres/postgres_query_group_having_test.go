//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationGroupBySingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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
}

func TestPostgresIntegrationHavingClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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
}

func TestPostgresIntegrationMultipleHavingClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
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
	if results[0].Avatar != "avatar1" {
		t.Errorf("Expected 'avatar1', got '%s'", results[0].Avatar)
	}
	if results[0].Count != 2 {
		t.Errorf("Expected count 2, got %d", results[0].Count)
	}
}

func TestPostgresIntegrationHavingWithSubqueryCallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping HAVING with subquery callback - subquery parameter numbering not yet implemented for PostgreSQL")
}

func TestPostgresIntegrationHavingWithSubqueryInArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping HAVING with subquery in args test - subquery parameter numbering not implemented for PostgreSQL")
}
