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
	if !strings.Contains(sql, "WHERE \"name\" = 'test'") {
		t.Error("Expected WHERE \"name\" = 'test'")
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

func TestSQLiteIntegrationQueryToRawSqlTableQualifiedColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Table-qualified columns should not be quoted by quoteWhereIdentifiers
	// The quoteIdentifier function handles them separately
	sql := query.Table("users").Where("users.id = ?", 1).ToRawSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("Generated SQL: %s", sql)
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	// The table-qualified column should be handled by quoteIdentifier, not quoteWhereIdentifiers
	if !strings.Contains(sql, "WHERE \"users\".id = 1") && !strings.Contains(sql, "WHERE users.id = 1") {
		t.Error("Expected WHERE with table-qualified column")
	}
}

func TestSQLiteIntegrationQueryToRawSqlFunctionCall(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Function calls should not be quoted
	sql := query.Table("users").Where("COUNT(id) > ?", 0).ToRawSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("Generated SQL: %s", sql)
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	// COUNT(id) should not be quoted
	if !strings.Contains(sql, "WHERE COUNT(id) > 0") {
		t.Error("Expected WHERE COUNT(id) > 0")
	}
}

func TestSQLiteIntegrationQueryToRawSqlBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// BETWEEN should be processed by quoteWhereIdentifiers
	// WhereBetween adds the column name without quoting, expecting the builder to handle it
	sql := query.Table("users").WhereBetween("id", 1, 10).ToRawSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("Generated SQL: %s", sql)
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"id\" BETWEEN 1 AND 10") {
		t.Error("Expected WHERE \"id\" BETWEEN 1 AND 10")
	}
}

func TestSQLiteIntegrationQueryToRawSqlIn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// IN should be handled by WhereIn which quotes the column before adding to query
	sql := query.Table("users").WhereIn("id", []any{1, 2, 3}).ToRawSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("Generated SQL: %s", sql)
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"id\" IN (1, 2, 3)") {
		t.Error("Expected WHERE \"id\" IN (1, 2, 3)")
	}
}

func TestSQLiteIntegrationQueryToRawSqlNull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// NULL should not be quoted
	sql := query.Table("users").Where("deleted_at IS NULL").ToRawSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("Generated SQL: %s", sql)
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"deleted_at\" IS NULL") {
		t.Error("Expected WHERE \"deleted_at\" IS NULL")
	}
}

func TestSQLiteIntegrationQueryToRawSqlMultipleConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// Multiple conditions with AND
	sql := query.Table("users").Where("name = ?", "test").Where("age > ?", 18).ToRawSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("Generated SQL: %s", sql)
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"name\" = 'test'") {
		t.Error("Expected WHERE \"name\" = 'test'")
	}
	if !strings.Contains(sql, "AND \"age\" > 18") {
		t.Error("Expected AND \"age\" > 18")
	}
}

func TestSQLiteIntegrationQueryToRawSqlOrCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	// OR condition
	sql := query.Table("users").Where("name = ?", "test").OrWhere("email = ?", "test@example.com").ToRawSql().Get(&models.User{})
	sql = strings.ReplaceAll(sql, "`", "\"")
	t.Logf("Generated SQL: %s", sql)
	if !strings.Contains(sql, "SELECT * FROM \"users\"") {
		t.Error("Expected SELECT * FROM \"users\"")
	}
	if !strings.Contains(sql, "WHERE \"name\" = 'test'") {
		t.Error("Expected WHERE \"name\" = 'test'")
	}
	if !strings.Contains(sql, "OR \"email\" = 'test@example.com'") {
		t.Error("Expected OR \"email\" = 'test@example.com'")
	}
}
