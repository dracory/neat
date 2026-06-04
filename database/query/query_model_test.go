package query

import (
	"context"
	"testing"

	"github.com/dracory/neat/database/driver"
)

type User struct {
	ID   uint
	Name string
}

func (User) TableName() string {
	return "users"
}

type TestAddress struct {
	ID     uint
	UserID uint
	Street string
}

func (TestAddress) TableName() string {
	return "addresses"
}

type Product struct {
	ID    uint
	Name  string
	Price float64
}

func TestModelAlwaysUpdatesTableName(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	q.Model(User{})
	if q.table != "users" {
		t.Errorf("Expected table 'users', got '%s'", q.table)
	}

	q.Model(TestAddress{})
	if q.table != "addresses" {
		t.Errorf("Expected table 'addresses', got '%s'", q.table)
	}
}

func TestResolveTableNameWithCustomMethod(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	tableName := q.resolveTableName(User{})
	if tableName != "users" {
		t.Errorf("Expected 'users', got '%s'", tableName)
	}

	tableName = q.resolveTableName(TestAddress{})
	if tableName != "addresses" {
		t.Errorf("Expected 'addresses', got '%s'", tableName)
	}
}

func TestResolveTableNameWithoutCustomMethod(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	tableName := q.resolveTableName(Product{})
	if tableName != "products" {
		t.Errorf("Expected 'products', got '%s'", tableName)
	}
}

func TestResolveTableNameWithPointer(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	tableName := q.resolveTableName(&User{})
	if tableName != "users" {
		t.Errorf("Expected 'users' from pointer, got '%s'", tableName)
	}
}

func TestResolveTableNameWithSlice(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	users := []User{}
	tableName := q.resolveTableName(users)
	if tableName != "users" {
		t.Errorf("Expected 'users' from slice, got '%s'", tableName)
	}

	userPtrs := []*User{}
	tableName = q.resolveTableName(userPtrs)
	if tableName != "users" {
		t.Errorf("Expected 'users' from pointer slice, got '%s'", tableName)
	}
}

func TestResolveTableNameWithNil(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)

	tableName := q.resolveTableName(nil)
	if tableName != "" {
		t.Errorf("Expected empty string for nil, got '%s'", tableName)
	}
}

func TestModelResetsQueryState(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewSQLite(), "", nil, nil)
	q.Table("old_table")
	q.Where("status = ?", "active")
	q.Select("id", "name")
	q.Limit(10)

	q.Model(User{})

	if q.table != "users" {
		t.Errorf("Expected table 'users' after Model(), got '%s'", q.table)
	}
	if len(q.wheres) != 0 {
		t.Errorf("Expected wheres to be reset after Model(), got %d", len(q.wheres))
	}
	if len(q.selects) != 0 {
		t.Errorf("Expected selects to be reset after Model(), got %d", len(q.selects))
	}
	if q.limit != nil {
		t.Errorf("Expected limit to be reset after Model(), got %v", q.limit)
	}
}
