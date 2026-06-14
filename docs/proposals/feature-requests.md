# Feature Requests Proposal

**Date**: June 1, 2026
**Status**: Open for Discussion

## Overview

This document outlines proposed feature improvements for Neat ORM, categorized by priority and impact. These proposals aim to enhance developer experience, close feature gaps with competing ORMs, and expand Neat's capabilities.

## High-Priority Feature Gaps

### 1. ManyToMany Relationships

**Description**: Add support for many-to-many relationships with pivot tables.

**Rationale**: 
- Currently missing from Neat (GORM has it)
- Critical for many real-world applications (e.g., user roles, tags, categories)
- Common requirement in most applications

**Proposed API**:
```go
type User struct {
    ID    uint
    Name  string
    Roles []Role `neat:"many_to_many:user_roles"`
}

type Role struct {
    ID          uint
    Name        string
    Permissions []Permission `neat:"many_to_many:role_permissions"`
}

// Usage
user.Roles().Attach(roleID)
user.Roles().Detach(roleID)
user.Roles().Sync([]uint{roleID1, roleID2})
db.Query().Model(&User{}).With("Roles").Find(&users)
```

**Implementation Details**:
- Automatic pivot table creation
- Sync, attach, detach operations
- Eager loading with `With()`
- Pivot table with additional columns (timestamps, custom fields)

**Estimated Effort**: High

---

### 2. Query Caching

**Description**: Add query result caching with TTL support.

**Rationale**:
- GORM has built-in caching support
- Improves performance for frequently accessed data
- Reduces database load

**Proposed API**:
```go
// Cache query for 5 minutes
db.Query().Model(&User{}).Cache(5*time.Minute).Where("id = ?", 1).First(&user)

// Cache with custom key
db.Query().Model(&User{}).Cache("user:1", 10*time.Minute).Where("id = ?", 1).First(&user)

// Invalidate cache
db.Query().Model(&User{}).CacheInvalidate("user:1")
```

**Implementation Details**:
- Pluggable cache backend (Redis, in-memory)
- Automatic cache invalidation on model updates
- Cache tags for bulk invalidation
- Configurable default TTL

**Estimated Effort**: Medium

---

### 3. Advanced Scopes

**Description**: Add global scopes and scope composition features.

**Rationale**:
- GORM has advanced scope support
- Enables reusable query logic
- Supports multi-tenant applications

**Proposed API**:
```go
// Global scope (applied to all queries)
func (u *User) GlobalScopes() []Scope {
    return []Scope{
        func(q Query) Query {
            return q.Where("deleted_at IS NULL")
        },
    }
}

// Scope composition
activeAdmins := db.Query().Model(&User{}).Scopes(Active, Admin).Find(&users)

// Dynamic scope parameters
func CreatedAfter(date time.Time) Scope {
    return func(q Query) Query {
        return q.Where("created_at > ?", date)
    }
}
```

**Implementation Details**:
- Global scopes defined on model
- Scope chaining and composition
- Dynamic scope parameters
- Scope removal for specific queries

**Estimated Effort**: Medium

---

## Developer Experience

### 4. Better Error Messages

**Description**: Improve error messages with SQL context and debugging information.

**Rationale**:
- Current errors can be cryptic
- Hard to debug complex queries
- Improves developer productivity

**Proposed Improvements**:
```go
// Before: "SQL error: near 'WHERE'"
// After: "SQL error: near 'WHERE' in query: SELECT * FROM users WHERE name = ? AND"

// Debug mode
db.Debug().Query().Model(&User{}).Where("name = ?", "John").First(&user)
// Output: [SQL] SELECT * FROM users WHERE name = 'John' LIMIT 1 [0.5ms]
```

**Implementation Details**:
- SQL query context in errors
- Parameter values in error messages
- Debug mode with query logging
- Execution time tracking
- Stack traces for query errors

**Estimated Effort**: Low

---

### 5. IDE Support

**Description**: Enhance IDE support with better autocomplete and validation.

**Rationale**:
- Improves developer productivity
- Reduces errors at development time
- Better code completion for query builder

**Proposed Features**:
- Go struct tags for better autocomplete
- VS Code extension for query building
- LSP integration for query validation
- Syntax highlighting for raw SQL
- Query builder snippets

**Implementation Details**:
- Custom Go language server
- Query builder IntelliSense
- Real-time SQL validation
- Database schema integration

**Estimated Effort**: High

---

### 6. CLI Tools

**Description**: Create command-line tools for common ORM operations.

**Rationale**:
- Streamlines development workflow
- Consistent with Laravel Artisan
- Reduces manual setup

**Proposed Commands**:
```bash
# Migrations
neat migrate:status
neat migrate:run
neat migrate:rollback
neat migrate:refresh
neat migrate:fresh

# Seeders
neat seed
neat seed:run UserSeeder
neat seed:rollback

# Database
neat db:reset
neat db:drop
neat db:create

# Models
neat make:model User
neat make:migration create_users_table
neat make:seeder UserSeeder
```

