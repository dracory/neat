package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// This example demonstrates advanced query builder features
func main() {
	if err := RunExample("sqlite://./example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates advanced query builder features
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create tables for the example
	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("age")
		blueprint.String("status")
		blueprint.Timestamp("created_at")
		blueprint.Timestamp("deleted_at").Nullable()
	})
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	err = db.Schema().Create("posts", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.Integer("user_id")
		blueprint.String("title")
		blueprint.Text("content")
	})
	if err != nil {
		return fmt.Errorf("failed to create posts table: %w", err)
	}

	err = db.Schema().Create("orders", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.Integer("user_id")
		blueprint.Decimal("amount")
	})
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	// Example 1: Join queries
	fmt.Println("=== Join Query ===")
	var results []map[string]any
	err = db.Query().
		Table("users").
		Select("users.name", "posts.title").
		Join("posts", "users.id", "=", "posts.user_id").
		Where("users.status = ?", "active").
		Get(&results)
	if err != nil {
		return fmt.Errorf("error in join query: %w", err)
	}
	fmt.Printf("Found %d results\n", len(results))

	// Example 2: Where conditions with OR
	fmt.Println("\n=== OR Conditions ===")
	var orResults []map[string]any
	err = db.Query().
		Table("users").
		Where("status = ?", "active").
		OrWhere("status = ?", "pending").
		Get(&orResults)
	if err != nil {
		return fmt.Errorf("error in OR query: %w", err)
	}
	fmt.Printf("Found %d users\n", len(orResults))

	// Example 3: WhereIn clause
	fmt.Println("\n=== WhereIn Clause ===")
	var inResults []map[string]any
	err = db.Query().
		Table("users").
		WhereIn("id", []any{1, 2, 3, 4, 5}).
		Get(&inResults)
	if err != nil {
		return fmt.Errorf("error in WhereIn query: %w", err)
	}
	fmt.Printf("Found %d users\n", len(inResults))

	// Example 4: WhereBetween clause
	fmt.Println("\n=== WhereBetween Clause ===")
	var betweenResults []map[string]any
	err = db.Query().
		Table("users").
		Where("age BETWEEN ? AND ?", 18, 30).
		Get(&betweenResults)
	if err != nil {
		return fmt.Errorf("error in WhereBetween query: %w", err)
	}
	fmt.Printf("Found %d users\n", len(betweenResults))

	// Example 5: WhereNull and WhereNotNull
	fmt.Println("\n=== WhereNull Clause ===")
	var nullResults []map[string]any
	err = db.Query().
		Table("users").
		Where("deleted_at IS NULL").
		Get(&nullResults)
	if err != nil {
		return fmt.Errorf("error in WhereNull query: %w", err)
	}
	fmt.Printf("Found %d non-deleted users\n", len(nullResults))

	// Example 6: Group and Having
	fmt.Println("\n=== Group and Having ===")
	var groupResults []map[string]any
	err = db.Query().
		Table("users").
		Select("status", "COUNT(*) as count").
		Group("status").
		Having("COUNT(*) > 0").
		Get(&groupResults)
	if err != nil {
		return fmt.Errorf("error in Group query: %w", err)
	}
	fmt.Printf("Found %d status groups\n", len(groupResults))

	// Example 7: OrderBy with multiple columns
	fmt.Println("\n=== Multiple OrderBy ===")
	var orderResults []map[string]any
	err = db.Query().
		Table("users").
		OrderBy("status", "asc").
		OrderBy("created_at", "desc").
		Limit(10).
		Get(&orderResults)
	if err != nil {
		return fmt.Errorf("error in OrderBy query: %w", err)
	}
	fmt.Printf("Found %d users\n", len(orderResults))

	// Example 8: Pagination with Offset
	fmt.Println("\n=== Pagination with Offset ===")
	page := 2
	perPage := 10
	offset := (page - 1) * perPage
	var paginatedResults []map[string]any
	err = db.Query().
		Table("users").
		Offset(offset).
		Limit(perPage).
		Get(&paginatedResults)
	if err != nil {
		return fmt.Errorf("error in pagination query: %w", err)
	}
	fmt.Printf("Found %d users on page %d\n", len(paginatedResults), page)

	// Example 9: Aggregation functions
	fmt.Println("\n=== Aggregation Functions ===")
	var count int64
	err = db.Query().Table("users").Count(&count)
	if err != nil {
		return fmt.Errorf("error in Count query: %w", err)
	}
	fmt.Printf("Total users: %d\n", count)

	var sum float64
	err = db.Query().Table("orders").Sum("amount", &sum)
	if err != nil {
		return fmt.Errorf("error in Sum query: %w", err)
	}
	fmt.Printf("Total order amount: %.2f\n", sum)

	// Example 10: Distinct
	fmt.Println("\n=== Distinct ===")
	var distinctResults []map[string]any
	err = db.Query().
		Table("users").
		Distinct("status").
		Get(&distinctResults)
	if err != nil {
		return fmt.Errorf("error in Distinct query: %w", err)
	}
	fmt.Printf("Found %d distinct status values\n", len(distinctResults))

	return nil
}
