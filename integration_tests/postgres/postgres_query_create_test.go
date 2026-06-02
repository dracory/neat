package postgres_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgreSQLIntegrationQueryCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestQueryCreateByStruct(t, db)
}

func TestPostgreSQLIntegrationQueryBatchCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestQueryBatchCreateByStruct(t, db)
}

func TestPostgreSQLIntegrationQueryCreateByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestQueryCreateByMap(t, db)
}

func TestPostgreSQLIntegrationQueryInsertGetIdByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestQueryInsertGetIdByStruct(t, db)
}

func TestPostgreSQLIntegrationQueryInsertGetIdByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestQueryInsertGetIdByMap(t, db)
}

func TestPostgreSQLIntegrationQueryInsertGetIdBigSerial(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.BigSerialUser{Name: "bigserial_user"}
	id, err := query.Model(&models.BigSerialUser{}).InsertGetId(&user)
	if err != nil {
		t.Errorf("InsertGetId with bigserial failed: %v", err)
	}
	if id == 0 {
		t.Error("ID should not be zero for bigserial")
	}
	if user.ID != int64(id) {
		t.Errorf("Expected ID %d, got %d", id, user.ID)
	}

	// Verify the record was inserted correctly
	var foundUser models.BigSerialUser
	err = query.Model(&models.BigSerialUser{}).Where("id = ?", id).First(&foundUser)
	if err != nil {
		t.Errorf("Failed to find bigserial user with ID %d: %v", id, err)
	}
	if foundUser.Name != "bigserial_user" {
		t.Errorf("Expected name 'bigserial_user', got '%s'", foundUser.Name)
	}
}

func TestPostgreSQLIntegrationQueryInsertGetIdBigSerialByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	id, err := query.Table("bigserial_users").InsertGetId(map[string]any{
		"name": "bigserial_map_user",
	})
	if err != nil {
		t.Errorf("InsertGetId by map with bigserial failed: %v", err)
	}
	if id == 0 {
		t.Error("ID should not be zero for bigserial")
	}

	var user models.BigSerialUser
	err = query.Model(&models.BigSerialUser{}).Where("id = ?", id).First(&user)
	if err != nil {
		t.Errorf("Failed to find bigserial user with ID %d: %v", id, err)
	}
	if user.Name != "bigserial_map_user" {
		t.Errorf("Expected name 'bigserial_map_user', got '%s'", user.Name)
	}
}
