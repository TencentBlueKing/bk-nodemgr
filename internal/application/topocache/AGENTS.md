|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)

|Scope:internal/application/topocache

|Overview:In-memory host distinct cache on the application (BFF) service|keyed by tenant-id × role-type (agent/proxy)|refreshed from backend every 1min by a background goroutine|serves /api/v3/topo/host/distinct without hitting backend DB per request
|Consumers:internal/application/router/api-v3/topo/host.go DistinctHost handler reads via Get()|internal/application/service/service.go constructs via New()|internal/application/options/capability.go owns the field and starts it

|Structure:internal/application/topocache:{cache.go,AGENTS.md}

|Where to look:Cache struct + New + Start + Get + run + sync + syncTenant + set + buildRoleCondition:cache.go
|Where to look:Role constants (RoleTypeAgent="agent", RoleTypeProxy="proxy"):cache.go
|Where to look:Sync interval constant (syncInterval=time.Minute) + concurrency cap (syncTenantConcurrency=10):cache.go

|Key Semantics:role_type="agent"→filters node_role in [agent,blank]|role_type="proxy"→filters node_role in [proxy]
|Key Semantics:Get(tenantID,roleType)→returns nil if cache not yet synced|handler returns empty result on nil
|Key Semantics:sync()→lists enabled tenants from mongo, fans out per-tenant backend calls via gopool, errors are per-tenant isolated
|Key Semantics:Start(nCtx)→non-blocking, launches goroutine, first sync immediate then every 1min|stops on nCtx.Done()
|Key Semantics:tenant-scoped context built via contextx.From(nCtx,contextx.WithTenantID(tenant.ID))|backend client forwards tenant-id as X-Bk-Tenant-Id header

|Conventions:Protect items map with sync.RWMutex|sync() takes WLock via set()|Get() takes RLock
|Conventions:Per-tenant parallel sync via gopool.NewPool() with SetLimit(syncTenantConcurrency=10)|failures logged and skipped, do not abort other tenants|gp.Wait() errors (incl. recovered panics) logged at Error
|Conventions:Tenant list sourced from `tenant` mongo collection via shared pkg/dao/mongo/tenant DAO (not tenant.GetAllTenantIDs, which is single-mode on app)
|Conventions:Global distinct by design — no bk_biz_id filtering (distinct values span all businesses in the tenant)
|Conventions:Exported symbols require English comments (project rule)
|Conventions:Structured logging via pkg/logger only

|Anti-patterns:Do NOT call Get() with role_type other than "agent"/"proxy"→returns nil silently
|Anti-patterns:Do NOT block Start() with synchronous sync|Start must return immediately
|Anti-patterns:Do NOT add bk_biz_id or other condition filtering to buildRoleCondition→cache is global per tenant by design
|Anti-patterns:Do NOT construct tenant context without contextx.From(nCtx,...)→backend header forwarding depends on ctx.TenantID()
|Anti-patterns:Do NOT call backendHandler.DistinctHost from the router handler directly→read from cache instead
|Anti-patterns:Do NOT purge stale tenant entries on tenant removal→harmless, next sync overwrites or entry simply goes unread

|Dependencies:pkg/thirdparty/backend:{IHandler.DistinctHost}
|Dependencies:pkg/dao/mongo/tenant:{IHandler.List,WithStatus}
|Dependencies:pkg/contextx:{IContext,From,WithTenantID,Background}
|Dependencies:pkg/runtime/gopool:{NewPool}
|Dependencies:pkg/logger:{G}
|Dependencies:pkg/types:{HostCondition,HostDynamicExactFields,NodeRole,HostDistinctResult,UnlimitedPage,Tenant}
