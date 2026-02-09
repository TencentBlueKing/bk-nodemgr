# FILE THIRDPARTY KNOWLEDGE BASE

## OVERVIEW

`file` package hosts third-party file-service integration adapters.

## WHERE TO LOOK

- Provider wrapper: `pkg/thirdparty/file/file.go`
- Business adapter: `pkg/thirdparty/file/handler.go`

## CONVENTIONS

- Keep external protocol details inside this package.
- Expose business-facing methods via handler abstraction.

## ANTI-PATTERNS

- Do not call third-party file endpoints directly from business modules.
