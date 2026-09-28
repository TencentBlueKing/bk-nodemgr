---
name: bk-nodemgr-error-handling
description: Use when writing, reviewing, refactoring, or debugging Go error handling in bk-nodemgr, especially code with err, fmt.Errorf, %w, resterrf.ErrWrap, errors.Is/As, logger.WithErr, router/service/storage error propagation, or any err == nil check. This skill owns bk-nodemgr error semantics; use bk-nodemgr-logger for logger API, Biz/Sys, fields, and levels.
---

# bk-nodemgr Error Handling

## Overview

This skill is the sole project owner for Go error handling in bk-nodemgr. It supersedes generic Go error-handling advice when code lives in this repository.

Core principle: errors are values. Error paths must be explicit `if err != nil` guard clauses; success stays on the mainline and is never wrapped in an `err == nil` branch.

## Ownership Boundary

This skill owns:

- Error creation, wrapping, inspection, aggregation, and propagation.
- The absolute `err == nil` ban and migration rules.
- REST error mapping through `resterrf.ErrWrap` and `ErrUnwrap`.
- The decision of whether the current layer should log, return, wrap, map, or recover.
- Panic/recover boundaries when they affect error semantics.

`bk-nodemgr-logger` owns logger API and formatting details: `Biz` vs `Sys`, context fields, `WithErr`, `With`, `WithDuration`, message shape, level choice, adapters, and logger noise control. Use this skill to decide whether an error should be logged at the current boundary; use `bk-nodemgr-logger` to decide exactly how to log it.

## When to Use

Use this skill for Go coding, review, refactoring, debugging, or audits involving:

- `err`, `error`, `fmt.Errorf`, `%w`, `errors.Is`, `errors.As`, or `errors.Join`.
- `resterrf.ErrWrap`, `errf.ErrWrap`, `ErrUnwrap`, REST response errors, or permission errors.
- `logger.G.*().WithErr(err)` when deciding logging responsibility or log-once behavior.
- Router -> service -> storage error propagation.
- `err == nil`, hidden nil-error helpers, ignored errors, panic/recover, or log-and-return.

Do not use this skill as the only guide for logger field naming or level formatting; load `bk-nodemgr-logger` too. Do not use it for readability-only control-flow questions that do not change error semantics; load `bk-nodemgr-code-style` too.

## Source Anchors

Read the relevant anchor before changing or judging a matching surface:

| Surface                  | Anchor                                                                        | Contract                                                                                              |
| ------------------------ | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| Repository hard rule     | `AGENTS.md`                                                                   | Highest-priority project constraints and scoped instruction policy                                    |
| REST error code chain    | `pkg/rest/errf/error.go`                                                      | `ErrWrap` joins REST code error with cause; `ErrUnwrap` maps returned errors to response code/details |
| REST handler boundary    | `pkg/rest/server/handler.go`                                                  | Handler return error is centrally unwrapped and converted to response                                 |
| REST response body       | `pkg/rest/server/request.go`                                                  | Error details and permission data are exposed from unwrapped errors                                   |
| Logger error field       | `pkg/logger/logger.go`, `pkg/logger/iface.go`                                 | `WithErr` records `err` as structured field; do not format the same error into message                |
| Mature router boundary   | `internal/backend/router/api-v3/node/agent/install.go`                        | Request decode/auth/service failures log once and return `resterrf.ErrWrap`                           |
| Similar router boundary  | `internal/backend/router/api-v3/node/proxy/install.go`                        | Confirms agent/proxy handler consistency                                                              |
| Service/system boundary  | `internal/backend/service/service.go`                                         | Startup/system operations wrap with `%w` and log with `Sys().WithErr` at responsibility boundaries    |
| Retry/poll migration     | `pkg/runtime/retrier/polling.go`, `pkg/runtime/retrier/expo_backoff.go`       | Historical success-on-nil patterns require semantic migration, not mechanical replacement             |
| Probe/fallback migration | `pkg/runtime/crypter/rsa.go`, `internal/backend/auth/v3/provider/provider.go` | Try-parse fallback needs `(value, ok)` or domain helpers, not hidden `err == nil`                     |
| Error chain tests        | `pkg/rest/server/request_test.go`                                             | Permission and wrapped-error behavior must stay stable                                                |

