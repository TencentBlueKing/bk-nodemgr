|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/thirdparty/usermanager
|Overview:UserManager/bk-user integration package|owns tenant listing and virtual-user login_name→tenant-scoped bk_username resolution
|Structure:pkg/thirdparty/usermanager:{handler.go,user_manager.go,types.go,handler_test.go}
|Where to look:business-facing adapter:handler.go:IHandler,HandlerMultiTenant,HandlerSingle; keep callers on scenario methods
|Where to look:raw bk-user HTTP calls:user_manager.go:cli,listTenant,batchLookupVirtualUser,getHeader; endpoint paths and APIGW headers stay here
|Where to look:wire payloads:types.go:RespCommon,BaseBroker,listTenantResp,batchLookupVirtualUserResp; third-party response structs stay private
|Where to look:behavior checks:handler_test.go:httptest coverage for tenant listing, APIGW errors, virtual-user lookup, validation
|Conventions:IHandler embeds access.ITenantVirtualUserResolver and exposes ListALLTenants; do not add raw endpoint-shaped methods for callers
|Conventions:multi-tenant tenant listing uses contextx.From(nCtx, contextx.WithTenantID(tenant.SystemTenantID)) before bk-user /open/tenants/
|Conventions:virtual-user lookup uses caller tenant from contextx.IContext via X-Bk-Tenant-Id and /open/tenant/virtual-users/-/lookup/?lookups=<loginName>&lookup_field=login_name
|Conventions:HandlerSingle keeps identity behavior for GetBKUsernameByLoginName and returns the default tenant list only
|Conventions:handler layer validates context/login input and maps bk-user response payloads into pkg/types or stable tenant resolver semantics
|Conventions:private bk-user response type names should avoid collisions with imported project packages such as pkg/tenant
|Anti-patterns:do not expose bk-user response structs or BaseBroker outside this package
|Anti-patterns:do not call cli/listTenant/batchLookupVirtualUser from business modules; go through IHandler/tenant resolver surfaces
|Anti-patterns:do not resolve tenant-scoped bk_username without checking context tenant id in multi-tenant mode
|Anti-patterns:do not add speculative resolver injection, background sync, cache, or frontend/proto fields from this package
|Commands:go test ./pkg/thirdparty/usermanager
