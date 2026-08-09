package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestCockroachDBIntegrationQueryCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{Name: "create_struct_user", Avatar: "avatar"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if user.ID == 0 {
		t.Error("User ID should be set after create")
	}

	var foundUser models.User
	err = query.Model(&models.User{}).Where("id = ?", user.ID).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}
	if foundUser.Name != "create_struct_user" {
		t.Errorf("Expected 'create_struct_user', got '%s'", foundUser.Name)
	}
}

func TestCockroachDBIntegrationQueryBatchCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "batch_user1", Avatar: "avatar1"},
		{Name: "batch_user2", Avatar: "avatar2"},
		{Name: "batch_user3", Avatar: "avatar3"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to batch create users: %v", err)
	}

	for _, u := range users {
		if u.ID == 0 {
			t.Error("User ID should be set after batch create")
		}
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).Where("name LIKE ?", "batch_user%").Find(&foundUsers)
	if err != nil {
		t.Fatalf("Failed to find users: %v", err)
	}
	if len(foundUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(foundUsers))
	}
}

func TestCockroachDBIntegrationQueryCreateByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	err := query.Table("users").Create(map[string]any{
		"name":   "map_user",
		"avatar": "map_avatar",
	})
	if err != nil {
		t.Fatalf("Failed to create user by map: %v", err)
	}

	var foundUser models.User
	err = query.Model(&models.User{}).Where("name = ?", "map_user").First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}
	if foundUser.Name != "map_user" {
		t.Errorf("Expected 'map_user', got '%s'", foundUser.Name)
	}
	if foundUser.Avatar != "map_avatar" {
		t.Errorf("Expected 'map_avatar', got '%s'", foundUser.Avatar)
	}
}

func TestCockroachDBIntegrationQueryInsertGetIdStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{Name: "insert_get_id_user", Avatar: "avatar"}
	id, err := query.Model(&models.User{}).InsertGetId(&user)
	if err != nil {
		t.Fatalf("Failed to insert and get ID: %v", err)
	}
	if id == 0 {
		t.Error("ID should be non-zero")
	}

	var foundUser models.User
	err = query.Model(&models.User{}).Where("id = ?", id).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}
	if foundUser.Name != "insert_get_id_user" {
		t.Errorf("Expected 'insert_get_id_user', got '%s'", foundUser.Name)
	}
}

func TestCockroachDBIntegrationQueryInsertGetIdMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	id, err := query.Table("users").InsertGetId(map[string]any{
		"name":   "insert_get_id_map",
		"avatar": "map_avatar",
	})
	if err != nil {
		t.Fatalf("Failed to insert and get ID by map: %v", err)
	}
	if id == 0 {
		t.Error("ID should be non-zero")
	}

	var foundUser models.User
	err = query.Model(&models.User{}).Where("id = ?", id).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}
	if foundUser.Name != "insert_get_id_map" {
		t.Errorf("Expected 'insert_get_id_map', got '%s'", foundUser.Name)
	}
}

func TestCockroachDBIntegrationQueryInsertGetIdBigSerial(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.BigSerialUser{Name: "bigserial_user"}
	id, err := query.Model(&models.BigSerialUser{}).InsertGetId(&user)
	if err != nil {
		t.Fatalf("Failed to insert and get bigserial ID: %v", err)
	}
	if id == 0 {
		t.Error("ID should be non-zero")
	}

	var foundUser models.BigSerialUser
	err = query.Model(&models.BigSerialUser{}).Where("id = ?", id).First(&foundUser)
	if err != nil {
		t.Fatalf("Failed to find bigserial user: %v", err)
	}
	if foundUser.Name != "bigserial_user" {
		t.Errorf("Expected 'bigserial_user', got '%s'", foundUser.Name)
	}
}
