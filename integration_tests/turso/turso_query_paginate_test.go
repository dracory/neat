package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTursoIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestPaginateFirstPage(t, db)
}

func TestTursoIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestPaginateSecondPage(t, db)
}

func TestTursoIntegrationPaginateLastPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestPaginateLastPage(t, db)
}

func TestTursoIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestPaginateWithConditions(t, db)
}

func TestTursoIntegrationPaginatePageBeyondBounds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestPaginatePageBeyondBounds(t, db)
}

func TestTursoIntegrationPaginateEmptyResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestPaginateEmptyResults(t, db)
}

func TestTursoIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestPaginateWithSelectAliases(t, db)
}

func TestTursoIntegrationCountWithSelectAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestCountWithSelectAlias(t, db)
}
