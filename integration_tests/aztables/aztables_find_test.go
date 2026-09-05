//go:build integration

package aztables

import (
	"testing"

	"github.com/dracory/neat/database"
)

// AzUser is a model mapped to Azure Table Storage entities.
// Table Storage uses PartitionKey + RowKey as the composite primary key
// instead of an auto-increment id, so this model does not have an ID field.
type AzUser struct {
	PartitionKey string `db:"PartitionKey"`
	RowKey       string `db:"RowKey"`
	Name         string `db:"Name"`
	Avatar       string `db:"Avatar"`
}

// TableName returns the Azure Table Storage table name for AzUser.
func (AzUser) TableName() string {
	return "azusers"
}

// setupFindTest creates a database connection and a fresh table for AzUser.
func setupFindTest(t *testing.T) *database.Database {
	db := SetupAztablesConnection(t)
	createTable(t, db, AzUser{}.TableName())
	t.Cleanup(func() { dropTable(t, db, AzUser{}.TableName()) })
	return db
}

// TestAztablesIntegrationFirst tests the ORM First operation via neat's
// query builder against Azure Table Storage.
func TestAztablesIntegrationFirst(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Create users
	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "first_user1", Avatar: "avatar1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "first_user2", Avatar: "avatar2"},
	}
	err := query.Model(&AzUser{}).Create(&users)
	if err != nil {
		queries := databaseConn.GetQueryLog()
		t.Logf("Query log: %v", queries)
		t.Fatalf("Failed to create users: %v", err)
	}

	// Test First — should get one record
	var user AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		First(&user)
	if err != nil {
		queries := databaseConn.GetQueryLog()
		t.Logf("Query log on SELECT error: %v", queries)
		t.Fatalf("Failed to get first user: %v", err)
	}

	if user.Name != "first_user1" && user.Name != "first_user2" {
		t.Errorf("Expected first_user1 or first_user2, got '%s'", user.Name)
	}
}

// TestAztablesIntegrationFind tests the ORM Find operation (multiple records).
func TestAztablesIntegrationFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Create users
	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk1", Name: "find_user1", Avatar: "avatar1"},
		{PartitionKey: "pk1", RowKey: "rk2", Name: "find_user2", Avatar: "avatar2"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "find_user3", Avatar: "avatar3"},
	}
	err := query.Model(&AzUser{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Test Find — should get all matching records in the partition
	var foundUsers []AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users: %v", err)
	}

	if len(foundUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(foundUsers))
	}
}

// TestAztablesIntegrationCreate tests the ORM Create operation (single and
// batch) via neat's query builder.
func TestAztablesIntegrationCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Test Create single record
	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "create_user", Avatar: "avatar"}
	err := query.Model(&AzUser{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify the record was created by querying it back
	var createdUser AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&createdUser)
	if err != nil {
		t.Fatalf("Failed to query created user: %v", err)
	}

	if createdUser.Name != "create_user" {
		t.Errorf("Expected 'create_user', got '%s'", createdUser.Name)
	}

	// Test Create multiple records
	users := []AzUser{
		{PartitionKey: "pk1", RowKey: "rk2", Name: "create_user1", Avatar: "avatar1"},
		{PartitionKey: "pk1", RowKey: "rk3", Name: "create_user2", Avatar: "avatar2"},
	}
	err = query.Model(&AzUser{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Verify the records were created
	var foundUsers []AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to query created users: %v", err)
	}

	if len(foundUsers) < 3 {
		t.Errorf("Expected at least 3 users, got %d", len(foundUsers))
	}
}

// TestAztablesIntegrationUpdate tests the ORM Update operation.
func TestAztablesIntegrationUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Create a user
	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "update_user", Avatar: "old_avatar"}
	err := query.Model(&AzUser{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test Update — change the Avatar field
	_, err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Update("Avatar", "new_avatar")
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	var updatedUser AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Avatar != "new_avatar" {
		t.Errorf("Expected 'new_avatar', got '%s'", updatedUser.Avatar)
	}
}

// TestAztablesIntegrationDelete tests the ORM Delete operation.
func TestAztablesIntegrationDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseConn := setupFindTest(t)
	query := databaseConn.Query()

	// Create a user
	user := AzUser{PartitionKey: "pk1", RowKey: "rk1", Name: "delete_user", Avatar: "avatar"}
	err := query.Model(&AzUser{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test Delete
	_, err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		Delete(&AzUser{})
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify deletion — First should fail (return no rows)
	var deletedUser AzUser
	err = query.Model(&AzUser{}).
		Where("PartitionKey = ?", "pk1").
		Where("RowKey = ?", "rk1").
		First(&deletedUser)
	if err == nil {
		t.Error("Expected error for deleted user, got nil")
	}
}
