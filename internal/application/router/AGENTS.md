|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:internal/application/router
|Overview:Gin-based HTTP router for application (BFF) service|proxies most requests to backend/file via pkg/thirdparty|serves Vue3 frontend|handles Excel template upload/download
|Structure:internal/application/router:{admin,api-v3,healthz,web}
|Structure:api-v3:{api-v3.go,cipher/rsa,node/{agent,proxy,workflow},notice,pkg/{event,publish,release,upload},plugin/workflow,policy/config,process,topo}
|Where to look:add new API domain:api-v3/+api-v3.go:create package, add Load() call
|Where to look:add route to existing domain:api-v3/{domain}/{domain}.go:register in Load(), implement handler
|Where to look:service wiring:internal/application/service/:where LoadBasicAPIs is called
|Where to look:handler deps:internal/application/options/:options.Capability carries thirdparty handlers
|Where to look:proto types:pkg/proto/application/api/v3/:request binding and response conversion
|Where to look:backend proxy client:pkg/thirdparty/backend/:backend.IHandler proxies to backend service
|Where to look:file proxy client:pkg/thirdparty/file/:file.IHandler proxies to file service
|Where to look:frontend settings:internal/application/frontsetting/:IFrontSetting for index.html template vars
|Conventions:BFF proxy pattern=application router does NOT access storage/DAO directly|always proxy to backend/file via pkg/thirdparty HTTP clients
|Conventions:handler structure=type handler{rg *gin.RouterGroup, backendHandler backend.IHandler, fileHandler file.IHandler}
|Conventions:handler flow=proto request→BindJSON→ConvertToTypes→backendHandler.SomeMethod(rCtx,...)→ConvertFromTypes→proto response
|Conventions:error code for proxy failures=resterrf.ThirdpartyRequestFailed (not resterrf.Aborted)
|Conventions:four handler signatures=1) standard JSON: restserver.Handler(h.Action) 2) file download: restserver.FileHandler(h.Action) returns *restserver.FileResponse 3) stream download: restserver.StreamHandler(h.Action) returns *restserver.StreamResponse 4) file upload: rCtx.ParseFileForm(req)
|Conventions:single entry point=only LoadBasicAPIs (no split like backend's LoadCallbackAPIs/LoadProxyAPIs)
|Conventions:web router=web/ serves Vue3 frontend index.html with Go template variables from frontsetting.IFrontSetting|handles tenant ID resolution for multi-tenant mode
|Conventions:Excel template handling=node/agent/template.go and policy/config/template.go generate/parse Excel via excelize/v2|only handlers with significant in-handler logic (column definitions, validation, data mapping)
|Conventions:parallel backend calls=some handlers (topo/host.go, pkg/release/) use gopool.Go() to fan out parallel requests to backend for data enrichment
|Conventions:MVP development=first commit: proto+route+handler returning empty/mock response (no backend call yet, entire request-response path compiles)|second commit: add backend proxy call+basic response conversion (real data flows)|third commit: error handling for proxy failures+parallel calls optimization+logging/metrics
|Conventions:reference-first=before adding new handler, find 2-3 similar BFF handlers as pattern reference|focus on: proxy call patterns, parallel enrichment, error handling
|Anti-patterns:no direct access to pkg/dao/mongo or internal/backend/storage (always proxy via pkg/thirdparty)|no resterrf.Aborted for proxy failures (use resterrf.ThirdpartyRequestFailed)|no restserver.StreamHandler for non-binary endpoints (only for release downloads)|no business logic in handlers (delegate to thirdparty clients or keep minimal, template generation is exception)|if handler loops over items calling backend repeatedly→backend needs aggregation API|if handler merges data from multiple backend calls→consider parallel gopool.Go() pattern
|Unique Styles:web/ serves HTML with Go template rendering (not JSON)|notice/ uses GET (not POST) for announcement retrieval|pkg/release/ injects both backend.IHandler and file.IHandler (only package needing two thirdparty clients)|admin/ is empty middleware placeholder (no routes registered)|Excel template handlers contain hardcoded Chinese column names (intentional for user-facing exports)
