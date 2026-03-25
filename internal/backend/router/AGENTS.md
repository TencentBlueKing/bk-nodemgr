# Backend Router Knowledge Base

## Where to Look

| Task                         | Location                      | Notes                                                                 |
|------------------------------|-------------------------------|-----------------------------------------------------------------------|
| Add new API domain           | `api-v3/` + `api-v3.go`       | Create package, add `Load()` call in `api-v3.go`                      |
| Add route to existing domain | `api-v3/{domain}/{domain}.go` | Register in `Load()`, implement handler method                        |
| Understand callback flow     | `api-v3/callback/`            | See `callback/README.md` for routing convention                       |
| IAM provider extension       | `api-v3/iam/v3/provider/`     | See `provider/AGENTS.md` for full guide                               |
| Service wiring               | `internal/backend/service/`   | Where `LoadBasicAPIs`/`LoadCallbackAPIs`/`LoadProxyAPIs` are called   |
| Handler dependencies         | `internal/backend/options/`   | `options.Capability` struct carries all injected deps                 |
| Proto request/response types | `pkg/proto/backend/api/v3/`   | Request binding and response conversion                               |
| Error code definitions       | `pkg/rest/errf/`              | `resterrf.InvalidParameter`, `resterrf.DBExecCmdFailed`, etc.         |
| REST framework               | `pkg/rest/server/`            | `restserver.Handler`, `restserver.IContext`, `restserver.FileHandler` |

## Conventions

### Package Skeleton (every route package)

```go
type handler struct {
rg      *gin.RouterGroup
// injected dependencies from options.Capability
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
return &handler{
rg: rg.Group("/<path>"),
// wire deps from capability
}
}

func Load(rg *gin.RouterGroup, capability *options.Capability) {
h := newHandler(rg, capability)
h.rg.POST("/<action>", restserver.Handler(h.ActionName))
}
```

### Two Handler Signatures

1. Standard API handler (most routes):

```go
func (h *handler) ActionName(rCtx restserver.IContext) (interface{}, error) {
req := new(protoBackend.SomeReq)
if err := rCtx.BindJSON(req); err != nil {
logger.G.Biz(rCtx).WithErr(err).Error("failed to <action>, failed to decode request body")
return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
}
// business logic via manager/storage
resp := new(protoBackend.SomeResp)
resp.ConvertFromTypes(result)
return resp.GetData(), nil
}
```

2. Raw gin handler (callback/workflow/node only):

```go
func (h *handler) ActionName(gCtx *gin.Context) {
nCtx := contextx.New(gCtx)
req := new(protoCallback.SomeReq)
if err := gCtx.BindJSON(req); err != nil {
logger.G.Biz(nCtx).WithErr(err).Error("failed to ...")
gCtx.JSON(http.StatusBadRequest, err)
return
}
// ...
gCtx.JSON(http.StatusOK, nil)
}
```

### Error Handling

- Wrap errors with `resterrf.ErrWrap(code, err)` — never return raw errors.
- Common codes: `resterrf.InvalidParameter`, `resterrf.DBExecCmdFailed`, `resterrf.BackendOperateFailed`.
- Always log before returning error: `logger.G.Biz(rCtx).WithErr(err).Error("failed to ...")`.

### Logging

- Business context: `logger.G.Biz(rCtx)` (carries request trace).
- System context (background goroutines): `logger.G.Sys()`.
- Structured fields: `.With("key", value)`.
- Success actions get `.Info()` log with identifying fields.

### Auth

- Auth middleware added at first-level router only (see `README.md`).
- `restserver.MiddlewareAuth(authIdentity)` parses user info into `rest.Context`.
- Access user via `rCtx.BKUsername()`, `rCtx.TenantID()`.
- For `api-v3/node/agent` permission checks, build auth resources via package-local helper functions and reuse them across handlers (for example, `buildBizResources(...)` and `buildNetworkUnitResources(...)`). Do not inline resource construction inside handlers when the helper-based pattern applies.

### Routes

- Nearly all routes use `POST` — including reads (`/list`, `/get`, `/distinct`).
- Exceptions: `GET /healthz`, `GET /get_manual_script/:os_type/:operation_instance_id`.
- Proxy uses `Any` for catch-all forwarding.

### Type Flow

```
proto request → BindJSON → convert to pkg/types → manager/storage → convert to proto response → GetData()
```

- Use `req.ConvertXxxToTypes()` for request conversion.
- Use `resp.ConvertXxxFromTypes()` for response conversion.

## Anti-Patterns

- Do not return raw errors without `resterrf.ErrWrap`.
- Do not use proto structs as business models — convert to `pkg/types` at the router boundary.
- Do not add auth middleware below first-level router groups.
- Do not mix `restserver.Handler` and raw gin handler styles within the same package (callback/node is the sole
  exception for historical reasons).
- Do not place business logic in handlers — delegate to manager/storage layers.
- Do not skip logging on error paths.

## Unique Styles

- `callback/workflow/node/` uses raw `*gin.Context` handlers instead of `restserver.Handler` — intentional for installer
  callback endpoints that need direct HTTP control.
- `proxy/` uses `gin.Any("/*path")` catch-all to forward relay messages to backend callback endpoints.
- `api-v3.go` splits loading into three entry points (`LoadBasicAPIs`, `LoadCallbackAPIs`, `LoadProxyAPIs`) with
  separate middleware chains.
- Background event recording uses `go func()` with `contextx.New(context.Background())` for fire-and-forget topo events.
- `callback/workflow/` maintains separate `node/` and `plugin/` packages despite overlapping API shapes — intentional
  per `callback/workflow/README.md`.
