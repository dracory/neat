//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func seedValueTestData(t *testing.T, db interface{}) {
	var query interface {
		Model(interface{}) interface{ Create(interface{}) error }
	}
	switch v := db.(type) {
	case interface {
		Query() interface {
			Model(interface{}) interface{ Create(interface{}) error }
		}
	}:
		query = v.Query()
	default:
		query = db
	}

	users := []models.User{
		{Name: "value_user_1", Avatar: "avatar_1"},
		{Name: "value_user_2", Avatar: "avatar_2"},
	}
	for _, user := range users {
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}
}

func TestMysqlIntegrationQueryValueBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedValueTestData(t, db)

	var name string
	err := db.Query().Model(&models.User{}).OrderBy("id", "asc").Value("name", &name)
	if err != nil {
		t.Errorf("Value basic failed: %v", err)
	}
	if name != "value_user_1" {
		t.Errorf("Expected 'value_user_1', got '%s'", name)
	}
}

func TestMysqlIntegrationQueryValueWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedValueTestData(t, db)

	var name string
	err := db.Query().Model(&models.User{}).Where("avatar = ?", "avatar_2").Value("name", &name)
	if err != nil {
		t.Errorf("Value with where failed: %v", err)
	}
	if name != "value_user_2" {
		t.Errorf("Expected 'value_user_2', got '%s'", name)
	}
}

func TestMysqlIntegrationQueryToSqlValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	seedValueTestData(t, db)

	var name string
	sql := db.Query().Model(&models.User{}).Where("id = ?", 1).ToSql().Value("name", &name)
	if sql == "" {
		t.Error("SQL should not be empty")
	}
}
