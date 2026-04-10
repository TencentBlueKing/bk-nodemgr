|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/networkunit
|Overview:Fixed-collection Mongo DAO for topology network units|record-level tenant filtering over shared collection `networkunit`|global network area records stay visible across tenants via `tenantFilter`
|Structure:pkg/dao/mongo/networkunit:{handler.go,networkunit.go,table.go,options.go,constants.go,handler_test.go,README.md}
|Where to look:public handler API+tenant checks:pkg/dao/mongo/networkunit/handler.go:{IHandler,New,Count,List,Get,Create,UpdateMany,DeleteMany,GetNetworkUnitDistributionByNetworkAreaID}
|Where to look:collection schema+unique key:pkg/dao/mongo/networkunit/table.go:{TableName,NetworkUnit,Links,Link,Endpoints,CustomDeployConfig,UniqueFields,UniqueKey,TableNetworkUnit}
|Where to look:field keys+query options:pkg/dao/mongo/networkunit/{constants.go,options.go}:{FieldKey*,With/WithoutNetworkUnitID,With/WithoutNetworkAreaID,With/WithoutIsDirect,With/WithoutFuzzyNetworkUnitName,With/WithoutGeneration}
|Where to look:DAO internals:pkg/dao/mongo/networkunit/networkunit.go:{newDao,GetIndexes,create,updateMany,deleteMany,tenantFilter,getNetworkUnitDistributionByNetworkAreaID}
|Where to look:type contract:update field mask:pkg/types/topo.go:{NetworkUnit,NetworkUnitUpdateFields,NewNetworkUnitUpdateFields}
|Where to look:upstream usage:internal/backend/storage/topo/{storage.go,topo.go,iface.go}:storage composes this package and forwards `types.NetworkUnitUpdateFields`
|Where to look:transport conversions:pkg/proto/{application,backend}/api/v3/networkunit.go:proto update requests convert to/from `types.NetworkUnitUpdateFields`
|Where to look:tests+package intent:pkg/dao/mongo/networkunit/{handler_test.go,README.md}:real Mongo integration tests and package boundary notes
|Conventions:document envelope=base.TableBroker[*NetworkUnit] with `basic.*` + `data.*`|inner data implements `UniqueFields`/`UniqueKey`
|Conventions:domain boundary=export `pkg/types.NetworkUnit` at handler layer|keep BSON structs, field keys, and Mongo update details inside this package
|Conventions:collection model=single shared collection `networkunit`|not `TableName(tenantID)` per tenant|tenant isolation is record-level via `tenantFilter`
|Conventions:`tenantFilter(tenantID)` must preserve `$or{tenant_id==current tenant|networkarea_id==base.GlobalNetworkAreaID}` semantics|global network area belongs to system tenant but is readable by all tenants
|Conventions:read paths start from `base.AliveFilter()` then apply package `OptFn` and tenant filter|`Count`/`List`/`Get` require `nCtx.CheckTenantID()`
|Conventions:create path uses `counter.Generate(nCtx, TableName())` to allocate `NetworkUnitID`|validate tenant match, non-empty name, and non-negative `NetworkAreaID` before insert
|Conventions:update contract is field-mask driven by `types.NetworkUnitUpdateFields`|`handler.UpdateMany` only updates enabled fields from `generateNetworkUnitUpdates`
|Conventions:update implementation has two layers|handler path uses `dao.UpdateFieldsBulk` for partial `$set` updates|`dao.updateMany`/`buildUpdateManyParams` uses `base.BuildUpsertParam` full-document replacement and should not be conflated with handler partial updates
|Conventions:delete path is soft delete via `base.BuildDeleteParam()`|`DeleteMany` binds tenant-aware filter and must not remove other tenants' records
|Conventions:aggregation path=`getNetworkUnitDistributionByNetworkAreaID` performs `$match` → group by `networkarea_id` → sort by `_id`|keep distribution semantics storage-only
|Conventions:custom deploy config crosses boundary as `map[string]CustomDeployConfig` in DAO and `map[criteria.OSType]...` in `pkg/types`|conversion helpers own the key translation
|Conventions:follow existing package style when extending filters|current `options.go` still uses typed `base.WithInt64Values` helpers; do not mix in unrelated modernization refactors here
|Dependencies:shared DAO base=pkg/dao/mongo/base|ID allocation=pkg/dao/mongo/counter|domain contract=pkg/types/topo.go|main caller=internal/backend/storage/topo
|Tests:pkg/dao/mongo/networkunit/handler_test.go loads `.env` and connects real Mongo|covers Get/Count/List/Create/UpdateMany/DeleteMany|fixtures include tenant-local and `system_tenant` global network-area records
|Anti-patterns:do not add business orchestration, policy checks, or cross-entity joins here|complex topology logic belongs in `internal/backend/storage/topo`
|Anti-patterns:do not bypass `IHandler` to query the collection directly|tenant visibility, ID allocation, conversion helpers, and soft-delete semantics must stay centralized
|Anti-patterns:do not break global network area visibility by replacing `tenantFilter` with tenant-only filters
|Anti-patterns:do not pass proto structs or transport-layer request types into DAO code|convert at proto/storage boundaries and use `pkg/types`
|Anti-patterns:do not describe `handler.UpdateMany` as full-document upsert|for handler callers it is field-scoped partial update controlled by `NetworkUnitUpdateFields`
|Anti-patterns:do not add cross-table queries, cross-db transactions, or async behavior; DAO scope stays single-collection and synchronous
|Related:shared patterns=pkg/dao/mongo/base/AGENTS.md|peer refs=pkg/dao/mongo/{host,node-workflow}/AGENTS.md