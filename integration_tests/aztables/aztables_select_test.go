//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationQuerySelectSpecificColumns verifies Select() projects
// only the requested columns (OData $select).
func TestAztablesIntegrationQuerySelectSpecificColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "select_user", Avatar: "select_avatar"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var foundUser AzUser
	err := query.Model(&AzUser{}).
		Select("Name").
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&foundUser)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	if foundUser.Name != "select_user" {
		t.Errorf("Expected 'select_user', got '%s'", foundUser.Name)
	}
	// Avatar should not be projected
	if foundUser.Avatar != "" {
		t.Errorf("Expected empty avatar (not projected), got '%s'", foundUser.Avatar)
	}
}

// TestAztablesIntegrationQuerySelectMultipleColumns verifies Select with
// multiple columns.
func TestAztablesIntegrationQuerySelectMultipleColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "multi_select", Avatar: "avatar_val"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var foundUser AzUser
	err := query.Model(&AzUser{}).
		Select("Name", "Avatar").
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&foundUser)
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	if foundUser.Name != "multi_select" {
		t.Errorf("Expected 'multi_select', got '%s'", foundUser.Name)
	}
	if foundUser.Avatar != "avatar_val" {
		t.Errorf("Expected 'avatar_val', got '%s'", foundUser.Avatar)
	}
}
