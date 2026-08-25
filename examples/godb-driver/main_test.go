package main

import (
	"testing"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/examples/godb-driver/data"
	_ "modernc.org/sqlite"
)

func TestRunExample(t *testing.T) {
	if err := RunExample(); err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestGODBExampleQueries(t *testing.T) {
	config := neat.DBConfig{
		Default: "go_db",
		Connections: map[string]neat.ConnectionConfig{
			"go_db": {
				Driver: contractsdb.DriverGODB,
				Tables: driver.Tables{
					"blogs":      data.Blogs,
					"categories": data.Categories,
				},
			},
		},
	}

	database, err := neat.New(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	var blogs []data.Blog
	err = database.Query().
		Model(&data.Blog{}).
		Where("category_id = ?", 2).
		OrderBy("id", "asc").
		Get(&blogs)
	if err != nil {
		t.Fatalf("failed to query blogs: %v", err)
	}

	if len(blogs) != 2 {
		t.Fatalf("expected 2 blogs, got %d", len(blogs))
	}
	if blogs[0].Title != "Go Tips" || blogs[1].Title != "Advanced Patterns" {
		t.Errorf("unexpected blogs: %v", blogs)
	}

	var joined []BlogWithCategory
	err = database.Query().
		Table("blogs").
		LeftJoin("categories ON blogs.category_id = categories.id").
		Select("blogs.id, blogs.title, categories.name AS category_name").
		OrderBy("blogs.id", "asc").
		Get(&joined)
	if err != nil {
		t.Fatalf("failed to query join: %v", err)
	}

	if len(joined) != 3 {
		t.Fatalf("expected 3 joined rows, got %d", len(joined))
	}
	if joined[0].CategoryName != "General" || joined[1].CategoryName != "Programming" {
		t.Errorf("unexpected joined results: %v", joined)
	}
}
