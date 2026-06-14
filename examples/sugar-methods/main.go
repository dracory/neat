package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// User represents a user model
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:255"`
	Email string `gorm:"size:255;uniqueIndex"`
	Age   int
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

func main() {
	if err := RunExample("sqlite://./sugar-methods-example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates all sugar methods
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create users table
	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("age")
	})
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create sample users
	sampleUsers := []User{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
		{Name: "Diana", Email: "diana@example.com", Age: 28},
		{Name: "Eve", Email: "eve@example.com", Age: 32},
	}

	for _, user := range sampleUsers {
		err := db.Query().Table("users").Create(user)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	fmt.Println("=== Sugar Methods Examples ===\n")

	// Example 1: CountAsVar - Count records
	fmt.Println("1. CountAsVar - Count total users:")
	count, err := db.Query().Table("users").CountAsVar()
	if err != nil {
		return fmt.Errorf("CountAsVar failed: %w", err)
	}
	fmt.Printf("   Total users: %d\n\n", count)

	// Example 2: SumAsVar - Sum column values
	fmt.Println("2. SumAsVar - Sum of all ages:")
	totalAge, err := db.Query().Table("users").SumAsVar("age")
	if err != nil {
		return fmt.Errorf("SumAsVar failed: %w", err)
	}
	fmt.Printf("   Total age: %.0f\n\n", totalAge)

	// Example 3: AvgAsVar - Average of column values
	fmt.Println("3. AvgAsVar - Average age:")
	avgAge, err := db.Query().Table("users").AvgAsVar("age")
	if err != nil {
		return fmt.Errorf("AvgAsVar failed: %w", err)
	}
	fmt.Printf("   Average age: %.1f\n\n", avgAge)

	// Example 4: MinAsVar - Minimum value
	fmt.Println("4. MinAsVar - Minimum age:")
	minAge, err := db.Query().Table("users").MinAsVar("age")
	if err != nil {
		return fmt.Errorf("MinAsVar failed: %w", err)
	}
	fmt.Printf("   Minimum age: %.0f\n\n", minAge)

	// Example 5: MaxAsVar - Maximum value
	fmt.Println("5. MaxAsVar - Maximum age:")
	maxAge, err := db.Query().Table("users").MaxAsVar("age")
	if err != nil {
		return fmt.Errorf("MaxAsVar failed: %w", err)
	}
	fmt.Printf("   Maximum age: %.0f\n\n", maxAge)

	// Example 6: ExistsAsVar - Check if records exist
	fmt.Println("6. ExistsAsVar - Check if users over 30 exist:")
	exists, err := db.Query().Table("users").Where("age > ?", 30).ExistsAsVar()
	if err != nil {
		return fmt.Errorf("ExistsAsVar failed: %w", err)
	}
	fmt.Printf("   Users over 30 exist: %t\n\n", exists)

	// Example 7: PluckAsVar - Get single column values
	fmt.Println("7. PluckAsVar - Get all email addresses:")
	emailsAny, err := db.Query().Table("users").PluckAsVar("email")
	if err != nil {
		return fmt.Errorf("PluckAsVar failed: %w", err)
	}
	fmt.Println("   Emails:")
	for _, emailAny := range emailsAny {
		if email, ok := emailAny.(string); ok {
			fmt.Printf("   - %s\n", email)
		}
	}
	fmt.Println()

	// Example 8: ValueAsVar - Get single column value from first record
	fmt.Println("8. ValueAsVar - Get email of first user:")
	emailAny, err := db.Query().Table("users").Where("age = ?", 25).ValueAsVar("email")
	if err != nil {
		return fmt.Errorf("ValueAsVar failed: %w", err)
	}
	email := emailAny.(string)
	fmt.Printf("   Email: %s\n\n", email)

	// Example 9: FirstAsVar - Get first record
	fmt.Println("9. FirstAsVar - Get first user over 25:")
	userAny, err := db.Query().Table("users").Where("age > ?", 25).FirstAsVar()
	if err != nil {
		return fmt.Errorf("FirstAsVar failed: %w", err)
	}
	user := userAny.(User)
	fmt.Printf("   User: %s (Age: %d)\n\n", user.Name, user.Age)

	// Example 10: FindOneAsVar - Alias for FirstAsVar
	fmt.Println("10. FindOneAsVar - Find one user (Sequelize-style):")
	userAny, err = db.Query().Table("users").Where("name = ?", "Bob").FindOneAsVar()
	if err != nil {
		return fmt.Errorf("FindOneAsVar failed: %w", err)
	}
	user = userAny.(User)
	fmt.Printf("    User: %s (Age: %d)\n\n", user.Name, user.Age)

	// Example 11: GetAsVar - Get all records
	fmt.Println("11. GetAsVar - Get all users over 28:")
	usersAny, err := db.Query().Table("users").Where("age > ?", 28).GetAsVar()
	if err != nil {
		return fmt.Errorf("GetAsVar failed: %w", err)
	}
	fmt.Println("    Users:")
	for _, userAny := range usersAny {
		if user, ok := userAny.(User); ok {
			fmt.Printf("    - %s (Age: %d)\n", user.Name, user.Age)
		}
	}
	fmt.Println()

	// Example 12: AllAsVar - Alias for GetAsVar (Django-style)
	fmt.Println("12. AllAsVar - Get all users (Django-style):")
	usersAny, err = db.Query().Table("users").AllAsVar()
	if err != nil {
		return fmt.Errorf("AllAsVar failed: %w", err)
	}
	fmt.Printf("    Total users: %d\n\n", len(usersAny))

	// Example 13: FindAllAsVar - Alias for GetAsVar (Sequelize-style)
	fmt.Println("13. FindAllAsVar - Find all users (Sequelize-style):")
	usersAny, err = db.Query().Table("users").FindAllAsVar()
	if err != nil {
		return fmt.Errorf("FindAllAsVar failed: %w", err)
	}
	fmt.Printf("    Total users: %d\n\n", len(usersAny))

	// Example 14: FindAsVar - Find with conditions
	fmt.Println("14. FindAsVar - Find users with conditions:")
	usersAny, err = db.Query().Table("users").FindAsVar("age BETWEEN ? AND ?", 28, 35)
	if err != nil {
		return fmt.Errorf("FindAsVar failed: %w", err)
	}
	fmt.Println("    Users aged 28-35:")
	for _, userAny := range usersAny {
		if user, ok := userAny.(User); ok {
			fmt.Printf("    - %s (Age: %d)\n", user.Name, user.Age)
		}
	}
	fmt.Println()

	fmt.Println("=== All Examples Completed Successfully ===")
	return nil
}
