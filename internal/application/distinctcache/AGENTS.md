|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)

|Scope:internal/application/distinctcache

|Overview:In-memory distinct caches on the application (BFF) service|two caches refreshed by a single background goroutine every syncInterval (5min):host distinct (keyed by tenant-id × role-type agent/proxy) and process distinct (keyed by tenant-id × plugin-name)|serves /topo/host/distinct and /process/distinct (plugin_name path) without hitting backend DB per request

|Consumers:internal/application/router/api-v3/topo/host.go DistinctHost handler reads via GetHost()|internal/application/router/api-v3/process/distinct.go Distinct handler reads via GetProcess() (plugin_name path) or falls back to backend (conditions path)|internal/application/service/service.go constructs via New()|internal/application/options/capability.go owns the field and starts it

|Structure:internal/application/distinctcache:{cache.go,cache_test.go,AGENTS.md}

|Where to look:Cache struct + New + Start + run + sync + buildTenantCtx:cache.go
|Where to look:Host cache — syncHost + syncHostTenant + setHost + GetHost + buildRoleCondition:cache.go
|Where to look:Process cache — syncProcess + syncProcessTenant + setProcess + GetProcess:cache.go
|Where to look:Role constants (RoleTypeAgent="agent", RoleTypeProxy="proxy"):cache.go
|Where to look:Sync interval constant (syncInterval=5*time.Minute) + concurrency cap (syncTenantConcurrency=10):cache.go

|Key Semantics:host role_type="agent"→filters node_role in [agent,blank]|host role_type="proxy"→filters node_role in [proxy]
|Key Semantics:GetHost(tenantID,roleType)→returns nil if cache not yet synced|handler returns empty result on nil
|Key Semantics:GetProcess(tenantID,pluginName)→returns nil if cache not yet synced|handler returns empty result on nil
|Key Semantics:process sync discovers plugin names via DistinctProcess(selector=PluginName only)|then fetches full distinct per plugin via DistinctProcess(allSet selector, condition=plugin_name)
|Key Semantics:sync()→lists enabled tenants from mongo once, then calls syncHost() and syncProcess() sequentially
|Key Semantics:syncHost/syncProcess→fans out per-tenant backend calls via gopool, errors are per-tenant isolated
|Key Semantics:Start(nCtx)→non-blocking, launches goroutine, first sync immediate then every syncInterval|stops on nCtx.Done()
|Key Semantics:tenant-scoped context built via buildTenantCtx(nCtx,tenantID)|backend client forwards tenant-id as X-Bk-Tenant-Id header

|Conventions:Host items map protected by hostMu (sync.RWMutex)|Process items map protected by processMu (sync.RWMutex)
|Conventions:Per-tenant parallel sync via gopool.NewPool() with SetLimit(syncTenantConcurrency=10)|failures logged and skipped, do not abort other tenants|gp.Wait() errors (incl. recovered panics) logged at Error
|Conventions:Tenant list sourced from `tenant` mongo collection via shared pkg/dao/mongo/tenant DAO (not tenant.GetAllTenantIDs, which is single-mode on app)
|Conventions:Global distinct by design — no bk_biz_id filtering (distinct values span all businesses in the tenant)
|Conventions:Process per-plugin errors are isolated — failed plugin skipped, successful plugins still cached
|Conventions:Exported symbols require English comments (project rule)
|Conventions:Structured logging via pkg/logger only

|Anti-patterns:Do NOT call GetHost() with role_type other than "agent"/"proxy"→returns nil silently
|Anti-patterns:Do NOT block Start() with synchronous sync|Start must return immediately
|Anti-patterns:Do NOT add bk_biz_id or other condition filtering to buildRoleCondition→cache is global per tenant by design
|Anti-patterns:Do NOT construct tenant context without buildTenantCtx()→backend header forwarding depends on ctx.TenantID()
|Anti-patterns:Do NOT call backendHandler.DistinctHost/DistinctProcess from the router handler directly when cache path applies→read from cache instead
|Anti-patterns:Do NOT purge stale tenant entries on tenant removal→harmless, next sync overwrites or entry simply goes unread

|Dependencies:pkg/thirdparty/backend:{IHandler.DistinctHost, IHandler.DistinctProcess}
|Dependencies:pkg/dao/mongo/tenant:{IHandler.List,WithStatus}
|Dependencies:pkg/contextx:{IContext,From,WithTenantID,WithBKUsername,Background,WithCancel}
|Dependencies:pkg/access:{GetVirtualUser}
|Dependencies:pkg/runtime/gopool:{NewPool}
|Dependencies:pkg/logger:{G}
|Dependencies:pkg/types:{HostCondition,HostDynamicExactFields,NodeRole,HostDistinctResult,ProcessCondition,ProcessExactFields,ProcessDistinctSelector,ProcessDistinctResult,NewProcessDistinctSelectorAllSet,UnlimitedPage,Tenant}
