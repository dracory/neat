package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

// User is a plain struct representing a user in the system.
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
	if err := RunAdvancedExample(); err != nil {
		log.Fatalf("Advanced example failed: %v", err)
	}
}

// RunAdvancedExample demonstrates advanced array-driver queries: JOINs
// between two array sources, GROUP BY with aggregates, HAVING, and
// aggregate functions (Count, Sum).
//
// Note: The query builder's Group() method currently rejects dotted
// identifiers like "users.name" (treats them as invalid). To work around
// this, GROUP BY uses the column alias from the Select clause instead.
func RunAdvancedExample() error {
	database, err := newDatabase()
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Printf("failed to close database: %v", err)
		}
	}()

	if err := ExampleJoinTwoArrays(database); err != nil {
		return err
	}
	if err := ExampleLeftJoinWithMissingData(database); err != nil {
		return err
	}
	if err := ExampleGroupByWithCount(database); err != nil {
		return err
	}
	if err := ExampleGroupByWithSumAndHaving(database); err != nil {
		return err
	}
	if err := ExampleAggregateCount(database); err != nil {
		return err
	}
	if err := ExampleAggregateSum(database); err != nil {
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
				Driver: "array",
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
		// Carol (user 3) has no orders — useful for LEFT JOIN demo
	}
}

// UserOrder is the result struct for JOIN queries between users and orders.
type UserOrder struct {
	UserName  string  `db:"user_name"`
	UserEmail string  `db:"user_email"`
	OrderID   int     `db:"order_id"`
	Total     float64 `db:"total"`
	Status    string  `db:"status"`
}

// ExampleJoinTwoArrays demonstrates an INNER JOIN between two array-backed
// sources. Each source is populated as a separate table with a known name
// (via .Table()), then joined using a standard SQL JOIN clause.
//
// The key insight: you must use .Table("name") to give each source a known
// table name, then reference those names in the Join() clause. Both tables
// must be populated before the JOIN query runs — the array driver populates
// on first Model() call, so we touch the second source with a dummy query
// first.
//
//	userSource := neat.NewArraySourceFrom(users).Table("users")
//	orderSource := neat.NewArraySourceFrom(orders).Table("orders")
//
//	database.Query().
//	    Model(userSource).
//	    Join("orders ON users.id = orders.user_id").
//	    Select("users.name as user_name, ...").
//	    Get(&results)
func ExampleJoinTwoArrays(database *neat.Database) error {
	fmt.Println("=== JOIN: Users + Orders ===")

	userSource := neat.NewArraySourceFrom(users()).Table("users")
	orderSource := neat.NewArraySourceFrom(orders()).Table("orders")

	// Populate the orders table first — the array driver creates the table
	// on the first Model() call. Both tables must exist before the JOIN.
	var dummy []Order
	if err := database.Query().Model(orderSource).Get(&dummy); err != nil {
		return fmt.Errorf("failed to populate orders table: %w", err)
	}

	var results []UserOrder
	err := database.Query().
		Model(userSource).
		Join("orders ON users.id = orders.user_id").
		Select("users.name as user_name, users.email as user_email, orders.id as order_id, orders.total, orders.status").
		OrderBy("users.name", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to join users and orders: %w", err)
	}

	for _, r := range results {
		fmt.Printf("  %s <%s> — Order #%d: $%.2f (%s)\n", r.UserName, r.UserEmail, r.OrderID, r.Total, r.Status)
	}
	return nil
}

// ExampleLeftJoinWithMissingData demonstrates a LEFT JOIN where one user
// (Carol) has no orders. With LEFT JOIN, Carol still appears in the results
// with NULL order fields, which are scanned into pointer fields (*int,
// *float64) on the result struct.
//
//	database.Query().
//	    Model(userSource).
//	    LeftJoin("orders ON users.id = orders.user_id").
//	    Select("users.name as user_name, orders.id as order_id, orders.total").
//	    Get(&results)
func ExampleLeftJoinWithMissingData(database *neat.Database) error {
	fmt.Println("\n=== LEFT JOIN: Users with no orders still appear ===")

	userSource := neat.NewArraySourceFrom(users()).Table("users")
	orderSource := neat.NewArraySourceFrom(orders()).Table("orders")

	var dummy []Order
	if err := database.Query().Model(orderSource).Get(&dummy); err != nil {
		return fmt.Errorf("failed to populate orders table: %w", err)
	}

	type UserWithOrderInfo struct {
		UserName string   `db:"user_name"`
		OrderID  *int     `db:"order_id"` // pointer handles NULL from LEFT JOIN
		Total    *float64 `db:"total"`
	}

	var results []UserWithOrderInfo
	err := database.Query().
		Model(userSource).
		LeftJoin("orders ON users.id = orders.user_id").
		Select("users.name as user_name, orders.id as order_id, orders.total").
		OrderBy("users.name", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to left join: %w", err)
	}

	for _, r := range results {
		if r.OrderID == nil {
			fmt.Printf("  %s — no orders\n", r.UserName)
		} else {
			fmt.Printf("  %s — Order #%d: $%.2f\n", r.UserName, *r.OrderID, *r.Total)
		}
	}
	return nil
}

