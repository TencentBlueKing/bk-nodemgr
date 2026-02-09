# INTERNAL KNOWLEDGE BASE

## OVERVIEW

`internal/` is the service-domain layer: routers, managers, service lifecycle, and storage orchestration per service.

## STRUCTURE

```
internal/
|- backend/       # router/api-v3, manager, dpmgr, storage, service
|- application/   # router/api-v3, service
|- file/          # router/api-v3, manager, storage, service
`- relay/         # router/api-v3/admin/healthz, handler, manager, service
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| API endpoint behavior | `internal/*/router/api-v3/**` | Request bind/validation, response shaping, route trees |
| Business workflows/actions | `internal/backend/manager/**` | Workflow definitions and action execution flow |
| Service lifecycle | `internal/*/service/**` | HTTP server startup, dependency wiring, background workers |
| Service-specific persistence | `internal/*/storage/**` | Coordinates `pkg/dao/mongo` access by domain |
| Relay install/detect flow | `internal/relay/handler/**` | SSH/WMI paths and callback/report handling |

## CONVENTIONS

- Keep service boundaries explicit; avoid cross-service imports unless truly shared and stable.
- Router layer should convert transport types early and delegate core behavior to manager/storage layers.
- Storage in `internal` orchestrates service rules; low-level CRUD remains in `pkg/dao/mongo`.

## ANTI-PATTERNS

- Do not put generic utilities here if they are service-agnostic (`pkg/` is for shared code).
- Do not use proto structs as long-lived business models; convert via `pkg/proto/*` helpers.
- Avoid coupling `backend`, `file`, `application`, and `relay` logic directly to each other.
