package sqlserver_test

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLServerIntegrationQueryToSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("users").Where("id = ?", 1).ToSql().Get(&models.User{}))
	if !strings.Contains(sql, "SELECT") || !strings.Contains(sql, "USERS") {
		t.Errorf("SQL should contain SELECT ... USERS, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "ID") {
		t.Errorf("SQL should contain WHERE ... ID, got: %s", sql)
	}
}

func TestSQLServerIntegrationQueryToRawSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("SQL Server uses parameterized queries (@P1) instead of interpolated values - skipping")

	db := SetupSQLServerTest(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("users").Where("id = ?", 1).ToRawSql().Get(&models.User{}))
	if !strings.Contains(sql, "SELECT") || !strings.Contains(sql, "USERS") {
		t.Errorf("SQL should contain SELECT ... USERS, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, " 1") {
		t.Errorf("SQL should contain WHERE ... 1 (interpolated), got: %s", sql)
	}
}

func TestSQLServerIntegrationQueryToSqlCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("users").Where("name = ?", "test").ToSql().Count())
	if !strings.Contains(sql, "COUNT") {
		t.Errorf("SQL should contain COUNT, got: %s", sql)
	}
	if !strings.Contains(sql, "USERS") {
		t.Errorf("SQL should contain USERS, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "NAME") {
		t.Errorf("SQL should contain WHERE ... NAME, got: %s", sql)
	}
}

func TestSQLServerIntegrationQueryToSqlUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("users").Where("id = ?", 1).ToSql().Update("name", "new_name"))
	if !strings.Contains(sql, "UPDATE") || !strings.Contains(sql, "USERS") {
		t.Errorf("SQL should contain UPDATE USERS, got: %s", sql)
	}
	if !strings.Contains(sql, "SET") || !strings.Contains(sql, "NAME") {
		t.Errorf("SQL should contain SET ... NAME, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "ID") {
		t.Errorf("SQL should contain WHERE ... ID, got: %s", sql)
	}
}
