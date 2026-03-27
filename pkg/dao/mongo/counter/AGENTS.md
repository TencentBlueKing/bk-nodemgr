|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)

|Scope:pkg/dao/mongo/counter

|Overview:Global atomic sequence generator backed by a single shared MongoDB `counter` collection. Not tenant-scoped.
|Consumers:Any DAO needing monotonic IDs (e.g. configpolicy) injects `counter.Handler` at construction time.

|Structure:pkg/dao/mongo/counter:{handler.go,counter.go,table.go,constant.go,handler_test.go}

|Where to look:Public API (Handler interface):handler.go
|Where to look:MongoDB FindOneAndUpdate+$inc logic:counter.go:{generate,generateN,buildGenerateParam}
|Where to look:Document shape (Counter struct, TableCounter):table.go
|Where to look:BSON field key constants:constant.go:{FieldKeyKey,FieldKeySequence}
|Where to look:Integration tests (requires real MongoDB via .env):handler_test.go

|Key Semantics:Generate(key)→returns value BEFORE $inc by 1|first call on new key returns 0
|Key Semantics:GenerateN(key,n)→atomic $inc by n in single call|caller owns [returned+1..returned+n]
|Key Semantics:Upsert→key auto-created via $setOnInsert if absent

|Conventions:Always use FieldKeyKey/FieldKeySequence constants for BSON paths (never string literals)
|Conventions:Always use base.FieldKeyIsDeleted/base.FieldKeyUpdatedAt/base.FieldKeyCreatedAt for basic.* paths
|Conventions:All public methods accept contextx.IContext|key must be non-empty|n must be >0
|Conventions:Index bootstrap via sync.Once in getDao()
|Conventions:Exported symbols require English comments (project rule)

|Anti-patterns:Do NOT implement business rules here (priority ordering, ID allocation strategy)→caller's job
|Anti-patterns:Do NOT hardcode BSON paths as string literals→use constant.go or base constants
|Anti-patterns:Do NOT add tenant-scoping→counter collection is intentionally global
|Anti-patterns:Do NOT call GenerateN with n<=0→returns error
|Anti-patterns:Do NOT assume sequence starts at specific value across test runs→collection is persistent

|Dependencies:pkg/dao/mongo/base:{AliveFilter,TableBroker,FieldKeyIsDeleted,FieldKeyUpdatedAt,FieldKeyCreatedAt}
|Dependencies:pkg/contextx:IContext
|Dependencies:pkg/logger:structured logging
