# CONFIG POLICY DAO KNOWLEDGE BASE

## Why AGENTS.md?

Shared maintenance rules for this package reduce collaboration cost and regression risk when touching BSON shapes, filters, or conversions.

## OVERVIEW

`configpolicy` is the **MongoDB DAO** for configuration policies: tenant-scoped collections, CRUD/count/list, and conversion between persisted documents and `pkg/types.ConfigPolicy`.

**In scope:** document shape (`RawData` under `data.raw`), query `OptFn` builders, BSON field keys for sort/filter, handler wiring on `base.IOrm`, and unit tests for persistence behavior.

**Out of scope:** policy **matching**, **priority merge**, **workflow**, or **HTTP/proto** handling — those live in `internal/*/storage` and above. This package only persists and queries what callers pass.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `configpolicy.go` | `dao` + `newDao`, embeds `base.IOrm`, auto-assigns `ConfigPolicyID` and `Priority` via `counter` on create, index definition |
| `handler.go` | `IHandler` — `Count` / `List` / `Get` / `Create` / `UpdateMany` / `DeleteMany` / `EnableMany` / `DisableMany`; tenant `sync.Map`; `types` ↔ `ConfigPolicy` conversion |
| `table.go` | `ConfigPolicy` document type, `RawData`, `Scope`; implements `base.IData` (`UniqueFields`, etc.); counter key constants (`tableNamePrefix`, `counterKeyPriority`) |
| `constants.go` | BSON paths for filters/sort (`FieldKey*`, always under `data.raw.*` or `data.version`) |
| `options.go` | `OptFn` filters: `WithBizID`, `WithConfigPolicyType`, `WithEnabled`, `WithEnabledScope`, fuzzy name/operator, etc. |
| `handler_test.go` | DAO integration-style tests for create/update/list |

Related: generic ORM and filters — `pkg/dao/mongo/base` (`AliveFilter`, `OptFn`, `ParsePage`).

## CONVENTIONS

- **Context:** All entrypoints use `contextx.IContext`; call `nCtx.CheckTenantID()` where the handler already does.
- **Filters:** Compose with `base.AliveFilter()` then apply each `OptFn` — same pattern as other `pkg/dao/mongo/*` handlers.
- **Field keys:** Use `FieldKey*` constants for any Mongo path used in sort/filter; they must stay aligned with `RawData` / `table.go` BSON tags (`data.raw.<field>`).
- **Types boundary:** Public API of `IHandler` uses `*types.ConfigPolicy`; conversion stays in `handler.go` (or small helpers there), not in storage/business layers duplicated ad hoc.
- **Comments:** Exported symbols use English comments (project rule).
- **Indexes:** Add or adjust indexes in `dao.GetIndexes()` when new query patterns need them; keep in sync with real usage.

## ANTI-PATTERNS

- Do **not** implement merge-by-priority, default config templates, or “which policy wins” logic here — that belongs in `internal/backend/storage/configpolicy` (or similar).
- Do **not** bypass `TableBroker` / `base.IOrm` patterns established in this package; avoid raw collection access except inside the embedded ORM path.
- Do **not** change BSON/json tags on `RawData` without updating `FieldKey*` constants and any consumers (storage sorts, list filters).
- Do **not** introduce service-specific or HTTP types into this package — keep `pkg/types` + BSON only.
- Do **not** duplicate filter field strings — always go through `constants.go` for `data.raw.*` paths.
