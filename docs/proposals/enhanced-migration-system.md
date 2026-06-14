# Enhanced Migration System

**Date**: June 14, 2026
**Status**: Open for Discussion
**Priority**: Medium
**Author**: Neat ORM Team

## Overview

This proposal aims to enhance the current migration system in Neat ORM by adding advanced features inspired by standalone migration packages and modern best practices. The current system provides solid basic migration functionality but could benefit from several developer experience improvements, enhanced debugging capabilities, and additional safety features that would bring it on par with standalone migration solutions.

## Background

### The Standalone Migrate Package

The [`github.com/dracory/migrate`](https://github.com/dracory/migrate) package is a lightweight, database-agnostic migration framework for Go applications by the same organization. It demonstrates modern migration best practices and provides:

**Key Features**:
- Lexicographical migration ordering by ID
- Transactional execution (each migration in a transaction)
- Rollback support
- Customizable logging with slog
- Duplicate detection
- Automatic builtin migrations (creates tracking table automatically)
- Performance tracking (started_at, completed_at columns)
- Batch grouping
- Custom table names (configurable, default: `migration_tracker`)
- Migration ID formats:
  - `YYYY_MM_DD_HHMM_description` (default, e.g., `2026_03_21_1200_create_users_table`)
  - `YYYY_MM_DD_NNN_description` (sequence-based, e.g., `2026_03_21_001_create_users_table`)
  - `none` (no prefix validation)
- Context support for cancellation
- Comprehensive validation functions
- Environment variable support (`MIGRATE_TABLE_NAME`)

**Migration Interface**:
```go
type MigrationInterface interface {
    ID() string                                // Migration ID
    Description() string                        // Human-readable description
    Up(ctx context.Context, tx *sql.Tx) error   // Apply migration
    Down(ctx context.Context, tx *sql.Tx) error // Rollback migration
}
```

**Migrator Interface**:
```go
type MigratorInterface interface {
    AddMigration(migration MigrationInterface) error
    AddMigrations(migrations []MigrationInterface) error
    Up(ctx context.Context) error
    Down(ctx context.Context) error
    Status(ctx context.Context) error
    GetStatus(ctx context.Context) ([]MigrationStatus, error)
    GetHistory(ctx context.Context) ([]MigrationRecord, error)
}
```

**Constants and Types**:
```go
type NamingFormat string

const (
    NamingFormatPrefixYYYY_MM_DD_HHMM NamingFormat = "YYYY_MM_DD_HHMM"
    NamingFormatPrefixYYYY_MM_DD_NNN  NamingFormat = "YYYY_MM_DD_NNN"
    NamingFormatPrefixNone            NamingFormat = "none"
)
```

**Why Not Use It Directly?**
While the standalone package is excellent, Neat ORM's migration system has unique requirements:
1. **Deep Schema Builder Integration**: Neat's migrations use the schema DSL, not raw SQL
2. **Type Safety**: Schema builder provides type-safe migrations
3. **Existing Ecosystem**: Large user base with existing migrations
4. **Global Registry Pattern**: Neat uses a registration pattern that works without files
5. **Zero External Dependencies**: Neat aims to be self-contained

**This Proposal**: Takes inspiration from the standalone package's best practices while adapting them to Neat's architecture and maintaining compatibility with existing patterns.

### Current Neat Architecture

Neat ORM's migration system is built on a solid foundation with the following architecture:

**Core Components**:
1. **Migrator** (`database/migration/migrator.go`) - Main orchestrator for migration operations
2. **Repository** (`database/migration/repository.go`) - Handles persistence of migration records
3. **Global Registry** - Thread-safe registration system for migrations
4. **Schema Integration** - Deep integration with Neat's schema builder

**Existing Strengths**:
- Database-agnostic design (SQLite, MySQL, PostgreSQL, SQL Server, Oracle, Turso)
- Global migration registry with thread-safe access
- Works with or without migration files on disk
- SQL injection protection in table names
- Batch-based migration tracking
- Step-based and batch-based rollback
- Integration with Neat's schema builder for type-safe migrations

**Design Philosophy**:
The current system prioritizes simplicity, type safety, and deep integration with Neat's ecosystem over standalone flexibility. This proposal aims to enhance these strengths while adding commonly requested features.

### Current Migration System Features

**Migration Management**:
```go
type Migrator interface {
    Create(name string) error                  // Create new migration file
    Run() error                                 // Run pending migrations
    Rollback(step, batch int) error            // Rollback migrations (step or batch)
    Fresh() error                               // Drop all tables and re-run migrations
    Reset() error                               // Rollback all and re-run
    Status() ([]Status, error)                 // Get migration status
}
```

**Current Migration Table Schema**:
```sql
CREATE TABLE migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,      -- SQLite syntax (adapts per database)
    migration VARCHAR(255) NOT NULL,           -- Migration file name (e.g., "1717080000_create_users_table")
    batch INTEGER NOT NULL,                    -- Batch number for grouping
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP  -- When migration was executed
)
```

**Migration File Structure**:
```go
package migrations

import (
    contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// Up applies the migration
func Up(schema contractsschema.Schema) error {
    return nil
}

// Down rolls back the migration
func Down(schema contractsschema.Schema) error {
    return nil
}
```

**Registration System**:
```go
// Thread-safe global registry
migration.RegisterMigration("1717080000_create_users_table", migration.Migration{
    Up:   Up,
    Down: Down,
})
```

**Existing Capabilities**:
- Unix timestamp-based naming convention (e.g., `1717080000_create_users_table.go`)
- Configurable migration table name via config
- Protection against SQL injection in table names
- Automatic migration table creation
- Batch tracking for grouped migrations
- Flexible rollback (by step or batch)
- Status reporting with batch information
- Works with or without physical migration files
- Thread-safe global registry
- Database-specific SQL generation

**Current Limitations**:
1. **Migration ID Format**: Unix timestamps are not human-readable
2. **No Performance Metrics**: Cannot track execution time per migration
3. **Limited Logging**: No structured logging support (no slog integration)
4. **No Descriptions**: Migrations lack descriptive metadata
5. **No Validation**: Migration ID format is not validated
6. **Limited History**: Cannot view detailed execution history with timing
7. **No Transaction Wrapping**: Migrations don't run in explicit transactions
8. **No Duplicate Detection**: No built-in protection against duplicate migration names
9. **Manual Format Enforcement**: No tooling to enforce naming conventions

## Alignment with Standalone Migrate Package

This proposal aligns Neat's migration system with the proven patterns from [`github.com/dracory/migrate`](https://github.com/dracory/migrate):

| Feature | Standalone Package | Current Neat | Proposed Neat | Status |
|---------|-------------------|--------------|---------------|---------|
| **ID Format** | `YYYY_MM_DD_HHMM_description` (default) | `unix_timestamp` | `YYYY_MM_DD_HHMM_description` (default) | ✅ Aligned |
| **Alt Format** | `YYYY_MM_DD_NNN_description` | Not supported | `YYYY_MM_DD_NNN_description` (date format) | ✅ Aligned |
| **No Validation** | `none` format | Not supported | `custom` format | ✅ Similar |
| **Description** | Required method | Not supported | Optional (comment or struct) | ✅ Added |
| **Performance Tracking** | `started_at`, `completed_at` columns | Only `created_at` | Added `started_at`, `completed_at` | ✅ Aligned |
| **Transactions** | Automatic per migration | No explicit wrapping | Optional per migration | ✅ Added |
| **Logging** | slog support | None | Neat's log interface | ✅ Added |
| **Duplicate Detection** | Built-in | Silent overwrite | Added with error | ✅ Added |
| **Batch Tracking** | Yes (string column) | Yes (integer column) | Yes (integer column) | ✅ Same concept |
| **Context Support** | Yes (`ctx context.Context`) | No | Not in Phase 1 | 🔄 Future |
| **Transaction Parameter** | Yes (`tx *sql.Tx`) | Uses schema with ORM | Uses schema with ORM | ❌ Different by design |
| **Validation Functions** | `ValidateMigrationID`, `ValidateTableName` | Basic table name only | Add `ValidateMigrationID` | ✅ Added |
| **Interface Design** | 4 methods (ID, Description, Up, Down) | 2 functions (Up, Down) + registration | Keep functions + add description | ✅ Adapted |
| **GetStatus/GetHistory** | Separate methods | Single `Status()` | Add `GetHistory()` | ✅ Added |
| **Builtin Migrations** | Automatic creation of tracking table | Manual schema creation | Auto-upgrade on first run | ✅ Similar |
| **Table Name** | `migration_tracker` (default) | `migrations` (default) | `migrations` (keep current) | ❌ Different default |
| **Env Variables** | `MIGRATE_TABLE_NAME` | Config-based | Config-based | ❌ Different approach |

### Key Differences (By Design)

1. **Schema Builder vs Raw SQL**: 
   - Standalone: `Up(ctx, tx)` receives `*sql.Tx` for raw SQL
   - Neat: `Up(schema)` receives schema builder for type-safe operations
   - **Why**: Neat's schema builder provides type safety and database abstraction

2. **Function-Based vs Interface-Based**:
   - Standalone: Struct with 4 methods
   - Neat: Two functions + registration
   - **Why**: Simpler for users, matches Neat's patterns

3. **Context Parameter**:
   - Standalone: Context for cancellation
   - Neat: Not yet implemented
   - **Why**: Future enhancement, Phase 1 focuses on core features

### Shared Philosophy

Both systems share core principles:
- **Immutable IDs**: Never change migration IDs after application
- **Lexicographical Ordering**: Migrations run in ID order
- **Transactional Safety**: Changes are atomic
- **Performance Monitoring**: Track execution time
- **Rollback Support**: Can undo migrations
- **Batch Grouping**: Group related migrations

## Proposed Enhancements

### 1. Structured Migration ID Formats

**Current Behavior**: Unix timestamp format (e.g., `1717080000_description`)

**Problem**: 
- Timestamps like `1717080000` are not human-readable
- Hard to determine when a migration was created
- Difficult to organize migrations chronologically by visual inspection

**Solution**: Support multiple ID format options with validation

**Proposed API**:
```go
// In config/database.go or similar
type MigrationConfig struct {
    Table    string            `mapstructure:"table"`
    IDFormat MigrationIDFormat `mapstructure:"id_format"`
}

type MigrationIDFormat string

const (
    MigrationIDFormatDateTime MigrationIDFormat = "datetime"   // 2026_06_14_1200_description (default)
    MigrationIDFormatDate     MigrationIDFormat = "date"       // 2026_06_14_001_description (sequence)
    MigrationIDFormatUnix     MigrationIDFormat = "unix"       // 1717080000_description (legacy)
    MigrationIDFormatCustom   MigrationIDFormat = "custom"     // User-defined format
)

// Configuration
config := map[string]any{
    "database": map[string]any{
        "migrations": map[string]any{
            "table":     "migrations",
            "id_format": "datetime",  // or "unix", "date", "custom"
        },
    },
}
```

**Implementation in Create Method**:
```go
func (m *Migrator) Create(name string) error {
    format := m.config.GetString("database.migrations.id_format", "datetime")
    
    var prefix string
    switch MigrationIDFormat(format) {
    case MigrationIDFormatDateTime:
        prefix = time.Now().Format("2006_01_02_1504") // YYYY_MM_DD_HHMM
    case MigrationIDFormatDate:
        prefix = m.generateDateSequence() // 2026_06_14_001
    case MigrationIDFormatUnix:
        prefix = fmt.Sprintf("%d", time.Now().Unix())
    case MigrationIDFormatCustom:
        fallthrough
    default:
        prefix = time.Now().Format("2006_01_02_1504") // Default to datetime
    }
    
    migrationName := prefix + "_" + strings.ReplaceAll(name, " ", "_")
    // ... create file
}
```

**Benefits**:
- Human-readable migration IDs by default
- Chronological ordering preserved
- Better organization and debugging
- Easier code reviews
- Can still use unix format if preferred

**Examples**:
```
DateTime: 2026_06_14_1200_create_users_table (default)
Date:     2026_06_14_001_create_users_table
          2026_06_14_002_add_email_to_users
Unix:     1717080000_create_users_table (legacy, opt-in)
```

### 2. Enhanced Migration Tracking Table

**Current Schema**:
```sql
CREATE TABLE migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    migration VARCHAR(255) NOT NULL,
    batch INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```

**Problem**: 
- No performance metrics (can't identify slow migrations)
- No description field (hard to understand purpose without reading code)
- No execution timing data

**Proposed Enhanced Schema**:
```sql
CREATE TABLE migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    migration VARCHAR(255) NOT NULL,
    batch INTEGER NOT NULL,
    description TEXT,                           -- New: Migration description
    started_at DATETIME,                        -- New: Migration start time
    completed_at DATETIME,                      -- New: Migration completion time
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```

**Migration Path for Existing Users**:
```go
// Auto-upgrade migration table on first run with new version
func (r *Repository) CreateRepository() error {
    if r.RepositoryExists() {
        // Check if new columns exist, add them if not
        return r.upgradeRepositorySchema()
    }
    // Create with new schema...
}

func (r *Repository) upgradeRepositorySchema() error {
    // Check if description column exists
    // If not, add nullable columns for backward compatibility
    alterSQL := fmt.Sprintf(`
        ALTER TABLE %s ADD COLUMN description TEXT;
        ALTER TABLE %s ADD COLUMN started_at DATETIME;
        ALTER TABLE %s ADD COLUMN completed_at DATETIME;
    `, r.table, r.table, r.table)
    
    // Execute with error handling for already-existing columns
    // ...
}
```

**Updated Repository Log Method**:
```go
func (r *Repository) Log(migrationName string, batch int, description string, startedAt, completedAt time.Time) error {
    insertSQL := fmt.Sprintf(
        "INSERT INTO %s (migration, batch, description, started_at, completed_at, created_at) VALUES (?, ?, ?, ?, ?, ?)",
        r.table,
    )
    _, err := r.orm.Query().Exec(insertSQL, migrationName, batch, description, startedAt, completedAt, time.Now())
    return err
}
```

**Benefits**:
- Performance monitoring per migration
- Better debugging for slow migrations
- Documentation within the database
- Audit trail with precise timing
- Backward compatible with nullable columns

### 3. Migration Descriptions

**Current Behavior**: No way to add descriptive metadata to migrations

**Problem**: 
- Must read migration code to understand purpose
- Status output only shows file names
- Difficult to understand migration history

**Solution**: Add optional description support via comments or dedicated method

**Approach A: Comment-Based (Simpler, No Interface Changes)**:
```go
package migrations

import contractsschema "github.com/dracory/neat/contracts/database/schema"

// Description: Create users table with email and timestamps
// This migration adds the initial users table structure

// Up applies the migration
func Up(schema contractsschema.Schema) error {
    schema.Create("users", func(table schema.Blueprint) {
        table.ID()
        table.String("email").Unique()
        table.Timestamps()
    })
    return nil
}

// Down rolls back the migration
func Down(schema contractsschema.Schema) error {
    schema.DropIfExists("users")
    return nil
}
```

**Approach B: Interface-Based (More Structured, Requires Changes)**:
```go
// Update Migration struct
type Migration struct {
    Description string // New field
    Up          func(schema contractsschema.Schema) error
    Down        func(schema contractsschema.Schema) error
}

// Registration with description
migration.RegisterMigration("2026_06_14_120000_create_users_table", migration.Migration{
    Description: "Create users table with email and timestamps",
    Up:          Up,
    Down:        Down,
})
```

**Usage in Status Command**:
```go
status, err := migrator.Status()
for _, s := range status {
    fmt.Printf("%s: %s (Batch: %d, Ran: %v)\n", 
        s.Name, s.Description, s.Batch, s.Ran)
}

// Output:
// 2026_06_14_120000_create_users_table: Create users table with email and timestamps (Batch: 1, Ran: true)
```

**Benefits**:
- Self-documenting migrations
- Better status and history output
- Easier code reviews
- Improved team communication
- Optional (backward compatible)

**Recommendation**: Start with Approach A (comment-based) for simplicity, consider Approach B if structured access is needed

### 4. Performance Tracking and Execution History

**Current Behavior**: Only tracks when migration was logged (created_at), not execution time

**Problem**: 
- Cannot identify slow migrations
- No metrics for optimization
- Missing execution audit trail

**Solution**: Track start/end times and provide history methods

**Implementation in Migrator**:
```go
func (m *Migrator) Run() error {
    // ... existing setup code ...
    
    for _, name := range names {
        if contains(ran, name) {
            continue
        }
        
        migrationFunc := migrationFiles[name]
        
        // Track performance
        startedAt := time.Now()
        
        // Run Up method
        if err := migrationFunc.Up(m.schema); err != nil {
            return fmt.Errorf("failed to run migration %s: %w", name, err)
        }
        
        completedAt := time.Now()
        duration := completedAt.Sub(startedAt)
        
        // Log with timing data
        description := m.extractDescription(name) // Parse from comments or registration
        if err := m.repository.Log(name, batch, description, startedAt, completedAt); err != nil {
            return fmt.Errorf("failed to log migration: %w", err)
        }
        
        // Optional: Log to application logger
        if logger := m.getLogger(); logger != nil {
            logger.Infof("Migration completed: %s (Duration: %v)", name, duration)
        }
    }
    
    return nil
}
```

**New Contract Methods**:
```go
// In contracts/migration/repository.go
type File struct {
    ID          uint
    Migration   string
    Batch       int
    Description string    // New
    StartedAt   time.Time // New
    CompletedAt time.Time // New
}

type Repository interface {
    // Existing methods...
    Log(file string, batch int, description string, startedAt, completedAt time.Time) error // Updated signature
    
    // New methods
    GetHistory() ([]File, error)                    // Get full history with timing
    GetMigrationTiming(name string) (*File, error)  // Get specific migration timing
}

// In contracts/migration/migrator.go
type Status struct {
    Name        string
    Batch       int
    Ran         bool
    Description string        // New
    Duration    time.Duration // New
    LastRun     time.Time     // New
}
```

**Usage Examples**:
```go
// Get detailed history
history, err := migrator.Repository().GetHistory()
for _, record := range history {
    duration := record.CompletedAt.Sub(record.StartedAt)
    fmt.Printf("%s - %s - Duration: %v\n", 
        record.Migration, record.Description, duration)
}

// Get enhanced status
status, err := migrator.Status()
for _, s := range status {
    if s.Ran {
        fmt.Printf("%s: %s (Batch: %d, Duration: %v, Last Run: %s)\n",
            s.Name, s.Description, s.Batch, s.Duration, s.LastRun.Format(time.RFC3339))
    }
}
```

**Benefits**:
- Identify performance bottlenecks
- Monitor migration execution trends
- Debug slow migrations in production
- Audit trail for compliance
- Optimization insights

### 5. Structured Logging with slog Support

**Current Behavior**: No structured logging in migration system

**Problem**:
- Cannot track migration execution in application logs
- No visibility into migration progress
- Difficult to debug in production
- No integration with application logging

**Solution**: Add optional slog integration for structured logging

**Configuration**:
```go
// In config
config := map[string]any{
    "database": map[string]any{
        "migrations": map[string]any{
            "table":   "migrations",
            "logging": map[string]any{
                "enabled": true,
                "level":   "info", // debug, info, warn, error
            },
        },
    },
}
```

**Implementation**:
```go
// In migrator.go
type Migrator struct {
    config     config.Config
    orm        contractsorm.Orm
    repository *Repository
    schema     *schema.Schema
    paths      []string
    logger     log.Log // Neat's log interface
}

func (m *Migrator) Run() error {
    if err := m.repository.CreateRepository(); err != nil {
        return fmt.Errorf("failed to create migration repository: %w", err)
    }
    
    m.log("info", "Starting migrations", map[string]any{})
    
    // ... existing code ...
    
    for _, name := range names {
        if contains(ran, name) {
            continue
        }
        
        migrationFunc := migrationFiles[name]
        
        m.log("info", "Running migration", map[string]any{
            "migration": name,
            "batch":     batch,
        })
        
        startedAt := time.Now()
        
        if err := migrationFunc.Up(m.schema); err != nil {
            m.log("error", "Migration failed", map[string]any{
                "migration": name,
                "error":     err.Error(),
            })
            return fmt.Errorf("failed to run migration %s: %w", name, err)
        }
        
        completedAt := time.Now()
        duration := completedAt.Sub(startedAt)
        
        m.log("info", "Migration completed", map[string]any{
            "migration": name,
            "duration":  duration.String(),
            "batch":     batch,
        })
        
        // ... log to repository ...
    }
    
    m.log("info", "All migrations completed", map[string]any{
        "batch": batch,
    })
    
    return nil
}

func (m *Migrator) log(level, message string, data map[string]any) {
    if !m.config.GetBool("database.migrations.logging.enabled", false) {
        return
    }
    
    configLevel := m.config.GetString("database.migrations.logging.level", "info")
    
    // Use Neat's existing log.Log interface
    if m.logger == nil {
        return
    }
    
    // Format message with data
    msg := message
    if len(data) > 0 {
        msg = fmt.Sprintf("%s %v", message, data)
    }
    
    switch level {
    case "debug":
        if configLevel == "debug" {
            m.logger.Debug(msg)
        }
    case "info":
        if configLevel == "debug" || configLevel == "info" {
            m.logger.Info(msg)
        }
    case "warn":
        m.logger.Warning(msg)
    case "error":
        m.logger.Error(msg)
    }
}
```

**Log Output Examples**:
```
[INFO] Starting migrations
[INFO] Running migration {"migration":"2026_06_14_1200_create_users_table","batch":1}
[INFO] Migration completed {"migration":"2026_06_14_1200_create_users_table","duration":"45ms","batch":1}
[INFO] Running migration {"migration":"2026_06_14_1201_add_indexes","batch":1}
[INFO] Migration completed {"migration":"2026_06_14_1201_add_indexes","duration":"120ms","batch":1}
[INFO] All migrations completed {"batch":1}
```

**Benefits**:
- Production visibility
- Integration with existing logging infrastructure
- Debugging support
- Audit trail in logs
- Performance monitoring
- Optional (disabled by default for backward compatibility)

### 6. Migration ID Validation

**Current Behavior**: No validation of migration ID formats

**Problem**: 
- Inconsistent naming across team
- No enforcement of conventions
- Potential for sorting issues
- Hard to catch errors early

**Solution**: Add comprehensive validation matching the standalone package

**Implementation** (inspired by `github.com/dracory/migrate`):
```go
// Validation function
func ValidateMigrationID(id string, format MigrationIDFormat) error {
    if len(id) > 255 {
        return fmt.Errorf("migration ID too long (max 255 characters)")
    }
    
    if len(id) == 0 {
        return fmt.Errorf("migration ID cannot be empty")
    }
    
    // For "custom" format, only validate length and non-empty
    if format == MigrationIDFormatCustom {
        return nil
    }
    
    parts := strings.Split(id, "_")
    
    // Minimum 5 parts: YYYY, MM, DD, [HHMM|NNN], description
    if len(parts) < 5 {
        return fmt.Errorf("migration ID must have at least 5 parts separated by underscores")
    }
    
    switch format {
    case MigrationIDFormatDateTime:
        return validateDateTimeFormat(id, parts)
    case MigrationIDFormatDate:
        return validateDateFormat(id, parts)
    case MigrationIDFormatUnix:
        return validateUnixFormat(id, parts)
    default:
        return fmt.Errorf("unknown migration ID format: %s", format)
    }
}

func validateDatePart(parts []string) error {
    if len(parts[0]) != 4 || len(parts[1]) != 2 || len(parts[2]) != 2 {
        return fmt.Errorf("date parts must be YYYY_MM_DD format")
    }
    
    year, err := strconv.Atoi(parts[0])
    if err != nil {
        return fmt.Errorf("year must be numeric: %w", err)
    }
    
    month, err := strconv.Atoi(parts[1])
    if err != nil {
        return fmt.Errorf("month must be numeric: %w", err)
    }
    if month < 1 || month > 12 {
        return fmt.Errorf("month must be between 01 and 12, got %02d", month)
    }
    
    day, err := strconv.Atoi(parts[2])
    if err != nil {
        return fmt.Errorf("day must be numeric: %w", err)
    }
    if day < 1 || day > 31 {
        return fmt.Errorf("day must be between 01 and 31, got %02d", day)
    }
    
    // Validate actual calendar date
    dateStr := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])
    _, err = time.Parse("2006-01-02", dateStr)
    if err != nil {
        return fmt.Errorf("invalid calendar date: %w", err)
    }
    
    return nil
}

func validateTimePart(part string) error {
    if len(part) != 4 {
        return fmt.Errorf("time part must be 4 digits (HHMM)")
    }
    
    num, err := strconv.Atoi(part)
    if err != nil {
        return fmt.Errorf("time part must be numeric: %w", err)
    }
    
    hour := num / 100
    minute := num % 100
    
    if hour < 0 || hour > 23 {
        return fmt.Errorf("hour must be between 00 and 23, got %02d", hour)
    }
    if minute < 0 || minute > 59 {
        return fmt.Errorf("minute must be between 00 and 59, got %02d", minute)
    }
    
    return nil
}

func validateSequencePart(part string) error {
    if len(part) != 3 {
        return fmt.Errorf("sequence part must be 3 digits (NNN)")
    }
    
    sequence, err := strconv.Atoi(part)
    if err != nil {
        return fmt.Errorf("sequence part must be numeric: %w", err)
    }
    
    if sequence < 0 || sequence > 999 {
        return fmt.Errorf("sequence must be between 000 and 999, got %03d", sequence)
    }
    
    return nil
}

func validateDescription(parts []string) error {
    description := strings.Join(parts[4:], "_")
    if len(description) == 0 {
        return fmt.Errorf("description cannot be empty")
    }
    if len(description) > 200 {
        return fmt.Errorf("description too long (max 200 characters)")
    }
    return nil
}

func validateDateTimeFormat(id string, parts []string) error {
    // Format: 2026_06_14_1200_description
    if err := validateDatePart(parts); err != nil {
        return err
    }
    if err := validateTimePart(parts[3]); err != nil {
        return err
    }
    return validateDescription(parts)
}

func validateDateFormat(id string, parts []string) error {
    // Format: 2026_06_14_001_description
    if err := validateDatePart(parts); err != nil {
        return err
    }
    if err := validateSequencePart(parts[3]); err != nil {
        return err
    }
    return validateDescription(parts)
}

func validateUnixFormat(id string, parts []string) error {
    // Format: 1717080000_description
    if len(parts) < 2 {
        return fmt.Errorf("unix format must be: timestamp_description")
    }
    
    timestamp := parts[0]
    if _, err := strconv.ParseInt(timestamp, 10, 64); err != nil {
        return fmt.Errorf("invalid unix timestamp: %s", timestamp)
    }
    
    description := strings.Join(parts[1:], "_")
    if description == "" {
        return fmt.Errorf("description cannot be empty")
    }
    
    return nil
}

// ValidateTableName ensures the table name contains only safe characters
// (matching standalone package implementation)
func ValidateTableName(name string) error {
    if len(name) == 0 {
        return fmt.Errorf("table name cannot be empty")
    }
    if len(name) > 64 {
        return fmt.Errorf("table name too long (max 64 characters)")
    }
    
    // First character must be a letter or underscore
    firstRune := rune(name[0])
    if !unicode.IsLetter(firstRune) && firstRune != '_' {
        return fmt.Errorf("table name must start with a letter or underscore")
    }
    
    // All characters must be alphanumeric or underscore
    for _, r := range name {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
            return fmt.Errorf("table name contains invalid characters (only alphanumeric and underscore allowed)")
        }
    }
    
    return nil
}
```

**Usage in Create Method**:
```go
func (m *Migrator) Create(name string) error {
    // ... generate migration name ...
    
    // Validate format if enabled
    if m.config.GetBool("database.migrations.validate_format", true) {
        format := MigrationIDFormat(m.config.GetString("database.migrations.id_format", "unix"))
        if err := ValidateMigrationID(migrationName, format); err != nil {
            return fmt.Errorf("invalid migration ID: %w", err)
        }
    }
    
    // ... create file ...
}
```

**Benefits**:
- Consistent naming across team
- Early error detection
- Enforced conventions
- Better organization
- Optional (can be disabled)

### 7. Duplicate Detection

**Current Behavior**: No protection against duplicate migration names in registry

**Problem**:
- Duplicate registrations silently overwrite
- Hard to debug registration issues
- Team coordination problems

**Solution**: Add duplicate detection with clear error messages

**Implementation**:
```go
// In migrator.go - Update RegisterMigration
func RegisterMigration(name string, migration Migration) error {
    registryMutex.Lock()
    defer registryMutex.Unlock()
    
    if _, exists := migrationRegistry[name]; exists {
        return fmt.Errorf("migration '%s' is already registered", name)
    }
    
    migrationRegistry[name] = migration
    return nil
}

// For backward compatibility, keep panic version
func MustRegisterMigration(name string, migration Migration) {
    if err := RegisterMigration(name, migration); err != nil {
        panic(err)
    }
}

// In migration files, update usage:
func init() {
    migration.MustRegisterMigration("2026_06_14_120000_create_users_table", migration.Migration{
        Up:   Up,
        Down: Down,
    })
}
```

**Benefits**:
- Early detection of conflicts
- Clear error messages
- Better team collaboration
- Easier debugging

### 8. Explicit Transaction Support

**Current Behavior**: Individual operations may be auto-committed depending on database driver

**Problem**:
- No guaranteed transaction boundaries
- Partial migrations can leave database in inconsistent state
- Harder to rollback on errors

**Solution**: Wrap each migration in an explicit transaction

**Implementation**:
```go
func (m *Migrator) Run() error {
    // ... existing setup code ...
    
    for _, name := range names {
        if contains(ran, name) {
            continue
        }
        
        migrationFunc := migrationFiles[name]
        
        // Start transaction
        tx := m.orm.Begin()
        if tx.Error() != nil {
            return fmt.Errorf("failed to start transaction for migration %s: %w", name, tx.Error())
        }
        
        // Create schema with transaction
        txSchema := schema.NewSchema(m.config, tx)
        
        startedAt := time.Now()
        
        // Run Up method within transaction
        if err := migrationFunc.Up(txSchema); err != nil {
            tx.Rollback()
            m.log("error", "Migration failed, rolled back", map[string]any{
                "migration": name,
                "error":     err.Error(),
            })
            return fmt.Errorf("failed to run migration %s: %w", name, err)
        }
        
        completedAt := time.Now()
        
        // Log migration within same transaction
        // ... repository.Log call ...
        
        // Commit transaction
        if err := tx.Commit().Error(); err != nil {
            return fmt.Errorf("failed to commit transaction for migration %s: %w", name, err)
        }
        
        m.log("info", "Migration committed", map[string]any{
            "migration": name,
            "duration":  completedAt.Sub(startedAt).String(),
        })
    }
    
    return nil
}
```

**Rollback with Transactions**:
```go
func (m *Migrator) Rollback(step, batch int) error {
    // ... get files to rollback ...
    
    for i := len(files) - 1; i >= 0; i-- {
        file := files[i]
        
        // Start transaction for rollback
        tx := m.orm.Begin()
        if tx.Error() != nil {
            return fmt.Errorf("failed to start transaction for rollback %s: %w", file.Migration, tx.Error())
        }
        
        txSchema := schema.NewSchema(m.config, tx)
        
        // Run Down method
        if err := migrationFunc.Down(txSchema); err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to rollback migration %s: %w", file.Migration, err)
        }
        
        // Remove from repository
        if err := m.repository.Delete(file.Migration); err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to delete migration from repository: %w", err)
        }
        
        // Commit
        if err := tx.Commit().Error(); err != nil {
            return fmt.Errorf("failed to commit rollback for %s: %w", file.Migration, err)
        }
    }
    
    return nil
}
```

**Configuration Option**:
```go
config := map[string]any{
    "database": map[string]any{
        "migrations": map[string]any{
            "use_transactions": true, // Default: true
        },
    },
}
```

**Benefits**:
- Atomic migrations (all-or-nothing)
- Automatic rollback on errors
- Consistent database state
- Safer production deployments
- Can be disabled for specific cases (DDL limitations)

### 9. Enhanced Rollback Capabilities

**Current Behavior**: Rollback supports step-based and batch-based rollback

**Existing API**:
```go
// Current capabilities
err := migrator.Rollback(3, 0)    // Rollback last 3 migrations (step-based)
err := migrator.Rollback(0, 5)    // Rollback all migrations in batch 5
err := migrator.Rollback(0, 0)    // Rollback last batch
err := migrator.Reset()            // Rollback all and re-run
```

**Problem**: 
- Cannot rollback specific migration by name
- No "rollback to" specific point
- Limited flexibility

**Solution**: Add convenience methods for common rollback patterns

**Proposed Additional Methods**:
```go
// New convenience methods (additions to existing Rollback)

// RollbackToMigration rolls back all migrations after specified one
func (m *Migrator) RollbackToMigration(name string) error {
    files, err := m.repository.GetMigrations()
    if err != nil {
        return err
    }
    
    // Find target migration index
    targetIndex := -1
    for i, file := range files {
        if file.Migration == name {
            targetIndex = i
            break
        }
    }
    
    if targetIndex == -1 {
        return fmt.Errorf("migration not found: %s", name)
    }
    
    // Get migrations to rollback (after target)
    toRollback := files[targetIndex+1:]
    
    // Rollback in reverse order
    for i := len(toRollback) - 1; i >= 0; i-- {
        // ... perform rollback ...
    }
    
    return nil
}

// RollbackMigration rolls back a specific migration (if it's the last applied)
func (m *Migrator) RollbackMigration(name string) error {
    last, err := m.repository.GetLast()
    if err != nil {
        return err
    }
    
    // Check if migration is in last batch
    found := false
    for _, file := range last {
        if file.Migration == name {
            found = true
            break
        }
    }
    
    if !found {
        return fmt.Errorf("migration %s is not in the last batch and cannot be rolled back individually", name)
    }
    
    // Rollback (same as Rollback(0, 0) but filtered to specific migration)
    // ... implementation ...
}

// RollbackBatches rolls back specified number of batches
func (m *Migrator) RollbackBatches(count int) error {
    for i := 0; i < count; i++ {
        if err := m.Rollback(0, 0); err != nil {
            return err
        }
    }
    return nil
}
```

**Usage Examples**:
```go
// Rollback to a specific migration (everything after it)
err := migrator.RollbackToMigration("2026_06_14_1200_create_users_table")

// Rollback specific migration (if in last batch)
err := migrator.RollbackMigration("2026_06_14_1205_add_indexes")

// Rollback last 3 batches
err := migrator.RollbackBatches(3)

// Existing methods still work
err := migrator.Rollback(3, 0)    // Last 3 migrations
err := migrator.Rollback(0, 5)    // Batch 5
err := migrator.Rollback(0, 0)    // Last batch
err := migrator.Reset()            // All migrations
```

**Benefits**:
- More flexible rollback options
- Better migration management
- Easier recovery from issues
- Backward compatible (adds new methods)

### 10. Migration Locking for Concurrent Execution

**Current Behavior**: No protection against concurrent migration execution

**Problem**:
- Multiple processes can run migrations simultaneously
- Race conditions in distributed deployments
- Potential database corruption
- No coordination mechanism

**Solution**: Add advisory locking mechanism

**Implementation**:
```go
// In repository.go
func (r *Repository) AcquireLock(timeout time.Duration) (bool, error) {
    driver := r.config.GetString(fmt.Sprintf("database.connections.%s.driver", r.orm.Name()))
    
    lockSQL := ""
    switch driver {
    case "postgres":
        // PostgreSQL advisory lock
        lockSQL = "SELECT pg_try_advisory_lock(hashtext('migrations'))"
    case "mysql":
        // MySQL GET_LOCK
        lockSQL = fmt.Sprintf("SELECT GET_LOCK('migrations', %d)", int(timeout.Seconds()))
    case "sqlserver":
        // SQL Server app lock
        lockSQL = "EXEC sp_getapplock @Resource='migrations', @LockMode='Exclusive', @LockOwner='Session', @LockTimeout=?"
    case "sqlite", "turso":
        // SQLite uses database-level locking, no additional lock needed
        return true, nil
    default:
        // For unsupported databases, proceed without lock
        return true, nil
    }
    
    var locked bool
    err := r.orm.Query().Raw(lockSQL).Scan(&locked)
    return locked, err
}

func (r *Repository) ReleaseLock() error {
    driver := r.config.GetString(fmt.Sprintf("database.connections.%s.driver", r.orm.Name()))
    
    releaseSQL := ""
    switch driver {
    case "postgres":
        releaseSQL = "SELECT pg_advisory_unlock(hashtext('migrations'))"
    case "mysql":
        releaseSQL = "SELECT RELEASE_LOCK('migrations')"
    case "sqlserver":
        releaseSQL = "EXEC sp_releaseapplock @Resource='migrations', @LockOwner='Session'"
    case "sqlite", "turso":
        return nil // No explicit unlock needed
    default:
        return nil
    }
    
    _, err := r.orm.Query().Exec(releaseSQL)
    return err
}

// In migrator.go
func (m *Migrator) Run() error {
    // Acquire lock with timeout
    timeout := time.Duration(m.config.GetInt("database.migrations.lock_timeout", 60)) * time.Second
    locked, err := m.repository.AcquireLock(timeout)
    if err != nil {
        return fmt.Errorf("failed to acquire migration lock: %w", err)
    }
    if !locked {
        return fmt.Errorf("could not acquire migration lock (another process is running migrations)")
    }
    defer m.repository.ReleaseLock()
    
    m.log("info", "Acquired migration lock", map[string]any{})
    
    // ... existing migration code ...
    
    return nil
}
```

**Configuration**:
```go
config := map[string]any{
    "database": map[string]any{
        "migrations": map[string]any{
            "lock_enabled": true,  // Default: true
            "lock_timeout": 60,    // Seconds to wait for lock
        },
    },
}
```

**Benefits**:
- Safe concurrent deployments
- Prevents race conditions
- Kubernetes/distributed friendly
- Database-native locking
- Optional (can be disabled)

## Migration Path

### Phase 1: Foundation Enhancements (Week 1-2)

**Goal**: Add backward-compatible infrastructure improvements

**Tasks**:
1. Enhance migration table schema (add nullable columns: description, started_at, completed_at)
2. Update Repository interface and implementation:
   - Modify `Log()` method signature to accept timing and description
   - Add automatic schema upgrade in `CreateRepository()`
3. Add performance tracking in `Run()` and `Rollback()`
4. Update Status struct in contracts
5. Write comprehensive tests

**Deliverables**:
- Enhanced schema with automatic migration
- Performance tracking working
- All tests passing
- Backward compatibility maintained

### Phase 2: Developer Experience (Week 3-4)

**Goal**: Improve migration creation and management

**Tasks**:
1. Implement multiple ID format support:
   - Add MigrationIDFormat types
   - Update `Create()` method with format logic
   - Add date-sequence generation
2. Add validation functions and integrate into `Create()`
3. Implement duplicate detection in `RegisterMigration()`
4. Add structured logging support
5. Update documentation and examples

**Deliverables**:
- Multiple ID formats working
- Validation preventing invalid names
- Duplicate detection active
- Logging integrated
- Documentation updated

### Phase 3: Advanced Features (Week 5-6)

**Goal**: Add safety and reliability features

**Tasks**:
1. Implement explicit transaction wrapping
2. Add migration locking mechanism
3. Implement additional rollback convenience methods
4. Add description extraction (comment-based)
5. Performance testing and optimization

**Deliverables**:
- Transactional migrations
- Concurrent-safe execution
- Enhanced rollback options
- Complete feature set
- Performance benchmarks

### Phase 4: Polish and Release (Week 7)

**Goal**: Finalize and document

**Tasks**:
1. Comprehensive testing (all databases)
2. Documentation updates:
   - Migration guide
   - Configuration reference
   - Best practices
   - Upgrade guide
3. Example applications
4. Release notes

**Deliverables**:
- Complete test coverage
- Full documentation
- Examples
- Release ready

## Configuration Reference

**Complete Configuration Example**:
```go
config := map[string]any{
    "database": map[string]any{
        "migrations": map[string]any{
            // Table configuration
            "table": "migrations",                    // Default: "migrations"
            
            // ID format
            "id_format": "datetime",                  // Options: "datetime", "date", "unix", "custom"
                                                       // Default: "datetime"
            
            // Validation
            "validate_format": true,                  // Validate migration IDs
                                                       // Default: true
            
            // Transactions
            "use_transactions": true,                 // Wrap migrations in transactions
                                                       // Default: true
            
            // Locking
            "lock_enabled": true,                     // Enable migration locking
                                                       // Default: true
            "lock_timeout": 60,                       // Lock timeout in seconds
                                                       // Default: 60
            
            // Logging
            "logging": map[string]any{
                "enabled": true,                       // Enable structured logging
                                                       // Default: false
                "level": "info",                       // Log level: debug, info, warn, error
                                                       // Default: "info"
            },
        },
        
        "connections": map[string]any{
            "default": map[string]any{
                "driver": "postgres",
                "host":   "localhost",
                "port":   5432,
                // ... other connection settings
            },
        },
    },
}
```

## Backward Compatibility

**Status**: Breaking change with simple migration path

**What Changes**:
- Default migration ID format changes from `unix` to `datetime`
- Existing migrations with unix timestamps continue to work
- New migrations use datetime format by default

**Compatibility Strategy**:

1. **Schema Changes**:
   - New columns are nullable (description, started_at, completed_at)
   - Automatic schema upgrade on first run
   - Existing data remains valid
   - No data migration required

2. **Configuration**:
   - DateTime format is now default
   - Users can opt back to unix format if needed
   - All other new config options have sensible defaults
   - Logging disabled by default
   - Transactions enabled by default (can be disabled)
   - Locking enabled by default (can be disabled)

3. **API Compatibility**:
   - Existing `Migrator` interface unchanged
   - Existing `Repository` methods work as before
   - New methods are additions, not replacements
   - Registration API backward compatible

4. **Migration Files**:
   - Existing migration files work without modification
   - No changes required to Up/Down functions
   - Descriptions are optional
   - Legacy registrations still work

**Upgrade Path**:

```go
// Step 1: Update Neat ORM dependency
go get -u github.com/dracory/neat

// Step 2: Existing migrations continue to work
// First run automatically upgrades migration table

// Step 3: New migrations use datetime format automatically
migrator.Create("add user roles")
// Creates: 2026_06_14_1200_add_user_roles.go

// Step 4: (Optional) Continue using unix format
config := map[string]any{
    "database": map[string]any{
        "migrations": map[string]any{
            "id_format": "unix", // Opt back to unix format
        },
    },
}
```

**Migration Coexistence**:
```
database/migrations/
├── 1717080000_create_users_table.go       # Old unix format (still works)
├── 1717080100_add_email_index.go          # Old unix format (still works)
├── 2026_06_14_1200_create_posts.go        # New datetime format
└── 2026_06_14_1201_add_timestamps.go      # New datetime format
```

All migrations run in chronological order regardless of format, as lexicographical sorting works for both.

**For Users Who Want to Avoid Breaking Changes**:
Simply set the ID format to unix in configuration before creating new migrations:
```go
config := map[string]any{
    "database": map[string]any{
        "migrations": map[string]any{
            "id_format": "unix",
        },
    },
}
```

## Benefits Summary

### For Developers

1. **Better Developer Experience**:
   - Human-readable migration IDs (datetime format)
   - Clear validation error messages
   - Self-documenting migrations (descriptions)
   - Structured logging for debugging

2. **Improved Productivity**:
   - Less time deciphering timestamps
   - Faster debugging with performance metrics
   - Better code reviews with descriptions
   - Fewer deployment issues with locking

3. **Enhanced Safety**:
   - Transactional migrations prevent partial state
   - Locking prevents concurrent execution
   - Validation catches errors early
   - Duplicate detection prevents conflicts

### For Operations

1. **Production Reliability**:
   - Safe concurrent deployments (locking)
   - Atomic migrations (transactions)
   - Performance monitoring (timing data)
   - Audit trail (complete history)

2. **Debugging and Monitoring**:
   - Identify slow migrations
   - Track execution patterns
   - Structured logging integration
   - Historical performance data

3. **Compliance and Auditing**:
   - Complete execution history
   - Timing and performance records
   - Clear audit trail
   - Description metadata

### For Teams

1. **Better Collaboration**:
   - Consistent naming conventions
   - Self-documenting migrations
   - Easier code reviews
   - Clear team standards

2. **Knowledge Sharing**:
   - Descriptive migrations
   - Better organization
   - Easier onboarding
   - Clear history

### Technical Benefits

1. **Flexibility**:
   - Multiple ID formats
   - Configurable behavior
   - Optional features
   - Extensible design

2. **Reliability**:
   - Transaction support
   - Locking mechanism
   - Error handling
   - Rollback safety

3. **Maintainability**:
   - Clear code structure
   - Good test coverage
   - Comprehensive documentation
   - Backward compatibility

## Implementation Effort

**Estimated Effort**: 6-7 weeks (1 developer)

**Breakdown by Phase**:

| Phase | Component | Estimated Days |
|-------|-----------|----------------|
| **Phase 1** | Schema enhancements | 2 |
| | Repository updates | 3 |
| | Performance tracking | 2 |
| | Testing | 3 |
| | **Subtotal** | **10 days (2 weeks)** |
| **Phase 2** | ID format support | 3 |
| | Validation implementation | 2 |
| | Duplicate detection | 1 |
| | Logging integration | 2 |
| | Testing & Documentation | 2 |
| | **Subtotal** | **10 days (2 weeks)** |
| **Phase 3** | Transaction wrapping | 3 |
| | Locking mechanism | 4 |
| | Enhanced rollback | 2 |
| | Description extraction | 1 |
| | Testing & Optimization | 2 |
| | **Subtotal** | **12 days (2.4 weeks)** |
| **Phase 4** | Database testing | 2 |
| | Documentation | 3 |
| | Examples | 1 |
| | Release prep | 1 |
| | **Subtotal** | **7 days (1.4 weeks)** |
| **Total** | | **39 days (~7.8 weeks)** |

**Risk Factors**:
- Transaction support varies by database (some DDL may not be transactional)
- Locking mechanisms are database-specific
- Testing across all supported databases (SQLite, MySQL, PostgreSQL, SQL Server, Oracle, Turso)
- Backward compatibility testing with existing migrations

**Mitigation**:
- Transactions and locking can be disabled per database if needed
- Comprehensive test suite for each database
- Gradual rollout with feature flags
- Beta testing with community

## Alternatives Considered

### Alternative A: Minimal Enhancements Only

**Approach**: Only add performance tracking and descriptions, skip format changes

**Pros**:
- Much faster to implement (2-3 weeks)
- Lower risk
- Simpler testing
- Less code to maintain

**Cons**:
- Doesn't address human-readability of timestamps
- Misses opportunity for significant improvement
- Community may still request format options later
- Half-measure that doesn't fully solve the problem

**Verdict**: Not recommended. The ID format is a major pain point that should be addressed.

### Alternative B: Breaking Change with New System

**Approach**: Design completely new migration system from scratch

**Pros**:
- Clean slate, modern design
- No backward compatibility constraints
- Could simplify codebase
- Opportunity for architectural improvements

**Cons**:
- Breaking change for all users
- Requires migration of migrations
- High risk of disruption
- Significant effort for users to upgrade
- Could fragment ecosystem
- Community pushback likely

**Verdict**: Not recommended. Neat ORM is production-ready and stability is more important than perfection.

### Alternative C: Use External Migration Package

**Approach**: Integrate with standalone migration library (e.g., golang-migrate/migrate)

**Pros**:
- Proven, mature solution
- Large feature set
- Active community
- Less maintenance burden
- Well-documented

**Cons**:
- Additional dependency
- Less integration with Neat's schema builder
- Different API than current system
- Migration path complex
- Loses control over features
- May not fit Neat's design philosophy
- Users would need to learn new system

**Verdict**: Not recommended. Neat's integration with its schema builder is a key differentiator. An external package would lose this tight integration.

### Alternative D: CLI-Based Migration Management

**Approach**: Build separate CLI tool for advanced features

**Pros**:
- Keeps core library simple
- Advanced features opt-in via CLI
- Easier to extend
- Could provide richer UX

**Cons**:
- Additional tool to install
- Split functionality
- Harder to use programmatically
- More complex deployment
- Doesn't solve library-level issues

**Verdict**: Could complement this proposal but not replace it. Core improvements should be in the library.

## Recommendation

**Proceed with the proposed enhancements** as outlined in this document.

**Rationale**:
1. **Addresses Real Pain Points**: Human-readable IDs, performance tracking, and better logging are frequently requested
2. **Maintains Compatibility**: Zero breaking changes, smooth upgrade path
3. **Incremental Delivery**: Can ship features in phases
4. **Competitive**: Brings Neat on par with other modern ORMs
5. **Manageable Effort**: 7-8 weeks is reasonable for the value delivered
6. **Future-Proof**: Extensible design allows future enhancements

**Success Criteria**:
- All existing migrations work without modification
- New features can be adopted gradually
- Performance improves or stays the same
- Documentation is comprehensive
- Community feedback is positive
- All tests pass on all supported databases

## Open Questions

1. **ID Format Default**: ✅ **Resolved** - DateTime format will be the default, with unix available as opt-in

2. **Transaction Opt-Out**: Should transactions be opt-in or opt-out? Some databases have DDL transaction limitations.
   - **Recommendation**: Opt-out (enabled by default), with automatic detection for problematic databases

3. **Description Source**: Comment-based parsing or struct-based approach for descriptions?
   - **Recommendation**: Start with comment-based for simplicity, consider struct-based in future if needed

4. **Locking Strategy**: Should locking be mandatory or optional? What about databases without advisory locks?
   - **Recommendation**: Optional with automatic fallback for databases without lock support

5. **Performance Threshold**: Should we add configurable slow migration warnings?
   - **Recommendation**: Yes, add optional threshold (e.g., 5 seconds) with warning log

6. **Migration Dependencies**: Should we support explicit dependencies between migrations?
   - **Recommendation**: Not in this phase; migrations already run in order. Can be future enhancement.

7. **Dry Run Mode**: Should we add a dry-run option to preview migrations without executing?
   - **Recommendation**: Good idea, but defer to future enhancement to limit scope

8. **Rollback Confirmation**: Should destructive rollbacks require confirmation in production?
   - **Recommendation**: Add configuration option for confirmation callback, disabled by default

9. **Migration Metadata**: Should we store additional metadata (git hash, deployer, environment)?
   - **Recommendation**: Defer to future enhancement, focus on core features first

10. **CLI Integration**: Should these features be exposed via CLI commands?
    - **Recommendation**: If Neat develops CLI tooling, yes. For now, focus on library API.

## Related Proposals and Issues

- **[Sugar Methods for API Usability](sugar-methods-for-usability.md)** - May influence migration API ergonomics
- **[Feature Requests](feature-requests.md)** - Community requests for migration improvements
- **Schema Builder Enhancements** - Consider how migration features integrate with schema DSL
- **Testing Infrastructure** - Migration testing patterns and best practices

## Community Feedback

Before finalizing this proposal, seek community feedback on:
1. ID format preferences (unix vs datetime vs date-sequence)
2. Most needed features (prioritization)
3. Configuration preferences
4. Migration workflow pain points
5. Production deployment concerns

**Feedback Channels**:
- GitHub Discussions
- Issue tracker for specific feature requests
- Community Discord/Slack (if available)
- User survey

## Next Steps

1. **Community Review** (1-2 weeks):
   - Share proposal with community
   - Gather feedback and concerns
   - Refine based on input

2. **Proof of Concept** (1 week):
   - Build prototype of Phase 1 features
   - Validate approach
   - Identify technical challenges

3. **Final Approval** (1 week):
   - Present refined proposal to maintainers
   - Get approval to proceed
   - Finalize timeline

4. **Implementation** (7-8 weeks):
   - Execute phased rollout
   - Regular check-ins and demos
   - Community beta testing

5. **Release** (1 week):
   - Final testing
   - Documentation review
   - Version release
   - Announcement

**Total Timeline**: Approximately 11-13 weeks from approval to release

## Conclusion

This proposal aims to significantly enhance Neat ORM's migration system while maintaining 100% backward compatibility. The phased approach allows for incremental delivery and validation, while the comprehensive feature set brings Neat on par with modern migration solutions.

The key differentiators are:
- **Zero breaking changes** - existing code continues to work
- **Gradual adoption** - features can be enabled one at a time
- **Deep integration** - leverages Neat's schema builder
- **Production-ready** - safety features like locking and transactions
- **Developer-friendly** - human-readable IDs and better tooling

**Success depends on**:
- Community engagement and feedback
- Thorough testing across all databases
- Comprehensive documentation
- Smooth upgrade experience
- Delivered value exceeding implementation cost

With approval, implementation can begin immediately with Phase 1, delivering incremental value while working toward the complete vision.