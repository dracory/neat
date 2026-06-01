package turso

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTursoIntegrationQueryCountBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.QueryCountBasic(t, db)
}

func TestTursoIntegrationQueryCountWithTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.QueryCountWithTable(t, db)
}
