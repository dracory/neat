package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTursoIntegrationDottedColumn_OrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	if db == nil {
		t.Skip("Turso not available")
	}
	common.TestDottedColumnOrderBy(t, db)
}

func TestTursoIntegrationDottedColumn_OrderByDesc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	if db == nil {
		t.Skip("Turso not available")
	}
	common.TestDottedColumnOrderByDesc(t, db)
}

func TestTursoIntegrationDottedColumn_GroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	if db == nil {
		t.Skip("Turso not available")
	}
	common.TestDottedColumnGroupBy(t, db)
}

func TestTursoIntegrationDottedColumn_WhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	if db == nil {
		t.Skip("Turso not available")
	}
	common.TestDottedColumnWhereColumn(t, db)
}

func TestTursoIntegrationDottedColumn_Distinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	if db == nil {
		t.Skip("Turso not available")
	}
	common.TestDottedColumnDistinct(t, db)
}

func TestTursoIntegrationDottedColumn_GroupByHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	if db == nil {
		t.Skip("Turso not available")
	}
	common.TestDottedColumnGroupByHaving(t, db)
}
