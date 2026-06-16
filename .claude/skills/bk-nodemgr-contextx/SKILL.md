---
name: bk-nodemgr-contextx
description: Use when editing or reviewing bk-nodemgr code that uses pkg/contextx, contextx.IContext, tenant/user/message propagation, FromContext, value-preserving cancel/timeout/deadline helpers, or background work with WithoutCancel.
---

# bk-nodemgr contextx

## Overview

`pkg/contextx` is bk-nodemgr's shared request context model. It embeds Go `context.Context` and adds stable project identity fields, value propagation helpers, and OpenTelemetry span attributes.

Use this skill to preserve tenant, user, login name, message ID, values, trace context, and cancellation semantics through bk-nodemgr service, storage, workflow, async, and third-party paths. Combine it with `golang-context` for generic Go cancellation/deadline rules; keep this skill focused on bk-nodemgr's project contract.

## When to Use

Use this skill when work touches:

- `pkg/contextx/**` implementation, docs, or tests.
- Function signatures using `contextx.IContext` or converting from raw `context.Context`.
- `contextx.New`, `contextx.From`, `contextx.FromContext`, or `contextx.Background`.
- `contextx.WithCancel`, `contextx.WithTimeout`, `contextx.WithDeadline`, or `contextx.WithoutCancel`.
- Tenant, bk username, login name, message ID, or context value propagation.
- Storage, workflow, scheduler, goasync, relay, third-party, logger, or DAO code that depends on project context values.
- Code that considers replacing `contextx.IContext` with raw `context.Context`.

## When Not to Use

Do not use this skill as the primary guide for:

- Generic Go context questions with no bk-nodemgr identity/value propagation.
- Goroutine ownership, pool error handling, or shared-state safety; use the concurrency/gopool skills.
- Retry primitive selection; use `bk-nodemgr-retrier`.
- Logging field/level policy; use `bk-nodemgr-logger` when logging is the main concern.
- Adding service-specific metadata, auth policy, workflow state, tenant rules, or business validation into `pkg/contextx`.

## Source Anchors

Read these before changing behavior:

- `pkg/contextx/AGENTS.md`: package contract, conventions, and anti-patterns.
- `pkg/contextx/README.md`: intent, boundaries, and usage constraints.
- `pkg/contextx/iface.go`: `IContext`, `IContextValues`, and `IContextInfo` interfaces.
- `pkg/contextx/context.go`: constructors, wrappers, span attributes, value preservation, and `WithoutCancel`.
- `pkg/contextx/info.go`: identity fields, `Values`, `GetValue`, `Check*`, and clone behavior.

Representative caller anchors:

- `pkg/goasync/handler.go`: `contextx.WithoutCancel` for async work that outlives request cancellation.
- `pkg/workflow/worker.go`: raw `context.Context` converted into `contextx.IContext`, action timeouts, and background refresh contexts.
- `pkg/workflow/trigger_handler.go`: workflow trigger paths using project context.
- `pkg/scheduler/AGENTS.md`, `pkg/basestorage/AGENTS.md`, `internal/backend/storage/AGENTS.md`: package-level context expectations.
- DAO and third-party call sites under `pkg/dao/**`, `pkg/thirdparty/**`, `pkg/relayhandler/**`, and `pkg/rest/**`.

## Quick Reference

| Need | Use | Contract |
| --- | --- | --- |
| Project context with tenant/user/message fields | `contextx.IContext` | Required where identity metadata is part of the call contract |
| Build from raw context | `contextx.New` or `contextx.FromContext` | Centralizes span attributes and project info extraction |
| Derive from existing project context | `contextx.From` | Copies tenant, bk username, login name, message ID, and values before applying overrides |
| Add cancellation | `contextx.WithCancel` | Preserves project values while deriving cancellable context |
| Add timeout/deadline | `contextx.WithTimeout` / `contextx.WithDeadline` | Prefer over stdlib wrappers on `IContext` paths when values must survive |
| Background work outliving request | `contextx.WithoutCancel` | Keeps project info while detaching cancellation from the parent request |
| Validate required identity | `CheckTenantID`, `CheckBKUsername`, `CheckLoginName`, `CheckMessageID` | Call at validation boundaries instead of assuming fields exist |
| Generic cancellation theory | `golang-context` | Use as companion background, not a replacement for project context rules |

## Core Patterns

### Convert raw entry contexts at the boundary

Some framework callbacks only provide `context.Context`. Convert once, then pass `contextx.IContext` through project-owned layers.

```go
nCtx := contextx.New(ctx, contextx.WithMessageID(messageID))
if err := stg.GetOperationInstanceBriefData(nCtx, operationInstanceID); err != nil {
    return fmt.Errorf("get operation instance brief data: %w", err)
}
```

Do not replace existing `contextx.IContext` boundaries with raw `context.Context` when storage, logger, tenant checks, or third-party calls need project metadata.

### Preserve values when deriving cancellation or timeout

```go
timeoutCtx, cancel := contextx.WithTimeout(contextx.From(nCtx), actionDef.Timeout())
defer cancel()
```

Use stdlib `context.WithTimeout` only when the downstream path does not need `Values`, identity helpers, or project context behavior.

### Detach cancellation only for deliberate async work

```go
task := &Task{
    nCtx:  contextx.WithoutCancel(nCtx),
    runFn: runFn,
}
```

`WithoutCancel` is for work that must outlive the incoming request while retaining project identity and trace/log context. It is not a shortcut for ignoring cancellation in request-bound work.

## Common Mistakes

- Replacing `contextx.IContext` with raw `context.Context` in service/storage paths that require tenant, user, login name, or message ID.
- Calling plain `context.WithTimeout` / `context.WithDeadline` on an `IContext` path and losing project value semantics for downstream code.
- Constructing context info ad hoc instead of using `New`, `From`, or `FromContext`.
- Assuming `TenantID`, `BKUsername`, `LoginName`, or `MessageID` always exist; use the `Check*` guards at validation boundaries.
- Adding service-specific metadata, auth decisions, workflow states, or business validation rules to `pkg/contextx`.
- Duplicating OpenTelemetry attribute injection in callers instead of keeping it centralized in `context.go`.
- Treating this skill as generic Go context guidance and skipping source-anchor searches.

## Verification Checklist

- Search project usage: `rg "contextx\.(IContext|New|From|FromContext|WithCancel|WithTimeout|WithDeadline|WithoutCancel|Background)|pkg/contextx" internal pkg`.
- Confirm whether the path requires tenant/user/message metadata before changing signatures from `contextx.IContext` to raw `context.Context`.
- Confirm derived contexts preserve project values when downstream logger/storage/third-party code expects them.
- Confirm every cancel function from `WithCancel`, `WithTimeout`, or `WithDeadline` is called on all paths unless ownership is explicitly transferred.
- Confirm required identity fields are checked at request/service/storage boundaries, not assumed deep inside helpers.
- If changing `pkg/contextx` behavior, inspect representative workflow, goasync, storage, logger, and third-party callers.

## Evals

Pressure prompts live in `evals/evals.json`. Keep run outputs, timing, grading, benchmarks, and review artifacts outside git unless explicitly requested.

## Cross-References

- `golang-context`: generic Go cancellation, timeout, deadline, and context value rules.
- `golang-concurrency`: goroutine ownership and cancellation observation.
- `bk-nodemgr-gopool`: grouped goroutine error handling and caller-owned context propagation.
- `bk-nodemgr-retrier`: retry loops that must respect project context cancellation and timeout.
- `bk-nodemgr-logger`: project logging conventions when context values flow into logs.
