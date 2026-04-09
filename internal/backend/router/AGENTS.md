|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:internal/backend/router
|Overview:Backend router adapters own route registration|request bind/validation|response shaping|auth checks|delegate business logic to manager/storage
|Structure:internal/backend/router:{admin,api-v3,healthz}
|Child AGENTS:internal/backend/router/api-v3/iam/v3/provider/AGENTS.md
|Where to look:new API domain:internal/backend/router/api-v3|api-v3.go:add package+Load() wiring
|Where to look:add route to existing domain:internal/backend/router/api-v3/{domain}/{domain}.go:register in Load()+implement handler
|Where to look:callback flow:internal/backend/router/api-v3/callback/README.md:raw gin callback conventions
|Where to look:service wiring:internal/backend/service/:LoadBasicAPIs|LoadCallbackAPIs|LoadProxyAPIs
|Where to look:handler deps:internal/backend/options/capability.go:manager|storage|authorizer deps live here
|Where to look:proto req/resp:pkg/proto/backend/api/v3/:request binding and boundary conversion
|Where to look:error codes:pkg/rest/errf/:InvalidParameter|DBExecCmdFailed|BackendOperateFailed|PermissionDenied
|Where to look:REST stack:pkg/rest/server/:Handler|IContext|FileHandler
|Conventions:package skeleton=handler{rg,deps}+newHandler(rg.Group(...),capability)+Load(register routes)
|Conventions:standard handler=BindJSON→log+ErrWrap invalid params→delegate→ConvertFromTypes/GetData
|Conventions:raw gin handlers only for callback/workflow/node paths needing direct HTTP control
|Conventions:error handling=always log before return|always resterrf.ErrWrap|never return raw errors
|Conventions:logging=logger.G.Biz(rCtx) for request scope|logger.G.Sys() for background|success paths log Info with identifiers
|Conventions:auth=middleware only at first-level router|user via rCtx.BKUsername/TenantID|reuse package-local permission helpers across handlers
|Conventions:before coding read relevant module + analogous handler in same service/layer; prefer existing router helpers and pkg/proto converters over parallel implementations
|Conventions:API MVP flow=first commit: proto+route+handler returning empty/mock data (entire chain compiles and runs)|second commit: wire storage/DAO for core logic|third commit: add error handling and boundary conditions|follow docs/api/API接口开发流程.md
|Conventions:API reference=before adding new API, find similar endpoints (e.g. GetProcessDistributionByHostID, GetHostDistributionByNetworkAreaID)|reuse patterns instead of redesigning
|Routes:mostly POST including reads|exceptions GET /healthz and /get_manual_script/:os_type/:operation_instance_id|proxy uses gin.Any("/*path")
|Type Flow:proto request→BindJSON→pkg/types→manager/storage→proto response→GetData
|Anti-patterns:no raw errors|no proto structs as router business models|no auth middleware below first-level groups|no mixed restserver/raw gin styles except callback/node historical paths|no business logic in handlers|no skipped error logs|no duplicate helpers before checking current router and pkg/proto patterns
|Unique Styles:api-v3.go splits LoadBasicAPIs|LoadCallbackAPIs|LoadProxyAPIs|callback/workflow/node uses raw gin|proxy forwards catch-all relay callbacks|background topo events use go func()+contextx.New(context.Background())
