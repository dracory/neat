package sqlserver_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLServerIntegrationWhereIn verifies that WhereIn() filters rows to only
// those whose ID appears in the provided slice, returning exactly 2 of 3 users.
func TestSQLServerIntegrationWhereIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "where_in_user1", Avatar: "avatar1"},
		{Name: "where_in_user2", Avatar: "avatar2"},
		{Name: "where_in_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "where_in_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 2 {
		t.Fatalf("Expected at least 2 created users, got %d", len(createdUsers))
	}

	var foundUsers []models.User
	ids := []any{createdUsers[0].ID, createdUsers[1].ID}
	err := query.Model(&models.User{}).WhereIn("id", ids).Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with WhereIn: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}
}

// TestSQLServerIntegrationOrWhereIn verifies that OrWhereIn() broadens a WHERE
// clause: rows matching the base condition OR whose ID is in the slice are
// returned, yielding 2 results even though the base condition matches nothing.
func TestSQLServerIntegrationOrWhereIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "or_where_in_user1", Avatar: "avatar1"},
		{Name: "or_where_in_user2", Avatar: "avatar2"},
		{Name: "or_where_in_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "or_where_in_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 2 {
		t.Fatalf("Expected at least 2 created users, got %d", len(createdUsers))
	}

	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("id = ?", -1).OrWhereIn("id", []any{createdUsers[0].ID, createdUsers[1].ID}).Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with OrWhereIn: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}
}

// TestSQLServerIntegrationWhereNotIn verifies that WhereNotIn() excludes rows
// whose ID appears in the provided slice, returning only the one user not in
// the exclusion list.
func TestSQLServerIntegrationWhereNotIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "where_not_in_user1", Avatar: "avatar1"},
		{Name: "where_not_in_user2", Avatar: "avatar2"},
		{Name: "where_not_in_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "where_not_in_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 2 {
		t.Fatalf("Expected at least 2 created users, got %d", len(createdUsers))
	}

	var foundUsers []models.User
	ids := []any{createdUsers[0].ID, createdUsers[1].ID}
	err := query.Model(&models.User{}).WhereNotIn("id", ids).Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with WhereNotIn: %v", err)
	}

	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user, got %d", len(foundUsers))
	}
}
