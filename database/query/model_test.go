package query_test

import (
	"testing"

	"github.com/dracory/neat/database/query"
)

type User struct {
	ID   uint
	Name string
}

type Address struct {
	ID     uint
	Name   string
	UserID uint
}

func (User) TableName() string {
	return "users"
}

func (Address) TableName() string {
	return "addresses"
}

func TestModelAlwaysUpdatesTableName(t *testing.T) {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))

	// Set model to User using the public Model() method
	w.Q.Model(&User{})
	if w.GetTable() != "users" {
		t.Errorf("Expected table 'users', got '%s'", w.GetTable())
	}

	// Set model to Address on the same query object
	w.Q.Model(&Address{})
	if w.GetTable() != "addresses" {
		t.Errorf("Expected table 'addresses' after second Model() call, got '%s'", w.GetTable())
	}
}

func TestModelWithNilValue(t *testing.T) {
	w := query.WrapQuery(query.NewTestQuery(nil, nil, query.MakeDBConfig(), nil))

	w.Q.Model(nil)
	if w.GetTable() != "" {
		t.Errorf("Expected empty table for nil model, got '%s'", w.GetTable())
	}
}
