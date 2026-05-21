# Soft Deletes

This document describes the soft delete functionality in Neat ORM.

## What are Soft Deletes?

Soft deletes allow you to mark records as deleted without actually removing them from the database. This is useful for:
- Data recovery
- Audit trails
- Compliance requirements

## Adding Soft Delete Capability

Embed the SoftDeletes struct in your model:

```go
type User struct {
    neat.SoftDeletes
    ID   uint
    Name string
}
```

This adds a `deleted_at` timestamp column to your model.

## Soft Deleting Records

When you delete a record with soft deletes enabled, it sets the `deleted_at` timestamp instead of removing the record:

```go
db.Query().Where("id", 1).Delete(&user)
```

## Querying with Soft Deletes

### Default Behavior (Excludes Soft-Deleted Records)

By default, queries exclude soft-deleted records:

```go
var users []User
db.Query().Get(&users) // Excludes soft-deleted records
```

### Including Soft-Deleted Records

To include soft-deleted records in queries:

```go
var users []User
db.Query().WithTrashed().Get(&users)
```

### Querying Only Soft-Deleted Records

To query only soft-deleted records:

```go
var users []User
db.Query().OnlyTrashed().Get(&users)
```

## Restoring Soft-Deleted Records

To restore a soft-deleted record:

```go
db.Query().Restore(&user)
```

## Force Deleting (Permanent Deletion)

To permanently delete a record (bypass soft delete):

```go
db.Query().ForceDelete(&user)
```

## Checking if a Record is Soft-Deleted

```go
if user.IsDeleted() {
    fmt.Println("User is soft-deleted")
}
```

## Getting the Deleted At Timestamp

```go
deletedAt := user.GetDeletedAt()
```

## Customizing the Soft Delete Column

You can customize the soft delete column name:

```go
type User struct {
    neat.SoftDeletes
    ID   uint
    Name string
}

func (u *User) DeletedAtColumn() string {
    return "archived_at"
}
```

## Note

This documentation is a placeholder and will be expanded as the soft delete system is fully implemented.
