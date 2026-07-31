# Backend Storage Knowledge Base

## Overview

Service-level storage orchestration layer between router/manager and `pkg/dao/mongo`. Wraps DAO with metrics, input guards, domain logic, and background monitoring.

## Structure

```
storage/
├── metrics.go              # MetricData + Metric() helper (deprecated, use basestorage.WrapFn)
├── monitor.go              # Prometheus monitor: request_total, request_duration, slow_request
├── cipher/                 # Asymmetric encryption key storage
├── configpolicy/           # Config policy + config policy events
├── credit/                 # Host credential storage (encrypt/decrypt)
├── deploypolicy/           # Deploy policy CRUD + domain discovery/refresh
├── globalsettings/         # Global settings CRUD
├── node/                   # Node deployment + node workflow + workflow monitoring
├── plugin/                 # Plugin deployment/workflow/process/config + workflow monitoring
├── release/                # Agent/proxy/plugin/cert/bintool releases + package events
├── tenant/                 # Tenant data access
├── topo/                   # Topology: host, networkarea, networkunit, accesspoint, events, GSE domain
└── workflow/               # Scheduled workflows, operations, triggers, stop-oper-inst domain
```

## Where to Look

| Task | Location | Notes |
|------|----------|-------|
| Add new storage domain | Create new package under `storage/` | Follow 4-file skeleton below |
| Add DAO method to existing domain | `{domain}/iface.go` + `{domain}/dao_xxx.go` | Interface first, then implement |
| Add domain logic (multi-table) | `{domain}/iface.go` + `{domain}/domain_xxx.go` | `IDomainXxx` interface |
| Understand metric instrumentation | `metrics.go` + `monitor.go` | Legacy pattern; new code uses `basestorage.WrapFn` |
| Service wiring | `internal/backend/service/service.go` | `initialStorages()` creates all Storage instances |
| Capability injection | `internal/backend/options/capability.go` | Holds all `IStorage` refs, passed to router/manager |
| Low-level CRUD | `pkg/dao/mongo/` | DAO handlers per collection |
| Base storage framework | `pkg/basestorage/` | `basestorage.Interface`, `InitStorage`, `WrapFn` |

## Conventions

### File Responsibilities (per package)

| File | Purpose |
|------|---------|
| `iface.go` | `IStorage` (full interface), `IDaoXxx` (single-table), `IDomainXxx` (cross-table domain) |
| `storage.go` | Package entry: `Storage` struct, `NewStorage()`, `initDao()`, `check()`, metric constants, exported methods only |
| `dao_xxx.go` | Single-table data access (unexported implementation methods) |
| `domain_xxx.go` | Cross-table domain logic (unexported implementation methods) |
| `error.go` | Package-specific sentinel errors (optional, only if needed) |

### Interface Hierarchy

```
IStorage
├── basestorage.Interface          # Start/Stop/Health lifecycle
├── IDaoXxx                        # Single-table: Create/List/Get/Update/Delete
└── IDomainXxx                     # Cross-table: business-specific queries
```

- `IStorage` composes all sub-interfaces — used at initialization time.
- `IDaoXxx` / `IDomainXxx` are used as narrow injection targets in router handlers.

### Storage Constructor Pattern

```go
const StorageName = "domain_name"

func NewStorage(client *mongo.Client, database string) (*Storage, error) {
    if client == nil {
        return nil, errors.New("mongo client is nil")
    }
    s := &Storage{
        Storage: basestorage.Storage{
            Name:     StorageName,
            Database: client.Database(database),
        },
    }
    err := basestorage.InitStorage(&s.Storage,
        basestorage.WithStartFunc(s.initDao),
        basestorage.WithCheckFunc(s.check))
    // ...
}
```

### Metric Instrumentation (two patterns)

1. Legacy (older packages like `topo`, `node`):
```go
func (s *Storage) SomeMethod(nCtx contextx.IContext, ...) (result Type, err error) {
    metric := s.metric().Start("operation_name")
    defer metric.End(err)
    result, err = s.someMethodImpl(nCtx, ...)
    return result, err
}
```

2. Modern (newer packages like `deploypolicy`):
```go
func (s *Storage) SomeMethod(nCtx contextx.IContext, ...) (Type, error) {
    var result Type
    err := s.WrapFn(nCtx, "operation_name", func(nCtx contextx.IContext) error {
        // ...
    })
    return result, err
}
```

Prefer `basestorage.WrapFn` for new code — `Metric()` is deprecated.

### Method Naming

| Prefix | Meaning | Maps to DAO |
|--------|---------|-------------|
| `Create` / `CreateMany` | Insert | `Create` / `CreateMany` |
| `Update` / `UpdateMany` | Modify | `Update` / `UpdateMany` |
| `Delete` / `DeleteMany` | Remove | `Delete` / `DeleteMany` |
| `Store` | Insert with transform (encrypt, compress) | Custom |
| `Load` | Read with transform (decrypt, decompress) | Custom |
| `List` | Paginated query | `List` |
| `Get` | Single-item fetch | `Get` |
| `Count` | Count query | `Count` |
| `Distinct` | Distinct field values | `Distinct` |
| `Exist` | Existence check | `Exist` |
| `Upsert` / `UpsertMany` | Insert-or-update | `UpsertMany` |

### Input Guards

Storage methods intercept invalid inputs that callers shouldn't worry about:
- `nil` context → return `basestorage.ErrNilContent()` or `base.ErrInvalidContext()`
- Empty slices → return `nil` (no-op, not error)
- `nil` data → return `basestorage.ErrUpsertNilData()`

### Context Threading

All methods take `contextx.IContext` (not `context.Context`). Provides `BKUsername()`, `TenantID()`, trace propagation.

### Background Monitoring

`node/` and `plugin/` storage packages register scheduler tasks to monitor workflow status. Pattern:
- `registerScheduler()` called in `NewStorage()`
- Periodic tasks poll recent workflows and update status
- Uses `sync.RWMutex` for concurrent workflow map access

## Anti-Patterns

- Do not put exported method implementations in `dao_xxx.go` or `domain_xxx.go` — only `storage.go` has exported methods.
- Do not bypass metric instrumentation — every exported method must record metrics.
- Do not pass raw `context.Context` — always use `contextx.IContext`.
- Do not return errors for empty input slices — silently return `nil`.
- Do not place generic/shared logic here — that belongs in `pkg/dao/mongo` or `pkg/basestorage`.
- Do not use `Metric()` in new code — use `basestorage.WrapFn` instead (see `metrics.go` deprecation comment).

## Unique Styles

- Exported methods in `storage.go` are thin wrappers: metric + guard + delegate to unexported impl in `dao_xxx.go` / `domain_xxx.go`.
- `credit/` uses `Store`/`Load` naming for encrypt/decrypt credential operations.
- `workflow/` has `domain_stop_oper_inst.go` with indexed polling for cross-backend operation instance stopping.
- `plugin/error.go` exists as a placeholder — package-level sentinel errors are optional.
- `globalsettings/` has no `iface.go` — interfaces are defined inline in `globalsettings.go` (legacy, don't follow for new packages).
