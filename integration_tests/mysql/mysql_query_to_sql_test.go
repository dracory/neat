//go:build integration

package mysql

import (
	"strings"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLIntegrationQueryToSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	t.Run("ToSql", func(t *testing.T) {
		sql := query.Table("users").Where("id = ?", 1).ToSql().Get(&models.User{})
		if !strings.Contains(strings.ToUpper(sql), "SELECT * FROM `USERS`") {
			t.Error("SQL should contain SELECT * FROM USERS")
		}
		if !strings.Contains(strings.ToUpper(sql), "WHERE `ID` = ?") {
			t.Error("SQL should contain WHERE ID = ?")
		}
	})

	t.Run("ToRawSql", func(t *testing.T) {
		sql := query.Table("users").Where("id = ?", 1).ToRawSql().Get(&models.User{})
		if !strings.Contains(strings.ToUpper(sql), "SELECT * FROM `USERS`") {
			t.Error("SQL should contain SELECT * FROM USERS")
		}
		if !strings.Contains(strings.ToUpper(sql), "WHERE `ID` = 1") {
			t.Error("SQL should contain WHERE ID = 1")
		}
	})

	t.Run("ToSql Count", func(t *testing.T) {
		sql := query.Table("users").Where("name = ?", "test").ToSql().Count()
		if !strings.Contains(strings.ToUpper(sql), "SELECT COUNT(*)") {
			t.Error("SQL should contain SELECT COUNT(*)")
		}
		if !strings.Contains(strings.ToUpper(sql), "FROM `USERS`") {
			t.Error("SQL should contain FROM USERS")
		}
		if !strings.Contains(strings.ToUpper(sql), "WHERE `NAME` = ?") {
			t.Error("SQL should contain WHERE NAME = ?")
		}
	})

	t.Run("ToSql Update", func(t *testing.T) {
		sql := query.Table("users").Where("id = ?", 1).ToSql().Update("name", "new_name")
		if !strings.Contains(strings.ToUpper(sql), "UPDATE `USERS`") {
			t.Error("SQL should contain UPDATE USERS")
		}
		if !strings.Contains(strings.ToUpper(sql), "SET `NAME`=?") {
			t.Error("SQL should contain SET NAME=?")
		}
		if !strings.Contains(strings.ToUpper(sql), "WHERE `ID` = ?") {
			t.Error("SQL should contain WHERE ID = ?")
		}
	})
}
