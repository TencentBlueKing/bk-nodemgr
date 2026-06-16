---
name: bk-nodemgr-conv
description: Use when editing or reviewing bk-nodemgr code that uses pkg/runtime/conv, manual map/slice/struct/string/number/bool coercion, proto or workflow payload shaping, or duplicate conversion helpers.
---

# bk-nodemgr conv

## Overview

`pkg/runtime/conv` owns portable, business-agnostic conversion helpers for bk-nodemgr. Use this skill to keep conversion behavior centralized, stable, and source-anchored instead of adding caller-local helpers with subtly different nil, overflow, duplicate-key, or ordering semantics.

## When to Use

Use this skill when work touches:

- `pkg/runtime/conv/**` implementation, docs, or tests.
- Manual conversion between primitives, strings, numbers, bools, structs, maps, slices, or pointers.
- Workflow/action payload shaping such as `conv.MapToStruct(ctx.Data.Content, param)`.
- Proto/API/third-party data shaping where the conversion itself is reusable and business-neutral.
- Caller helpers that duplicate `ToInt64`, `MapToStruct`, `StructToMap`, `SliceToMap`, `SliceToSlice`, `SliceUnique`, `SliceIntersect`, `MapKeyToSlice`, `MapValueToSlice`, `StringToBool`, `ToBool`, or `NonEmptyOr`.

## When Not to Use

Do not route these decisions to `pkg/runtime/conv`:

- Business enum mapping, workflow state transitions, auth policy, tenant/product rules, or frontend/API-specific defaulting.
- Validation that depends on domain meaning rather than type shape.
- Logging, config loading, storage, network calls, or cross-service imports.
- Compatibility shims for one caller when the behavior is not portable.

## Source Anchors

Read these before changing behavior:

- `pkg/runtime/conv/AGENTS.md`: hard package boundary and semantic constraints.
- `pkg/runtime/conv/README.md`: package intent and user-facing scope.
- `pkg/runtime/conv/conv.go`: current helper implementations.
- `pkg/runtime/conv/conv_test.go`: nil, pointer, overflow, duplicate-key, ordering, and panic-recovery contracts.
- `.claude/skills/code-review/references/go-standards.md`: project review expectation to prefer common conversion helpers.

Representative caller anchors:

- `internal/backend/auth/scope.go`: `ToInt64`, `SliceUnique`, `SliceIntersect` for auth scope shaping while auth rules remain outside conv.
- `internal/backend/manager/workflowdef/node/action_select_relay_host.go`: `MapToStruct`, `SliceToSlice`, and `SliceToMap` in workflow action data flow.
- `pkg/proto/**`, `internal/*`, `pkg/thirdparty/**`: broad conversion call sites; search before inventing helpers.

## Quick Reference

| Need | Prefer | Contract to Check |
| --- | --- | --- |
| Mixed numeric/string value to `int64` | `ToInt64`, `ToInt64Default` | nil, pointer, json.Number, overflow, NaN/Inf |
| Request/action map into params struct | `MapToStruct` | destination must be struct pointer; caller handles returned error |
| Struct params into map | `StructToMap` or `StructToMapIgnoreError` | use ignore variant only when the conversion is known safe |
| Unique/intersection slice shaping | `SliceUnique`, `SliceIntersect` | first-seen and left-side order stability |
| Slice to map by key | `SliceToMap` | duplicate key returns error; panic recovery returns nil map + error |
| Merge maps | `MapUnion`, `MapUnionIgnoreConflict` | conflict error vs right-side overwrite |
| Map keys or values as slice | `MapKeyToSlice`, `MapValueToSlice` | key order is sorted where defined |
| String/bool/default coercion | `StringToBool`, `ToBool`, `ToBoolDefault`, `NonEmptyOr` | accepted text/numeric forms and default semantics |

## Core Patterns

### Keep domain rules at the caller boundary

Use `conv` for shape conversion, then apply business decisions in the owning layer.

```go
param := new(ActionParam)
if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
    return fmt.Errorf("parse action param: %w", err)
}

// Business validation stays here, not in pkg/runtime/conv.
if param.Policy == "" {
    return errors.New("policy is required")
}
```

### Prefer existing helpers over caller-local loops

Before adding a new helper, search for an existing conversion shape:

```bash
rg "conv\.SliceToMap|conv\.SliceToSlice|conv\.MapToStruct|conv\.ToInt64" internal pkg
```

Add to `pkg/runtime/conv` only when the behavior is portable across modules and can be specified without domain nouns.

## Common Mistakes

- Adding API or workflow-specific defaulting to `pkg/runtime/conv`. Put that in the handler, service, workflow action, or proto converter that owns the domain contract.
- Using `StructToMapIgnoreError` because it is convenient. Use it only when errors are structurally impossible or already proven by tests.
- Reimplementing slice/map helpers in callers, then drifting on duplicate-key or ordering behavior.
- Changing string/error contracts without updating `conv_test.go` and caller expectations.
- Treating nil and empty slices/maps as interchangeable without checking the existing helper contract.

## Verification Checklist

- Search for analogous usage before adding or changing helpers: `rg "pkg/runtime/conv|conv\." internal pkg test`.
- If behavior changes, update `pkg/runtime/conv/conv_test.go` in the same change.
- Check nil, pointer chains, overflow/underflow, duplicate keys, order stability, and recovered panic paths.
- Keep imports dependency-light; do not import `internal/*`, project service packages, storage, logging, or network clients.
- For caller changes, verify errors are wrapped at the caller boundary and domain validation remains outside `conv`.

## Evals

Pressure prompts live in `evals/evals.json`. Keep run outputs, timing, grading, benchmarks, and review artifacts outside git unless explicitly requested.

## Cross-References

- `golang-data-structures`: slice/map semantics and allocation tradeoffs.
- `golang-safety`: nil, panic, overflow, and defensive copying concerns.
- `golang-testing`: table-driven coverage when changing helper behavior.
- `code-review`: project-level review checklist that already treats `pkg/runtime/conv` as a standard conversion surface.
