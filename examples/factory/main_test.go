package main_test

import (
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/factory"
)

func TestRunExample(t *testing.T) {
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

// User mirrors the example model for assertions
type User struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Email    string `db:"email"`
	Age      int    `db:"age"`
	Status   string `db:"status"`
	IsActive bool   `db:"is_active"`
}

func TestFactory_RunExample_UserCount(t *testing.T) {
	// RunExample creates: 1 + 3 (bulk) + 1 + 1 (quiet) + 1 (final) = 7 persisted; Make() is not persisted
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("age")
		blueprint.String("status")
		blueprint.Boolean("is_active")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Single create
	u := &User{Name: "John", Email: "john@example.com", Age: 30, Status: "active", IsActive: true}
	_, err = db.Factory().Table("users").Create(u)
	if err != nil {
		t.Fatalf("single create failed: %v", err)
	}

	var count int64
	if err = db.Query().Table("users").Count(&count); err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 after single create, got %d", count)
	}

	// Bulk create: +3
	tmpl := &User{Name: "Bulk", Email: "bulk@example.com", Age: 25, Status: "pending", IsActive: false}
	_, err = db.Factory().Table("users").Count(3).Create(tmpl)
	if err != nil {
		t.Fatalf("bulk create failed: %v", err)
	}

	if err = db.Query().Table("users").Count(&count); err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 4 {
		t.Errorf("expected 4 after bulk create, got %d", count)
	}

	// CreateQuietly: +1
	quiet := &User{Name: "Quiet", Email: "quiet@example.com", Age: 28, Status: "active", IsActive: true}
	_, err = db.Factory().Table("users").CreateQuietly(quiet)
	if err != nil {
		t.Fatalf("create quietly failed: %v", err)
	}

	if err = db.Query().Table("users").Count(&count); err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 after CreateQuietly, got %d", count)
	}
}

func TestFactory_Make_NotPersisted(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("age")
		blueprint.String("status")
		blueprint.Boolean("is_active")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	makeUser := &User{Name: "Make User", Email: "make@example.com", Age: 22}
	_, err = db.Factory().Make(makeUser)
	if err != nil {
		t.Fatalf("Make failed: %v", err)
	}

	// Make must NOT persist to DB
	var count int64
	if err = db.Query().Table("users").Count(&count); err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows after Make (not persisted), got %d", count)
	}
}

func TestFactory_BulkMake_Count(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	tmpl := &User{Name: "Bulk", Email: "bulk@example.com", Age: 30}
	result, err := db.Factory().Count(3).Make(tmpl)
	if err != nil {
		t.Fatalf("bulk Make failed: %v", err)
	}

	users, ok := result.([]*User)
	if !ok {
		t.Fatalf("expected []*User from Make, got %T", result)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 in-memory users from bulk Make, got %d", len(users))
	}
}
