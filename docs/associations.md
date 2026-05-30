# Associations

This document describes the model associations (relationships) in Neat ORM.

## Supported Relationships

- BelongsTo
- HasMany
- HasOne
- PolymorphicBelongsTo
- PolymorphicHasMany

## BelongsTo

A BelongsTo relationship defines a many-to-one relationship where a model belongs to another model.

```go
type Post struct {
    ID     uint
    Title  string
    UserID uint
    User   User // BelongsTo
}

type User struct {
    ID    uint
    Name  string
    Posts []Post // HasMany
}
```

### Querying BelongsTo

```go
// Eager load the relationship
db.Query().With("user").Where("id", 1).First(&post)

// Lazy load the relationship
db.Query().Load(&post, "user")
```

## HasMany

A HasMany relationship defines a one-to-many relationship where a model has many related models.

```go
type User struct {
    ID    uint
    Name  string
    Posts []Post // HasMany
}

type Post struct {
    ID     uint
    Title  string
    UserID uint
    User   User // BelongsTo
}
```

### Querying HasMany

```go
// Eager load the relationship
db.Query().With("posts").Where("id", 1).First(&user)

// Lazy load the relationship
db.Query().Load(&user, "posts")
```

## HasOne

A HasOne relationship defines a one-to-one relationship where a model has one related model.

```go
type User struct {
    ID      uint
    Name    string
    Profile Profile // HasOne
}

type Profile struct {
    ID     uint
    UserID uint
    Bio    string
    User   User // BelongsTo
}
```

### Querying HasOne

```go
// Eager load the relationship
db.Query().With("profile").Where("id", 1).First(&user)

// Lazy load the relationship
db.Query().Load(&user, "profile")
```

## Eager Loading

Load relationships when querying the parent model:

```go
// Load a single relationship
db.Query().With("posts").Where("id", 1).First(&user)

// Load multiple relationships
db.Query().With("posts").With("profile").Where("id", 1).First(&user)

// Nested relationships
db.Query().With("posts.comments").Where("id", 1).First(&user)
```

## Lazy Loading

Load relationships on-demand after querying:

```go
// Load a relationship
db.Query().Load(&user, "posts")

// Load only if not already loaded
db.Query().LoadMissing(&user, "posts")
```

## Association Operations

### Append

Add a related model:

```go
db.Query().Association("posts").Append(&user, &post)
```

### Replace

Replace all related models:

```go
db.Query().Association("posts").Replace(&user, posts)
```

### Delete

Remove a related model:

```go
db.Query().Association("posts").Delete(&user, &post)
```

### Clear

Remove all related models:

```go
db.Query().Association("posts").Clear(&user)
```

### Count

Count related models:

```go
count, err := db.Query().Association("posts").Count(&user)
```

## PolymorphicBelongsTo

A PolymorphicBelongsTo relationship allows a model to belong to multiple different model types. This is useful when a single model can be associated with various parent models.

```go
type Comment struct {
    ID              uint
    Body            string
    CommentableID   uint   `db:"commentable_id"`
    CommentableType string `db:"commentable_type"`
}

type Post struct {
    ID       uint
    Title    string
    Content  string
    Comments []*Comment
}

type Video struct {
    ID       uint
    Title    string
    URL      string
    Comments []*Comment
}
```

In this example, a `Comment` can belong to either a `Post` or a `Video`. The `CommentableID` stores the ID of the parent model, and `CommentableType` stores the type name (e.g., "Post" or "Video").

### Querying PolymorphicBelongsTo

```go
// Associate a comment with a post
comment := Comment{Body: "Great post!"}
db.Query().Association("Commentable").Append(&comment, &post)

// Find the parent model
var parent Post
db.Query().Association("Commentable").Find(&parent, &comment)
```

## PolymorphicHasMany

A PolymorphicHasMany relationship allows a model to have many related models that can belong to multiple different model types.

```go
type Post struct {
    ID       uint
    Title    string
    Content  string
    Comments []*Comment
}

type Video struct {
    ID       uint
    Title    string
    URL      string
    Comments []*Comment
}

type Comment struct {
    ID              uint
    Body            string
    CommentableID   uint   `db:"commentable_id"`
    CommentableType string `db:"commentable_type"`
}
```

In this example, both `Post` and `Video` can have many `Comments`. The polymorphic fields on the `Comment` model track which parent model each comment belongs to.

### Querying PolymorphicHasMany

```go
// Append comments to a post
comment1 := Comment{Body: "First comment"}
comment2 := Comment{Body: "Second comment"}
db.Query().Association("Comments").Append(&post, &comment1, &comment2)

// Find all comments for a post
var comments []Comment
db.Query().Association("Comments").Find(&comments, &post)

// Count comments
count := db.Query().Association("Comments").Count(&post)
```

## Note

This documentation is a placeholder and will be expanded as the association system is fully implemented.
