# BKREPO KNOWLEDGE BASE

## OVERVIEW

`bkrepo` encapsulates BKRepo interactions and related data/handler adapters.

## WHERE TO LOOK

- Core client and file operations: `pkg/thirdparty/bkrepo/bkrepo.go`, `file.go`, `file_group.go`
- Handler adapter: `pkg/thirdparty/bkrepo/handler.go`
- Types/header helpers: `pkg/thirdparty/bkrepo/types.go`, `header.go`

## CONVENTIONS

- Keep BKRepo request/response details in this package.
- Expose scenario APIs via handler instead of raw endpoints.

## ANTI-PATTERNS

- Do not let upstream modules depend on BKRepo protocol models.
