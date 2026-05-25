//go:build disabled

package postgres

import (
	"testing"
	"github.com/dracory/neat/integration_tests/models"
)

// TestPostgreSQLIntegrationWhereIn tests WhereIn operation
func TestPostgreSQLIntegrationWhereIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "where_in_user1", Avatar: "avatar1"},
		{Name: "where_in_user2", Avatar: "avatar2"},
		{Name: "where_in_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "where_in_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test WhereIn with multiple IDs
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

// TestPostgreSQLIntegrationOrWhereIn tests OrWhereIn operation
func TestPostgreSQLIntegrationOrWhereIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "or_where_in_user1", Avatar: "avatar1"},
		{Name: "or_where_in_user2", Avatar: "avatar2"},
		{Name: "or_where_in_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "or_where_in_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test OrWhereIn
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("id = ?", -1).OrWhereIn("id", []any{createdUsers[0].ID, createdUsers[1].ID}).Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with OrWhereIn: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}
}

// TestPostgreSQLIntegrationWhereNotIn tests WhereNotIn operation
func TestPostgreSQLIntegrationWhereNotIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "where_not_in_user1", Avatar: "avatar1"},
		{Name: "where_not_in_user2", Avatar: "avatar2"},
		{Name: "where_not_in_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "where_not_in_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test WhereNotIn
	var foundUser models.User
	err := query.Model(&models.User{}).Where("id = ?", createdUsers[2].ID).WhereNotIn("id", []any{createdUsers[0].ID, createdUsers[1].ID}).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find user with WhereNotIn: %v", err)
	}

	if foundUser.ID != createdUsers[2].ID {
		t.Errorf("Expected user[2], got %d", foundUser.ID)
	}
}

// TestPostgreSQLIntegrationOrWhereNotIn tests OrWhereNotIn operation
func TestPostgreSQLIntegrationOrWhereNotIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "or_where_not_in_user1", Avatar: "avatar1"},
		{Name: "or_where_not_in_user2", Avatar: "avatar2"},
		{Name: "or_where_not_in_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "or_where_not_in_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test OrWhereNotIn
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("id = ?", -1).OrWhereNotIn("id", []any{createdUsers[0].ID, createdUsers[1].ID}).Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with OrWhereNotIn: %v", err)
	}

	// Should find user[2] since it's not in the excluded list
	user2Found := false
	for _, user := range foundUsers {
		if user.ID == createdUsers[2].ID {
			user2Found = true
		}
	}

	if !user2Found {
		t.Error("Expected to find user[2] with OrWhereNotIn")
	}
}

// TestPostgreSQLIntegrationWhereBetween tests WhereBetween operation
func TestPostgreSQLIntegrationWhereBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "where_between_user1", Avatar: "avatar1"},
		{Name: "where_between_user2", Avatar: "avatar2"},
		{Name: "where_between_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "where_between_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test WhereBetween with IDs
	var foundUsers []models.User
	err := query.Model(&models.User{}).WhereBetween("id", createdUsers[0].ID, createdUsers[2].ID).Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with WhereBetween: %v", err)
	}

	if len(foundUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(foundUsers))
	}
}

// TestPostgreSQLIntegrationWhereNotBetween tests WhereNotBetween operation
func TestPostgreSQLIntegrationWhereNotBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "where_not_between_user1", Avatar: "avatar1"},
		{Name: "where_not_between_user2", Avatar: "avatar2"},
		{Name: "where_not_between_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "where_not_between_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test WhereNotBetween - exclude first two users
	var foundUser models.User
	err := query.Model(&models.User{}).Where("name = ?", "where_not_between_user3").WhereNotBetween("id", createdUsers[0].ID, createdUsers[1].ID).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find user with WhereNotBetween: %v", err)
	}

	if foundUser.ID != createdUsers[2].ID {
		t.Errorf("Expected user[2], got %d", foundUser.ID)
	}
}

// TestPostgreSQLIntegrationOrWhereBetween tests OrWhereBetween operation
func TestPostgreSQLIntegrationOrWhereBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "or_where_between_user1", Avatar: "avatar1"},
		{Name: "or_where_between_user2", Avatar: "avatar2"},
		{Name: "or_where_between_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "or_where_between_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test OrWhereBetween
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("name = ?", "or_where_between_user3").OrWhereBetween("id", createdUsers[0].ID, createdUsers[1].ID).Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with OrWhereBetween: %v", err)
	}

	if len(foundUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(foundUsers))
	}
}

