|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/process-v2
|Overview:Tenant-scoped Mongo DAO for node manager v2 compatible plugin process records|collection=process_v2_<tenantID>|owns BSON persistence, tenant table routing, CRUD/info updates, and pkg/types.Process conversion for the v2 compatible process
|Out of scope:process lifecycle policy|GSE query strategy|plugin package selection|workflow orchestration|HTTP/proto/front contracts|v3 process records (see pkg/dao/mongo/process)
|Structure:pkg/dao/mongo/process-v2:{handler.go,processv2.go,table.go,constants.go,options.go,README.md}
|Where to look:public API+tenant routing+types conversion:handler.go:{IHandler,New,tenantDao,Create,List,Get,Exist,Delete,Update,UpdateManyInfo,convertProcessToTypes,convProcessFromTypes,convProcessInfoFromTypes}
|Where to look:per-tenant dao+indexes:processv2.go:{newDao,GetIndexes}:compound index={data.host_id,data.name}
|Where to look:collection name+document contract:table.go:{TableName,Process,UniqueFields,UniqueKey,Table}:collection=process_v2_<tenantID>|unique={data.host_id,data.name}|UniqueKey returns "<hostID>:<name>"
|Where to look:BSON field paths:constants.go:FieldKey* paths under data.* and nested {info,identity,controller,resource,monitor_policy}
|Where to look:query builders:options.go:{WithHostID,WithPluginName}
|Where to look:shared DTOs:pkg/types:{Process,ProcessInfo,ProcessInfoDelta,ProcessStatus,Page}
|Type Flow:pkg/types.Process↔handler converters↔Process BSON data↔base.TableBroker↔Mongo process_v2_<tenantID>
|Conventions:all exported handler entrypoints use contextx.IContext and CheckTenantID before tenant table access
|Conventions:tenantDao is the only daoMap mutation point; keep physical table naming through TableName(tenantID) and index bootstrap through EnsureIndexes()
|Conventions:read filters=base.AliveFilter()→OptFn chain|List computes total with Count then applies base.ParsePage(page)
|Conventions:handler boundary accepts/returns pkg/types.Process; Mongo BSON structs stay package-private
|Conventions:Update writes the whole data document, so callers must pass a complete process when optional fields matter|UpdateManyInfo only rewrites the info block
|Conventions:Delete is a hard delete through base ORM, mirroring the v3 process table, because uninstall must not leave a stale v2 record
|Conventions:in v2 compatibility mode the process name equals the plugin name; do not add a separate process-name field
|Tests:this package has no tests yet|add unit tests for converters and real Mongo tests with //go:build integration and support.RequireMongoDatabase(t) when behavior grows
|Commands:default=GOTOOLCHAIN=local go test ./pkg/dao/mongo/process-v2 -count=1
|Dependencies:pkg/dao/mongo/base(TableBroker/IOrm/AliveFilter/OptFn/page/errors)|pkg/contextx(tenant context)|pkg/types(domain DTOs)|pkg/logger|MongoDB Go driver
|Anti-patterns:no process lifecycle/business policy or workflow orchestration in this DAO; keep service decisions above storage
|Anti-patterns:no proto/HTTP/front structs in this package; public boundary stays pkg/types
|Anti-patterns:no raw collection access outside IHandler/tenantDao/base ORM path; tenant isolation and table naming must remain centralized
|Anti-patterns:no handwritten dotted BSON paths when constants.go already owns FieldKey* paths
|Anti-patterns:never reuse the v3 process table or v3 DAO for v2 compatible processes; the two generations must stay isolated
|Related:shared DAO rules:pkg/dao/mongo/base/AGENTS.md|v3 counterpart:pkg/dao/mongo/process/AGENTS.md|domain model:pkg/types/process.go
