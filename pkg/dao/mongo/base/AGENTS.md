# BASE DAO KNOWLEDGE BASE

## OVERVIEW

`base` is the foundational DAO toolkit for all MongoDB collection handlers. It provides the generic ORM (`Orm[P, T]`), common filter builders (`OptFn`), BSON parameter builders, pagination, metric instrumentation, data broker structs, sentinel errors, and validation helpers. Every `pkg/dao/mongo/*` handler depends on this package.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `orm.go` | Generic `Orm[P, T]` implementing `IOrm` — CRUD, Count, List, Distinct, HardDelete, BulkUpdate |
| `broker.go` | `TableBroker[T]`, `BasicInfo`, `IData` interface, `TableChangeEventBroker` |
| `filter.go` | `OptFn` type and filter builders: `WithValues`, `WithoutValues`, `WithFuzzyValues`, `WithRegexMatch`, `WithElemMatch`, `WithTimeRange`, `AliveFilter` |
| `param_builder.go` | BSON update/upsert/delete/push/pull param builders (`BuildUpsertParam`, `BuildDeleteParam`, etc.) |
| `page.go` | `ParsePage` — converts `types.Page` to `mongo FindOptions` with sort/skip/limit |
| `metrics.go` | DAO-level Prometheus metric recording via `pkg/dao/mongo.Monitor` |
| `errors.go` | Sentinel error constructors: `ErrInvalidContext`, `ErrRecordNoFound`, `ErrEmptyParamData`, etc. |
| `constant.go` | Shared BSON field keys (`basic.created_at`, `basic.updated_at`, `basic.is_deleted`) and constants |
| `validate.go` | `CheckTenantIDMatched` — multi-tenant ID validation |
| `options.go` | Convenience `OptFn` wrappers (e.g. `WithUpdateAtTimeRange`) |

## CONVENTIONS

### TableBroker Envelope

All collections use `TableBroker[T]` as the document envelope:
```
{ "basic": { "created_at", "updated_at", "is_deleted" }, "data": <T> }
```
The `IData` interface requires `UniqueKey()` and `UniqueFields()` on the inner data type.

### Generic ORM Pattern

Each domain handler creates a `dao` struct implementing `base.IDao` (GetClient, GetTableName, GetIndexes), then embeds `base.IOrm` via `base.NewOrm[*T, T](dao)`. The ORM handles metric recording, cursor iteration, and soft-delete semantics automatically.

### Filter Composition

Filters are built by chaining `OptFn` functions over `AliveFilter()`:
```go
filter := base.AliveFilter()
for _, opt := range opts {
    filter = opt(filter)
}
```

### Soft Delete

`DeleteMany` sets `basic.is_deleted = true` (soft delete). `HardDelete` / `HardDeleteMany` permanently remove documents and reject empty filters as a safety guard.

### Metric Instrumentation

Every ORM method records operation metrics (duration, data length, errors) via `metricData`. The monitor is initialized once in `init()`.

### Deprecated APIs

- `WithInt64Values`, `WithStringValues`, `WithoutStringValues` → use generic `WithValues` / `WithoutValues`
- `BuildUpdateField` (param_builder) → use `Orm.UpdateField`
- `ErrInvalidID` → use per-package errors

## ANTI-PATTERNS

- Do not bypass `TableBroker` envelope — all documents must wrap data inside `basic` + `data`.
- Do not call `HardDelete` / `HardDeleteMany` with empty filters — it will return an error.
- Do not use deprecated typed filter functions (`WithInt64Values`, `WithStringValues`) — use `WithValues[T]`.
- Do not duplicate metric logic in domain handlers — the ORM handles it.
- Do not pass raw `context.Context` to ORM methods — use `contextx.IContext`.