// TestPostgreSQLIntegrationOrWhereNotBetween tests OrWhereNotBetween operation
func TestPostgreSQLIntegrationOrWhereNotBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "or_where_not_between_user1", Avatar: "avatar1"},
		{Name: "or_where_not_between_user2", Avatar: "avatar2"},
		{Name: "or_where_not_between_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "or_where_not_between_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test OrWhereNotBetween
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("name = ?", "or_where_not_between_user3").OrWhereNotBetween("id", createdUsers[0].ID, createdUsers[1].ID).Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with OrWhereNotBetween: %v", err)
	}

	// Should find user[3] since it matches the first condition
	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user, got %d", len(foundUsers))
	}

	if foundUsers[0].ID != createdUsers[2].ID {
		t.Errorf("Expected user[2], got %d", foundUsers[0].ID)
	}
}

// TestPostgreSQLIntegrationWhereNull tests WhereNull operation
func TestPostgreSQLIntegrationWhereNull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users with and without bio
	bio := "test_bio"
	users := []models.User{
		{Name: "where_null_user1", Avatar: "avatar1", Bio: &bio},
		{Name: "where_null_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "where_null_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test WhereNull
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("name = ?", "where_null_user2").WhereNull("bio").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with WhereNull: %v", err)
	}

	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user, got %d", len(foundUsers))
	}

	if foundUsers[0].ID != createdUsers[1].ID {
		t.Errorf("Expected user[1], got %d", foundUsers[0].ID)
	}
}

// TestPostgreSQLIntegrationWhereNotNull tests WhereNotNull operation
func TestPostgreSQLIntegrationWhereNotNull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users with and without bio
	bio := "test_bio"
	users := []models.User{
		{Name: "where_not_null_user1", Avatar: "avatar1", Bio: &bio},
		{Name: "where_not_null_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "where_not_null_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test WhereNotNull
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("name = ?", "where_not_null_user1").WhereNotNull("bio").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with WhereNotNull: %v", err)
	}

	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user, got %d", len(foundUsers))
	}

	if foundUsers[0].ID != createdUsers[0].ID {
		t.Errorf("Expected user[0], got %d", foundUsers[0].ID)
	}
}

// TestPostgreSQLIntegrationOrWhereNull tests OrWhereNull operation
func TestPostgreSQLIntegrationOrWhereNull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users with and without bio
	bio := "test_bio"
	users := []models.User{
		{Name: "or_where_null_user1", Avatar: "avatar1", Bio: &bio},
		{Name: "or_where_null_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Test OrWhereNull
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("name = ?", "or_where_null_user1").OrWhereNull("bio").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with OrWhereNull: %v", err)
	}

	if len(foundUsers) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(foundUsers))
	}
}

// TestPostgreSQLIntegrationOrWhere tests OrWhere operation
func TestPostgreSQLIntegrationOrWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "or_where_user1", Avatar: "avatar1"},
		{Name: "or_where_user2", Avatar: "avatar2"},
		{Name: "or_where_user3", Avatar: "avatar3"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Test OrWhere
	var foundUsers []models.User
	err := query.Model(&models.User{}).Where("name = ?", "or_where_user1").OrWhere("avatar = ?", "avatar2").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users with OrWhere: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}
}

// TestPostgreSQLIntegrationWhereColumnOperator tests where with column operator variations
func TestPostgreSQLIntegrationWhereColumnOperator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create users
	users := []models.User{
		{Name: "column_op_user1", Avatar: "avatar1"},
		{Name: "column_op_user2", Avatar: "avatar2"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	// Get the created users to get their IDs
	var createdUsers []models.User
	if err := query.Model(&models.User{}).Where("name LIKE ?", "column_op_user%").Find(&createdUsers); err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	// Test where with different operators
	var foundUser models.User
	err := query.Model(&models.User{}).Where("name = ?", "column_op_user1").First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if foundUser.ID != createdUsers[0].ID {
		t.Errorf("Expected user[0], got %d", foundUser.ID)
	}
}
