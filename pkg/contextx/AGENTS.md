# CONTEXTX KNOWLEDGE BASE

## OVERVIEW

`pkg/contextx` provides the shared request context model for bk-nodemgr.
It extends Go `context.Context` with stable identity fields (`tenant_id`, `bk_username`, `login_name`, `message_id`) and value propagation helpers, and injects OpenTelemetry span attributes during context construction.

This package is infrastructure-only. It must not contain service/domain logic.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `iface.go` | Core interfaces: `IContext`, `IContextValues`, `IContextInfo` |
| `context.go` | Constructors and wrappers: `New`, `From`, `FromContext`, `WithCancel`, `WithTimeout`, `WithDeadline`, `WithoutCancel`, `Background` |
| `info.go` | Context identity/value carrier (`Info`) and field validation (`Check*`) |
| `README.md` | Package intent, boundaries, and usage constraints |

## CONVENTIONS

- Prefer `pkg/contextx.IContext` at service/internal boundaries when tenant/user/message metadata is required.
- Create or derive contexts through package APIs (`New`, `From`, `FromContext`) instead of ad-hoc struct construction.
- Use `contextx.WithTimeout` / `contextx.WithDeadline` / `contextx.WithCancel` when you need cancellation while preserving existing context values.
- Run explicit guards (`CheckTenantID`, `CheckBKUsername`, `CheckLoginName`, `CheckMessageID`) at validation boundaries instead of assuming fields exist.
- Keep span-attribute behavior centralized in `context.go`; avoid duplicate tracing field injection in callers.
- Keep map propagation semantics stable (`WithValues`, clone behavior in `From`/`FromContext`); avoid mutation patterns that break caller expectations.

## DEPENDENCIES

- Go stdlib `context` + `time` for lifecycle APIs.
- OpenTelemetry (`go.opentelemetry.io/otel/{trace,attribute}`) for context-derived span attributes.

## ANTI-PATTERNS

- Do not add service-specific metadata keys or business validation rules in this package.
- Do not bypass this package by replacing `IContext` with raw `context.Context` where tenant/user identity is mandatory.
- Do not use plain `context.WithTimeout/WithDeadline` on `IContext` paths if you need to preserve `Values()` and identity fields.
- Do not expose mutable internal state patterns that allow callers to accidentally corrupt shared context value semantics.
- Do not duplicate context models in other packages; extend this package additively when cross-service context fields are truly shared.
