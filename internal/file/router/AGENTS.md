# File Router Knowledge Base

## Overview

Gin-based HTTP router for the file service. Handles package upload, publish, download, and cross-node file transfer.

## Structure

```
router/
├── api-v3/
│   ├── api-v3.go              # LoadBasicAPIs, LoadDownloadAPIs (separate middleware chains)
│   ├── download/              # Binary file download (agent/proxy/plugin/cert/bintool/installer)
│   ├── publish/               # Release publishing (upload → release record creation)
│   ├── transfer/              # Cross-node file transfer launch + query
│   └── upload/                # Origin package upload (multipart file form)
├── admin/                     # Admin routes (empty — middleware placeholder)
└── healthz/                   # Health check (GET /healthz)
```

## Where to Look

| Task | Location | Notes |
|------|----------|-------|
| Add new file domain | `api-v3/` + `api-v3.go` | Create package, add `Load()` call |
| Add download endpoint | `api-v3/download/` | Use `restserver.FileHandler` |
| Add upload endpoint | `api-v3/upload/` | Use `rCtx.ParseFileForm(req)` |
| Service wiring | `internal/file/service/` | Where `LoadBasicAPIs`/`LoadDownloadAPIs` are called |
| Handler dependencies | `internal/file/options/` | `options.Capability` carries `Manager` + `StorageTopo` |
| Proto types | `pkg/proto/file/api/v3/` | Request binding and response conversion |
| Manager layer | `internal/file/manager/` | `manager.IManager` — all business logic |

## Conventions

### Package Skeleton

Same as backend router: `handler` struct + `newHandler()` + `Load()`. Key difference: handler injects `manager.IManager` (single manager interface) instead of multiple storage interfaces.

```go
type handler struct {
    rg      *gin.RouterGroup
    manager manager.IManager
}
```

### Three Handler Signatures

1. Standard JSON handler (publish, transfer):
```go
func (h *handler) Action(rCtx restserver.IContext) (interface{}, error)
// Uses: restserver.Handler(h.Action)
```

2. File download handler (download):
```go
func (h *handler) Action(rCtx restserver.IContext) (*restserver.FileResponse, error)
// Uses: restserver.FileHandler(h.Action)
// Returns FileResponse with Data (io.Reader), Size, FileName, ContentType
```

3. File upload handler (upload):
```go
func (h *handler) Action(rCtx restserver.IContext) (interface{}, error) {
    req := new(protoFile.SomeReq)
    fileHeader, err := rCtx.ParseFileForm(req)  // multipart form parse
    file, err := fileHeader.Open()
    defer file.Close()
    // ...
}
// Uses: restserver.Handler(h.Action)
```

### Two Entry Points

- `LoadBasicAPIs` — full API surface (download + publish + transfer + upload) with auth middleware.
- `LoadDownloadAPIs` — download-only subset, potentially with different/lighter middleware.

### Error Codes

- `resterrf.InvalidParameter` — bad request / parse failure.
- `resterrf.Aborted` — manager operation failed (upload/publish/transfer).

### Transfer Domain

`transfer/` is unique to the file service — launches async file transfers to target hosts via GSE. Handler injects both `manager.IManager` and `topo.IStorage` (for host lookup).

## Anti-Patterns

- Do not return `*restserver.FileResponse` from non-download handlers — only download uses `FileHandler`.
- Do not skip `fileHeader.Open()` / `defer file.Close()` in upload handlers.
- Do not place business logic in handlers — delegate to `manager.IManager`.
- Do not use `restserver.StreamHandler` here — that's application router only.

## Unique Styles

- All download handlers return `*restserver.FileResponse` with `restserver.MIMETypeBin` content type.
- Upload handlers use `rCtx.ParseFileForm(req)` to bind both JSON fields and multipart file in one call.
