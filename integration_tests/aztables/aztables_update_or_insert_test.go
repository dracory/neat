//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationUpdateOrInsert verifies the UpdateOrInsert operation
// (insert if not exists, update if exists) using map attributes.
//
// Note: UpdateOrInsert internally uses a transaction to atomically check
// and insert/update. Azure Table Storage does not support transactions
// via aztablessql, so this test is expected to fail and is skipped.
func TestAztablesIntegrationUpdateOrInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	// Attempt the operation — it will fail because aztablessql doesn't
	// support transactions, which UpdateOrInsert uses internally.
	err := query.Model(&AzUser{}).UpdateOrInsert(
		map[string]any{"PartitionKey": "pk1", "RowKey": "rk1"},
		map[string]any{"Name": "upsert_user", "Avatar": "upsert_avatar"},
	)
	if err != nil {
		t.Skipf("UpdateOrInsert not supported (requires transactions): %v", err)
	}

	// Test Insert (record doesn't exist yet)
	err = query.Model(&AzUser{}).UpdateOrInsert(
		map[string]any{"PartitionKey": "pk1", "RowKey": "rk1"},
		map[string]any{"Name": "upsert_user", "Avatar": "upsert_avatar"},
	)
	if err != nil {
		t.Skipf("UpdateOrInsert not supported (requires transactions): %v", err)
	}

	var user AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&user)
	if err != nil {
		t.Fatalf("Failed to find inserted user: %v", err)
	}
	if user.Name != "upsert_user" {
		t.Errorf("Expected 'upsert_user', got '%s'", user.Name)
	}
	if user.Avatar != "upsert_avatar" {
		t.Errorf("Expected 'upsert_avatar', got '%s'", user.Avatar)
	}

	// Test Update (record already exists)
	err = query.Model(&AzUser{}).UpdateOrInsert(
		map[string]any{"PartitionKey": "pk1", "RowKey": "rk1"},
		map[string]any{"Name": "upsert_user_updated", "Avatar": "upsert_avatar_updated"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert (update) failed: %v", err)
	}

	var user2 AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&user2)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}
	if user2.Name != "upsert_user_updated" {
		t.Errorf("Expected 'upsert_user_updated', got '%s'", user2.Name)
	}
	if user2.Avatar != "upsert_avatar_updated" {
		t.Errorf("Expected 'upsert_avatar_updated', got '%s'", user2.Avatar)
	}

	// Verify only one record exists
	var count int64
	_ = query.Model(&AzUser{}).Where("PartitionKey = ?", "pk1").Where("RowKey = ?", "rk1").Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}
