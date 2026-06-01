package turso

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/common"
	"github.com/dracory/neat/integration_tests/models"
)

func TestTursoIntegrationQueryValueBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedValueTestData(t, db)

	var name string
	err := db.Query().Model(&models.User{}).OrderBy("id", "asc").Value("name", &name)
	if err != nil {
		t.Errorf("Value basic failed: %v", err)
	}
	if name != "value_user_1" {
		t.Errorf("Expected 'value_user_1', got '%s'", name)
	}
}

func TestTursoIntegrationQueryValueWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedValueTestData(t, db)

	var name string
	err := db.Query().Model(&models.User{}).Where("avatar = ?", "avatar_2").Value("name", &name)
	if err != nil {
		t.Errorf("Value with where failed: %v", err)
	}
	if name != "value_user_2" {
		t.Errorf("Expected 'value_user_2', got '%s'", name)
	}
}

func TestTursoIntegrationQueryValueWithOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedValueTestData(t, db)

	var name string
	err := db.Query().Model(&models.User{}).OrderBy("id", "desc").Value("name", &name)
	if err != nil {
		t.Errorf("Value with order failed: %v", err)
	}
	if name != "value_user_2" {
		t.Errorf("Expected 'value_user_2', got '%s'", name)
	}
}

func TestTursoIntegrationQueryValueNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedValueTestData(t, db)

	var name string
	err := db.Query().Model(&models.User{}).Where("name = ?", "non_existent").Value("name", &name)
	if err != nil {
		t.Errorf("Value not found failed: %v", err)
	}
	if name != "" {
		t.Error("Expected empty name")
	}
}

func TestTursoIntegrationQueryToSqlValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)

	var name string
	sql := db.Query().Model(&models.User{}).WithTrashed().Where("id = ?", 1).ToSql().Value("name", &name)
	sql = strings.ReplaceAll(sql, "`", "\"")
	if !strings.Contains(sql, "SELECT name FROM \"users\"") {
		t.Error("Expected SELECT name FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"id\" = ?") {
		t.Error("Expected WHERE \"id\" = ?")
	}
	if !strings.Contains(sql, "LIMIT 1") {
		t.Error("Expected LIMIT 1")
	}
}
