# GSE KNOWLEDGE BASE

## OVERVIEW

`gse` provides standardized integration with GSE services.

## WHERE TO LOOK

- API wrapper: `pkg/thirdparty/gse/gse.go`, `pkg/thirdparty/gse/gse_proc.go`
- Handler adapter: `pkg/thirdparty/gse/handler.go`
- Data types: `pkg/thirdparty/gse/types.go`

## CONVENTIONS

- Keep GSE API invocation and conversions in this package.
- Expose scenario-level methods, not raw GSE client entrypoints.

## ANTI-PATTERNS

- Do not leak GSE-native models into `internal/*` callers.
