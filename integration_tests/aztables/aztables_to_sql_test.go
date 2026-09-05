//go:build integration

package aztables

import (
	"strings"
	"testing"
)

// TestAztablesIntegrationQueryToSql verifies ToSql generates a SELECT with
// unquoted identifiers for the aztables dialect.
func TestAztablesIntegrationQueryToSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupAztablesConnection(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("azusers").Where("PartitionKey = ?", "pk1").ToSql().Get(&AzUser{}))
	if !strings.Contains(sql, "SELECT") || !strings.Contains(sql, "AZUSERS") {
		t.Errorf("SQL should contain SELECT ... AZUSERS, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "PARTITIONKEY") {
		t.Errorf("SQL should contain WHERE ... PARTITIONKEY, got: %s", sql)
	}
}

// TestAztablesIntegrationQueryToRawSql verifies ToRawSql interpolates values.
func TestAztablesIntegrationQueryToRawSql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupAztablesConnection(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("azusers").Where("PartitionKey = ?", "pk1").ToRawSql().Get(&AzUser{}))
	if !strings.Contains(sql, "SELECT") || !strings.Contains(sql, "AZUSERS") {
		t.Errorf("SQL should contain SELECT ... AZUSERS, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "PARTITIONKEY") {
		t.Errorf("SQL should contain WHERE ... PARTITIONKEY, got: %s", sql)
	}
}

// TestAztablesIntegrationQueryToSqlCount verifies ToSql for Count generates
// SELECT COUNT(*).
func TestAztablesIntegrationQueryToSqlCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupAztablesConnection(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("azusers").Where("PartitionKey = ?", "pk1").ToSql().Count())
	if !strings.Contains(sql, "COUNT") {
		t.Errorf("SQL should contain COUNT, got: %s", sql)
	}
	if !strings.Contains(sql, "AZUSERS") {
		t.Errorf("SQL should contain AZUSERS, got: %s", sql)
	}
}

// TestAztablesIntegrationQueryToSqlUpdate verifies ToSql for Update generates
// UPDATE ... SET ... WHERE.
func TestAztablesIntegrationQueryToSqlUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupAztablesConnection(t)
	query := db.Query()

	sql := strings.ToUpper(query.Table("azusers").Where("PartitionKey = ?", "pk1").ToSql().Update("Name", "new_name"))
	if !strings.Contains(sql, "UPDATE") || !strings.Contains(sql, "AZUSERS") {
		t.Errorf("SQL should contain UPDATE AZUSERS, got: %s", sql)
	}
	if !strings.Contains(sql, "SET") || !strings.Contains(sql, "NAME") {
		t.Errorf("SQL should contain SET ... NAME, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "PARTITIONKEY") {
		t.Errorf("SQL should contain WHERE ... PARTITIONKEY, got: %s", sql)
	}
}
