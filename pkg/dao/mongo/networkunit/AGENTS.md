|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/networkunit
|Overview:Tenant-partitioned Mongo DAO for topology network units|each operation routes to `networkunit_<tenantID>`|`data.tenant_id` remains the ownership and write-validation field
|Structure:pkg/dao/mongo/networkunit:{handler.go,networkunit.go,table.go,options.go,constants.go,table_test.go,handler_validation_test.go,handler_test.go,README.md}
|Where to look:public handler API+tenant checks:pkg/dao/mongo/networkunit/handler.go:{IHandler,New,Count,List,Get,Create,UpdateMany,DeleteMany,GetNetworkUnitDistributionByNetworkAreaID}
|Where to look:collection schema+unique key:pkg/dao/mongo/networkunit/table.go:{TableName,NetworkUnit,Links,Link,Endpoints,CustomDeployConfig,UniqueFields,UniqueKey,TableNetworkUnit}
|Where to look:field keys+query options:pkg/dao/mongo/networkunit/{constants.go,options.go}:{FieldKey*,With/WithoutNetworkUnitID,With/WithoutNetworkAreaID,With/WithoutIsDirect,With/WithoutFuzzyNetworkUnitName,With/WithoutGeneration}
|Where to look:DAO internals:pkg/dao/mongo/networkunit/networkunit.go:{newDao,GetIndexes,create,deleteMany,getNetworkUnitDistributionByNetworkAreaID}
|Where to look:type contract:update field mask:pkg/types/topo.go:{NetworkUnit,NetworkUnitUpdateFields,NewNetworkUnitUpdateFields}
|Where to look:upstream usage:internal/backend/storage/topo/{storage.go,topo.go,iface.go}:storage composes this package and forwards `types.NetworkUnitUpdateFields`
|Where to look:transport conversions:pkg/proto/{application,backend}/api/v3/networkunit.go:proto update requests convert to/from `types.NetworkUnitUpdateFields`
|Where to look:tests+package intent:pkg/dao/mongo/networkunit/{table_test.go,handler_validation_test.go,handler_test.go,README.md}:table naming and validation unit tests, real Mongo integration tests, package boundary notes
|Conventions:document envelope=base.TableBroker[*NetworkUnit] with `basic.*` + `data.*`|inner data implements `UniqueFields`/`UniqueKey`
|Conventions:domain boundary=export `pkg/types.NetworkUnit` at handler layer|keep BSON structs, field keys, and Mongo update details inside this package
|Conventions:collection model=`TableName(tenantID)`→`networkunit_<tenantID>`|handler owns `sync.Map` DAO cache|only `tenantDao` may create/cache tenant DAOs and ensure indexes
|Conventions:read paths require `nCtx.CheckTenantID()` then select `tenantDao(nCtx.TenantID())` and apply `base.AliveFilter()` + package `OptFn`|do not add record-level `tenantFilter`
|Conventions:create path writes through current tenant DAO but uses fixed counter namespace `networkunit` so `NetworkUnitID` stays globally unique|validate tenant match, non-empty name, and non-negative `NetworkAreaID` before insert
|Conventions:`types.DefaultNetworkAreaID`/`data.networkarea_id=0` means tenant-local default/direct network area|do not treat id 0 as cross-tenant or system-tenant visible
|Conventions:update contract is field-mask driven by `types.NetworkUnitUpdateFields`|`handler.UpdateMany` only updates enabled fields from `generateNetworkUnitUpdates`
|Conventions:update implementation uses `dao.UpdateOneFieldBulk` for partial `$set` updates selected by the field mask
|Conventions:delete path is soft delete via `base.BuildDeleteParam()`|tenant collection routing is the isolation boundary
|Conventions:aggregation path=`getNetworkUnitDistributionByNetworkAreaID` performs `$match` → group by `networkarea_id` → sort by `_id`|keep distribution semantics storage-only
|Conventions:custom deploy config crosses boundary as `map[string]CustomDeployConfig` in DAO and `map[criteria.OSType]...` in `pkg/types`|conversion helpers own the key translation
|Conventions:follow existing package style when extending filters|current `options.go` still uses typed `base.WithInt64Values` helpers; do not mix in unrelated modernization refactors here
|Dependencies:shared DAO base=pkg/dao/mongo/base|ID allocation=pkg/dao/mongo/counter|domain contract=pkg/types/topo.go|main caller=internal/backend/storage/topo
|Tests:pkg/dao/mongo/networkunit/handler_test.go uses `//go:build integration` + `testsuite/support.RequireMongoDatabase`|fixtures are per-test|covers physical collection routing, tenant isolation, global counter IDs, CRUD, aggregation
|Anti-patterns:do not add business orchestration, policy checks, or cross-entity joins here|complex topology logic belongs in `internal/backend/storage/topo`
|Anti-patterns:do not bypass `IHandler` to query tenant collections directly|collection routing, ID allocation, conversion helpers, and soft-delete semantics must stay centralized
|Anti-patterns:do not reintroduce shared-collection `tenantFilter`, tenant-local counter keys, cross-tenant id 0 visibility, or runtime dual-read/dual-write compatibility
|Anti-patterns:do not pass proto structs or transport-layer request types into DAO code|convert at proto/storage boundaries and use `pkg/types`
|Anti-patterns:do not describe `handler.UpdateMany` as full-document upsert|for handler callers it is field-scoped partial update controlled by `NetworkUnitUpdateFields`
|Anti-patterns:do not add cross-table queries, cross-db transactions, or async behavior; each operation stays within one tenant collection and synchronous
|Related:shared patterns=pkg/dao/mongo/base/AGENTS.md|tenant DAO refs=pkg/dao/mongo/{business,cipher,configpolicy,topoevent}