**Implementation Details**:
- Cobra-based CLI
- Interactive mode
- Configuration file support
- Environment variable support

**Estimated Effort**: Medium

---

## Performance

### 7. Query Optimization

**Description**: Add automatic query optimization features.

**Rationale**:
- Prevent N+1 query problems
- Improve eager loading efficiency
- Connection pool recommendations

**Proposed Features**:
```go
// Automatic query batching
db.Query().Model(&User{}).With("Posts.Comments").Find(&users)
// Automatically batches queries instead of N+1

// Eager loading optimization
db.Query().Model(&User{}).With("Posts", "Profile").Find(&users)
// Optimizes to single query with joins where possible

// Connection pool recommendations
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

**Implementation Details**:
- Query batching for eager loading
- Automatic join optimization
- Connection pool health checks
- Performance monitoring

**Estimated Effort**: High

---

### 8. Benchmark Suite

**Description**: Create comprehensive benchmark suite.

**Rationale**:
- Performance comparisons with GORM
- Database-specific performance tests
- Memory usage profiling
- Regression detection

**Proposed Benchmarks**:
- Query building overhead
- Connection pooling efficiency
- Eager loading performance
- Transaction overhead
- Memory allocation patterns

**Implementation Details**:
- Go benchmark framework
- Continuous benchmarking
- Performance regression detection
- Database-specific benchmarks

**Estimated Effort**: Medium

---

## Testing

### 11. Property-Based Testing

**Description**: Add property-based tests for query builder.

**Rationale**:
- Catches edge cases
- Validates query correctness
- Prevents SQL injection

**Proposed Tests**:
- Query builder commutativity
- WHERE clause associativity
- JOIN order independence
- Parameter binding correctness

**Implementation Details**:
- Use `quicktest` or `gopter`
- Property definitions for query operations
- Fuzz testing for SQL injection

**Estimated Effort**: Medium

---

### 12. Snapshot Testing

**Description**: Add snapshot testing for query results and migrations.

**Rationale**:
- Regression testing for queries
- Migration schema validation
- Test data management

**Proposed Features**:
```go
// Query result snapshots
assert.QuerySnapshot(t, db.Query().Model(&User{}).Where("id = ?", 1))

// Migration schema snapshots
assert.SchemaSnapshot(t, migration)
```

**Implementation Details**:
- Snapshot file format
- Update mechanism
- CI/CD integration

**Estimated Effort**: Low

---

## Documentation

### 13. Migration Guides

**Description**: Create comprehensive migration guides from other ORMs.

**Rationale**:
- Lowers barrier to adoption
- Addresses common concerns
- Provides practical examples

**Proposed Guides**:
- "From GORM to Neat"
- "From Django to Neat"
- "From Sequelize to Neat"
- "From TypeORM to Neat"
- "From Laravel Eloquent to Neat"

**Implementation Details**:
- Side-by-side code comparisons
- Common patterns mapping
- Migration checklist
- Troubleshooting guide

**Estimated Effort**: Medium

---

### 14. Interactive Examples

**Description**: Add interactive examples and playground integration.

**Rationale**:
- Hands-on learning
- Quick experimentation
- Better documentation

**Proposed Features**:
- Go Playground integration
- Live query builder demo
- API reference with runnable examples
- Interactive tutorials

**Implementation Details**:
- Go Playground buttons in docs
- Embedded code editors
- Live database connections
- Example repository

**Estimated Effort**: Medium

---

## Advanced Features

### 15. Database-Specific Features

**Description**: Add support for database-specific features.

**Rationale**:
- Leverage database capabilities
- Better performance for specific databases
- Feature parity with native drivers

**Proposed Features**:
- PostgreSQL: JSONB operators, array operations, full-text search
- MySQL: Full-text search, spatial functions, window functions
- SQLite: FTS5, R-Tree, generated columns
- SQL Server: Spatial data, hierarchyid

**Implementation Details**:
- Database-specific query methods
- Type-safe operators
- Feature detection
- Fallback for unsupported databases

**Estimated Effort**: High

---

### 16. Event System

**Description**: Add comprehensive event system for model and query lifecycle.

**Rationale**:
- Hooks for business logic
- Audit logging
- Cache invalidation
- Notifications

**Proposed Events**:
```go
// Model events
user.On("beforeSave", func(u *User) error {
    return u.Validate()
})
user.On("afterSave", func(u *User) error {
    return cache.Invalidate("user:%d", u.ID)
})

