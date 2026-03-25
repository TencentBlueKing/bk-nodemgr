# WORKFLOW KNOWLEDGE BASE

## OVERVIEW

`pkg/workflow/` provides the shared workflow runtime engine used by multiple services.

It owns workflow orchestration primitives such as controller lifecycle, operation/action execution, trigger handling, state transitions, logging/trace, and metric hooks.

Service-specific business actions must stay in `internal/*/manager/workflowdef/**`; `pkg/workflow` should remain generic runtime infrastructure.

## STRUCTURE

```
pkg/workflow/
|- manager.go                 # workflow manager entry, create/run/load orchestration
|- worker.go                  # worker loop and workflow execution pipeline
|- controller.go              # controller abstraction and runtime coordination
|- storage.go                 # persistence boundary for workflow runtime state
|- trace.go                   # workflow trace context propagation helpers
|- logger_adaptor.go          # bridge runtime logs to project logger
|- trigger_handler.go         # trigger event dispatch and handling
|- action/                    # action definition and runtime instance contracts
|- operation/                 # operation definition and instance orchestration
|- trigger/                   # trigger metadata and trigger model definitions
|- common/                    # common errors and message payload contracts
`- metric/                    # workflow metrics and monitoring integration
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Workflow execution flow | `worker.go`, `controller.go`, `manager.go` | Main runtime path from scheduling to action execution |
| Action/operation model changes | `action/**`, `operation/**` | Keep definition contracts and instance behavior aligned |
| Trigger-driven behavior | `trigger_handler.go`, `trigger/**` | Ensure trigger metadata and handler logic are consistent |
| Persistence/state transitions | `storage.go`, `operation/instance.go`, `action/instance.go` | Validate status transitions and retry semantics |
| Runtime observability | `trace.go`, `logger_adaptor.go`, `metric/**` | Preserve log/tracing/metrics signal continuity |

## CONVENTIONS

- Keep runtime contracts stable and backward-compatible; many workflow definitions depend on these interfaces.
- Separate infra/runtime concerns from domain policy; avoid importing service-layer logic into `pkg/workflow`.
- Prefer explicit error propagation with context (`fmt.Errorf("...: %w", err)`) to preserve root causes.
- Keep action and operation status transitions deterministic and idempotent under retries.
- Any change to execution order, retry behavior, or state mutation should include/adjust tests.

## ANTI-PATTERNS

- Do not embed service-specific workflow rules in `pkg/workflow`.
- Do not couple runtime primitives to a single workflow definition from `internal/backend/...`.
- Do not add hidden side effects in logger/metric adapters that can change workflow behavior.
- Do not break trace/context propagation across action-operation-trigger boundaries.

## VERIFICATION

```bash
# package tests
go test ./pkg/workflow/...

# optional broader safety check for workflow users
go test ./internal/backend/manager/workflowdef/...
```
