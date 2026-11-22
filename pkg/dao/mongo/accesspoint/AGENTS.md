|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/accesspoint
|Overview:Tenant-partitioned Mongo DAO for topology access points|each operation routes to `accesspoint_<tenantID>`|`data.tenant_id` remains the ownership and write-validation field
|Structure:pkg/dao/mongo/accesspoint:{handler.go,accesspoint.go,table.go,options.go,constants.go,table_test.go,convert_test.go,handler_test.go,README.md}
|Where to look:public handler API+tenant checks:pkg/dao/mongo/accesspoint/handler.go:{IHandler,New,Count,List,Get,Create,CreateMany,UpdateMany,DeleteMany}
|Where to look:collection schema+unique key:pkg/dao/mongo/accesspoint/table.go:{TableName,AccessPoint,Endpoints,UniqueFields,UniqueKey,TableAccessPoint}
|Where to look:field keys+query options:pkg/dao/mongo/accesspoint/{constants.go,options.go}:{FieldKey*,With/WithoutAccessPointID,With/WithoutNetworkAreaID}
|Where to look:DAO internals:pkg/dao/mongo/accesspoint/accesspoint.go:{newDao,GetIndexes,create,createMany,updateMany,deleteMany,buildUpdateManyParams,buildDeleteManyParams}
|Where to look:types↔BSON conversion:pkg/dao/mongo/accesspoint/handler.go:{convertAccessPointFromTypes,convertAccessPointToTypes}|covered by convert_test.go
|Where to look:type contract:pkg/types/topo.go:{AccessPoint,AccessPointList}|upstream usage:internal/backend/storage/topo/{storage.go,topo.go,gse_domain.go}
|Where to look:transport conversions:pkg/proto/{application,backend}/api/v3/{networkunit.go,accesspoint.go}:proto AccessPoint↔types.AccessPoint
|Conventions:document envelope=base.TableBroker[*AccessPoint] with `basic.*` + `data.*`|inner data implements `UniqueFields`/`UniqueKey`
|Conventions:domain boundary=export `pkg/types.AccessPoint` at handler layer|keep BSON structs, field keys, and Mongo write models inside this package
|Conventions:collection model=`TableName(tenantID)`→`accesspoint_<tenantID>`|handler owns `sync.Map` DAO cache|only `tenantDao` may create/cache tenant DAOs and ensure indexes
|Conventions:read paths require `nCtx.CheckTenantID()` then select `tenantDao(nCtx.TenantID())` and apply `base.AliveFilter()` + package `OptFn`|do not add record-level `tenantFilter`
|Conventions:create/createMany/UpdateMany call `base.CheckTenantIDMatched` per item before database access|create/createMany draw `AccessPointID` from the global counter namespace `accesspoint`, unique across tenant collections
|Conventions:`UpdateMany` replaces the whole `data` document via `base.BuildUpsertParam` (filter=alive+`data.accesspoint_id`,upsert=false)|callers must send the complete access point
|Conventions:`CloudSecurityGroupPoolID` persists as `cloud_security_group_pool_id`|empty string means no pool bound|a full update that sends empty clears an existing binding|legacy documents without the field read as empty
|Conventions:pool existence is not validated in the DAO|that check belongs to the router/service layer
|Conventions:delete path is soft delete via `base.BuildDeleteParam()`|tenant collection routing is the isolation boundary|`GetIndexes` currently declares no extra indexes
|Dependencies:shared DAO base=pkg/dao/mongo/base|ID allocation=pkg/dao/mongo/counter|domain contract=pkg/types/topo.go|main caller=internal/backend/storage/topo
|Tests:table_test.go+convert_test.go are unit tests|handler_test.go uses `//go:build integration` + `support.RequireMongoDatabase` for tenant routing, cross-tenant rejection, global counter IDs
|Anti-patterns:do not add business orchestration, policy checks, or cross-entity joins here|topology logic belongs in `internal/backend/storage/topo`
|Anti-patterns:do not bypass `IHandler` to query tenant collections directly|collection routing, conversion, tenant validation, and soft-delete semantics must stay centralized
|Anti-patterns:do not pass proto structs or transport-layer request types into DAO code|convert at proto/storage boundaries and use `pkg/types`
|Anti-patterns:do not reintroduce shared-collection `tenantFilter`, tenant-local counter keys, or cross-tenant id 0 visibility
|Anti-patterns:do not describe `UpdateMany` as a field-scoped partial update|it overwrites `data`
|Related:shared patterns=pkg/dao/mongo/base/AGENTS.md|tenant DAO refs=pkg/dao/mongo/{networkarea,networkunit}
