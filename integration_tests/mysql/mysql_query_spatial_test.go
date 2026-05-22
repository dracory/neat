//go:build integration

package mysql

import (
	"testing"
)

func TestMySQLIntegrationSpatial(t *testing.T) {
	t.Skip("ORM Raw() returns *Query struct which cannot be used as a map value in Create() — spatial inserts not yet supported")
}
