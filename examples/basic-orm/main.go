package main

import (
	"log"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/database/db"
)

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Create database connection using DBConfig
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	db, err := database.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Note: Schema builder requires config adapter (deferred implementation)
	// For now, this example demonstrates the basic ORM structure

	// Create a user (schema creation would be done via schema builder)
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// The actual query execution would work once database is properly connected
	log.Printf("Example user created: %+v", user)
	log.Println("This example demonstrates the basic ORM structure")
	log.Println("Full functionality requires database connection and schema setup")
}
