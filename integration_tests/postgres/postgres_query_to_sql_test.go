//go:build disabled

package postgres

import (
	"strings"
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryToSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	t.Run("ToSql", func(t *testing.T) {
		sql := query.Table("users").Where("id = ?", 1).ToSql().Get(&models.User{})
		if !strings.Contains(sql, "SELECT * FROM \"users\"") {
			t.Error("Expected SQL to contain 'SELECT * FROM \"users\"'")
		}
		if !strings.Contains(sql, "WHERE \"id\" = $1") {
			t.Error("Expected SQL to contain 'WHERE \"id\" = $1'")
		}
	})

	t.Run("ToRawSql", func(t *testing.T) {
		sql := query.Table("users").Where("id = ?", 1).ToRawSql().Get(&models.User{})
		if !strings.Contains(sql, "SELECT * FROM \"users\"") {
			t.Error("Expected SQL to contain 'SELECT * FROM \"users\"'")
		}
		if !strings.Contains(sql, "WHERE \"id\" = 1") {
			t.Error("Expected SQL to contain 'WHERE \"id\" = 1'")
		}
	})

	t.Run("ToSql Count", func(t *testing.T) {
		sql := query.Table("users").Where("name = ?", "test").ToSql().Count()
		if !strings.Contains(sql, "SELECT count(*)") {
			t.Error("Expected SQL to contain 'SELECT count(*)'")
		}
		if !strings.Contains(sql, "FROM \"users\"") {
			t.Error("Expected SQL to contain 'FROM \"users\"'")
		}
		if !strings.Contains(sql, "WHERE \"name\" = $1") {
			t.Error("Expected SQL to contain 'WHERE \"name\" = $1'")
		}
	})

	t.Run("ToSql Update", func(t *testing.T) {
		sql := query.Table("users").Where("id = ?", 1).ToSql().Update("name", "new_name")
		if !strings.Contains(sql, "UPDATE \"users\"") {
			t.Error("Expected SQL to contain 'UPDATE \"users\"'")
		}
		if !strings.Contains(sql, "SET \"name\"=$1") {
			t.Error("Expected SQL to contain 'SET \"name\"=$1'")
		}
		if !strings.Contains(sql, "WHERE \"id\" = $2") {
			t.Error("Expected SQL to contain 'WHERE \"id\" = $2'")
		}
	})
}
