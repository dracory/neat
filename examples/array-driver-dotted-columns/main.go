package main

import (
	"fmt"
	"log"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

// User is a plain struct representing a user.
type User struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

// Order is a plain struct representing an order placed by a user.
type Order struct {
	ID     int     `db:"id"`
	UserID int     `db:"user_id"`
	Total  float64 `db:"total"`
	Status string  `db:"status"`
}

func main() {
	if err := RunDottedColumnExample(); err != nil {
		log.Fatalf("Dotted column example failed: %v", err)
	}
}

// RunDottedColumnExample demonstrates table-qualified column references
// (e.g. "users.name") in ORDER BY, GROUP BY, WhereColumn, and Distinct
// clauses when joining two array-backed sources.
func RunDottedColumnExample() error {
	database, err := newDatabase()
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Printf("failed to close database: %v", err)
		}
	}()

	if err := ExampleOrderByDottedColumn(database); err != nil {
		return err
	}
	if err := ExampleGroupByDottedColumn(database); err != nil {
		return err
	}
	if err := ExampleWhereColumnDotted(database); err != nil {
		return err
	}
	if err := ExampleDistinctDottedColumn(database); err != nil {
		return err
	}

	return nil
}

// newDatabase creates an in-memory array-driver database connection.
func newDatabase() (*neat.Database, error) {
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
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	return database, nil
}

// users returns the shared slice of User structs.
func users() []User {
	return []User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
		{ID: 3, Name: "Carol", Email: "carol@example.com"},
	}
}

// orders returns the shared slice of Order structs.
func orders() []Order {
	return []Order{
		{ID: 101, UserID: 1, Total: 50.00, Status: "completed"},
		{ID: 102, UserID: 1, Total: 30.00, Status: "completed"},
		{ID: 103, UserID: 2, Total: 120.00, Status: "pending"},
		{ID: 104, UserID: 2, Total: 80.00, Status: "completed"},
	}
}

// populateBoth ensures both array-backed tables exist in the SQLite
// connection before running JOIN queries. The array driver populates
// a table on the first Model() call, so both tables must be touched
// before a JOIN references them.
func populateBoth(database *neat.Database) error {
	var d1 []User
	if err := database.Query().Model(neat.NewArraySourceFrom(users()).Table("users")).Get(&d1); err != nil {
		return fmt.Errorf("failed to populate users table: %w", err)
	}
	var d2 []Order
	if err := database.Query().Model(neat.NewArraySourceFrom(orders()).Table("orders")).Get(&d2); err != nil {
		return fmt.Errorf("failed to populate orders table: %w", err)
	}
	return nil
}

// ExampleOrderByDottedColumn demonstrates ORDER BY with a table-qualified
// column name ("users.name") in a JOIN query. Without dotted column
// support, the ORDER BY clause would be silently dropped.
//
//	database.Query().
//	    Model(userSource).
//	    Join("orders ON users.id = orders.user_id").
//	    Select("users.name as user_name, orders.total").
//	    OrderBy("users.name", "asc").
//	    Get(&results)
func ExampleOrderByDottedColumn(database *neat.Database) error {
	fmt.Println("=== ORDER BY with dotted column (users.name) ===")

	if err := populateBoth(database); err != nil {
		return err
	}

	userSource := neat.NewArraySourceFrom(users()).Table("users")

	type Result struct {
		UserName string  `db:"user_name"`
		Total    float64 `db:"total"`
	}

	var results []Result
	err := database.Query().
		Model(userSource).
		Join("orders ON users.id = orders.user_id").
		Select("users.name as user_name, orders.total").
		OrderBy("users.name", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to order by dotted column: %w", err)
	}

	for _, r := range results {
		fmt.Printf("  %s — $%.2f\n", r.UserName, r.Total)
	}
	return nil
}

// ExampleGroupByDottedColumn demonstrates GROUP BY with a table-qualified
// column name ("users.name") in a JOIN query with COUNT(*) aggregate.
//
//	database.Query().
//	    Model(orderSource).
//	    Join("users ON users.id = orders.user_id").
//	    Select("users.name as user_name, COUNT(*) as order_count").
//	    Group("users.name").
//	    Get(&results)
func ExampleGroupByDottedColumn(database *neat.Database) error {
	fmt.Println("\n=== GROUP BY with dotted column (users.name) ===")

	if err := populateBoth(database); err != nil {
		return err
	}

	orderSource := neat.NewArraySourceFrom(orders()).Table("orders")

	type Result struct {
		UserName string `db:"user_name"`
		Count    int64  `db:"order_count"`
	}

	var results []Result
	err := database.Query().
		Model(orderSource).
		Join("users ON users.id = orders.user_id").
		Select("users.name as user_name, COUNT(*) as order_count").
		Group("users.name").
		OrderBy("users.name", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to group by dotted column: %w", err)
	}

	for _, r := range results {
		fmt.Printf("  %s — %d order(s)\n", r.UserName, r.Count)
	}
	return nil
}

// ExampleWhereColumnDotted demonstrates WhereColumn with table-qualified
// column names on both sides of the comparison. This is the most common
// use case for JOIN queries — comparing columns across two tables.
//
//	database.Query().
//	    Model(userSource).
//	    Join("orders ON users.id = orders.user_id").
//	    WhereColumn("users.id", "=", "orders.user_id").
//	    Get(&results)
func ExampleWhereColumnDotted(database *neat.Database) error {
	fmt.Println("\n=== WhereColumn with dotted columns (users.id = orders.user_id) ===")

	if err := populateBoth(database); err != nil {
		return err
	}

	userSource := neat.NewArraySourceFrom(users()).Table("users")

	type Result struct {
		UserName string  `db:"user_name"`
		OrderID  int     `db:"order_id"`
		Total    float64 `db:"total"`
	}

	var results []Result
	err := database.Query().
		Model(userSource).
		Join("orders ON users.id = orders.user_id").
		WhereColumn("users.id", "=", "orders.user_id").
		Select("users.name as user_name, orders.id as order_id, orders.total").
		OrderBy("users.name", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to where column with dotted columns: %w", err)
	}

	for _, r := range results {
		fmt.Printf("  %s — Order #%d: $%.2f\n", r.UserName, r.OrderID, r.Total)
	}
	return nil
}

// ExampleDistinctDottedColumn demonstrates Distinct with a table-qualified
// column name. The distinct column is used in COUNT(DISTINCT col) aggregate
// queries — without dotted column support, the column would be silently
// dropped and the count would be wrong.
//
//	database.Query().
//	    Model(userSource).
//	    Join("orders ON users.id = orders.user_id").
//	    Distinct("users.name").
//	    Count(&count)
func ExampleDistinctDottedColumn(database *neat.Database) error {
	fmt.Println("\n=== Distinct with dotted column (users.name) ===")

	if err := populateBoth(database); err != nil {
		return err
	}

	userSource := neat.NewArraySourceFrom(users()).Table("users")

	// COUNT(DISTINCT users.name) — counts unique users who have orders.
	// Alice and Bob have orders, Carol doesn't → expect 2.
	var count int64
	err := database.Query().
		Model(userSource).
		Join("orders ON users.id = orders.user_id").
		Distinct("users.name").
		Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count distinct: %w", err)
	}

	fmt.Printf("  Distinct users with orders: %d (expected 2)\n", count)
	return nil
}
