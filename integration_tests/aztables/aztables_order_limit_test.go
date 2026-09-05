//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationLimitClause verifies that LIMIT caps the result set
// via the ORM query builder.
func TestAztablesIntegrationLimitClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "limit1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "limit2"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "limit3"},
		{PartitionKey: "pk1", RowKey: "rk4", Name: "limit4"},
		{PartitionKey: "pk1", RowKey: "rk5", Name: "limit5"},
	}
	if err := query.Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var found []AzUser
	err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Limit(3).
		Find(&found)
	if err != nil {
		t.Fatalf("Find with LIMIT failed: %v", err)
	}
	if len(found) != 3 {
		t.Errorf("Expected 3 rows with LIMIT 3, got %d", len(found))
	}
}

// TestAztablesIntegrationLimitOne verifies LIMIT 1 returns a single row.
func TestAztablesIntegrationLimitOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "limit_one1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "limit_one2"},
	}
	if err := query.Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var found []AzUser
	err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Limit(1).
		Find(&found)
	if err != nil {
		t.Fatalf("Find with LIMIT 1 failed: %v", err)
	}
	if len(found) != 1 {
		t.Errorf("Expected 1 row with LIMIT 1, got %d", len(found))
	}
}
