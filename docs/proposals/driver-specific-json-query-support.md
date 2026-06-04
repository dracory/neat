# Driver-Specific JSON Query Support

**Date**: June 3, 2026
**Status**: Proposal
**Priority**: Medium

## Problem Statement

Neat ORM currently provides JSON query methods (`WhereJsonContains`, `WhereJsonContainsKey`, `WhereJsonLength`) that use the `->` operator syntax. This works for MySQL, PostgreSQL, and SQLite but fails for Oracle and SQL Server which use different JSON function syntaxes.

### Current Limitations

- **Oracle**: Uses `JSON_VALUE`, `JSON_EXISTS`, `JSON_TABLE` functions instead of `->` operator
- **SQL Server**: Uses `JSON_VALUE`, `JSON_MODIFY` functions instead of `->` operator
- **Unified API**: No abstraction layer to translate JSON queries to database-specific syntax
- **Test Coverage**: Oracle and SQL Server JSON tests are skipped due to missing driver support

### Example of the Issue

```go
// Current implementation (works for MySQL/PostgreSQL/SQLite)
query.Model(&models.JsonData{}).WhereJsonContains("data->name", "json1").Find(&foundData)

// Fails on Oracle with: ORA-00936: missing expression
// Oracle expects: JSON_VALUE(data, '$.name') = 'json1'
```

## Proposed Solution

Implement a driver-specific grammar system similar to Laravel's approach, where each database driver has its own grammar class responsible for translating JSON query methods to database-specific SQL.

### Architecture

```
database/
  driver/
    grammar/
      base_grammar.go          // Base grammar interface
      mysql_grammar.go         // MySQL-specific implementations
      postgres_grammar.go      // PostgreSQL-specific implementations
      sqlite_grammar.go        // SQLite-specific implementations
      sqlserver_grammar.go     // SQL Server-specific implementations
      oracle_grammar.go        // Oracle-specific implementations
```

### Interface Design

```go
package grammar

// Grammar defines the interface for database-specific SQL generation
type Grammar interface {
    // JSON query methods
    CompileJsonContains(column string, value interface{}) string
    CompileJsonDoesntContain(column string, value interface{}) string
    CompileJsonContainsKey(column string) string
    CompileJsonDoesntContainKey(column string) string
    CompileJsonLength(column string, operator string, value interface{}) string
    CompileJsonPath(column string) string
    
    // Spatial query methods (future)
    CompileSpatialPoint(value string) string
    CompileSpatialDistance(column string, point string) string
}
```

### Implementation Examples

#### MySQL Grammar
```go
func (g *MySQLGrammar) CompileJsonContains(column string, value interface{}) string {
    return fmt.Sprintf("JSON_CONTAINS(%s, ?)", column)
}

func (g *MySQLGrammar) CompileJsonLength(column string, operator string, value interface{}) string {
    return fmt.Sprintf("JSON_LENGTH(%s) %s ?", column, operator)
}
```

#### Oracle Grammar
```go
func (g *OracleGrammar) CompileJsonContains(column string, value interface{}) string {
    // Translate data->name to JSON_VALUE(data, '$.name')
    path := g.extractJsonPath(column)
    return fmt.Sprintf("JSON_VALUE(%s, '%s') = ?", g.extractTable(column), path)
}

func (g *OracleGrammar) CompileJsonLength(column string, operator string, value interface{}) string {
    path := g.extractJsonPath(column)
    return fmt.Sprintf("JSON_VALUE(%s, '$.size(%s)') %s ?", g.extractTable(column), path, operator)
}
```

#### SQL Server Grammar
```go
func (g *SQLServerGrammar) CompileJsonContains(column string, value interface{}) string {
    path := g.extractJsonPath(column)
    return fmt.Sprintf("JSON_VALUE(%s, '%s') LIKE ?", column, path)
}
```

### Integration with Query Builder

