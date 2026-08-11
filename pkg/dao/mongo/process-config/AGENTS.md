|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/process-config
|Overview:Mongo DAO for process configuration records|tenant-scoped process_config_<tenantID> collections|persists config content/md5/path/main marker/custom context keyed by {name,process_name,host_id}|callers consume pkg/types.ProcessConfig
|Structure:pkg/dao/mongo/process-config:{processconfig.go,table.go,constants.go,options.go,handler.go,handler_test.go,handler_integration_test.go,README.md,.env}
|Where to look:collection+ORM+bulk upsert:processconfig.go:{newDao,GetIndexes,upsertMany,buildUpsertManyParams}:BulkWrite UpdateOneModel SetUpsert(true); GetIndexes currently empty
|Where to look:document schema:table.go:{TableName,ProcessConfig,UniqueFields,UniqueKey,TableScheduledWorkflow}:collection=process_config_<tenantID>|unique={data.name,data.process_name,data.host_id}
|Where to look:BSON field paths:constants.go:{FieldKeyName,FieldKeyProcessName,FieldKeyHostID,FieldKeyContent,FieldKeyMD5,FieldKeyIsMainConfig}:all under data.*
|Where to look:query builders:options.go:{WithName,WithoutName,WithProcessName,WithoutProcessName,WithHostID,WithoutHostID,WithIsMainConfig,WithProcessUniqueKeys}:compose over base.AliveFilter()
|Where to look:handler API+converters:handler.go:IHandler{Create,Get,Count,List,UpsertMany,DeleteMany}|convProcessConfig{ToTypes,FromTypes}
|Where to look:upstream orchestration:internal/backend/storage/plugin/dao_process_config.go:types.ProcessConfigCondition→OptFn mapping; fuzzy include/exclude unsupported there
|Where to look:shared DTOs:pkg/types:{ProcessConfig,ProcessConfigCondition,ProcessUniqueKey}
|Type Flow:pkg/types.ProcessConfig↔handler converters↔ProcessConfig BSON data↔base.TableBroker↔Mongo process_config_<tenantID>
|Conventions:all exported handler entrypoints use contextx.IContext and CheckTenantID before tenant table access
|Conventions:tenantDao is the only daoMap mutation point; keep table naming through TableName(tenantID)
|Conventions:build filters from base.AliveFilter() + OptFn chain; keep FieldKey* constants synchronized with BSON tags and converter mappings
|Conventions:UpsertMany identity is {name,process_name,host_id}; base.BuildUpsertParam writes the full data document, so callers must pass complete configs when preserving optional fields matters
|Conventions:preserve WithProcessUniqueKeys paired $or semantics for {host_id,process_name}; do not rewrite as independent $in filters that create cross-product matches
|Conventions:CustomConfigContext is map[string]any and must round-trip through converters; update handler_test.go when its shape changes
|Conventions:condition-to-option mapping and fuzzy unsupported policy belong in internal/backend/storage/plugin, not this DAO package
|Conventions:add indexes in GetIndexes() only for concrete read/write patterns; align keys with AliveFilter and current DeleteMany hard-delete behavior
|Dependencies:pkg/dao/mongo/base (TableBroker/IOrm/AliveFilter/OptFn/page/errors)|pkg/contextx (tenant context)|pkg/types (domain DTOs)|pkg/logger|MongoDB Go driver BulkWrite/UpdateOneModel
|Anti-patterns:no business policy/config rendering/process relationship logic here; keep orchestration in internal/backend/storage/plugin and above
|Anti-patterns:no service/proto/router/front types in this package; public boundary stays pkg/types
|Anti-patterns:no raw collection access outside dao/base ORM path; do not bypass IHandler for tenant collections
|Anti-patterns:no handwritten dotted BSON paths when constants.go already owns FieldKey* paths
|Anti-patterns:no schema field additions without synchronized updates to {table.go,constants.go,handler.go converters,handler_test.go} and storage condition mapping if filterable
|Anti-patterns:no changing DeleteMany HardDeleteMany semantics without checking callers, rollback, and data-retention expectations
|Tests:handler_test.go covers converter/CustomConfigContext round-trip|handler_integration_test.go covers CRUD/filter/upsert with Mongo integration env from .env
