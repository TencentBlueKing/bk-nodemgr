# BASESTORAGE KNOWLEDGE BASE

## OVERVIEW

`basestorage` provides the foundational lifecycle management framework for all storage layers in the system. It defines the `Interface` contract (
Start, CheckHealthz, Terminate), the `Storage` base struct with embedded lifecycle hooks, Prometheus metric instrumentation (`Monitor`), and the
`WrapFn` pattern for tracing/metric recording.

Every storage implementation in `internal/*/storage/*` embeds `basestorage.Storage` and uses `InitStorage` to register custom start/check/terminate
hooks.

## WHERE TO LOOK

| File         | Purpose                                                                                                                           |
|--------------|-----------------------------------------------------------------------------------------------------------------------------------|
| `base.go`    | `Interface`, `Storage` struct, `InitStorage`, lifecycle methods (Start, CheckHealthz, Terminate), `WrapFn` tracing/metric wrapper |
| `monitor.go` | `Monitor` struct, `NewMonitor`, `MetricHandle`, `MetricParam`, `MetricOperation` — Prometheus metric recording                    |
| `metrics.go` | Global metric singleton initialization and `Metric()` helper function                                                             |
| `error.go`   | Sentinel error constructors: `ErrUpsertNilData`, `ErrNilContent`, `ErrEmptyUniqueKey`, etc.                                       |
| `README.md`  | Package positioning and core method definitions                                                                                   |

## CONVENTIONS

### Storage Embedding Pattern

All storage implementations embed `basestorage.Storage` and initialize via `InitStorage`:

```go
type MyStorage struct {
basestorage.Storage
// domain-specific fields
}

func New(db *mongo.Database) (*MyStorage, error) {
s := &MyStorage{
Storage: basestorage.Storage{
Name:     "my_storage",
Database: db,
},
}

err := basestorage.InitStorage(&s.Storage,
basestorage.WithStartFunc(s.start),
basestorage.WithCheckFunc(s.check),
basestorage.WithTerminateFunc(s.terminate),
)

return s, err
}
```

### Lifecycle Hook Registration

Three hooks must be registered via `InitStorage`:

- `WithStartFunc(func() error)` — initialization logic (index creation, cache warming, etc.)
- `WithCheckFunc(func() error)` — health check logic (collection existence, data integrity, etc.)
- `WithTerminateFunc(func())` — cleanup logic (close connections, flush buffers, etc.)

### WrapFn Pattern for Operations

All storage operations should use `WrapFn` to automatically record metrics and traces:

```go
func (s *MyStorage) DoSomething(nCtx contextx.IContext, param *Param) error {
return s.WrapFn(nCtx, "DoSomething", func(ctx contextx.IContext) error {
// actual implementation
return nil
})
}
```

**IMPORTANT**: The closure pattern must follow this exact structure:

- Outer function signature: `(nCtx contextx.IContext, ...) error`
- Inner closure signature: `func(ctx contextx.IContext) error`
- Use the closure's `ctx` parameter (not the outer `nCtx`) inside the implementation

### Metric Recording

Metrics are automatically recorded by `WrapFn`:

- `{storage_name}_request_total` — counter for total requests
- `{storage_name}_request_duration` — histogram for request duration (ms)
- `{storage_name}_slow_request` — counter for slow requests (default threshold: 1s)

Labels: `name` (storage name), `operation` (function name), `success` (true/false)

### Context Requirements

- All lifecycle methods (`Start`, `CheckHealthz`) and wrapped operations require `contextx.IContext`
- Do not pass raw `context.Context` — use `contextx.IContext` or `contextx.FromContext(ctx)`
- `WrapFn` automatically creates a traced context via `contextx.FromContext(traceCtx)`

### Scheduler Integration

If a storage needs periodic tasks, embed `scheduler.Scheduler`:

```go
s.Scheduler = scheduler.New()
s.Scheduler.AddTask(...)
```

The scheduler is automatically started in `Start()` and terminated in `Terminate()`.

## ANTI-PATTERNS

- Do not bypass `InitStorage` — all three hooks (start/check/terminate) must be registered.
- Do not call lifecycle methods directly on embedded `Storage` — always call on the concrete storage type.
- Do not use raw `context.Context` in storage operations — use `contextx.IContext`.
- Do not implement storage operations without `WrapFn` — metrics and tracing will be missing.
- Do not violate the `WrapFn` closure pattern — the inner closure must use its own `ctx` parameter, not the outer `nCtx`.
- Do not create custom metric recording logic — `WrapFn` handles it automatically.
- Do not forget to check `IsRunning` before operations — storage may not be initialized.
- Do not call `Start()` multiple times — it will return an error.
- Do not call `Terminate()` on a non-running storage — it will return an error.

## RELATIONSHIP WITH OTHER PACKAGES

- **`pkg/dao/mongo/base`**: Provides ORM and collection-level operations; `basestorage` provides storage-level lifecycle management.
- **`internal/*/storage/*`**: All domain storage implementations embed `basestorage.Storage`.
- **`pkg/contextx`**: Required context type for all storage operations.
- **`pkg/scheduler`**: Optional integration for periodic tasks.
- **`pkg/metrics`**: Underlying Prometheus metric registration (wrapped by `Monitor`).
- **`pkg/logger`**: Used for lifecycle event logging.

## MIGRATION NOTES

Recent unification work (see `openspec/changes/archive/2026-03-04-storage-pattern-unification/`) standardized the `WrapFn` closure pattern across all
storage implementations. If you encounter old patterns like:

```go
// OLD (incorrect)
return s.WrapFn(nCtx, "Op", func (nCtx contextx.IContext) error {
// uses outer nCtx directly
})
```

Refactor to:

```go
// NEW (correct)
return s.WrapFn(nCtx, "Op", func (ctx contextx.IContext) error {
// uses closure's ctx parameter
})
```
