//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationUpdateByModel verifies updating via Model().
//
// Note: Azure Table Storage UPDATE requires WHERE on both PartitionKey
// AND RowKey — you cannot update by partition alone or by non-key columns.
func TestAztablesIntegrationUpdateByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "update_model", Avatar: "original"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// UPDATE requires WHERE PartitionKey = ? AND RowKey = ?
	res, err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Update("Avatar", "updated_avatar")
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	// Verify
	var updatedUser AzUser
	_ = query.Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Where("RowKey = ?", "rk1").First(&updatedUser)
	if updatedUser.Avatar != "updated_avatar" {
		t.Errorf("Expected 'updated_avatar', got '%s'", updatedUser.Avatar)
	}
}

// TestAztablesIntegrationUpdateByTable verifies updating via Table().
func TestAztablesIntegrationUpdateByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "table_update", Avatar: "original"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := query.Table("azusers").
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Update("Avatar", "table_updated")
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var found AzUser
	_ = query.Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Where("RowKey = ?", "rk1").First(&found)
	if found.Avatar != "table_updated" {
		t.Errorf("Expected 'table_updated', got '%s'", found.Avatar)
	}
}

// TestAztablesIntegrationUpdateMultipleColumns verifies updating multiple
// columns via a map.
func TestAztablesIntegrationUpdateMultipleColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "multi_col", Avatar: "original"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	res, err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Update(map[string]any{
			"Avatar": "multi_updated",
			"Name":   "multi_updated_name",
		})
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", res.RowsAffected)
	}

	var found AzUser
	_ = query.Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Where("RowKey = ?", "rk1").First(&found)
	if found.Name != "multi_updated_name" {
		t.Errorf("Expected 'multi_updated_name', got '%s'", found.Name)
	}
	if found.Avatar != "multi_updated" {
		t.Errorf("Expected 'multi_updated', got '%s'", found.Avatar)
	}
}

// TestAztablesIntegrationUpdateWithWhere verifies that UPDATE only affects
// the row matching PartitionKey + RowKey.
//
// Note: Azure Table Storage UPDATE only supports WHERE on PartitionKey,
// RowKey, and ETag. To update a specific entity, use both PartitionKey
// and RowKey in the WHERE clause.
func TestAztablesIntegrationUpdateWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "where_update", Avatar: "group1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "where_update", Avatar: "group2"},
	}
	if err := query.Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Update only rk1 using PartitionKey + RowKey
	_, err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Update("Name", "updated_group1")
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	var u1, u2 AzUser
	_ = query.Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Where("RowKey = ?", "rk1").First(&u1)
	_ = query.Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Where("RowKey = ?", "rk2").First(&u2)
	if u1.Name != "updated_group1" {
		t.Errorf("rk1: expected 'updated_group1', got '%s'", u1.Name)
	}
	if u2.Name != "where_update" {
		t.Errorf("rk2: expected 'where_update' (unchanged), got '%s'", u2.Name)
	}
}
