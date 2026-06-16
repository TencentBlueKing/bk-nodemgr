---
name: bk-nodemgr-retrier
description: Use when editing or reviewing bk-nodemgr code that uses pkg/runtime/retrier, polling loops, exponential backoff, fallback candidates, transient third-party retries, context-aware retry behavior, or retry observability.
---

# bk-nodemgr retrier

## Overview

`pkg/runtime/retrier` owns portable retry primitives for bk-nodemgr. It provides three different tools: `Polling` for condition checks until success or timeout, `ExpoBackoff` for bounded transient retries, and `Fallback` for ordered candidate attempts.

Use this skill to choose the right primitive, preserve existing context/error semantics, and keep business policy, logging decisions, and domain terminal-state handling in callers.

## When to Use

Use this skill when work touches:

- `pkg/runtime/retrier/**` implementation, docs, or tests.
- Workflow actions that wait for remote process, GSE, sub-workflow, or identifier states.
- Third-party, relay, or SSH calls that need finite transient retry.
- Ordered candidate fallback such as relay endpoints or other interchangeable targets.
- Retry callbacks, attempt counters, timeout behavior, context cancellation, or caller-side logging.
- Code that considers replacing an existing custom loop with `Polling`, `ExpoBackoff`, or `Fallback`.

## When Not to Use

Do not force `pkg/runtime/retrier` onto:

- Infinite loops with stateful delay reset after intermittent success.
- Business recovery workflows, terminal-state decisions, or action-specific status transitions.
- Non-idempotent operations unless the caller proves retry safety.
- Logging/reporting policy that belongs to the caller or `pkg/logger` usage.
- Generic retry advice not anchored in bk-nodemgr source paths.

## Source Anchors

Read these before changing behavior:

- `pkg/runtime/retrier/README.md`: package intent, idempotency warning, and caller logging boundary.
- `pkg/runtime/retrier/polling.go`: `PollingOpts`, default timeout/interval, attempt-0 behavior, timeout and context errors.
- `pkg/runtime/retrier/expo_backoff.go`: bounded retry count, exponential delay, max delay, jitter, and final error shape.
- `pkg/runtime/retrier/fallback.go`: ordered candidates, validator, callbacks, and last-error wrapping.
- `pkg/runtime/retrier/*_test.go`: current behavior for success, all failures, context cancellation, zero retries, callbacks, and invalid candidates.

Representative caller anchors:

- `internal/backend/manager/workflowdef/node/action_wait_gse_ready.go`: polling workflow state with action timeout.
- `internal/backend/manager/workflowdef/node/action_install_pre_ordered_plugins.go`: workflow action retry/wait usage.
- `internal/backend/manager/workflowdef/node/utils/ssh_installer_poller.go`: explicit example of not using `ExpoBackoff` for indefinite reset-on-success polling.
- `pkg/thirdparty/gse/**`: finite transient third-party retries.
- `pkg/relayhandler/**` and `pkg/sshx/sshx.go`: relay/SSH retry call sites.

## Quick Reference

| Need | Primitive | Key Contract |
| --- | --- | --- |
| Wait until a condition succeeds or times out | `Polling` | attempt 0 runs immediately; timeout returns `reach max timeout`; ctx cancellation returns `ctx.Err()` |
| Retry a transient operation a bounded number of times | `ExpoBackoff` | finite attempts, capped exponential delay, positive jitter, no sleep after final attempt |
| Try ordered candidates with optional validity checks | `Fallback[T]` | nil validator means all valid; callbacks fire around each candidate; all-failed wraps last error |
| Retry non-idempotent operation | usually neither | caller must prove safety before using any retry primitive |
| Infinite loop with delay reset after success | custom loop | use existing SSH poller rationale as anchor |

## Core Patterns

### Use `Polling` for condition checks

`Polling` fits workflow-style waits where success is represented by `fn(attempt) == nil`.

```go
polling := retrier.NewPolling(retrier.PollingOpts{
    Timeout:  act.Timeout(),
    Interval: time.Second,
})
if err := polling.Do(ctx, func(attempt int) error {
    return waitRemoteState(ctx)
}); err != nil {
    return fmt.Errorf("wait remote state: %w", err)
}
```

Do not expect timeout errors to preserve the last condition error; callers should add domain context when returning/logging.

### Use `ExpoBackoff` for finite transient retries

`ExpoBackoff` is appropriate when the operation is idempotent or safe-by-contract and should stop after a bounded number of attempts.

```go
backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOpts{
    MaxRetries:    5,
    BaseDelay:     time.Second,
    MaxDelay:      3 * time.Second,
    JitterPercent: 0.2,
})
if err := backoff.Do(ctx, callThirdParty); err != nil {
    return fmt.Errorf("call third party: %w", err)
}
```

Check tests before relying on edge behavior such as `MaxRetries == 0`.

### Use `Fallback` for ordered candidate attempts

`Fallback` is for trying candidates in order, not for encoding business priority policy inside the retrier package.

```go
fallback := retrier.NewFallback(candidates, isCandidateValid, retrier.FallbackOpts[Candidate]{
    OnAttempt: func(candidate Candidate) { reportAttempt(candidate) },
})
if err := fallback.Do(ctx, func(candidate Candidate) error {
    return callCandidate(ctx, candidate)
}); err != nil {
    return fmt.Errorf("try candidates: %w", err)
}
```

Keep candidate construction, logging level, and domain error translation in the caller.

## Common Mistakes

- Using `ExpoBackoff` for an indefinite loop that must reset delay after successful heartbeats. Keep a custom loop when the retry state is not a finite operation.
- Retrying operations that are not idempotent or not otherwise safe-by-contract.
- Hiding retry logging or business terminal-state decisions inside `pkg/runtime/retrier`.
- Assuming `Polling` returns the last error on timeout; it currently returns `reach max timeout`.
- Expecting `ExpoBackoff` final errors to support `errors.Is` for the last error; current all-failed error formats `last-err` with `%v`.
- Treating callbacks as policy hooks. They are observability/adaptation hooks around attempts, not a place to move domain decisions into runtime.

## Verification Checklist

- Search for analogous usage: `rg "retrier\.NewPolling|retrier\.NewExpoBackoff|retrier\.NewFallback" internal pkg`.
- State why the chosen primitive matches the retry shape: condition wait, bounded transient retry, or ordered fallback.
- Confirm context cancellation is passed into `Do` and into the retried operation where needed.
- Confirm retry safety: idempotent operation, read-only call, or explicit caller-side safety guarantee.
- Keep logging and domain error wrapping in the caller; use `bk-nodemgr-logger` when adding or reviewing logs.
- If changing runtime behavior, update the matching `pkg/runtime/retrier/*_test.go` and inspect representative workflow/third-party/relay/SSH callers.

## Evals

Pressure prompts live in `evals/evals.json`. Keep run outputs, timing, grading, benchmarks, and review artifacts outside git unless explicitly requested.

## Cross-References

- `bk-nodemgr-contextx`: project context propagation and value-preserving cancellation/timeout handling through retry operations.
- `golang-error-handling`: wrapping, `errors.Is/As`, and final error contracts.
- `golang-observability`: retry metrics/logging decisions at caller boundaries.
- `bk-nodemgr-logger`: project logging conventions for retry attempts and failures.
