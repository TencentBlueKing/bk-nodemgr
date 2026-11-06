|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/process-v2-config
|Overview:Mongo DAO for node manager v2 compatible plugin configuration|tenant-scoped process_v2_config_<tenantID> collections|persists config content/md5/path/main marker/custom context keyed by {name,process_name,host_id,is_main_config}|callers currently write only the main config file
|Out of scope:config rendering|main-vs-sub config source split|GSE push|workflow orchestration|v3 config records (see pkg/dao/mongo/process-config)
|Structure:pkg/dao/mongo/process-v2-config:{processv2config.go,table.go,constants.go,options.go,handler.go,README.md}
|Where to look:collection+ORM+bulk upsert:processv2config.go:{newDao,GetIndexes,upsertMany,buildUpsertManyParams}:BulkWrite UpdateOneModel SetUpsert(true)|GetIndexes returns one compound index over the four unique fields
|Where to look:document schema:table.go:{TableName,ProcessConfig,UniqueFields,UniqueKey,Table}:collection=process_v2_config_<tenantID>|unique={data.name,data.process_name,data.host_id,data.is_main_config}
|Where to look:BSON field paths:constants.go:{FieldKeyName,FieldKeyTemplateName,FieldKeyProcessName,FieldKeyHostID,FieldKeySet,FieldKeyContent,FieldKeyMD5,FieldKeyIsMainConfig}:all under data.*
|Where to look:query builders:options.go:{WithName,WithoutName,WithTemplateName,WithoutTemplateName,WithProcessName,WithoutProcessName,WithHostID,WithoutHostID,WithSet,WithoutSet,WithIsMainConfig,WithProcessUniqueKeys,WithProcessConfigUniqueKeys}:compose over base.AliveFilter()
|Where to look:v2 read paths:handler.go:{Get,Count,List}|Get filters by WithProcessConfigUniqueKeys; Count/List take OptFn chains that the storage adapter builds from types.ProcessConfigCondition
|Where to look:handler API+converters:handler.go:IHandler{Get,Count,List,UpsertMany,DeleteMany}|convProcessConfig{ToTypes,FromTypes}
|Where to look:shared DTOs:pkg/types:{ProcessConfig,ProcessUniqueKey,ProcessConfigUniqueKey,Page}
|Type Flow:pkg/types.ProcessConfig↔handler converter↔ProcessConfig BSON data↔base.TableBroker↔Mongo process_v2_config_<tenantID>
|Conventions:all exported handler entrypoints use contextx.IContext and CheckTenantID before tenant table access
|Conventions:tenantDao is the only daoMap mutation point; keep table naming through TableName(tenantID) and index bootstrap through EnsureIndexes()
|Conventions:UpsertMany identity is {name,process_name,host_id,is_main_config}; base.BuildUpsertParam writes the full data document, so callers must pass complete configs
|Conventions:preserve WithProcessUniqueKeys paired $or semantics for {host_id,process_name}; WithProcessConfigUniqueKeys follows the same shape and skips nil keys, so delete callers must reject empty/all-nil key lists themselves
|Conventions:only main config files are persisted today; the table keeps the v3 schema so sub-config storage can be added without migration|filtering main-vs-sub is the caller's decision, not this package's
|Conventions:DeleteMany is a hard delete and requires explicit filters; never call it with no options
|Dependencies:pkg/dao/mongo/base (TableBroker/IOrm/AliveFilter/OptFn/errors)|pkg/contextx (tenant context)|pkg/types (domain DTOs)|pkg/logger|MongoDB Go driver BulkWrite/UpdateOneModel
|Anti-patterns:no business policy/config rendering/process relationship logic here; keep orchestration in internal/backend/manager/workflowdef/pluginv2 and above
|Anti-patterns:no service/proto/router/front types in this package; public boundary stays pkg/types
|Anti-patterns:no raw collection access outside dao/base ORM path; do not bypass IHandler for tenant collections
|Anti-patterns:no handwritten dotted BSON paths when constants.go already owns FieldKey* paths
|Anti-patterns:no schema field additions without synchronized updates to {table.go,constants.go,handler.go converter}
|Anti-patterns:never reuse the v3 process config table or DAO for v2 compatible configs; the two generations must stay isolated
|Tests:handler_test.go covers convProcessConfig{ToTypes,FromTypes} and BSON-container normalization of CustomConfigContext|add //go:build integration Mongo tests when read paths or schema mapping grow
|Related:shared DAO rules:pkg/dao/mongo/base/AGENTS.md|v3 counterpart:pkg/dao/mongo/process-config/AGENTS.md|process v2 counterpart:pkg/dao/mongo/process-v2/AGENTS.md
