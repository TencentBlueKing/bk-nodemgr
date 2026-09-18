---
name: bk-nodemgr-gopool
description: Use when editing or reviewing bk-nodemgr code that uses pkg/runtime/gopool, goroutine fan-out/fan-in, bounded parallel execution, panic-to-error behavior, Wait handling, or shared result collection.
---

# bk-nodemgr gopool

## Overview

`pkg/runtime/gopool` is bk-nodemgr's small wrapper around `errgroup.Group` for grouped goroutine execution. It centralizes panic-to-error behavior, but it does not own cancellation, lifecycle shutdown, logging policy, or safe shared-state design.

Use this skill to keep fan-out/fan-in code explicit about `Go`, `Wait`, `SetLimit`, error propagation, and caller-owned concurrency safety.

## When to Use

Use this skill when work touches:

- `pkg/runtime/gopool/**` implementation or tests.
- Parallel backend calls in handlers or BFF-style aggregation.
- Service startup/shutdown fan-out where `gp.Wait()` returns aggregated errors.
- Workflow, file transfer, third-party, or manager code that launches multiple goroutines for one task.
- Code that ignores `gp.Wait()` or writes shared results from goroutines.
- Panic handling expectations for goroutines launched through `gopool.Go`.

## When Not to Use

Do not use `gopool` as a substitute for:

- Long-lived worker lifecycle management that needs explicit start/stop/drain semantics.
- Context cancellation propagation; pass and honor context in the goroutine body yourself.
- Safe shared-state design; protect shared maps/slices/counters in the caller.
- Fire-and-forget background jobs where no caller can own `Wait` and error handling.
- Generic Go concurrency teaching; anchor guidance in bk-nodemgr call sites.

## Source Anchors

Read these before changing behavior:

- `pkg/runtime/gopool/gopool.go`: `Pool`, `NewPool`, `Go`, `Wait`, `SetLimit`, and panic recovery.
- `pkg/runtime/gopool/gopool_test.go`: current panic-to-error and limit behavior coverage.
- `pkg/rest/server/server.go`: server start/shutdown fan-out returning `gp.Wait()`.
- `pkg/workflow/trigger_handler.go`: workflow fan-out with `gp.Wait()` error handling.
- `internal/application/router/AGENTS.md`: router guidance for parallel backend calls in BFF handlers.
- Representative callers under `internal/*/service`, `internal/backend/manager/**`, `internal/file/manager/**`, `pkg/thirdparty/file`, and workflow actions.

## Quick Reference

| Need | Use | Caller Responsibility |
| --- | --- | --- |
| Launch grouped goroutine | `gp.Go(func() error { ... })` | Return meaningful error; wrap/log at owning layer |
| Wait for group completion | `gp.Wait()` | Handle or intentionally ignore with a local reason |
| Bound active goroutines | `gp.SetLimit(n)` | Set before active goroutines; avoid changing limit while running |
| Panic conversion | built into `Go` | Treat returned panic error as task failure from `Wait` |
| Cancellation | context inside goroutine body | `gopool` does not derive/cancel context for siblings |
| Shared results | caller-owned synchronization | Use indexed writes, mutex, channels, or other safe ownership |

## Core Patterns

### Return `Wait` unless the work is explicitly best-effort

```go
gp := gopool.NewPool()
for _, item := range items {
    item := item
    gp.Go(func() error {
        return handle(ctx, item)
    })
}
if err := gp.Wait(); err != nil {
    return fmt.Errorf("handle items: %w", err)
}
```

Ignoring `Wait` is only acceptable when every goroutine handles its own errors and the caller documents why the result is best-effort.

### Keep cancellation and shared state outside the pool

`gopool` starts goroutines and aggregates errors. It does not make shared state safe.

```go
results := make([]Result, len(inputs))
for i, input := range inputs {
    i, input := i, input
    gp.Go(func() error {
        result, err := fetch(ctx, input)
        if err != nil {
            return err
        }
        results[i] = result // safe because each goroutine owns one index
        return nil
    })
}
```

For append, maps, counters, or multi-field mutation, add explicit synchronization or redesign ownership.

## Common Mistakes

- Calling `gp.Go` and returning before `gp.Wait` without documenting best-effort semantics.
- Expecting `gopool` to cancel sibling goroutines on error. The wrapper uses `errgroup.Group`, not `errgroup.WithContext`.
- Calling `SetLimit` after goroutines are active; upstream `errgroup` says the limit must not be modified while active goroutines exist.
- Mutating shared slices or maps from goroutines because `gopool` feels like a "safe pool". Only launch/error behavior is wrapped.
- Swallowing panic-derived errors from `Wait` and losing the stack snippet that `gopool.Go` preserved.
- Using `gopool` for service lifecycle daemons that need explicit shutdown contracts.

## Verification Checklist

- Search for analogous fan-out usage: `rg "gopool\.NewPool|\.Go\(func\(\) error|\.Wait\(\)" internal pkg`.
- Confirm every `Go` body returns errors intentionally and does not hide failures unless best-effort behavior is documented.
- Confirm `Wait` is handled, returned, or intentionally ignored with a local reason.
- Check shared result writes for index ownership, mutexes, channels, or other safe synchronization.
- Check context propagation inside goroutine bodies; do not assume pool-level cancellation.
- If changing pool semantics, update `pkg/runtime/gopool/gopool_test.go` and inspect representative REST/workflow/service callers.

## Evals

Pressure prompts live in `evals/evals.json`. Keep run outputs, timing, grading, benchmarks, and review artifacts outside git unless explicitly requested.

## Cross-References

- `golang-concurrency`: goroutine ownership, synchronization, and fan-out/fan-in patterns.
- `bk-nodemgr-contextx`: project context propagation and value-preserving cancellation/deadline handling inside goroutine bodies.
- `golang-error-handling`: wrapping and returning task errors.
- `golang-safety`: panic prevention, race avoidance, and shared mutable state.
