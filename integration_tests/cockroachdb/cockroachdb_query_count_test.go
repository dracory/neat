package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestCockroachDBIntegrationQueryCountBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.QueryCountBasic(t, db)
}
