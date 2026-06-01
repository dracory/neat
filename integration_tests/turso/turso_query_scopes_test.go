package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTursoIntegrationQueryScopesLocalWithoutParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithoutParameters(t, db)
}

func TestTursoIntegrationQueryScopesLocalWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithParameters(t, db)
}

func TestTursoIntegrationQueryScopesMultipleChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesMultipleChaining(t, db)
}
