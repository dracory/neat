//go:build integration

package aztables

import (
	"fmt"
	"testing"
)

// TestAztablesIntegrationRawExec verifies raw SQL Exec works for INSERT and
// DELETE through the neat query builder.
func TestAztablesIntegrationRawExec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	// Insert via raw Exec
	_, err := query.Exec(
		fmt.Sprintf("INSERT INTO %s (PartitionKey, RowKey, Name) VALUES (?, ?, ?)", AzUser{}.TableName()),
		"pk1", "rk1", "raw_user",
	)
	if err != nil {
		t.Fatalf("Raw INSERT failed: %v", err)
	}

	// Verify
	var user AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&user)
	if err != nil {
		t.Fatalf("Failed to find raw-inserted user: %v", err)
	}
	if user.Name != "raw_user" {
		t.Errorf("Expected 'raw_user', got '%s'", user.Name)
	}

	// Delete via raw Exec
	_, err = query.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE PartitionKey = ? AND RowKey = ?", AzUser{}.TableName()),
		"pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("Raw DELETE failed: %v", err)
	}

	// Verify deletion
	var deleted AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&deleted)
	if err == nil {
		t.Error("Expected error for deleted user, got nil")
	}
}

// TestAztablesIntegrationRawSelect verifies raw SQL Select via sqlDB.Query.
func TestAztablesIntegrationRawSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupFindTest(t)
	query := db.Query()

	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "raw_select_user", Avatar: "avatar"}
	if err := query.Model(&AzUser{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB(): %v", err)
	}

	rows, err := sqlDB.Query(
		fmt.Sprintf("SELECT Name FROM %s WHERE PartitionKey = ? AND RowKey = ?", AzUser{}.TableName()),
		"pk1", "rk1",
	)
	if err != nil {
		t.Fatalf("Raw SELECT failed: %v", err)
	}
	row := scanOneRow(t, rows)
	if row["Name"] != "raw_select_user" {
		t.Errorf("Expected 'raw_select_user', got '%v'", row["Name"])
	}
}
