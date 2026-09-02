|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/comments/docs/tech discussions
|Compression Rule:Follow pipe-index format|keep concise|no prose/code blocks
|Scope:pkg/dao/mongo/package-export
|Overview:Fixed-collection Mongo DAO for package export metadata|owns persistence, filtering, pagination, soft deletion, indexes, and pkg/types.PackageExport conversion
|In scope:collection naming|base.IOrm wiring|Data BSON contract|OptFn query builders|indexes|CRUD validation|types conversion|integration behavior
|Out of scope:export workflow orchestration|package aggregation|artifact generation/storage|download token/HTTP handling|availability derivation|cross-collection joins
|Structure:pkg/dao/mongo/package-export:{package_export.go,handler.go,table.go,constants.go,options.go,handler_test.go}
|Naming:directory/import path=`package-export`|Go package identifier=`packageexport`|Mongo collection=`package_export`|domain model=`types.PackageExport`
|Where to look:public API+validation+types conversion:handler.go:{IHandler,Handler,New,List,Create,Get,Delete,convertFromTypes,convertToTypes}
|Where to look:dao+base.IOrm+indexes:package_export.go:{newDao,dao,GetClient,GetTableName,GetIndexes}
|Where to look:collection+document contract+unique key:table.go:{TableName,Data,UniqueFields,UniqueKey,Table}
|Where to look:BSON field keys+filters:{constants.go,options.go}:{FieldKeyExportID,FieldKeyWorkflowID,FieldKeyTenantID,FieldKeyStorageKey,FieldKeyDownloadName,FieldKeySize,FieldKeyMD5,FieldKeyOperator,WithExportID,WithWorkflowID,WithTenantID,WithOperator}
|Where to look:domain model+conditions:pkg/types/{export.go,condition.go}:{PackageExport,PackageExportCondition,PackageExportExactFields}
|Where to look:backend caller:internal/backend/storage/pkg:{storage.go,dao_pkg_export.go,iface.go}
|Where to look:file-service caller:internal/file/storage/packageexport:{storage.go,dao_package_export.go,iface.go}
|Where to look:shared Mongo rules:pkg/dao/mongo/base/AGENTS.md|peer refs:pkg/dao/mongo/{package-deployment,package-workflow,release}/AGENTS.md
|Conventions:collection is fixed `package_export` via TableName|changing TableName is a persistence-contract change and requires caller/data migration review
|Conventions:document envelope=base.TableBroker with Data payload under `data` and lifecycle fields under `basic`|Data implements base.IData
|Conventions:ExportID is globally unique in the collection via UniqueFields/UniqueKey|current explicit secondary index=FieldKeyWorkflowID
|Conventions:handler boundary accepts/returns pkg/types.PackageExport|Mongo Data, TableBroker, dao, and base.IOrm details stay package-local
|Conventions:Create requires non-nil context, valid tenant context, non-nil data, and non-empty ExportID|always overwrite input TenantID with nCtx.TenantID()
|Conventions:Get/Delete identify records by ExportID only|List applies only explicit OptFn filters|do not silently add tenant scoping without reviewing both backend and file-service callers
|Conventions:read/delete filters start from base.AliveFilter()|Delete uses DeleteMany and therefore preserves base soft-delete semantics
|Conventions:List computes total with Count before applying base.ParsePage(page)|return types.Page sorting/pagination behavior unchanged
|Conventions:Mongo field paths must match table.go BSON tags under `data`|reuse FieldKey constants for filters/indexes instead of handwritten dotted paths
|Conventions:new indexes belong in dao.GetIndexes() and must correspond to actual lookup/filter behavior|verify uniqueness implications before changing UniqueFields
|Conventions:keep convertFromTypes and convertToTypes mappings symmetric when fields change|update pkg/types contract and both service condition converters together when required
|Conventions:exported Go symbols require English godoc comments|errors preserve base sentinel semantics and wrap database failures with `%w`
|Tests:handler_test.go is a real Mongo integration suite guarded by `//go:build integration`|uses testsuite/support.RequireMongoDatabase for isolated database lifecycle
|Tests:cover create/get field persistence, duplicate ExportID rejection, filters, pagination/sorting, tenant field persistence, and soft deletion
|Commands:default=GOTOOLCHAIN=local go test ./pkg/dao/mongo/package-export -count=1
|Commands:integration compile=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/package-export -run '^$' -count=1
|Commands:integration run=GOTOOLCHAIN=local NODEMGR_TEST_ENV_SOURCE=docker NODEMGR_TEST_DOCKER_NETWORK=bridge go test -tags=integration ./pkg/dao/mongo/package-export -count=1
|Anti-patterns:no workflow/package assembly/download policy in this DAO|no proto/HTTP structs|no cross-collection joins|no direct mongo.Connect
|Anti-patterns:no package-global clients or fixed test databases|no package-local env loading|no integration helpers outside build-tagged tests
|Anti-patterns:no bypass of Handler/base.IOrm to access arbitrary collections|no hard delete replacing current soft-delete behavior
|Anti-patterns:no change to ExportID uniqueness, BSON tags, TableName, or FieldKey paths without checking indexes, persisted data, converters, tests, and both storage callers
