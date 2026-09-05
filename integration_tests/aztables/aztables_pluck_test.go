//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationPluckSingleColumn verifies Pluck() collects values
// from a single column across multiple rows.
func TestAztablesIntegrationPluckSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "pluck1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "pluck2"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "pluck3"},
	}
	if err := query.Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var names []string
	err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Pluck("Name", &names)
	if err != nil {
		t.Fatalf("Pluck failed: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}
}