// Query events
db.Query().On("beforeExecute", func(q Query) error {
    log.Printf("Executing: %s", q.ToSQL())
    return nil
})
```

**Implementation Details**:
- Event emitter pattern
- Event priority
- Async event handlers
- Event propagation control

**Estimated Effort**: Medium

---

### 17. Validation

**Description**: Add built-in model validation.

**Rationale**:
- Data integrity at model level
- Consistent validation logic
- User-friendly error messages

**Proposed API**:
```go
type User struct {
    ID    uint
    Name  string `validate:"required,min=3,max=50"`
    Email string `validate:"required,email"`
    Age   int    `validate:"min=18,max=120"`
}

user := User{Name: "John", Email: "invalid"}
if err := db.Query().Model(&user).Validate(); err != nil {
    // Validation errors
}
```

**Implementation Details**:
- Struct tag validation
- Custom validation rules
- Localization support
- Validation groups

**Estimated Effort**: Medium

---

## Tooling Ecosystem

### 18. Admin Panel

**Description**: Create auto-generated admin interface.

**Rationale**:
- Quick CRUD operations
- Database management
- Development productivity

**Proposed Features**:
- Auto-generated CRUD UI
- Search and filtering
- Bulk operations
- User authentication
- Permission system

**Implementation Details**:
- Web-based admin panel
- Model introspection
- Customizable templates
- Plugin system

**Estimated Effort**: High

---

### 19. Schema Inspector

**Description**: Create visual schema inspection tool.

**Rationale**:
- Visualize database structure
- Understand relationships
- Documentation generation

**Proposed Features**:
- Visual schema viewer
- Relationship graph
- Migration diff viewer
- ER diagram export

**Implementation Details**:
- Web-based tool
- CLI version
- SVG/PNG export
- Interactive exploration

**Estimated Effort**: Medium

---

### 20. Query Analyzer

**Description**: Add query analysis and optimization recommendations.

**Rationale**:
- Performance tuning
- Slow query detection
- Best practices enforcement

**Proposed Features**:
```go
// EXPLAIN query analysis
analyzer := db.Query().Model(&User{}).Where("name = ?", "John").Explain()
analyzer.Suggestions() // "Add index on name column"

// Slow query detection
db.Query().SetSlowQueryThreshold(100 * time.Millisecond)
db.Query().On("slowQuery", func(q Query, duration time.Duration) {
    log.Printf("Slow query: %s took %v", q.ToSQL(), duration)
})
```

**Implementation Details**:
- EXPLAIN parsing
- Index recommendations
- Query pattern detection
- Performance metrics

**Estimated Effort**: Medium

---

## Priority Matrix

| Feature | Priority | Impact | Effort | ROI |
|---------|----------|--------|--------|-----|
| ManyToMany | High | High | High | High |
| Query Caching | Medium | Medium | Medium | Medium |
| Advanced Scopes | Medium | Medium | Medium | Medium |
| Better Error Messages | High | High | Low | Very High |
| IDE Support | Medium | High | High | Medium |
| CLI Tools | Medium | High | Medium | High |
| Query Optimization | High | High | High | High |
| Benchmark Suite | Medium | Medium | Medium | Medium |
| Sequelize Compatibility | Low | Low | Low | Low |
| TypeORM Compatibility | Low | Low | High | Low |
| Property-Based Testing | Medium | Medium | Medium | Medium |
| Snapshot Testing | Low | Medium | Low | Medium |
| Migration Guides | High | High | Medium | High |
| Interactive Examples | Medium | Medium | Medium | Medium |
| Database-Specific Features | Medium | High | High | Medium |
| Event System | Medium | Medium | Medium | Medium |
| Validation | Medium | High | Medium | High |
| Admin Panel | Low | Medium | High | Low |
| Schema Inspector | Low | Medium | Medium | Low |
| Query Analyzer | Medium | High | Medium | High |

---

## Implementation Roadmap

### Phase 1 (Quick Wins - Low Effort, High Impact)
- Better Error Messages
- Snapshot Testing
- Sequelize Compatibility
- Migration Guides

### Phase 2 (Core Features - Medium Effort, High Impact)
- ManyToMany Relationships
- CLI Tools
- Query Optimization
- Validation
- Query Analyzer

### Phase 3 (Enhanced Experience - Medium/High Effort)
- Query Caching
- Advanced Scopes
- Benchmark Suite
- Interactive Examples
- Event System
- Database-Specific Features

### Phase 4 (Ecosystem - High Effort)
- IDE Support
- TypeORM Compatibility
- Admin Panel
- Schema Inspector

---

## Open Questions

1. Should we prioritize ManyToMany over Query Caching?
2. Should CLI tools be part of core or separate package?
3. Should Admin Panel be a separate project?
4. What validation library should we use or build our own?
5. Should we support multiple cache backends out of the box?

---

## References

- GORM Documentation: https://gorm.io/docs/
- Django ORM Documentation: https://docs.djangoproject.com/en/stable/topics/db/queries/
- Sequelize Documentation: https://sequelize.org/
- TypeORM Documentation: https://typeorm.io/
- Laravel Eloquent Documentation: https://laravel.com/docs/eloquent
