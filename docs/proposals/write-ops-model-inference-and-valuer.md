# Write Operations: TableName Inference Without Model() and Valuer on Nil Values

**Date**: November 12, 2026
**Status**: Proposed
**Priority**: High
**Impact**: High
**Effort**: Low-Medium
**ROI**: Very High

## Description

Two friction points discovered while integrating neat into a real project (`goflow`, `pkg/flowadmin/storage`):

1. `Query.Create(&row)` fails with `table name cannot be empty` even when the struct implements `TableName() string`. Callers must currently write `.Model(&row{}).Create(&row)`. The same applies to `Update`, `Delete`, `Count`, and other write ops used without an explicit `.Model()` or `.Table()` call.
2. Custom column types implementing `driver.Valuer` (e.g. a `JSONMap` stored as TEXT) are not invoked when the value is nil — the driver receives `NULL`, causing `NOT NULL constraint failed` errors on non-nullable JSON columns.

## Rationale

- `TableName()` is already the documented mechanism for table resolution — reads infer the table from the destination struct, so writes silently failing without `.Model()` is inconsistent and surprising.
- `driver.Valuer`/`sql.Scanner` is the standard `database/sql` contract for custom column types; nil values should still call `Value()` (which conventionally returns a default such as `"{}"`), not bypass it.
- Both issues produce cryptic failures (`table name cannot be empty`, `NOT NULL constraint failed`) that are hard to attribute without debugging.

## Observed Behavior

```go
type nodeRow struct {
    ID   string  `db:"id"`
    Data JSONMap `db:"data"`
}
func (nodeRow) TableName() string { return "nodes" }

// Fails: "table name cannot be empty"
db.Query().Create(&nodeRow{ID: "n1"})

// Works
db.Query().Model(&nodeRow{}).Create(&nodeRow{ID: "n1"})

// Fails when Data is nil: "NOT NULL constraint failed: nodes.data"
// JSONMap.Value() returning "{}" is never called for nil maps
```

## Proposed Improvements

1. **Infer model/table on write ops**: when `.Model()`/`.Table()` was not called, `Create`, `Save`, `Update`, `Delete`, `InsertGetId`, `Restore`, `ForceDelete` should resolve the table from the value's `TableName()` method or struct-name convention — same inference `First`/`Find` already apply to the destination.
2. **Honor `driver.Valuer` for nil values**: when a field type implements `driver.Valuer`, call `Value()` even when the field is a nil map/slice, so the implementation decides the stored representation.
3. **Optional**: emit a warning in debug mode when `Create`/`Update` is called without a resolvable model, before the write is attempted.

## Implementation Details

- Table inference: `inferTableFromValue` / `inferTableFromConds` in `query_model.go` reuse `resolveTableName`; wired into `Create`, `InsertGetId`, `Save`, `Delete`, `SoftDelete`, `ForceDelete`/`HardDelete`, `Restore`, `First`, `Find`, and `Get`. `Update(column, value)` is excluded — it takes a column, not a model.
- Inference also sets `q.model` (when unset) from a struct-like value or a fresh element instance for slice destinations, so `SoftDeleteColumnNamer` capability checks, soft-delete WHERE filters, and observer dispatch behave exactly as with an explicit `.Model()`.
- Valuer handling: in `extractStructColumnsAndValues` (`builder_extract.go`), fields whose type implements `driver.Valuer` bypass both the slice/struct-kind skip and the `IsZero()` skip, so a nil `JSONMap` still reaches `database/sql`, which invokes `Value()` and stores `"{}"`. Note the field was previously dropped from the column list entirely (not bound as NULL).
- Regression tests: `database/query/query_write_inference_valuer_test.go` covers create/save/delete/first/find inference, soft-delete semantics on inferred models, and nil/non-nil map+slice Valuer fields.

## Workaround (current)

Always chain `.Model(&T{})` (or `.Table("name")`) before `Create`/`Update`/`Delete`, and ensure custom `driver.Valuer` fields are non-nil before insert (or mark the column `.Nullable()`).

## References

- Discovered while replacing GORM with neat in `goflow` (`pkg/flowadmin/storage/neat_storage.go`, commit pending)
- Related: `docs/proposals/better-error-messages.md` (error sanitization hides the real error unless unwrapped or debug is enabled)
