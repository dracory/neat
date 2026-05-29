# JSON Queries

This example demonstrates JSON query functionality for querying and updating JSON columns in your database.

## Features Demonstrated

- Querying JSON fields for specific values using `WhereJsonContains`
- Querying JSON arrays for values
- Checking if JSON keys exist using `WhereJsonContainsKey`
- Checking JSON array length using `WhereJsonLength`
- Array indexing in JSON queries
- Updating JSON fields with path notation
- Combining multiple JSON query conditions

## Running the Example

```bash
cd examples/json-queries
go run main.go
```

## Running Tests

```bash
cd examples/json-queries
go test -v
```

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)
- The example creates a `products` table with a JSON `attributes` column

## Database Support

JSON query functions are supported for:
- **SQLite**: Uses `json_extract`, `json_type`, and `json_array_length` functions
- **MySQL**: Uses `JSON_CONTAINS`, `JSON_CONTAINS_PATH`, and `JSON_LENGTH` functions
- **PostgreSQL**: Uses JSON functions (implementation may vary)

## Example JSON Path Syntax

- `column->field` - Access nested field
- `column->field->nested` - Access deeply nested field
- `column->array->0` - Access array element by index
