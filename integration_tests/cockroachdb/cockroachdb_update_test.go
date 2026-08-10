//go:build integration

package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestCockroachDBIntegrationUpdateByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "update_by_model_name", Avatar: "original_avatar"},
		{Name: "update_by_model_name", Avatar: "original_avatar1"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "update_by_model_name").Find(&createdUsers)
	if err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 2 {
		t.Fatalf("Expected at least 2 created users, got %d", len(createdUsers))
	}

	res, err := query.Model(&models.User{}).Where("name = ?", "update_by_model_name").Update("avatar", "updated_avatar")
	if err != nil {
		t.Fatalf("Failed to update users: %v", err)
	}

	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUsers[0].ID).First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	if updatedUser.Avatar != "updated_avatar" {
		t.Errorf("Expected avatar to be 'updated_avatar', got '%s'", updatedUser.Avatar)
	}
}

func TestCockroachDBIntegrationUpdateByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "update_by_table_name", Avatar: "original_avatar"},
		{Name: "update_by_table_name", Avatar: "original_avatar1"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "update_by_table_name").Find(&createdUsers)
	if err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 2 {
		t.Fatalf("Expected at least 2 created users, got %d", len(createdUsers))
	}

	res, err := query.Table("users").Where("name = ?", "update_by_table_name").Update("avatar", "updated_avatar")
	if err != nil {
		t.Fatalf("Failed to update users: %v", err)
	}

	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUsers[0].ID).First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	if updatedUser.Avatar != "updated_avatar" {
		t.Errorf("Expected avatar to be 'updated_avatar', got '%s'", updatedUser.Avatar)
	}
}

func TestCockroachDBIntegrationUpdateWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "update_where_name", Avatar: "original_avatar"},
		{Name: "update_where_name", Avatar: "original_avatar1"},
		{Name: "other_name", Avatar: "other_avatar"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "update_where_name").Find(&createdUsers)
	if err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 1 {
		t.Fatalf("Expected at least 1 created user, got %d", len(createdUsers))
	}

	var otherUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "other_name").First(&otherUser)
	if err != nil {
		t.Fatalf("Failed to get other user: %v", err)
	}

	res, err := query.Model(&models.User{}).Where("name = ?", "update_where_name").Update("avatar", "updated_avatar")
	if err != nil {
		t.Fatalf("Failed to update users: %v", err)
	}

	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUsers[0].ID).First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	if updatedUser.Avatar != "updated_avatar" {
		t.Errorf("Expected avatar to be 'updated_avatar', got '%s'", updatedUser.Avatar)
	}

	err = query.Model(&models.User{}).Where("id = ?", otherUser.ID).First(&otherUser)
	if err != nil {
		t.Fatalf("Failed to find other user: %v", err)
	}

	if otherUser.Avatar != "other_avatar" {
		t.Errorf("Expected avatar to be 'other_avatar', got '%s'", otherUser.Avatar)
	}
}

func TestCockroachDBIntegrationUpdateMultipleColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "multi_col_name", Avatar: "original_avatar"},
		{Name: "multi_col_name", Avatar: "original_avatar1"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var createdUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "multi_col_name").Find(&createdUsers)
	if err != nil {
		t.Fatalf("Failed to get created users: %v", err)
	}

	if len(createdUsers) < 1 {
		t.Fatalf("Expected at least 1 created user, got %d", len(createdUsers))
	}

	res, err := query.Model(&models.User{}).Where("name = ?", "multi_col_name").Update(map[string]any{
		"avatar": "multi_updated_avatar",
		"name":   "multi_updated_name",
	})
	if err != nil {
		t.Fatalf("Failed to update users: %v", err)
	}

	if res.RowsAffected != 2 {
		t.Errorf("Expected 2 rows affected, got %d", res.RowsAffected)
	}

	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUsers[0].ID).First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	if updatedUser.Avatar != "multi_updated_avatar" {
		t.Errorf("Expected avatar to be 'multi_updated_avatar', got '%s'", updatedUser.Avatar)
	}

	if updatedUser.Name != "multi_updated_name" {
		t.Errorf("Expected name to be 'multi_updated_name', got '%s'", updatedUser.Name)
	}
}

func TestCockroachDBIntegrationUpdateOrCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	var user models.User
	err := query.Model(&models.User{}).UpdateOrCreate(&user, models.User{Name: "update_or_create_user"}, models.User{Avatar: "update_or_create_avatar"})
	if err != nil {
		t.Fatalf("Failed to update or create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should be set after create")
	}

	if user.Avatar != "update_or_create_avatar" {
		t.Errorf("Expected avatar to be 'update_or_create_avatar', got '%s'", user.Avatar)
	}

	var user2 models.User
	err = query.Model(&models.User{}).UpdateOrCreate(&user2, models.User{Name: "update_or_create_user"}, models.User{Avatar: "update_or_create_avatar_updated"})
	if err != nil {
		t.Fatalf("Failed to update or create user: %v", err)
	}

	if user2.ID == 0 {
		t.Error("User ID should be set after update")
	}

	if user2.Avatar != "update_or_create_avatar_updated" {
		t.Errorf("Expected avatar to be 'update_or_create_avatar_updated', got '%s'", user2.Avatar)
	}

	var count int64
	err = query.Model(&models.User{}).Where("name = ?", "update_or_create_user").Count(&count)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}
}

func TestCockroachDBIntegrationSaveUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{Name: "save_update_user", Avatar: "original_avatar"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "save_update_user").First(&createdUser)
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	if createdUser.ID == 0 {
		t.Fatal("User ID should be set after create")
	}

	createdUser.Name = "save_update_user_updated"
	createdUser.Avatar = "save_updated_avatar"
	err = query.Model(&models.User{}).Save(&createdUser)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	var updatedUser models.User
	err = query.Model(&models.User{}).Where("id = ?", createdUser.ID).First(&updatedUser)
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	if updatedUser.Name != "save_update_user_updated" {
		t.Errorf("Expected name to be 'save_update_user_updated', got '%s'", updatedUser.Name)
	}

	if updatedUser.Avatar != "save_updated_avatar" {
		t.Errorf("Expected avatar to be 'save_updated_avatar', got '%s'", updatedUser.Avatar)
	}
}

func TestCockroachDBIntegrationSaveQuietly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{Name: "save_quietly_user", Avatar: "save_quietly_avatar"}
	err := query.Model(&models.User{}).SaveQuietly(&user)
	if err != nil {
		t.Fatalf("Failed to save quietly: %v", err)
	}

	if user.ID == 0 {
		t.Fatal("User ID should be set after save")
	}

	var foundUser models.User
	err = query.Model(&models.User{}).Where("id = ?", user.ID).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if foundUser.Name != "save_quietly_user" {
		t.Errorf("Expected name to be 'save_quietly_user', got '%s'", foundUser.Name)
	}

	if foundUser.Avatar != "save_quietly_avatar" {
		t.Errorf("Expected avatar to be 'save_quietly_avatar', got '%s'", foundUser.Avatar)
	}
}
