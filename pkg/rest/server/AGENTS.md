# AGENTS Instructions for pkg/rest/server

## Scope
These instructions apply to all files under `pkg/rest/server/`.

## Purpose
`pkg/rest/server` provides shared REST infrastructure for Node Manager services:
- Gin server bootstrap (`NewServer`, `Options`, `Start`)
- Middleware chain (context/auth/request-id/log/tracing)
- Handler adapters (`Handler`, `StdHandler`, `StreamHandler`, `FileHandler`)
- Request/context abstraction and standard API response types

## Design Boundaries
- Keep this package service-agnostic. Put business logic in `internal/<service>`.
- Keep external API stable because this package is shared by multiple services.
- Use `pkg/rest/errf` for response and error mapping.
- Propagate metadata via `IRequest`/`IRequestData` and `GenRestContext`.
- Preserve middleware execution order in `NewServer` unless the change is intentional and validated.

## Implementation Conventions
- Public exported identifiers require English comments.
- Use structured logging via `pkg/logger`.
- Prefer extending existing wrapper and middleware abstractions over ad-hoc Gin handlers.
- For auth integrations, implement `IAuthIdentity` and wire it through `MiddlewareAuth`.
- For request ID behavior, use `IRequestIDSetter` and `MiddlewareSetRequestID`.

## Anti-Patterns
- Do not add service-specific conditions or feature flags in this package.
- Do not bypass unified response helpers with direct custom JSON response shapes.
- Do not manipulate tenant/user/request metadata outside request/context abstractions.
- Do not hand-edit generated protobuf code, even when touched from this scope.

## Validation Commands
Run from repository root:
- `go test ./pkg/rest/server/...`
- `go test -race ./pkg/rest/server/...`
- `make lint` (when shared interfaces or behavior change)
- `make pre` (before merging broad-impact changes)

## Quick File Map
- `server.go`: server/options/static setup/startup
- `middleware.go`: auth/request-id/log/tracing middleware
- `handler.go`: HTTP handler adapters and response writers
- `request.go`: request parsing and API response helpers
- `context.go`: rest request to internal context conversion
- `types.go`: response/file/stream types and MIME helpers
- `request_test.go`: unit tests
