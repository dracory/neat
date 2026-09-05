//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationQueryCountBasic verifies the ORM Count method returns
// the correct number of entities in a partition.
func TestAztablesIntegrationQueryCountBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "count_user1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "count_user2"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "count_user3"},
	}
	if err := query.Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var count int64
	err := query.Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Count(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

// TestAztablesIntegrationQueryCountWithWhere verifies Count with a WHERE
// filter on a non-key column.
func TestAztablesIntegrationQueryCountWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "count_a", Avatar: "group1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "count_b", Avatar: "group1"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "count_c", Avatar: "group2"},
	}
	if err := query.Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var count int64
	err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("Avatar = ?", "group1").
		Count(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2 for group1, got %d", count)
	}
}

// TestAztablesIntegrationQueryCountEmpty verifies Count returns 0 for a
// partition with no entities.
func TestAztablesIntegrationQueryCountEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	var count int64
	err := query.Model(&AzUser{}).Where("PartitionKey = ?", "nonexistent").Count(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for empty partition, got %d", count)
	}
}