// OrderCount is the result struct for GROUP BY count queries.
type OrderCount struct {
	UserName string `db:"user_name"`
	Count    int64  `db:"order_count"`
}

// ExampleGroupByWithCount demonstrates GROUP BY with COUNT(*) aggregate.
// Groups orders by user and counts how many orders each user has.
//
// Note: The query builder's Group() method rejects dotted identifiers
// (e.g. "users.name"). Use the column alias from the Select clause
// instead ("user_name").
//
//	database.Query().
//	    Model(orderSource).
//	    Join("users ON users.id = orders.user_id").
//	    Select("users.name as user_name, COUNT(*) as order_count").
//	    Group("user_name").
//	    Get(&results)
func ExampleGroupByWithCount(database *neat.Database) error {
	fmt.Println("\n=== GROUP BY + COUNT: Orders per user ===")

	userSource := neat.NewArraySourceFrom(users()).Table("users")
	orderSource := neat.NewArraySourceFrom(orders()).Table("orders")

	var dummy []User
	if err := database.Query().Model(userSource).Get(&dummy); err != nil {
		return fmt.Errorf("failed to populate users table: %w", err)
	}

	var results []OrderCount
	err := database.Query().
		Model(orderSource).
		Join("users ON users.id = orders.user_id").
		Select("users.name as user_name, COUNT(*) as order_count").
		Group("user_name").
		OrderBy("user_name", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to group by with count: %w", err)
	}

	for _, r := range results {
		fmt.Printf("  %s — %d order(s)\n", r.UserName, r.Count)
	}
	return nil
}

// UserTotal is the result struct for GROUP BY SUM queries.
type UserTotal struct {
	UserName string  `db:"user_name"`
	TotalSum float64 `db:"total_sum"`
}

// ExampleGroupByWithSumAndHaving demonstrates GROUP BY with SUM aggregate
// and a HAVING clause to filter groups. Shows only users whose total
// order value exceeds $100.
//
//	database.Query().
//	    Model(orderSource).
//	    Join("users ON users.id = orders.user_id").
//	    Select("users.name as user_name, SUM(orders.total) as total_sum").
//	    Group("user_name").
//	    Having("total_sum > ?", 100).
//	    Get(&results)
func ExampleGroupByWithSumAndHaving(database *neat.Database) error {
	fmt.Println("\n=== GROUP BY + SUM + HAVING: Users with > $100 total ===")

	userSource := neat.NewArraySourceFrom(users()).Table("users")
	orderSource := neat.NewArraySourceFrom(orders()).Table("orders")

	var dummy []User
	if err := database.Query().Model(userSource).Get(&dummy); err != nil {
		return fmt.Errorf("failed to populate users table: %w", err)
	}

	var results []UserTotal
	err := database.Query().
		Model(orderSource).
		Join("users ON users.id = orders.user_id").
		Select("users.name as user_name, SUM(orders.total) as total_sum").
		Group("user_name").
		Having("total_sum > ?", 100).
		OrderBy("total_sum", "desc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to group by with sum and having: %w", err)
	}

	for _, r := range results {
		fmt.Printf("  %s — $%.2f total\n", r.UserName, r.TotalSum)
	}
	return nil
}

// ExampleAggregateCount demonstrates the Count() aggregate method on a
// single array source. Counts the total number of orders.
//
//	database.Query().
//	    Model(orderSource).
//	    Count(&count)
func ExampleAggregateCount(database *neat.Database) error {
	fmt.Println("\n=== Aggregate: Count all orders ===")

	orderSource := neat.NewArraySourceFrom(orders()).Table("orders")

	var count int64
	err := database.Query().
		Model(orderSource).
		Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count orders: %w", err)
	}

	fmt.Printf("  Total orders: %d\n", count)
	return nil
}

// ExampleAggregateSum demonstrates the Sum() aggregate method with a
// Where clause. Sums the total of completed orders only.
//
//	database.Query().
//	    Model(orderSource).
//	    Where("status = ?", "completed").
//	    Sum("total", &sum)
func ExampleAggregateSum(database *neat.Database) error {
	fmt.Println("\n=== Aggregate: Sum of completed orders ===")

	orderSource := neat.NewArraySourceFrom(orders()).Table("orders")

	var sum float64
	err := database.Query().
		Model(orderSource).
		Where("status = ?", "completed").
		Sum("total", &sum)
	if err != nil {
		return fmt.Errorf("failed to sum completed orders: %w", err)
	}

	fmt.Printf("  Completed orders total: $%.2f\n", sum)
	return nil
}
