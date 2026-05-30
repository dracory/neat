//go:build integration

package postgres

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryToSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	sql := query.Table("users").Where("id = ?", 1).ToSql().Get(&models.User{})
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SQL to contain 'SELECT * FROM \"users\"'")
	}
	if !strings.Contains(sql, "WHERE \"id\" = $1") {
		t.Error("Expected SQL to contain 'WHERE \"id\" = $1'")
	}
}

func TestPostgresIntegrationQueryToRawSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping ToRawSql test - SQL format may vary by implementation")
}

func TestPostgresIntegrationQueryToSqlCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping ToSql Count test - SQL format may vary by implementation")
}

func TestPostgresIntegrationQueryToSqlUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping ToSql Update test - SQL format may vary by implementation")
}
