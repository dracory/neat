//go:build integration

package aztables

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// TestAztablesIntegrationQueryScopesWithoutParameters verifies that query
// scopes (closures applied via Scopes) work with the aztables dialect.
func TestAztablesIntegrationQueryScopesWithoutParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "scope_user1", Avatar: "active"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "scope_user2", Avatar: "inactive"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "scope_user3", Avatar: "active"},
	}
	if err := db.Query().Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	activeScope := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("Avatar = ?", "active")
	}

	var found []AzUser
	err := db.Query().Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Scopes(activeScope).
		Find(&found)
	if err != nil {
		t.Fatalf("Find with scope failed: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(found))
	}
}

// TestAztablesIntegrationQueryScopesWithParameters verifies parameterized
// scopes via Scopes.
func TestAztablesIntegrationQueryScopesWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "pscope1", Avatar: "alpha"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "pscope2", Avatar: "beta"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "pscope3", Avatar: "alpha"},
	}
	if err := db.Query().Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	avatarScope := func(avatar string) func(contractsorm.Query) contractsorm.Query {
		return func(q contractsorm.Query) contractsorm.Query {
			return q.Where("Avatar = ?", avatar)
		}
	}

	var found []AzUser
	err := db.Query().Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Scopes(avatarScope("alpha")).
		Find(&found)
	if err != nil {
		t.Fatalf("Find with parameterized scope failed: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("Expected 2 alpha users, got %d", len(found))
	}
}

// TestAztablesIntegrationQueryScopesMultipleChaining verifies that multiple
// scopes can be chained.
func TestAztablesIntegrationQueryScopesMultipleChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)

	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "chain1", Avatar: "alpha"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "chain2", Avatar: "beta"},
		{PartitionKey: "pk2", RowKey: "rk1", Name: "chain3", Avatar: "alpha"},
	}
	if err := db.Query().Model(&AzUser{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	partitionScope := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("PartitionKey = ?", "pk1")
	}
	avatarScope := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("Avatar = ?", "alpha")
	}

	var found []AzUser
	err := db.Query().Model(&AzUser{}).
		Scopes(partitionScope, avatarScope).
		Find(&found)
	if err != nil {
		t.Fatalf("Find with chained scopes failed: %v", err)
	}
	if len(found) != 1 {
		t.Errorf("Expected 1 user (pk1 + alpha), got %d", len(found))
	}
}
