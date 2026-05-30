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

// BigSerialUser represents a user model with int64 ID for bigserial testing
type BigSerialUser struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (BigSerialUser) TableName() string {
	return "bigserial_users"
}

// Comment represents a comment model with polymorphic associations
// A comment can belong to a Post or a Video
type Comment struct {
	ID              uint      `db:"id"`
	Body            string    `db:"body"`
	CommentableID   uint      `db:"commentable_id"`
	CommentableType string    `db:"commentable_type"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func (Comment) TableName() string {
	return "comments"
}

// Post represents a post model
type Post struct {
	ID        uint       `db:"id"`
	Title     string     `db:"title"`
	Content   string     `db:"content"`
	Comments  []*Comment `db:"-"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

func (Post) TableName() string {
	return "posts"
}

// Video represents a video model
type Video struct {
	ID        uint       `db:"id"`
	Title     string     `db:"title"`
	URL       string     `db:"url"`
	Comments  []*Comment `db:"-"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

func (Video) TableName() string {
	return "videos"
}
