package sqlserver

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLServerIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("SQL Server cannot combine TOP and OFFSET in same query - skipping paginate tests")

	db := SetupSQLServerTest(t)
	common.TestPaginateFirstPage(t, db)
}

func TestSQLServerIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("SQL Server cannot combine TOP and OFFSET in same query - skipping paginate tests")

	db := SetupSQLServerTest(t)
	common.TestPaginateSecondPage(t, db)
}

func TestSQLServerIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("SQL Server cannot combine TOP and OFFSET in same query - skipping paginate tests")

	db := SetupSQLServerTest(t)
	common.TestPaginateWithConditions(t, db)
}

func TestSQLServerIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("SQL Server cannot combine TOP and OFFSET in same query - skipping paginate tests")

	db := SetupSQLServerTest(t)
	common.TestPaginateWithSelectAliases(t, db)
}
