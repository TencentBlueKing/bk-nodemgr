|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/packageevent
|OVERVIEW:Tenant-scoped Mongo DAO for PackageEvent release audit records|routes operations to `packageevent_<tenantID>`|owns BSON schema, filters, indexes, distinct queries, and pkg/types.PackageEvent conversion
|In scope:collection naming|base.IOrm wiring|PackageEvent BSON contract|OptFn query builders|tenant dao cache|global event_id allocation|integration behavior
|Out of scope:release business orchestration|package publishing workflow|HTTP/proto/front contracts|migration workers/runbooks|cross-tenant aggregation|old shared collection cleanup
|Structure:pkg/dao/mongo/packageevent:{handler.go,packageevent.go,tables.go,constants.go,options.go,handler_test.go,tenant_test.go,tenant_integration_test.go,README.md}
|WHERE TO LOOK:public API+tenant routing+types conversion:handler.go:{IHandler,Handler,New,tenantDao,List,Count,CreateMany,DistinctEventType,DistinctReleaseType,DistinctOsType,DistinctCPUArch,DistinctOperator}
|WHERE TO LOOK:per-tenant dao+base.IOrm+indexes:packageevent.go:{newDao,dao,GetClient,GetTableName,GetIndexes}:index supports default sort `{operate_time,event_id}` for alive records
|WHERE TO LOOK:collection name+document contract+unique key:tables.go:{tableNamePrefix,TableName,PackageEvent,UniqueFields,UniqueKey,TablePackageEvent}:collection=`packageevent_<tenantID>`|unique=data.event_id
|WHERE TO LOOK:BSON field keys+filters:{constants.go,options.go}:{FieldKeyEventID,FieldKeyEventType,FieldKeyReleaseType,FieldKeyOperateTime,With/WithoutEventType,With/WithoutReleaseType,WithOperateTimeRange}
|WHERE TO LOOK:domain contract:pkg/types/package_event.go:{PackageEvent,PackageEventCondition,PackageEventDistinctRequest,PackageEventDistinctResult,PackageEventTypeList}
|WHERE TO LOOK:upstream storage orchestration:internal/backend/storage/release/dao_event.go:{countPakcageEvent,listPackageEvent,createManyPackageEvent,distinctPackageEvent}
|WHERE TO LOOK:tests+package intent:pkg/dao/mongo/packageevent/{tenant_test.go,tenant_integration_test.go,handler_test.go,README.md}|shared Mongo rules:pkg/dao/mongo/base/AGENTS.md|peer ref:pkg/dao/mongo/networkunit/AGENTS.md
|CONVENTIONS:tenant boundary=contextx.IContext.CheckTenantID() then tenantDao(nCtx.TenantID())|never infer tenant from request fields or filters
|CONVENTIONS:collection model=`TableName(tenantID)`→`packageevent_<tenantID>`|handler owns `sync.Map` DAO cache|only `tenantDao` may create/cache tenant DAOs and ensure indexes
|CONVENTIONS:document envelope=base.TableBroker[*PackageEvent] with `basic.*` + `data.*`|inner data implements base.IData|`data.tenant_id` is internal ownership/write-validation field
|CONVENTIONS:create path injects empty PackageEvent.TenantID from context and rejects mismatches via base.CheckTenantIDMatched|`event_id` uses fixed counter namespace `packageevent` so IDs stay globally unique
|CONVENTIONS:read paths build base.AliveFilter() then apply package OptFn chain|List counts first, then uses base.ParsePage(page)|storage adds default sort `operate_time DESC,event_id DESC`
|CONVENTIONS:distinct enum paths convert strings through pkg/runtime/conv with criteria Validate()|do not use deprecated criteria StringListTo* helpers
|CONVENTIONS:Mongo field paths must align with tables.go BSON tags|reuse FieldKey constants for filters/indexes/sorts instead of duplicating strings
|CONVENTIONS:development data handling follows #2957 copy-only style: copy old `packageevent` into tenant collections and leave old `packageevent` untouched|no repo migration worker or release knob
|CONVENTIONS:handler boundary accepts/returns pkg/types.PackageEvent|Mongo structs, BSON paths, counter calls, and base.IOrm details stay package-private
|Tests:unit tests cover TableName, TenantID conversion, index contract, and tenant validation|real Mongo tests require `//go:build integration` + testsuite/support.RequireMongoDatabase|fixtures are isolated per generated database
|Commands:default=GOTOOLCHAIN=local go test ./pkg/dao/mongo/packageevent -count=1|compile=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/packageevent -run '^$' -count=1
|ANTI-PATTERNS:do not reintroduce shared-collection tenantFilter, runtime dual-read/dual-write, tenant-local counters, or cross-tenant listing/counting/distinct
|ANTI-PATTERNS:do not delete, rename, clear, or mutate old shared `packageevent` as part of this package|copy-only data handling stays operational, not code-driven
|ANTI-PATTERNS:do not add release workflow orchestration, package publishing rules, proto/front fields, HTTP request structs, or cross-entity joins to this DAO
|ANTI-PATTERNS:do not bypass Handler/tenantDao/base.IOrm to access arbitrary collections|tenant routing, ID allocation, conversion, and soft-delete semantics must stay centralized
|ANTI-PATTERNS:do not add package-local .env loading, direct mongo.Connect, package-global clients, fixed test databases, or order-dependent shared fixtures