## Quick Reference

| Situation                        | Do                                                                | Do Not                                                   |
| -------------------------------- | ----------------------------------------------------------------- | -------------------------------------------------------- |
| Normal error branch              | `if err != nil { return ... }`                                    | `if err == nil { ... } else { return ... }`              |
| Lower-layer failure              | `return fmt.Errorf("failed to fetch host: %w", err)`              | `return err` or `fmt.Errorf("failed: %v", err)`          |
| REST handler boundary            | `return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)`    | Return raw implementation errors from API handlers       |
| Error inspection                 | `errors.Is(err, target)` / `errors.As(err, &target)`              | Direct equality or type assertion against wrapped errors |
| Current boundary handles failure | `logger.G.Biz(ctx).WithErr(err).Error("failed to install agent")` | Log the same error again in every lower layer            |
| Logger message with error        | `WithErr(err).Error("failed to update host")`                     | `WithErr(err).Error("failed: %v", err)`                  |
| Parallel independent failures    | `errors.Join(errs...)` with stable wrapping context               | Concatenate error strings or discard all but one cause   |
| Expected business failure        | Return typed/sentinel/domain error and map at boundary            | `panic` or `recover` for normal control flow             |

## Absolute Rule: No `err == nil`

`err == nil` is prohibited repository-wide. This applies to production code, tests, the `tools` module, retry/poll loops, helper predicates, and generated surfaces. Existing occurrences are migration debt, not precedent and not exemptions.

Do not bypass this rule by:

- Hiding the comparison in helpers such as `isNilErr(err)` or `noError(err)`.
- Assigning `ok := err == nil`, `succeeded := err == nil`, or equivalent boolean aliases.
- Renaming `err` to avoid the grep.
- Adding `nolint`, build tags, comments, or generated-code excuses.
- Discarding errors with `_` or resetting an error to `nil` to keep a success branch.

Helpers are allowed only when they expose real domain or probing semantics, such as `parseLoginURL(raw) (url.URL, bool)` or `fileExists(path) bool`. The helper name must not merely restate error nilness.

## Replacement Patterns

| Original Intent         | Forbidden Shape                                            | Required Shape                                                                                                    |
| ----------------------- | ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Fail fast, then succeed | `if err == nil { success } else { return err }`            | `if err != nil { return ... }` followed by success mainline                                                       |
| Try parse one format    | `if value, err := parse(raw); err == nil { return value }` | `value, ok := parseRaw(raw); if ok { return value }` where helper owns probing semantics                          |
| Existence check         | `if _, err := os.Stat(path); err == nil { ... }`           | `if fileExists(path) { ... }`; helper must handle `os.IsNotExist` and unexpected errors explicitly                |
| Test expects success    | `if err == nil { ... }`                                    | `require.NoError(t, err)` or existing project assertion style, then assert returned values                        |
| Test expects failure    | `if err == nil { t.Fatal(...) }`                           | `require.Error(t, err)` or explicit `if err != nil` branch only when additional error checks follow               |
| Retry/poll success      | `if err == nil { return nil }`                             | Error-first loop that continues or returns on `err != nil`, or a domain helper returning `(done bool, err error)` |
| REST nil means OK       | `if err == nil { return OK }`                              | Keep nil-as-OK contract, but migrate implementation with explicit non-nil branch and final OK return              |
| Error chain inspection  | `if err == SomeErr`                                        | `if errors.Is(err, SomeErr)`                                                                                      |
| Typed error inspection  | `permErr := err.(PermissionError)`                         | `var permErr PermissionError; if errors.As(err, &permErr) { ... }`                                                |
| Lossy wrapping          | `fmt.Errorf("failed: %v", err)`                            | `fmt.Errorf("failed to do x: %w", err)`                                                                           |

## Error Flow by Layer

