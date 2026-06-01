package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryAggregateSum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateSum(t, db)
}

func TestPostgresIntegrationQueryAggregateSumWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateSumWithWhere(t, db)
}

func TestPostgresIntegrationQueryAggregateAvg(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateAvg(t, db)
}

func TestPostgresIntegrationQueryAggregateMax(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateMax(t, db)
}

func TestPostgresIntegrationQueryAggregateMin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateMin(t, db)
}

func TestPostgresIntegrationQueryAggregateGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateGroupBy(t, db)
}

func TestPostgresIntegrationQueryAggregateInvalidColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateInvalidColumn(t, db)
}

func TestPostgresIntegrationQueryAggregateNilPointer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateNilPointer(t, db)
}

func TestPostgresIntegrationQueryAggregateEmptyResult(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestAggregateEmptyResult(t, db)
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

	// PostgreSQL does not allow SUM on string columns, so we expect an error
	var sum float64
	err := query.Table("users").Where("name LIKE ?", "aggregate_user_%").Sum("name", &sum)
	if err == nil {
		t.Error("Expected error for SUM on string column, got nil")
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
