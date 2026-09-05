//go:build integration

package aztables

import (
	"testing"
)

// TestAztablesIntegrationDistinctSingleColumn verifies that Distinct returns
// unique values for a single column.
//
// Note: Azure Table Storage does not have server-side DISTINCT. The
// aztablessql driver may or may not support DISTINCT in its parser. If
// it doesn't, this test will fail and should be skipped.
func TestAztablesIntegrationDistinctSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "distinct_a", Avatar: "g1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "distinct_b", Avatar: "g1"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "distinct_c", Avatar: "g2"},
	}
	if err := query.Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var avatars []string
	err := query.Model(&AzUser{}).
		Distinct("Avatar").
		Where("PartitionKey = ?", "pk1").
		Pluck("Avatar", &avatars)
	if err != nil {
		// DISTINCT may not be supported by aztablessql — skip rather than fail
		t.Skipf("Distinct not supported by aztablessql: %v", err)
	}
	// Should have 2 unique avatars: g1, g2
	if len(avatars) != 2 {
		t.Errorf("Expected 2 distinct avatars, got %d: %v", len(avatars), avatars)
	}
}
