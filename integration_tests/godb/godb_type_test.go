//go:build integration

package godb

import (
	"testing"
	"time"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat"
	"github.com/dracory/neat/database/driver"
)

func TestGODB_TypePreservation_And_Nulls(t *testing.T) {
	type ComplexTypeStruct struct {
		ID       int64      `db:"id"`
		StrVal   string     `db:"str_val"`
		IntVal   *int       `db:"int_val"`
		FloatVal *float64   `db:"float_val"`
		BoolVal  *bool      `db:"bool_val"`
		TimeVal  *time.Time `db:"time_val"`
	}

	intNum := 100
	floatNum := 123.45
	bVal := true
	now := time.Now().UTC().Truncate(time.Second)

	complexData := []ComplexTypeStruct{
		{
			ID:       1,
			StrVal:   "Row 1",
			IntVal:   &intNum,
			FloatVal: &floatNum,
			BoolVal:  &bVal,
			TimeVal:  &now,
		},
		{
			ID:       2,
			StrVal:   "Row 2",
			IntVal:   nil,
			FloatVal: nil,
			BoolVal:  nil,
			TimeVal:  nil,
		},
	}

	config := neat.DBConfig{
		Default: "go_db",
		Connections: map[string]neat.ConnectionConfig{
			"go_db": {
				Driver: contractsdb.DriverGODB,
				Tables: driver.Tables{
					"complex_types": complexData,
				},
			},
		},
	}

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Row 1 verification
	var r1 ComplexTypeStruct
	err = db.Query().Table("complex_types").Where("id = ?", 1).First(&r1)
	if err != nil {
		t.Fatalf("failed to get row 1: %v", err)
	}

	if r1.StrVal != "Row 1" || r1.IntVal == nil || *r1.IntVal != 100 || r1.FloatVal == nil || *r1.FloatVal != 123.45 || r1.BoolVal == nil || !*r1.BoolVal || r1.TimeVal == nil || !r1.TimeVal.Equal(now) {
		t.Errorf("Row 1 was not correctly preserved: %+v", r1)
		if r1.TimeVal != nil {
			t.Errorf("now: %v, retrieved: %v", now, *r1.TimeVal)
		}
	}

	// Row 2 verification (NULL values)
	var r2 ComplexTypeStruct
	err = db.Query().Table("complex_types").Where("id = ?", 2).First(&r2)
	if err != nil {
		t.Fatalf("failed to get row 2: %v", err)
	}

	if r2.StrVal != "Row 2" || r2.IntVal != nil || r2.FloatVal != nil || r2.BoolVal != nil || r2.TimeVal != nil {
		t.Errorf("Row 2 was not correctly handled (expected NULL/nil fields): %+v", r2)
	}
}

func TestGODB_BothConfigStyles(t *testing.T) {
	// 1. Style A (Tables map)
	configA := neat.DBConfig{
		Default: "go_db_a",
		Connections: map[string]neat.ConnectionConfig{
			"go_db_a": {
				Driver: contractsdb.DriverGODB,
				Tables: driver.Tables{
					"users": usersData,
				},
			},
		},
	}
	dbA, err := neat.New(configA)
	if err != nil {
		t.Fatalf("Style A failed: %v", err)
	}
	defer func() { _ = dbA.Close() }()

	var countA int64
	err = dbA.Query().Table("users").Count(&countA)
	if err != nil || countA != 3 {
		t.Errorf("Style A query failed: %v, count=%d", err, countA)
	}

	// 2. Style B ([]Table slice)
	configB := neat.DBConfig{
		Default: "go_db_b",
		Connections: map[string]neat.ConnectionConfig{
			"go_db_b": {
				Driver: contractsdb.DriverGODB,
				Tables: []driver.Table{
					{Name: "users", Data: usersData},
				},
			},
		},
	}
	dbB, err := neat.New(configB)
	if err != nil {
		t.Fatalf("Style B failed: %v", err)
	}
	defer func() { _ = dbB.Close() }()

	var countB int64
	err = dbB.Query().Table("users").Count(&countB)
	if err != nil || countB != 3 {
		t.Errorf("Style B query failed: %v, count=%d", err, countB)
	}
}
