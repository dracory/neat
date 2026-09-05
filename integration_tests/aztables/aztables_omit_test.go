//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationQueryOmitDuringSelect verifies Omit() excludes
// columns from the SELECT projection.
func TestAztablesIntegrationQueryOmitDuringSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "omit_user", Avatar: "omit_avatar"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var foundUser AzUser
	err := query.Model(&AzUser{}).
		Omit("Avatar").
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&foundUser)
	if err != nil {
		t.Errorf("Omit during select failed: %v", err)
	}
	if foundUser.Name != "omit_user" {
		t.Errorf("Expected 'omit_user', got '%s'", foundUser.Name)
	}
	if foundUser.Avatar != "" {
		t.Errorf("Expected empty avatar (omitted), got '%s'", foundUser.Avatar)
	}
}

// TestAztablesIntegrationQueryOmitDuringUpdate verifies Omit() excludes
// columns from the UPDATE SET clause.
func TestAztablesIntegrationQueryOmitDuringUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "update_omit_user", Avatar: "original_avatar"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err := query.Model(&AzUser{}).
		Omit("Avatar").
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Update("Name", "updated_name")
	if err != nil {
		t.Errorf("Omit during update failed: %v", err)
	}

	var foundUser AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&foundUser)
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if foundUser.Name != "updated_name" {
		t.Errorf("Expected 'updated_name', got '%s'", foundUser.Name)
	}
	// Avatar should be preserved (was omitted from update)
	if foundUser.Avatar != "original_avatar" {
		t.Errorf("Expected 'original_avatar' (preserved), got '%s'", foundUser.Avatar)
	}
}
