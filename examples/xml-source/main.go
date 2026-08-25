package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

// User is a plain Go struct. The db tags map struct fields to XML columns.
// Column names come from attributes and leaf sub-elements in the XML.
type User struct {
	ID      int       `db:"id"`
	Name    string    `db:"name"`
	Email   string    `db:"email"`
	Active  bool      `db:"active"`
	Created time.Time `db:"created"`
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates querying an XML file via NewXmlFileSource.
func RunExample() error {
	// Resolve the path to the sample XML file
	xmlPath := resolvePath("data", "users.xml")

	// Create an array-driver database connection
	config := neat.DBConfig{
		Default: "array_db",
		Connections: map[string]neat.ConnectionConfig{
			"array_db": {
				Driver: contractsdb.DriverArray,
			},
		},
	}

	database, err := neat.New(config)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Printf("failed to close database: %v", err)
		}
	}()

	// Query the XML file — NewXmlFileSource parses it and wraps it as
	// an ArraySource. The table name is derived from the filename:
	// "users.xml" → "users"
	// Columns come from attributes (id) and leaf sub-elements (name, email,
	// active, created). Type inference converts strings to int, bool, time.
	fmt.Println("=== NewXmlFileSource — query an XML file ===")

	var users []User
	err = database.Query().
		Model(neat.NewXmlFileSource(xmlPath)).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	for _, u := range users {
		fmt.Printf("User #%d: %s <%s> (created: %s)\n", u.ID, u.Name, u.Email, u.Created.Format("2006-01-02"))
	}

	// Query without filter — all rows
	fmt.Println("\n=== All users ===")

	var allUsers []User
	err = database.Query().
		Model(neat.NewXmlFileSource(xmlPath)).
		OrderBy("id", "asc").
		Get(&allUsers)
	if err != nil {
		return fmt.Errorf("failed to query all users: %w", err)
	}

	for _, u := range allUsers {
		fmt.Printf("User #%d: %s (active: %v)\n", u.ID, u.Name, u.Active)
	}

	return nil
}

// resolvePath tries a path relative to the working directory, then falls
// back to a path relative to the repo root.
func resolvePath(parts ...string) string {
	p := filepath.Join(parts...)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return filepath.Join(append([]string{"examples", "xml-source"}, parts...)...)
}
