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

| Task                   | Location                      | Notes                                        |
| ---------------------- | ----------------------------- | -------------------------------------------- |
| Add/modify API fields  | `proto/*/api/v3/*.proto`      | Message/service/RPC definitions              |
| Proto coding rules     | `proto/README.md`             | Naming and integer type guidance             |
| Generated target code  | `pkg/proto/**`                | Includes `*.pb.go` and conversion helpers    |
| End-to-end API process | `docs/api/API接口开发流程.md` | Proto -> converter -> router -> storage flow |

## CONVENTIONS

- Use PascalCase for messages/services/RPC names.
- Use snake_case for proto field names.
- Keep integer defaults aligned with current proto conventions (`int64`/`uint64`).
- Regenerate after proto edits using repository-standard make commands.

|Language Policy:caller-facing Proto documentation comments use Chinese, overriding the general English-comment default for this scope; identifiers and established technical terms remain unchanged
|Proto documentation:explain caller-relevant purpose, business meaning, and verified limits; no symbol-name restatement or lint-filler comments; never invent requiredness/defaults/enums from types alone
|Proto workflow:use `bk-nodemgr-proto` when adding/changing/reviewing contracts, comments, naming, tags, or Swagger/APIGW projection|Reference:.agents/skills/bk-nodemgr-proto/SKILL.md
|Scope control:apply to new/changed interfaces and relevant comments, not historical bulk cleanup; preserve existing tags/categories; fix documentation at its Proto source and inspect generated Swagger content, not only generation success

## ANTI-PATTERNS

- Never hand-edit generated `*.pb.go` files (`Code generated ... DO NOT EDIT`).
- Do not treat proto structs as business-layer models in service logic.
- Do not skip converter maintenance in `pkg/proto/*` when proto fields change.
