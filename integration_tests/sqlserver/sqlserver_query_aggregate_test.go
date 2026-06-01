package sqlserver

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

// TestSQLServerIntegrationQueryAggregateSum verifies that Sum() returns the
// correct total of all seeded users' IDs.
func TestSQLServerIntegrationQueryAggregateSum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestAggregateSum(t, db)
}

// TestSQLServerIntegrationQueryAggregateSumWithWhere verifies that Sum() with a
// WHERE clause correctly sums only the rows matching the filter (avatar = "group1").
func TestSQLServerIntegrationQueryAggregateSumWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestAggregateSumWithWhere(t, db)
}

// TestSQLServerIntegrationQueryAggregateAvg verifies that Avg() returns a non-zero
// average of the seeded users' IDs.
func TestSQLServerIntegrationQueryAggregateAvg(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestAggregateAvg(t, db)
}

// TestSQLServerIntegrationQueryAggregateMax verifies that Max() returns a non-zero
// maximum ID from the seeded users.
func TestSQLServerIntegrationQueryAggregateMax(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestAggregateMax(t, db)
}

// TestSQLServerIntegrationQueryAggregateMin verifies that Min() returns a non-zero
// minimum ID from the seeded users.
func TestSQLServerIntegrationQueryAggregateMin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestAggregateMin(t, db)
}

// TestSQLServerIntegrationQueryAggregateGroupBy verifies that GROUP BY combined
// with SUM produces exactly two result rows (one per avatar group).
func TestSQLServerIntegrationQueryAggregateGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestAggregateGroupBy(t, db)
}
