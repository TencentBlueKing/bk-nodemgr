|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow .agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md from repo root (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/node-ops-history
|Overview:Tenant-scoped Mongo DAO for archived out-of-band settings operations|one operation and all host snapshots form one document
|Structure:{handler.go,node_ops_history.go,table.go,constants.go,options.go,handler_test.go}
|Where to look:public interfaces+validation+projection+domain conversion:handler.go:{IHandler,Handler,convertNodeOpsHistoryFieldSelectionToFields,convertNodeOpsHistoryToTypes,convertNodeOpsHistoryFromTypes}
|Where to look:collection+schema+identity:table.go:{TableName,Data,HostSnapshot,Table}|collection=node_ops_history_<tenantID>|unique=data.operate_id
|Where to look:ORM+indexes:node_ops_history.go:{newDao,dao,GetIndexes}|base.IOrm owns persistence and document envelope
|Where to look:filter/index/projection field paths:constants.go|query options:options.go|domain types+field selection:pkg/types/node_ops_history.go
|Where to look:creation validation+query adaptation:internal/backend/storage/node/dao_node_ops_history.go|archived business IDs:pkg/types/node_ops_history.go:NodeOpsHistoryInfo.BizIDs
|Conventions:validate context and tenant before DAO access|tenantDao routes exclusively by nCtx.TenantID() and caches per-tenant DAOs|Create rejects mismatched history.TenantID, empty snapshots and nil snapshot entries
|Conventions:Data contains operation metadata and embedded HostSnapshot values|base.TableBroker supplies basic.* and data.* envelope|keep FieldKey constants aligned with BSON tags
|Conventions:GetWithFields/ListWithFields use types.NodeOpsHistoryFieldSelection→FieldKey paths→base.IOrm projection|nil or all-false selection returns complete records|unselected scalars are zero values and unselected Snapshots is an empty slice|Snapshots selects the complete array
|Conventions:ListWithFields applies opts, pagination, total and fixed sort {operate_time DESC,operate_id ASC}|Count shares opts but never pagination or projection|all query paths start with base.AliveFilter()
|Conventions:WithBizID/WithoutBizID filter archived operation business IDs at data.biz_ids through base value options|empty values leave the filter unchanged|filters select complete operations, not individual elements of Snapshots|no snapshot host-ID queries or distinct
|Boundary:callers resolve permissions against archived NodeOpsHistoryInfo.BizIDs, not current-host business membership|storage consumes narrowed query conditions without interpreting authorization|callers must stop before querying when authorized business IDs are empty
|Conventions:reuse complete bidirectional converters for Create/GetWithFields/ListWithFields|preserve wrapped Mongo errors including mongo.ErrNoDocuments|do not log and return the same query error
|Tests:handler_test.go uses //go:build integration and testsuite/support.RequireMongoDatabase for isolated real MongoDB queries and cleanup|verify business filters, count/pagination, projection, persistence and tenant isolation|default tests must not connect external services
|Commands:default=GOTOOLCHAIN=local go test ./pkg/dao/mongo/node-ops-history|integration compile=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/node-ops-history -run '^$'|integration=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/node-ops-history -count=1
|Anti-patterns:no separate snapshot collection, business permission orchestration, raw proto types, duplicate projection machinery, cross-tenant reads, or redundant Get/List wrappers around WithFields methods
