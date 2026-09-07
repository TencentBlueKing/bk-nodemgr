|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/thirdparty/iamv4
|OVERVIEW:BlueKing IAM V4 APIGateway adapter for Node Manager runtime authorization; owns direct auth, batch auth by actions/resources, authorized resource scope query, permission apply URL, system token fetch, callback BasicAuth validation
|OVERVIEW:Expose `IHandler` as the only scenario-oriented entry point; keep V4 wire payloads, endpoint paths, token cache, and BasicAuth details inside this package
|WHERE TO LOOK:handler interface+scenario methods:pkg/thirdparty/iamv4/handler.go:{IHandler,Handler,New,IsAllowed,ResourcesAllowed,ActionsAllowed,ListAuthorizedResources,GetApplyURL,GetToken,IsBasicAuthAllowed}
|WHERE TO LOOK:raw APIGateway client+endpoint literals:pkg/thirdparty/iamv4/iamv4.go:{newClient,getHeader,directAuth,authByResources,authByActions,listAuthorizedResources,getApplyURL,getToken}
|WHERE TO LOOK:wire request/response models+validation:pkg/thirdparty/iamv4/types.go:{Config,Subject,Resource,DirectAuthRequest,AuthByResourcesRequest,AuthByActionsRequest,AuthorizedResourceRequest,ApplyURLRequest,TokenResponse}
|WHERE TO LOOK:callers+runtime adapter:internal/backend/auth/auth_iamv4.go|internal/backend/service/service.go:newIAMV4Handler|internal/backend/router/api-v3/iam/v4/v4.go
|CONVENTIONS:thirdparty boundary=`pkg/types` request contracts→`IHandler` methods→V4 wire DTOs→APIGateway client; callers must not construct V4 wire DTOs directly
|CONVENTIONS:endpoint path assembly stays in `iamv4.go`; business-facing conversion and token cache stay in `handler.go`; DTO shape and `Validate()` stay in `types.go`
|CONVENTIONS:preserve IAM V4 documented limits and semantics near this package; chunking/fallback policy belongs to `internal/backend/auth/auth_iamv4.go`, not low-level client methods
|CONVENTIONS:use constant-time compare for callback BasicAuth token/password checks; never log tokens, passwords, auth headers, or raw secrets
|CONVENTIONS:normalize IAM envelope failures through `RespCommon.IsFailed()` and wrapped errors; keep request validation before outbound calls
|ANTI-PATTERNS:do not expose raw V4 API structs to `internal/*` service/router/storage code
|ANTI-PATTERNS:do not spread `/v1/open/rbac/...` endpoint literals outside `iamv4.go`
|ANTI-PATTERNS:do not add Role/model registration automation here; `iamv4` is runtime APIGateway adapter only until registration workflow is explicitly designed
|ANTI-PATTERNS:do not add fallback authorization bypass or noop behavior in this package; disabled-mode selection belongs to service authorizer wiring
|ANTI-PATTERNS:do not duplicate IAM V3 cache/request types or mutate `pkg/thirdparty/iamv3`; V3/V4 packages are separate version boundaries
|COMMANDS:go test ./pkg/thirdparty/iamv4
