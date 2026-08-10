//go:build integration

package cockroachdb_test

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestCockroachDBIntegrationQueryToSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("users").Where("id = ?", 1).ToSql().Get(&models.User{}))
	if !strings.Contains(sql, "SELECT") || !strings.Contains(sql, "USERS") {
		t.Errorf("SQL should contain SELECT ... USERS, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "ID") {
		t.Errorf("SQL should contain WHERE ... ID, got: %s", sql)
	}
}

func TestCockroachDBIntegrationQueryToRawSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("users").Where("id = ?", 1).ToRawSql().Get(&models.User{}))
	if !strings.Contains(sql, "SELECT") || !strings.Contains(sql, "USERS") {
		t.Errorf("SQL should contain SELECT ... USERS, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "ID") {
		t.Errorf("SQL should contain WHERE ... ID, got: %s", sql)
	}
}

func TestCockroachDBIntegrationQueryToSqlCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
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

func TestCockroachDBIntegrationQueryToSqlUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
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
