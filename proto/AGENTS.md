# PROTO KNOWLEDGE BASE

## OVERVIEW

`proto/` is the source of truth for API contracts; generated Go artifacts land under `pkg/proto/**`.

## STRUCTURE

```
proto/
|- backend/api/v3/*.proto
|- application/api/v3/*.proto
`- file/api/v3/*.proto
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add/modify API fields | `proto/*/api/v3/*.proto` | Message/service/RPC definitions |
| Proto coding rules | `proto/README.md` | Naming and integer type guidance |
| Generated target code | `pkg/proto/**` | Includes `*.pb.go` and conversion helpers |
| End-to-end API process | `docs/api/API接口开发流程.md` | Proto -> converter -> router -> storage flow |

## CONVENTIONS

- Use PascalCase for messages/services/RPC names.
- Use snake_case for proto field names.
- Keep integer defaults aligned with current proto conventions (`int64`/`uint64`).
- Regenerate after proto edits using repository-standard make commands.

## ANTI-PATTERNS

- Never hand-edit generated `*.pb.go` files (`Code generated ... DO NOT EDIT`).
- Do not treat proto structs as business-layer models in service logic.
- Do not skip converter maintenance in `pkg/proto/*` when proto fields change.
