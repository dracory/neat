package tidb_test

import (
	"testing"
)

type SpatialModel struct {
	ID       uint   `db:"id"`
	Name     string `db:"name"`
	Location string `db:"location"`
}

func (SpatialModel) TableName() string {
	return "spatial_models"
}

func TestTiDBIntegrationSpatial(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// TiDB has limited spatial type support. The schema builder generates
	// a `point` column type that TiDB's parser rejects (Error 1064).
	// See: https://docs.pingcap.com/tidb/stable/mysql-compatibility
	t.Skip("TiDB has limited spatial type support; the POINT column type is not supported by the schema builder")
}
