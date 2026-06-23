# Array Driver Example

This example demonstrates how to use the `array` driver to query static or computed data using Neat's ORM.

## Overview

The `array` driver is useful for:
- Static data like roles, statuses, or country lists.
- Mocking data for tests.
- Querying computed datasets as if they were in a real database.

## How it works

1. Define a struct that implements `contractsorm.ArraySource`.
2. Configure a connection using the `array` driver.
3. Use the query builder normally by passing the `ArraySource` to `.Model()`.

Neat will automatically:
- Infer the schema from the first row of data.
- Create an in-memory SQLite table.
- Populate the table with the provided rows.
- Execute your queries against this in-memory table.

## Running the example

```bash
go run examples/array-driver/main.go
```
