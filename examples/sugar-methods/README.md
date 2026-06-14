# Sugar Methods Compatibility Example

This example demonstrates the sugar methods for Django and Sequelize compatibility, which provide familiar API patterns for developers coming from Python (Django) and JavaScript (Sequelize) backgrounds.

## Features Demonstrated

### Django-style Methods
- `Filter()` - Alias for `Where()`, provides Django QuerySet-style filtering
- `Exclude()` - Alias for `WhereNot()`, provides Django QuerySet-style exclusion
- `All()` - Alias for `Get()`, provides Django QuerySet-style retrieval

### Sequelize-style Methods
- `FindAll()` - Alias for `All()`/`Get()`, provides Sequelize-style retrieval
- `FindOne()` - Alias for `First()`, provides Sequelize-style single record retrieval
- `Destroy()` - Alias for `Delete()`, provides Sequelize-style deletion

### Method Aliases Chain
```
Django:    Filter() -> Where()
           Exclude() -> WhereNot()
           All() -> Get()

Sequelize: FindAll() -> All() -> Get()
           FindOne() -> First()
           Destroy() -> Delete()
```

## Running the Example

```bash
cd examples/sugar-methods
go run main.go
```

## Prerequisites

- SQLite database (default) or any supported database
- Neat ORM package installed

## Example Output

```
=== Seeded 5 products ===

=== Django-style: Filter() ===
Found 3 electronics using Filter():
  - Laptop ($1200)
  - Mouse ($25)
  - Monitor ($400)

=== Django-style: Exclude() ===
Found 3 non-furniture products using Exclude():
  - Laptop (Electronics)
  - Mouse (Electronics)
  - Monitor (Electronics)

=== Django-style: All() ===
Found 5 total products using All()

=== Sequelize-style: FindAll() ===
Found 4 active products using FindAll():
  - Laptop
  - Mouse
  - Desk
  - Monitor

=== Sequelize-style: FindOne() ===
First product using FindOne(): Laptop

=== Sequelize-style: Destroy() ===
Destroyed 1 product(s) using Destroy()

=== Chaining Django + Sequelize Methods ===
Found 2 expensive active electronics:
  - Laptop ($1200)
  - Monitor ($400)
```

## Compatibility Notes

All sugar methods are **aliases** - they simply call the underlying Neat methods:

| Sugar Method | Neat Method | Origin |
|--------------|-------------|--------|
| `Filter()` | `Where()` | Django |
| `Exclude()` | `WhereNot()` | Django |
| `All()` | `Get()` | Django |
| `FindAll()` | `All()` → `Get()` | Sequelize |
| `FindOne()` | `First()` | Sequelize |
| `Destroy()` | `Delete()` | Sequelize |

This means:
- All existing Neat functionality is preserved
- No performance overhead - just method delegation
- You can mix Laravel, Django, and Sequelize styles in the same query
- All methods work with the same query builder features (ordering, limits, etc.)
