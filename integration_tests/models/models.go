package models

import (
	"time"

	"github.com/dracory/neat/database/soft_delete"
)

// User represents a user model with associations
type User struct {
	soft_delete.SoftDeletes
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	Avatar    string    `db:"avatar"`
	Bio       *string   `db:"bio"`
	Votes     int       `db:"votes"`
	Address   *Address  `db:"-"`
	Books     []*Book   `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Address represents an address model
type Address struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	UserID    uint      `db:"user_id"`
	User      *User     `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (Address) TableName() string {
	return "addresses"
}

// Book represents a book model
type Book struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	UserID    uint      `db:"user_id"`
	User      *User     `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// People represents a people model
type People struct {
	ID        uint      `db:"id"`
	Body      string    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// JsonData represents a JSON data model
type JsonData struct {
	ID        uint      `db:"id"`
	Data      string    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
