# XMLDB Driver Example

This example demonstrates how to use Neat's `xmldb` driver to point at a directory of XML files and query them as database tables.

The directory serves as the database, and each `.xml` file is represented as a table named after its filename (without the extension). Attributes and leaf sub-elements across the rows define the column schema, with type inference and widening done automatically.

## Files

- `data/users.xml`: A standard XML file containing `<user>` entries (`"users"` table).
- `data/products.xml`: A standard XML file containing `<product>` entries (`"products"` table).
- `data/orders.xml`: A standard XML file containing `<order>` entries (`"orders"` table).

## How to Run

You can run this example using the standard Go toolchain:

```bash
go run examples/xmldb-driver/main.go
```

## Running the Tests

To run the unit tests for this example:

```bash
go test -v ./examples/xmldb-driver/...
```
