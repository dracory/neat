package sqlite

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryToSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	sql := query.Table("users").Where("id = ?", 1).ToSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("ToSql output: %s", sql)
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"id\" = ?") {
		t.Error("Expected WHERE \"id\" = ?")
	}
}

func TestSQLiteIntegrationQueryToRawSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	sql := query.Table("users").Where("id = ?", 1).ToRawSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"id\" = 1") {
		t.Error("Expected WHERE \"id\" = 1")
	}
}

func TestSQLiteIntegrationQueryToSqlCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	sql := query.Table("users").Where("name = ?", "test").ToSql().Count()
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("ToSql Count output: %s", sql)
	if !strings.Contains(sql, "SELECT COUNT(*)") {
		t.Error("Expected SELECT COUNT(*)")
	}
	if !strings.Contains(sql, "FROM \"users\"") {
		t.Error("Expected FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"name\" = ?") {
		t.Error("Expected WHERE \"name\" = ?")
	}
}

func TestSQLiteIntegrationQueryToRawSqlCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	sql := query.Table("users").Where("name = ?", "test").ToRawSql().Count()
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("Generated SQL: %s", sql)
	if !strings.Contains(sql, "SELECT COUNT(*)") {
		t.Error("Expected SELECT COUNT(*)")
	}
	if !strings.Contains(sql, "FROM \"users\"") {
		t.Error("Expected FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE name = 'test'") {
		t.Error("Expected WHERE name = 'test'")
	}
}

func TestSQLiteIntegrationQueryToSqlUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	sql := query.Table("users").Where("id = ?", 1).ToSql().Update("name", "new_name")
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("ToSql Update output: %s", sql)
	if !strings.Contains(sql, "UPDATE \"users\"") {
		t.Error("Expected UPDATE \"users\"")
	}
	if !strings.Contains(sql, "SET \"name\" = ?") {
		t.Error("Expected SET \"name\" = ?")
	}
	if !strings.Contains(sql, "WHERE \"id\" = ?") {
		t.Error("Expected WHERE \"id\" = ?")
	}
}

func TestSQLiteIntegrationQueryToSqlDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	sql := query.Table("users").Where("id = ?", 1).ToSql().Delete()
	sql = strings.ReplaceAll(sql, "`", "\"")
	if strings.Contains(sql, "UPDATE") {
		if !strings.Contains(sql, "UPDATE \"users\" SET \"deleted_at\"") {
			t.Error("Expected UPDATE \"users\" SET \"deleted_at\"")
		}
	} else {
		if !strings.Contains(sql, "DELETE FROM \"users\"") {
			t.Error("Expected DELETE FROM \"users\"")
		}
	}
	if !strings.Contains(sql, "WHERE \"id\" = ?") {
		t.Error("Expected WHERE \"id\" = ?")
	}
}
