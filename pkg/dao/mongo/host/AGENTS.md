# HOST DAO KNOWLEDGE BASE

## OVERVIEW

`host` is the MongoDB DAO for the host collection. It exposes `IHandler` (multi-tenant facade) and a per-tenant `dao` that embeds `base.IOrm[*Host, Host]`. Inner documents use `base.TableBroker`-style envelopes: `basic.*` plus `data` holding the `Host` payload (`pkg/dao/mongo/host/table.go`). Types round-trip to `pkg/types.Host` in `handler.go`.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `handler.go` | `IHandler` / `handler` — tenant routing, `tenantDao`, CRUD and queries, `convertHostFromTypes` / `convertHostToTypes` |
| `host.go` | Per-tenant `dao`, index definitions, bulk upsert/update param builders (`buildUpsertManyParams`, `buildUpsertStaticManyParams`, `buildUpdateDynamicManyParams`) |
| `table.go` | BSON `Host`, `HostStatic`, `HostDynamic`; `TableHost` broker type |
| `constants.go` | `FieldKey*` paths under `data.*` (and `FieldKeyOperationUpdatedAt` for list sort) |
| `options.go` | `OptFn` wrappers (`WithHostID`, biz/role/status filters, etc.) |
| `handler_test.go` | Integration-style tests (requires MongoDB + `.env`); includes `TouchOperationUpdatedAt` sort checks |
| `README.md` | Package intent and boundaries (Chinese) |

## CONVENTIONS

### Multi-tenant collections

Physical collection name is `TableName(tenantID)` (see `table.go`). Handlers always resolve the tenant from `contextx.IContext` before touching `daoMap`.

### Write paths and `basic.updated_at`

| API | Effect on document |
|-----|---------------------|
| `UpsertMany` | Full `data` replace via `base.BuildUpsertParam` + `basic.updated_at` |
| `UpsertStaticMany` | `$set` static + `basic.updated_at`; `$setOnInsert` dynamic on insert |
| `UpdateDynamicMany` | `$set` dynamic + `basic.updated_at` |
| `UpdateDynamicFields` | Partial dynamic field `$set` + `basic.updated_at` (via field map builder) |
| `TouchOperationUpdatedAt` | `$set` **only** `data.operation_updated_at` via `Collection.UpdateMany`; does **not** change `basic.updated_at` |

### Business operation timestamp (`operation_updated_at`)

- Stored at `data.operation_updated_at` (`FieldKeyOperationUpdatedAt`). DAO `Host` uses `*time.Time` with `omitempty`; `types.Host` uses `time.Time` (zero = unset).
- **Call `TouchOperationUpdatedAt` from user/API or explicit workflow paths** when a business operation should move the host ahead in “recently operated” list ordering. Do **not** call it from CMDB static sync or agent sync jobs.
- Implementation uses `AliveFilter()` + `WithHostID(...)` and a single `UpdateMany` (multi-ID uses `$in`). This intentionally avoids `IOrm.UpdateField`, because `base.buildUpdateField` also sets `basic.updated_at` and would mix sync-time ordering with business ordering.

### Indexes

`GetIndexes()` includes partial indexes on `(static.biz_id, dynamic.node_role, …)` for common list filters. One compound index orders by `FieldKeyOperationUpdatedAt` then `basic.updated_at` to support host list sort (`operation_updated_at desc`, `updated_at desc`).

### Queries

List/count/exist use `OptFn` over `base.AliveFilter()`. Pagination and sort strings are interpreted by `base.ParsePage` (field paths must match `FieldKey*` / `base.FieldKey*`).

## ANTI-PATTERNS

- Do not use `IOrm.UpdateField` to implement “touch business operation time only” — it refreshes `basic.updated_at`.
- Do not assume `UpsertMany` merges nested fields: `BuildUpsertParam` replaces the whole `data` document; omitting optional fields can drop existing keys (see package design docs for CMDB vs full upsert call sites).
- Do not bypass `IHandler` to open arbitrary collections — tenant and table naming must stay consistent.
- Do not pass raw `context.Context` where `contextx.IContext` is required (tenant checks).

## RELATED

- Shared DAO patterns: `pkg/dao/mongo/base/AGENTS.md`
- Domain host type: `pkg/types/host.go`
