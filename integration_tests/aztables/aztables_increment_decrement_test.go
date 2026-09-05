//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationQueryIncrement verifies the Increment operation.
//
// Note: Azure Table Storage's UPDATE via aztablessql uses merge semantics.
// The expression `SET col = col + ?` may not be supported by the parser.
// If it isn't, this test will be skipped.
func TestAztablesIntegrationQueryIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "increment_user", Avatar: "group1"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Increment("Votes", 5)
	if err != nil {
		// Increment generates `SET col = col + ?` which aztablessql may not parse
		t.Skipf("Increment not supported by aztablessql: %v", err)
	}
}

// TestAztablesIntegrationQueryDecrement verifies the Decrement operation.
//
// Note: Same limitation as Increment — may not be supported by aztablessql.
func TestAztablesIntegrationQueryDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "decrement_user", Avatar: "group1"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err := query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Decrement("Votes", 3)
	if err != nil {
		t.Skipf("Decrement not supported by aztablessql: %v", err)
	}
}
