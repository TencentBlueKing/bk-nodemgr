|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/access
|Overview:Shared virtual-user access helpers for bk-nodemgr; provides startup-configured system login name and tenant-scoped BK username resolution; does not own auth/permission policy
|Boundary:pkg/access owns virtual user accessor and tenant-scoped virtual user helper|pkg/tenant owns tenant mode, tenant ID provider, and tenant user resolver|internal/* owns startup wiring and service orchestration
|Structure:pkg/access:{access.go,AGENTS.md}
|Where to look:virtual user config:access.go:{SetVirtualUser,GetVirtualUser,defaultVirtualUser}
|Where to look:tenant-scoped system user:access.go:{GetVirtualUserBKUsername}:delegates to tenant.GetBKUsernameByLoginName
|Where to look:startup wiring:internal/backend/service/service.go:{initialStaticsConfigs}:calls SetVirtualUser from configured CMDB user
|Where to look:representative callers:internal/backend/dpmgr/executor.go|internal/backend/manager/schedule_manager.go|internal/application/distinctcache/cache.go
|Conventions:SetVirtualUser is a startup-only wiring hook; do not call it from request/background business paths
|Conventions:GetVirtualUser returns the configured login name fallbacking to bk-nodemgr default; callers needing tenant identity must prefer GetVirtualUserBKUsername(nCtx)
|Conventions:GetVirtualUserBKUsername requires contextx.IContext so tenant/user/message/cancel semantics flow into pkg/tenant resolver
|Conventions:keep tenant-scoped username resolution delegated to pkg/tenant; do not duplicate resolver logic or tenant mode branching here
|Conventions:preserve sync.Once singleton semantics for virtual user configuration; tests outside this package must not reset package state
|Anti-patterns:no service-specific auth/access policy in pkg/access|no DB/thirdparty clients here|no raw login-name substitution where tenant-scoped BK username is required|no runtime mutation of virtual user after startup|no bypassing pkg/tenant resolver|no speculative access helpers without current caller semantics
|Verification:go test ./pkg/access|if virtual user semantics change:go test ./internal/backend/service ./internal/backend/dpmgr ./internal/backend/manager ./internal/application/distinctcache
