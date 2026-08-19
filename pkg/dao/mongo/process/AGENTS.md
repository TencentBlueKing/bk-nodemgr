|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/process
|Overview:Tenant-scoped Mongo DAO for Process records|owns process BSON persistence, tenant table routing, CRUD/status/info updates, filters, distribution/distinct queries, and pkg/types.Process conversion
|Out of scope:process lifecycle policy|plugin package selection|workflow orchestration|HTTP/proto/front contracts|business result derivation
|Structure:pkg/dao/mongo/process:{handler.go,process.go,table.go,constants.go,options.go,handler_test.go,handler_integration_test.go,README.md}
|Where to look:public API+tenant routing+types conversion:handler.go:{IHandler,New,tenantDao,Create,Count,List,Get,Delete,Exist,Update,UpdateInfo,UpdateManyInfo,UpdateManyHostBizID,convProcessFromTypes,convertProcessToTypes}
|Where to look:per-tenant dao+indexes+aggregate queries:process.go:{newDao,GetIndexes,GetProcessDistributionByHostID,GetDistinctProcessName,GetDistinctPluginName}
|Where to look:collection name+document contract:table.go:{TableName,Process,UniqueFields,UniqueKey,TableProcess}:collection=process_<tenantID>|unique={data.host_id,data.name}|UniqueFields returns {host_id,plugin_name}
|Where to look:BSON field paths:constants.go:FieldKey* paths under data.* and nested {info,identity,controller,resource,monitor_policy}
|Where to look:query builders:options.go:{WithHostID,WithBizID,WithPluginName,WithFuzzyName,WithGroup,WithPkgName,WithFuzzyPkgName,WithGeneration,WithPlatformOS,WithPlatformArch,WithInfoStatus,WithInfoAgentID,WithInfoVersion}
|Where to look:shared DTOs:pkg/types:{Process,ProcessInfo,ProcessIdentity,ProcessController,ProcessResource,ProcessMonitorPolicy,ProcessStatus,ProcessRestartType,Page}
|Type Flow:pkg/types.Process↔handler converters↔Process BSON data↔base.TableBroker↔Mongo process_<tenantID>
|Conventions:all exported handler entrypoints use contextx.IContext and CheckTenantID before tenant table access
|Conventions:tenantDao is the only daoMap mutation point; keep physical table naming through TableName(tenantID) and index bootstrap through EnsureIndexes()
|Conventions:read filters=base.AliveFilter()→apply OptFn chain|List computes total with Count then applies base.ParsePage(page)
|Conventions:handler boundary accepts/returns pkg/types.Process; Mongo Process structs and BSON details stay package-private
|Conventions:Create/Update paths must preserve pkg/types.Process nested fields through converters; update handler_test.go when converter round-trip shape changes
|Conventions:Delete is soft delete through base ORM; post-delete reads should remain excluded by AliveFilter()
|Conventions:add indexes in GetIndexes() only for concrete query patterns; align keys with FieldKey* constants and AliveFilter usage
|Tests:unit converter tests stay in handler_test.go|real Mongo tests require //go:build integration and support.RequireMongoDatabase(t) in handler_integration_test.go
|Commands:default=GOTOOLCHAIN=local go test ./pkg/dao/mongo/process -count=1|compile=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/process -run '^$' -count=1
|Dependencies:pkg/dao/mongo/base(TableBroker/IOrm/AliveFilter/OptFn/page/errors)|pkg/contextx(tenant context)|pkg/types(domain DTOs)|pkg/logger|MongoDB Go driver
|Anti-patterns:no process lifecycle/business policy or workflow orchestration in this DAO; keep service decisions above storage
|Anti-patterns:no proto/HTTP/front structs in this package; public boundary stays pkg/types
|Anti-patterns:no raw collection access outside IHandler/tenantDao/base ORM path; tenant isolation and table naming must remain centralized
|Anti-patterns:no handwritten dotted BSON paths when constants.go already owns FieldKey* paths
|Anti-patterns:no direct mongo.Connect, package-local env loading, bare MONGO_* keys, package-global DB/client state, or sync.Once fixtures in integration tests
|Related:shared DAO rules:pkg/dao/mongo/base/AGENTS.md|process package intent:pkg/dao/mongo/process/README.md|domain model:pkg/types/process.go
