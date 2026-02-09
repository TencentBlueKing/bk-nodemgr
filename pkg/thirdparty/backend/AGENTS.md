# BACKEND THIRDPARTY KNOWLEDGE BASE

## OVERVIEW

`backend` package provides standardized integration with backend module APIs.

## WHERE TO LOOK

- Core client/wrapper: `pkg/thirdparty/backend/backend.go`
- Handler and API adapters: `pkg/thirdparty/backend/handler.go`
- Domain API files: `node_*.go`, `plugin*.go`, `process.go`, `release.go`, `topo.go`
- Shared types: `pkg/thirdparty/backend/types.go`

## CONVENTIONS

- Use `pkg/types` for data exchange where possible to shield backend protocol details.
- Keep this package focused on atomic API calls and retries.

## ANTI-PATTERNS

- Do not place scenario workflow/business orchestration in this package.
