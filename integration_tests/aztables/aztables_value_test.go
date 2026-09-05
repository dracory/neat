//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationQueryValueBasic verifies Value() returns a single
// column value from the first matching row.
func TestAztablesIntegrationQueryValueBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "value_user", Avatar: "value_avatar"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var name string
	err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Value("Name", &name)
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
	if name != "value_user" {
		t.Errorf("Expected 'value_user', got '%s'", name)
	}
}

// TestAztablesIntegrationQueryValueWithWhere verifies Value() with a WHERE
// filter on a non-key column.
func TestAztablesIntegrationQueryValueWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "vw1", Avatar: "alpha"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "vw2", Avatar: "beta"},
	}
	if err := query.Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var avatar string
	err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("Name = ?", "vw2").
		Value("Avatar", &avatar)
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
	if avatar != "beta" {
		t.Errorf("Expected 'beta', got '%s'", avatar)
	}
}