```go
// In database/query/query.go
type Query struct {
    grammar grammar.Grammar
    // ... existing fields
}

func (q *Query) WhereJsonContains(column string, value interface{}) *Query {
    sql := q.grammar.CompileJsonContains(column, value)
    return q.WhereRaw(sql, value)
}

func (q *Query) WhereJsonLength(column string, operator string, value interface{}) *Query {
    sql := q.grammar.CompileJsonLength(column, operator, value)
    return q.WhereRaw(sql, value)
}
```

## Implementation Plan

### Phase 1: Grammar Infrastructure (Week 1-2)
1. Create `database/driver/grammar` package
2. Define `Grammar` interface with JSON methods
3. Implement base grammar with default implementations
4. Add grammar initialization to driver setup

### Phase 2: Driver Implementations (Week 3-4)
1. Implement MySQL grammar (migrate existing logic)
2. Implement PostgreSQL grammar (migrate existing logic)
3. Implement SQLite grammar (migrate existing logic)
4. Implement SQL Server grammar
5. Implement Oracle grammar

### Phase 3: Query Builder Integration (Week 5)
1. Update query builder to use grammar for JSON methods
2. Update all existing JSON query calls
3. Add grammar selection based on driver type

### Phase 4: Testing (Week 6)
1. Add unit tests for each grammar implementation
2. Update integration tests for all drivers
3. Enable Oracle and SQL Server JSON tests
4. Add cross-driver compatibility tests

### Phase 5: Documentation (Week 7)
1. Update API documentation
2. Add driver-specific JSON query examples
3. Document grammar extension points for custom drivers

## Benefits

1. **Unified API**: Single API works across all database drivers
2. **Extensibility**: Easy to add support for new database drivers
3. **Maintainability**: Database-specific logic isolated in grammar classes
4. **Test Coverage**: Full test coverage for all drivers including Oracle and SQL Server
5. **Future-Proof**: Foundation for spatial queries and other database-specific features

## Migration Path

### Backward Compatibility
- Existing JSON query methods remain unchanged in API
- Only internal implementation changes
- No breaking changes for end users

### Deprecation (Optional)
- Could deprecate direct `->` operator usage in favor of explicit JSON methods
- Add warnings when using `->` operator on non-supporting databases

## Effort Estimate

- **Phase 1**: 2 weeks (architecture and infrastructure)
- **Phase 2**: 2 weeks (driver implementations)
- **Phase 3**: 1 week (integration)
- **Phase 4**: 1 week (testing)
- **Phase 5**: 1 week (documentation)

**Total**: 7 weeks for full implementation

## Risks and Mitigations

### Risk 1: Complex Path Extraction
- **Issue**: Extracting JSON paths from `data->name` format can be complex
- **Mitigation**: Use regex-based path extraction with comprehensive test coverage

### Risk 2: Performance Impact
- **Issue**: Additional abstraction layer may impact query performance
- **Mitigation**: Grammar methods are simple string operations, minimal overhead

### Risk 3: Driver-Specific Edge Cases
- **Issue**: Each database has unique JSON function behaviors
- **Mitigation**: Comprehensive integration tests for each driver

## Future Enhancements

1. **Spatial Query Support**: Extend grammar system for spatial queries (SDO_GEOMETRY, etc.)
2. **Full Text Search**: Add grammar methods for database-specific full-text search
3. **Window Functions**: Add grammar support for database-specific window function syntax
4. **Custom Functions**: Allow users to register custom grammar methods for custom SQL functions

## References

- Laravel Database Query Builder: https://laravel.com/docs/11.x/queries#json-where-clauses
- MySQL JSON Functions: https://dev.mysql.com/doc/refman/8.0/en/json-functions.html
- PostgreSQL JSON Functions: https://www.postgresql.org/docs/current/functions-json.html
- SQLite JSON Functions: https://www.sqlite.org/json1.html
- Oracle JSON Functions: https://docs.oracle.com/en/database/oracle/oracle-database/21/adxjs/json-in-oracle-database.html
- SQL Server JSON Functions: https://learn.microsoft.com/en-us/sql/t-sql/functions/json-functions-transact-sql
