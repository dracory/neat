# Advanced Queries

This example demonstrates advanced query builder features for complex database operations.

## Features Demonstrated

- Join queries with multiple tables
- OR conditions in WHERE clauses
- WhereIn for filtering by a list of values
- WhereBetween for range queries
- WhereNull and WhereNotNull for null checks
- GroupBy and Having clauses
- Multiple OrderBy clauses
- Pagination with Offset and Limit
- Aggregation functions (Count, Avg, Sum)
- Subqueries

## Running the Example

```bash
cd examples/advanced-queries
go run main.go
```

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)
- Appropriate tables with sample data (users, posts, orders)
