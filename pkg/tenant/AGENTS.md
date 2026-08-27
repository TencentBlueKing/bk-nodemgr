|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/tenant
|Overview:Shared tenant mode and tenant ID provider contracts for bk-nodemgr; hides single/multiple tenant differences from callers; does not own tenant permission policy or virtual-user resolution
|Boundary:pkg/tenant owns mode constants, tenant ID constants, tenant ID provider interface, and startup provider wiring|pkg/access owns tenant-scoped virtual-user resolver|internal/* owns service orchestration and DB-backed/provider installation|pkg/contextx owns context identity/value propagation
|Structure:pkg/tenant:{tenant.go,tenant_test.go,README.md,AGENTS.md}
|Where to look:mode contract:tenant.go:{Mode,ModeSingle,ModeMultiple,SetMode,GetMode,Validate}
|Where to look:tenant IDs:tenant.go:{SingleModeTenantID,SystemTenantID,ITenantIDProvider,ListEnabledTenantIDs,SetTenantIDProvider}
|Where to look:tests:tenant_test.go:default provider, nil provider, duplicate provider install
|Conventions:callers must use ListEnabledTenantIDs(nCtx) instead of hardcoding enabled tenant ID lists or bypassing this package
|Conventions:provider contracts use contextx.IContext so DB-backed/external providers can preserve tenant/user/message/cancel semantics
|Conventions:SetMode and SetTenantIDProvider are startup wiring hooks; both are once-only globals and must not be changed during request handling
|Conventions:single mode returns SingleModeTenantID only; multiple-mode/system initialization may use SystemTenantID only when endpoint/workflow semantics require it
|Conventions:error handling follows project guard style: if err != nil early return with %w context at caller-owned boundary
|Conventions:tests may reset package globals only inside pkg/tenant tests; production code must not reset tenantStorage or tenantMode
|Anti-patterns:no service-specific tenant DB/query logic in pkg/tenant|no permission/auth policy here|no raw identifier/common helper bypass for tenant IDs|no reintroducing no-error tenant list APIs|no mutable provider swaps after startup|no speculative tenant constants without current workflow/API semantics
|Verification:go test ./pkg/tenant|if ListEnabledTenantIDs signature/semantics change:go test ./internal/backend/storage/plugin ./internal/backend/storage/node ./internal/backend/storage/workflow ./internal/backend/manager ./pkg/thirdparty/cmdb
