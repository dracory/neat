package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTursoIntegrationQueryDeleteByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestQueryDeleteByModel(t, db)
}

func TestTursoIntegrationQueryDeleteByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestQueryDeleteByTable(t, db)
}

func TestTursoIntegrationQueryDeleteByModelWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.TestQueryDeleteByModelWithWhere(t, db)
}
