package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/database/schema"
)

// User is the model we will observe.
type User struct {
	ID        uint       `gorm:"column:id;primaryKey"`
	Name      string     `gorm:"column:name"`
	Email     string     `gorm:"column:email"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (User) TableName() string { return "users" }

// UserObserver reacts to User lifecycle events.
// It implements the required Observer interface plus the optional
// ObserverWithCreating, ObserverWithUpdating, and ObserverWithDeleting interfaces.
type UserObserver struct {
	events []string
}

// Creating is called before a user is inserted.
func (o *UserObserver) Creating(event contractsorm.Event) error {
	o.events = append(o.events, "creating")
	fmt.Println("[observer] Creating: about to insert a new user")
	return nil
}

// Created is called after a user is inserted.
func (o *UserObserver) Created(event contractsorm.Event) error {
	o.events = append(o.events, "created")
	fmt.Println("[observer] Created: new user inserted")
	return nil
}

// Updating is called before a user is updated.
func (o *UserObserver) Updating(event contractsorm.Event) error {
	o.events = append(o.events, "updating")
	fmt.Println("[observer] Updating: about to update a user")
	return nil
}

// Updated is called after a user is updated.
func (o *UserObserver) Updated(event contractsorm.Event) error {
	o.events = append(o.events, "updated")
	fmt.Println("[observer] Updated: user updated")
	return nil
}

// Deleting is called before a user is deleted.
func (o *UserObserver) Deleting(event contractsorm.Event) error {
	o.events = append(o.events, "deleting")
	fmt.Println("[observer] Deleting: about to delete a user")
	return nil
}

// Deleted is called after a user is deleted.
func (o *UserObserver) Deleted(event contractsorm.Event) error {
	o.events = append(o.events, "deleted")
	fmt.Println("[observer] Deleted: user deleted")
	return nil
}

// ForceDeleted is called after a user is permanently deleted.
func (o *UserObserver) ForceDeleted(event contractsorm.Event) error {
	o.events = append(o.events, "force_deleted")
	fmt.Println("[observer] ForceDeleted: user permanently deleted")
	return nil
}

func main() {
	if err := RunExample("sqlite://./example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates observer registration and lifecycle event callbacks.
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	// Create the users table
	err = db.Schema().Create("users", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("name")
		bp.String("email")
		bp.Timestamps()
		bp.SoftDeletes()
	})
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	observer := &UserObserver{}

	// Register the observer at the database level.
	// All subsequent db.Query() calls will have this observer active.
	db.Observe(&User{}, observer)

	// Example 1: Creating / Created events
	fmt.Println("=== Example 1: Create ===")
	user := &User{Name: "Alice", Email: "alice@example.com"}
	err = db.Query().Model(&User{}).Create(user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Printf("User created with ID: %d\n", user.ID)

	// Example 2: Updating / Updated events
	fmt.Println("\n=== Example 2: Update ===")
	user.Name = "Alice Smith"
	_, err = db.Query().Model(&User{}).Where("id = ?", user.ID).Update(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Example 3: Deleting / Deleted events (soft delete)
	fmt.Println("\n=== Example 3: Soft Delete ===")
	_, err = db.Query().Model(&User{}).Delete(user)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Example 4: ForceDeleted event
	fmt.Println("\n=== Example 4: Force Delete ===")
	_, err = db.Query().Model(&User{}).ForceDelete(user)
	if err != nil {
		return fmt.Errorf("failed to force delete user: %w", err)
	}

	// Example 5: WithoutEvents — observer is NOT called
	fmt.Println("\n=== Example 5: Create WithoutEvents (observer skipped) ===")
	silent := &User{Name: "Bob", Email: "bob@example.com"}
	err = db.Query().WithoutEvents().Model(&User{}).Create(silent)
	if err != nil {
		return fmt.Errorf("failed to create silently: %w", err)
	}
	fmt.Printf("User created quietly with ID: %d (no observer events fired)\n", silent.ID)

	fmt.Printf("\nAll observer events fired: %v\n", observer.events)
	return nil
}
