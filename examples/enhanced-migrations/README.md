# Enhanced Migration System

This example demonstrates the enhanced migration system features in Neat ORM, including:

- **Multiple Migration ID Formats**: Support for datetime, date, unix timestamp, and custom formats
- **Migration Descriptions**: Human-readable descriptions for better migration tracking
- **Performance Tracking**: Track execution time with `started_at` and `completed_at` timestamps
- **Transaction Support**: Optional transactional execution of migrations
- **Duplicate Detection**: Built-in protection against duplicate migration names
- **Format Validation**: Configurable validation of migration ID formats

## Features Demonstrated

- Creating migrations with different ID formats (datetime, date, unix, custom)
- Adding descriptions to migrations for better documentation
- Tracking migration execution time and performance
- Using transactions for safer migrations
- Validating migration ID formats
- Viewing migration history with detailed metadata

## Migration ID Formats

### DateTime Format (Default)
Format: `YYYY_MM_DD_HHMM_description`
Example: `2026_06_15_1200_create_users_table`

### Date Format
Format: `YYYY_MM_DD_NNN_description` (NNN is a sequence number)
Example: `2026_06_15_001_create_users_table`

### Unix Format (Legacy)
Format: `unix_timestamp_description`
Example: `1717080000_create_users_table`

### Custom Format
Format: No prefix validation
Example: `any_custom_name_you_want`

## Running the Example

```bash
cd examples/enhanced-migrations
go run main.go
```

## Running Tests

```bash
cd examples/enhanced-migrations
go test -v
```

## Configuration

The enhanced migration system can be configured via the config:

```go
// Set migration ID format (default: "datetime")
config.Set("database.migrations.id_format", "datetime")

// Enable/disable format validation (default: true)
config.Set("database.migrations.validate_format", true)

// Enable/disable transactions (default: true)
config.Set("database.migrations.use_transactions", true)

// Set custom migration table name (default: "migrations")
config.Set("database.migrations.table", "migrations")
```

## Migration Table Schema

The enhanced migration table includes additional columns for tracking metadata:

```sql
CREATE TABLE migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    migration VARCHAR(255) NOT NULL,
    batch INTEGER NOT NULL,
    description TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)
- Neat ORM library
