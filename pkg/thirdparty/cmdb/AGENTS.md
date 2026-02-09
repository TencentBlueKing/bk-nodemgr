# CMDB KNOWLEDGE BASE

## OVERVIEW

`cmdb` provides standardized integration with CMDB, including model conversion and adapter logic.

## WHERE TO LOOK

- Core client and wrappers: `pkg/thirdparty/cmdb/cmdb.go`, `pkg/thirdparty/cmdb/scope.go`
- Handler and adapters: `pkg/thirdparty/cmdb/handler.go`, `pkg/thirdparty/cmdb/combine.go`
- Supporting types: `pkg/thirdparty/cmdb/types.go`

## CONVENTIONS

- Keep CMDB payload mapping localized in this package.
- Enforce permission-sensitive handling in adapter layer.
- Use scenario APIs rather than exposing raw CMDB endpoint shapes.

## ANTI-PATTERNS

- Do not embed CMDB concepts directly into business modules.
