# PKG KNOWLEDGE BASE

## OVERVIEW

`pkg/` hosts shared platform libraries reused across services: type models, DAO, REST stack, workflow engine, third-party clients, and proto adapters.

`pkg/` should contain cross-service and business-agnostic capabilities; domain orchestration stays in `internal/<service>`.

## STRUCTURE

```
pkg/
|- types/             # shared domain payloads
|- dao/mongo/         # Mongo collections + handlers
|- rest/              # server/client/context/header/errf
|- workflow/          # workflow runtime, action/operation/trigger
|- proto/             # generated types + conversion helpers
|- thirdparty/        # external system clients (cmdb, iam, apigw, ...)
`- logger/            # structured logging abstraction
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Shared model additions | `pkg/types` | Default cross-layer representation |
| Mongo data access changes | `pkg/dao/mongo/**` | Keep collection schema/queries centralized |
| API middleware/client stack | `pkg/rest/**` | Shared HTTP and context behavior |
| Workflow runtime behavior | `pkg/workflow/**` | Action/operation/trigger/scheduler internals |
| Third-party integration | `pkg/thirdparty/**` | Encapsulate external APIs and auth strategy |
| Protocol conversion | `pkg/proto/**` | Proto lifecycle should stay scoped here |

## CONVENTIONS

- Prefer `pkg/types` as business-layer data contract instead of raw proto structs.
- Keep DAO logic storage-oriented; service-specific orchestration belongs in `internal/*/storage`.
- Reuse `pkg/logger` and `pkg/rest/*` utilities for consistency.
- Before moving code into `pkg`, confirm: (1) multiple services need it, and (2) it is not service-specific business logic.

## BOUNDARY CHECKLIST

- **Should be in `pkg`**: shared data types, third-party SDK wrappers, reusable infra code (SSH/Redis abstractions), generic workflow/runtime helpers.
- **Should not be in `pkg`**: service policy logic, route-specific behavior, feature rules tied to one domain.

## PKG VS RUNTIME

- `pkg`: shared code that may use internal system concepts (for example host/network-related domain identifiers).
- `pkg/runtime`: pure runtime primitives that can be extracted independently (for example conversion helpers, pools, retry primitives).

## ANTI-PATTERNS

- Do not move service-specific policy logic into `pkg`.
- Do not edit generated `pkg/proto/**/*.pb.go` files.
- Do not bypass converter methods in `pkg/proto/*/README.md` contracts.
