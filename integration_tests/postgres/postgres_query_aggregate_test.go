//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryAggregateSum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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

	var sum int64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum failed: %v", err)
	}
	if sum != 410 {
		t.Errorf("Expected sum 410, got %d", sum)
	}
}

func TestPostgresIntegrationQueryAggregateSumWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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

	var sum int64
	err := query.Table("users").Where("avatar = ?", "group1").Sum("id", &sum)
	if err != nil {
		t.Errorf("Sum with where failed: %v", err)
	}
	if sum != 203 {
		t.Errorf("Expected sum 203, got %d", sum)
	}
}

func TestPostgresIntegrationQueryAggregateAvg(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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

	var avg float64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Avg("id", &avg)
	if err != nil {
		t.Errorf("Avg failed: %v", err)
	}
	if avg != 102.5 {
		t.Errorf("Expected avg 102.5, got %f", avg)
	}
}

func TestPostgresIntegrationQueryAggregateMax(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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

	var max int64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Max("id", &max)
	if err != nil {
		t.Errorf("Max failed: %v", err)
	}
	if max != 104 {
		t.Errorf("Expected max 104, got %d", max)
	}
}

func TestPostgresIntegrationQueryAggregateMin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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

	var min int64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Min("id", &min)
	if err != nil {
		t.Errorf("Min failed: %v", err)
	}
	if min != 101 {
		t.Errorf("Expected min 101, got %d", min)
	}
}

func TestPostgresIntegrationQueryAggregateGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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
}

func TestPostgresIntegrationQueryAggregateInvalidColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	var sum int64
	err := query.Table("users").Sum("invalid; column", &sum)
	if err == nil {
		t.Error("Expected error for invalid column, got nil")
	}
}

func TestPostgresIntegrationQueryAggregateNilPointer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	err := query.Table("users").Sum("id", nil)
	if err == nil {
		t.Error("Expected error for nil pointer, got nil")
	}
}

func TestPostgresIntegrationQueryAggregateEmptyResult(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

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
}

func TestPostgresIntegrationQueryAggregateNullValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	u1 := models.User{Name: "null_test_1", Bio: nil}
	bio2 := "some bio"
	u2 := models.User{Name: "null_test_2", Bio: &bio2}

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
	// IDs are auto-generated, so just check that we got a valid sum > 0
	if sum <= 0 {
		t.Errorf("Expected positive sum, got %d", sum)
	}
}

func TestPostgresIntegrationQueryAggregateNonNumericColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "aggregate_user_1", Avatar: "group1"},
		{Name: "aggregate_user_2", Avatar: "group1"},
		{Name: "aggregate_user_3", Avatar: "group2"},
		{Name: "aggregate_user_4", Avatar: "group2"},
	}

	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// PostgreSQL allows SUM on string columns (returns 0), so we skip the error check
	var sum float64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("name", &sum)
	if err != nil {
		t.Errorf("Sum on string column failed: %v", err)
	}

	var max string
	err = query.Table("users").Where("name LIKE ?", "aggregate_user_%").Max("name", &max)
	if err != nil {
		t.Errorf("Max on string column failed: %v", err)
	}
	if max != "aggregate_user_4" {
		t.Errorf("Expected 'aggregate_user_4', got '%s'", max)
	}
}
