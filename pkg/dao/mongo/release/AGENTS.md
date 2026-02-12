# RELEASE DAO KNOWLEDGE BASE

## OVERVIEW

`release` provides MongoDB data access for release records (agent, proxy, plugin, cert, bintool, etc.). Each release type maps to a separate collection (`release_{type}`), managed through a single `Handler` with lazy DAO initialization via `sync.Map`. The package converts between internal `Release` models and `types.Release` at the handler boundary.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `handler.go` | `Handler` struct, `IHandler` / `ISwitcher` / `IDistinctor` interfaces, all exported CRUD methods, type converters |
| `release.go` | Internal `dao` struct (implements `base.IDao`), `newDao`, `upsertMany`, `buildUpsertManyParams` |
| `table.go` | `Release` data model (implements `base.IData`), `TableName()`, `UniqueFields`, `UniqueKey` |
| `constants.go` | BSON field key constants (`data.generation`, `data.type`, `data.version`, etc.) |
| `options.go` | `OptFn` filter builders: `WithName`, `WithType`, `WithVersion`, `WithPlatform`, `WithEnabled`, etc. |

## CONVENTIONS

### Dynamic Table Naming

Collections are named `release_{releaseType}`. `TableName()` generates the name; `Handler.releaseTypeDao()` lazily creates and caches a `dao` per release type using `sync.Map`.

### Handler as Facade

`Handler` is the only exported entry point. It implements `IHandler` which composes `ISwitcher` (enable/default toggles) and `IDistinctor` (distinct queries). All methods take `releaseType` as the first domain parameter to route to the correct collection.

### Type Conversion Boundary

Internal `Release` ↔ `types.Release` conversion happens exclusively in `handler.go` via `convertReleaseToTypes` / `convertReleaseFromTypes`. Callers never see the internal model.

### Filter Composition

Same pattern as `base`: chain `OptFn` over `base.AliveFilter()`. The `WithPlatform` option handles single vs. multi-platform with `$or` queries.

### Upsert Strategy

`UpsertMany` matches on the 6-field unique key (name, generation, type, cpu_arch, os_type, version) and uses `base.BuildUpsertParam` for `$set` + `$setOnInsert` semantics.

### ORM Embedding

`dao` embeds `base.IOrm[*Release, Release]` for standard CRUD. Custom operations (like `upsertMany`) are added as unexported methods on `dao`.

## ANTI-PATTERNS

- Do not expose the internal `Release` model outside this package — always convert to `types.Release`.
- Do not edit `Handler.daoMap` directly — only `releaseTypeDao()` manages DAO lifecycle.
- Do not add indexes in `GetIndexes()` without considering all release type collections share the same index set.
- Do not skip `nCtx` nil checks — every handler method must guard against nil context.
