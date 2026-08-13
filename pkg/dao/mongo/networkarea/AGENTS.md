|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/networkarea
|Overview:Tenant-partitioned Mongo DAO for topology network areas|each operation routes to `networkarea_<tenantID>`|`data.tenant_id` remains the ownership and write-validation field
|Structure:pkg/dao/mongo/networkarea:{handler.go,networkarea.go,table.go,options.go,constants.go,table_test.go,handler_validation_test.go,handler_test.go,README.md}
|Where to look:public handler API+tenant checks:pkg/dao/mongo/networkarea/handler.go:{IHandler,New,Count,List,Get,UpsertMany,UpdateMany,DeleteMany}
|Where to look:collection schema+tenant table naming:pkg/dao/mongo/networkarea/table.go:{TableName,NetworkArea,UniqueFields,UniqueKey,TableNetworkArea}
|Where to look:field keys+query options:pkg/dao/mongo/networkarea/{constants.go,options.go}:{FieldKey*,With/WithoutNetworkAreaID,With/WithoutCloudVendor,With/WithoutFuzzyNetworkAreaName}
|Where to look:DAO internals:pkg/dao/mongo/networkarea/networkarea.go:{newDao,GetIndexes,upsertMany,updateMany,deleteMany,build*Params}
|Where to look:type contract:pkg/types/topo.go:{NetworkArea,DefaultNetworkAreaID}|`DefaultNetworkAreaID` is tenant-local default/direct area id 0
|Where to look:upstream usage:internal/backend/storage/topo/{storage.go,topo.go,iface.go}:storage composes this package for topology business flow
|Where to look:tests+package intent:pkg/dao/mongo/networkarea/{table_test.go,handler_validation_test.go,handler_test.go,README.md}:table naming, validation unit tests, real Mongo integration tests, package boundary notes
|Conventions:document envelope=base.TableBroker[*NetworkArea] with `basic.*` + `data.*`|inner data implements `UniqueFields`/`UniqueKey`
|Conventions:domain boundary=export `pkg/types.NetworkArea` at handler layer|keep BSON structs, field keys, Mongo write models, and tenant table routing inside this package
|Conventions:collection model=`TableName(tenantID)`→`networkarea_<tenantID>`|handler owns `sync.Map` DAO cache|only `tenantDao` may create/cache tenant DAOs and ensure indexes
|Conventions:read paths require nil-context guard + `nCtx.CheckTenantID()` then select `tenantDao(nCtx.TenantID())` and apply `base.AliveFilter()` + package `OptFn`|do not add record-level `tenantFilter`
|Conventions:`types.DefaultNetworkAreaID`/`data.networkarea_id=0` means tenant-local default/direct network area|each tenant collection may contain its own id 0 record
|Conventions:write paths route through current tenant DAO and validate every item `data.tenant_id` equals `nCtx.TenantID()` before database access
|Conventions:bulk upsert/update/delete filters use network-area id and soft-delete state only|tenant collection routing is the isolation boundary
|Conventions:old fixed `networkarea` collection is migration source only|new runtime reads/writes must never target fixed `networkarea`
|Dependencies:shared DAO base=pkg/dao/mongo/base|domain contract=pkg/types/topo.go|main caller=internal/backend/storage/topo|tenant DAO refs=pkg/dao/mongo/{business,networkunit}
|Tests:pkg/dao/mongo/networkarea/handler_test.go uses `//go:build integration` + `testsuite/support.RequireMongoDatabase`|fixtures are per-test|covers physical collection routing, tenant isolation, id 0, CRUD
|Anti-patterns:do not add business orchestration, policy checks, cross-entity joins, migrations, or dual-read/dual-write compatibility here|complex topology logic belongs in `internal/backend/storage/topo`
|Anti-patterns:do not bypass `IHandler` to query tenant collections directly|collection routing, conversion helpers, tenant validation, and soft-delete semantics must stay centralized
|Anti-patterns:do not reintroduce fixed-collection runtime access, shared-collection `tenantFilter`, system-tenant global area visibility, or obsolete global-area constants
|Anti-patterns:do not pass proto structs or transport-layer request types into DAO code|convert at proto/storage boundaries and use `pkg/types`
|Anti-patterns:do not add cross-table queries, cross-db transactions, async behavior, feature flags, or Helm knobs; each operation stays within one tenant collection and synchronous
