//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgreSQLIntegrationQueryCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "create_user"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	var createdUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "create_user").First(&createdUser)
	if err != nil {
		t.Errorf("Failed to query created user: %v", err)
	}
	if createdUser.ID == 0 {
		t.Error("ID should be set after create")
	}
}

func TestPostgreSQLIntegrationQueryBatchCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "batch_create_user_1"},
		{Name: "batch_create_user_2"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Errorf("Batch create failed: %v", err)
	}
	var foundUsers []models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "batch_create_user%").Find(&foundUsers)
	if err != nil {
		t.Errorf("Failed to query created users: %v", err)
	}
	if len(foundUsers) < 2 {
		t.Error("Should have created at least 2 users")
	}
}

func TestPostgreSQLIntegrationQueryCreateByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	userMap := map[string]any{
		"name": "create_user_map",
	}
	err := query.Table("users").Create(userMap)
	if err != nil {
		t.Errorf("Create by map failed: %v", err)
	}
}

func TestPostgreSQLIntegrationQueryInsertGetIdByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{Name: "insert_get_id_user"}
	id, err := query.Model(&models.User{}).InsertGetId(&user)
	if err != nil {
		t.Errorf("InsertGetId failed: %v", err)
	}
	if id == 0 {
		t.Error("ID should not be zero")
	}
	if user.ID != id {
		t.Errorf("Expected ID %d, got %d", id, user.ID)
	}
}

func TestPostgreSQLIntegrationQueryInsertGetIdByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	id, err := query.Table("users").InsertGetId(map[string]any{
		"name": "insert_get_id_by_map_name",
	})
	if err != nil {
		t.Errorf("InsertGetId by map failed: %v", err)
	}
	if id == 0 {
		t.Error("ID should not be zero")
	}

	var user models.User
	err = query.Model(&models.User{}).Where("id = ?", id).First(&user)
	if err != nil {
		t.Errorf("Failed to find user with ID %d: %v", id, err)
	}
	if user.Name != "insert_get_id_by_map_name" {
		t.Errorf("Expected name 'insert_get_id_by_map_name', got '%s'", user.Name)
	}
}