| Layer                                                  | Responsibility                                                                                                                               |
| ------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/runtime/**`, `pkg/types`, conversion helpers      | Validate local contract and return wrapped errors with context. Do not log unless the package owns final handling.                           |
| `pkg/dao/**`, storage helpers                          | Preserve database/DAO cause with `%w`; define stable not-found/permission semantics for callers. Avoid duplicate `ERROR` logs.               |
| `internal/*/service`, managers, workflow orchestration | Add current operation context with `%w`; decide whether a background/system flow is the handling boundary.                                   |
| `internal/*/router/api-v3`                             | Bind/validate/auth/call service, log request-boundary failures once, and map to `resterrf.ErrWrap`.                                          |
| `pkg/rest/server`                                      | Central response conversion through `ErrUnwrap`, `AbortWithJSONError`, and `AbortWithJSONPermDenied`. Do not bypass this contract.           |
| `cmd/*` startup/shutdown                               | Startup initialization failures may log with `Sys().WithErr` or return fatal startup errors. Business request handling must not use `Fatal`. |

## REST Boundary

API handlers should return project REST errors, not raw implementation errors:

```go
if err := rCtx.BindJSON(req); err != nil {
    logger.G.Biz(rCtx).WithErr(err).Error("failed to install agent, failed to decode request body")
    return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
}
```

Use `resterrf.InvalidParameter` for request boundary validation, `resterrf.PermissionDenied` for auth failures, `resterrf.BackendOperateFailed` for backend operation failures, and narrower project codes when the existing module already uses them. Search analogous handlers before choosing a code.

`ErrWrap` is a single-layer REST mapping. Do not wrap repeatedly at multiple layers just because an error crosses multiple functions; add `%w` context internally, then map once at the transport boundary.

## Logging Boundary

Log only at a boundary that owns handling responsibility:

- Request/business boundary: use `logger.G.Biz(ctx).WithErr(err).Error(...)` when the request flow cannot continue.
- System/background boundary: use `logger.G.Sys()` or `logger.G.Sys().Ctx(ctx)` when no user request owns the flow.
- Lower layers: return `fmt.Errorf("context: %w", err)` and let the caller decide.
- Recoverable condition: use `Warn` only when the flow can recover without outside intervention.
- Interrupted or unrecoverable flow: use `Error`.
- Business request handling: never use `Fatal`.

Do not log and return the same error at every layer. If a lower layer adds context and returns, the upper boundary should be the place that logs. If a boundary logs because it consumes or terminates the flow, do not force another caller to log the same chain again.

## Message and Wrapping Rules

- Error messages and log messages are English, stable, lowercase unless a proper noun requires otherwise, and do not end with punctuation.
- Use action-oriented context: `failed to fetch networkunits`, `invalid metadata`, `permission denied`.
- Wrap internal causes with `%w`, not `%v` or `%s`.
- At logger call sites, prefer `WithErr(err)`; do not duplicate the error in the message.
- Error-code fields must include both `err-code` and `err-msg` when they are logged together.
- Keep high-cardinality values in structured fields, not in the stable message.

## Migration Catalog for Existing `err == nil`

Every `err == nil` match must receive exactly one classification before migration:

| Category                  | Typical Location                                                | Migration Rule                                                                                                                       |
| ------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| Production success branch | Router/service/storage/business logic                           | Rewrite to `if err != nil` guard and keep success on mainline                                                                        |
| Retry/poll success        | `pkg/runtime/retrier`, installer retry loops                    | Introduce explicit `(done bool, err error)` semantics or restructure the loop without a nil-error success branch                     |
| Probe/fallback parse      | `json.Unmarshal`, `time.Parse`, `strconv.Parse*`, `x509.Parse*` | Move probing into a domain helper returning `(value, ok)`; helper must handle unexpected errors intentionally                        |
| Existence/cache hit       | `os.Stat`, cache `Get`, optional resources                      | Move into domain helper such as `fileExists`, `loadCachedToken`, or `resourceAvailable`; distinguish not-found from unexpected error |
| Tests/assertions          | `_test.go`                                                      | Use `require.NoError`, `require.Error`, `assert.NoError`, `assert.Error`, or existing local assertion style                          |
| REST nil contract         | `ErrUnwrap(nil)`-style API contracts                            | Preserve external semantics; implement with non-nil branch first and final OK/default return                                         |
| Generated code            | `pkg/proto/**` generated output                                 | Do not hand-edit generated files; fix generator/source if a migration task explicitly approves it                                    |
| False positive            | Comments or non-error text                                      | Remove if stale or classify as no code change; do not count it as compliant code                                                     |

Migration work must be planned by category. Do not run a repository-wide regex replacement.

## Common Mistakes

| Mistake                                        | Why It Fails                                                        | Fix                                                        |
| ---------------------------------------------- | ------------------------------------------------------------------- | ---------------------------------------------------------- |
| `if err == nil { ... }`                        | Success branch hides error handling and pushes happy path rightward | Guard `if err != nil` first, then continue mainline        |
| `return err` from lower layer                  | Caller loses operation context                                      | `return fmt.Errorf("failed to ...: %w", err)`              |
| `%v` while wrapping                            | Error chain is lost; `errors.Is/As` cannot work                     | Use `%w` for internal wrapping                             |
| `logger.WithErr(err).Error("failed: %v", err)` | Error is duplicated and message becomes unstable                    | `WithErr(err).Error("failed to ...")`                      |
| Log-and-return at every layer                  | Same failure appears as multiple `ERROR` logs                       | Lower layers wrap; owning boundary logs once               |
| Direct comparison with wrapped errors          | Joined/wrapped errors are not matched                               | Use `errors.Is` or `errors.As`                             |
| Panic for expected failures                    | Breaks normal project error mapping and request handling            | Return error and map at boundary                           |
| Fix generated `err == nil` by hand             | Generated files drift from source                                   | Change generator/source only in a dedicated migration task |

## Verification Checklist

Before completing a change that touches Go error handling:

```bash
rg -n "err\s*==\s*nil" --glob '*.go'
rg -n "fmt\.Errorf\([^\n]*%[vs]" cmd internal pkg tools --glob '*.go'
rg -n "WithErr\([^\)]*\)\.Error\([^\n]*%[wv].*err" cmd internal pkg tools --glob '*.go'
rg -n "logger\.G\.(Biz|Sys).*WithErr|resterrf\.ErrWrap|errf\.ErrWrap" cmd internal pkg tools --glob '*.go'
```

Interpretation rules:

1. Any `err == nil` introduced or modified by the current diff is blocking and must be migrated.
2. Existing untouched matches are migration debt; report them only when relevant to the current review.
3. Every migration match must be classified with the catalog above before editing.
4. Raw match count must equal the sum of classified categories in a migration task.
5. Do not mechanically replace `err == nil`; preserve the original contract and add tests only when the user requested behavior migration or a real bugfix provides the scenario.
6. Check `%w`, `errors.Is/As`, `resterrf.ErrWrap`, and `WithErr` consistency against analogous local code.
7. If logging behavior changes, load `bk-nodemgr-logger` and verify Biz/Sys, fields, levels, and message format there.

## Eval Prompts

Use these prompts when validating this skill later. Do not create eval workspaces unless explicitly asked.

1. Review a new `internal/backend/router/api-v3` handler that uses `if err == nil { ... } else { return resterrf.ErrWrap(...) }`; require a blocking finding and guard-style rewrite.
2. Review a helper that returns `fmt.Errorf("failed: %v", err)`; require `%w`, operation context, and chain-safe inspection guidance.
3. Review a handler that calls `logger.G.Biz(rCtx).WithErr(err).Error("failed: %v", err)` and then returns `resterrf.ErrWrap`; require stable message and log-once boundary analysis.
4. Refactor a `json.Unmarshal` probing fallback to remove `err == nil` without changing fallback semantics.
5. Review a `_test.go` file with `if err == nil { t.Fatal(...) }`; require project assertion style and no `err == nil` exception for tests.
6. Review parallel workflow code that discards all but one error; require `errors.Join` only when failures are independent and caller can inspect the chain.

## Cross-References

- `bk-nodemgr-how-to`: project-first skill routing; should load this skill before generic Go error guidance.
- `bk-nodemgr-code-style`: readability, function shape, and happy-path linearity around error branches.
- `bk-nodemgr-logger`: logger API, Biz/Sys choice, structured fields, levels, adapters, and message formatting.
- `bk-nodemgr-gopool`: fan-out/fan-in behavior and panic-to-error contracts; this skill owns returned error semantics.
- `bk-nodemgr-retrier`: retry primitive selection; this skill owns removal of `err == nil` success branches.
- `code-review`: PR review workflow and severity reporting; this skill owns the error-handling contract being checked.
- `golang-error-handling`: generic Go fallback only outside bk-nodemgr or for language-level details not covered here.
