# Application Router Knowledge Base

## Overview

Gin-based HTTP router for the application (BFF) service. Proxies most requests to backend/file services via `pkg/thirdparty` HTTP clients. Also serves the Vue3 frontend and handles Excel template upload/download.

## Structure

```
router/
├── api-v3/
│   ├── api-v3.go              # LoadBasicAPIs (single entry point)
│   ├── cipher/rsa/            # RSA public key endpoint
│   ├── node/
│   │   ├── agent/             # Agent install/upgrade/reconfig/restart/uninstall + Excel template
│   │   ├── proxy/             # Proxy install/upgrade/reconfig/restart/uninstall
│   │   └── workflow/          # Node workflow + operation queries
│   ├── notice/                # Announcement endpoint
│   ├── pkg/
│   │   ├── event/             # Package events
│   │   ├── publish/           # Release publishing (proxied to file service)
│   │   ├── release/           # Release CRUD + download (proxied to backend + file)
│   │   └── upload/            # Package upload (proxied to file service)
│   ├── plugin/                # Plugin list/memo + plugin workflow
│   │   └── workflow/          # Plugin workflow + operation queries
│   ├── policy/config/         # Config policy CRUD + Excel template
│   ├── process/               # Process list/distinct/distribution
│   └── topo/                  # Topology: host, networkarea, networkunit, graph, events, constants
├── admin/                     # Admin routes (empty — middleware placeholder)
├── healthz/                   # Health check (GET /healthz)
└── web/                       # Frontend index.html serving (GET /)
```

## Where to Look

| Task | Location | Notes |
|------|----------|-------|
| Add new API domain | `api-v3/` + `api-v3.go` | Create package, add `Load()` call |
| Add route to existing domain | `api-v3/{domain}/{domain}.go` | Register in `Load()`, implement handler |
| Service wiring | `internal/application/service/` | Where `LoadBasicAPIs` is called |
| Handler dependencies | `internal/application/options/` | `options.Capability` carries thirdparty handlers |
| Proto types | `pkg/proto/application/api/v3/` | Request binding and response conversion |
| Backend proxy client | `pkg/thirdparty/backend/` | `backend.IHandler` — proxies to backend service |
| File proxy client | `pkg/thirdparty/file/` | `file.IHandler` — proxies to file service |
| Frontend settings | `internal/application/frontsetting/` | `IFrontSetting` for index.html template vars |

## Conventions

### BFF Proxy Pattern (key difference from backend)

Application router does NOT access storage/DAO directly. Instead, handlers proxy to backend or file services via thirdparty HTTP clients:

```go
type handler struct {
    rg             *gin.RouterGroup
    backendHandler backend.IHandler   // proxies to backend service
    fileHandler    file.IHandler      // proxies to file service (pkg/release only)
}
```

Handler flow:
```
proto request → BindJSON → ConvertToTypes → backendHandler.SomeMethod(rCtx, ...) → ConvertFromTypes → proto response
```

Error code for proxy failures: `resterrf.ThirdpartyRequestFailed` (not `resterrf.Aborted`).

### Four Handler Signatures

1. Standard JSON handler (most routes):
```go
func (h *handler) Action(rCtx restserver.IContext) (interface{}, error)
// Uses: restserver.Handler(h.Action)
```

2. File download handler (Excel template):
```go
func (h *handler) Action(rCtx restserver.IContext) (*restserver.FileResponse, error)
// Uses: restserver.FileHandler(h.Action)
```

3. Stream download handler (release downloads):
```go
func (h *handler) Action(rCtx restserver.IContext) (*restserver.StreamResponse, error)
// Uses: restserver.StreamHandler(h.Action)
// Streams binary from file service without buffering entire file
```

4. File upload handler (Excel template, package upload):
```go
func (h *handler) Action(rCtx restserver.IContext) (interface{}, error) {
    fileHeader, err := rCtx.ParseFileForm(req)
    // ...
}
```

### Single Entry Point

Only `LoadBasicAPIs` — no split like backend's `LoadCallbackAPIs`/`LoadProxyAPIs`.

### Web Router

`web/` serves the Vue3 frontend `index.html` with Go template variables injected from `frontsetting.IFrontSetting`. Handles tenant ID resolution for multi-tenant mode.

### Excel Template Handling

`node/agent/template.go` and `policy/config/template.go` generate/parse Excel files using `excelize/v2`. These are the only handlers with significant in-handler logic (column definitions, validation, data mapping).

### Parallel Backend Calls

Some handlers (e.g., `topo/host.go`, `pkg/release/`) use `gopool.Go()` to fan out parallel requests to the backend service for data enrichment.

## Anti-Patterns

- Do not access `pkg/dao/mongo` or `internal/backend/storage` from this router — always proxy via `pkg/thirdparty`.
- Do not use `resterrf.Aborted` for proxy failures — use `resterrf.ThirdpartyRequestFailed`.
- Do not use `restserver.StreamHandler` for non-binary endpoints — only for release downloads.
- Do not place business logic in handlers — delegate to thirdparty clients or keep it minimal (template generation is the exception).

## Unique Styles

- `web/` is unique to application — serves HTML with Go template rendering, not JSON.
- `notice/` uses `GET` (not `POST`) for announcement retrieval — one of the few GET endpoints.
- `pkg/release/` injects both `backend.IHandler` and `file.IHandler` — the only package needing two thirdparty clients.
- `admin/` is an empty middleware placeholder — no routes registered (same as file service).
- Excel template handlers contain hardcoded Chinese column names (`templateNameInnerIP = "内网 IPv4"`) — intentional for user-facing exports.
