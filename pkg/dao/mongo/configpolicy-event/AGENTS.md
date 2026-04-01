|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/configpolicy-event
|Overview:MongoDB DAO for config policy events|tenant-scoped collections|CRUD+count+list+distinct|types↔BSON conversion for pkg/types.ConfigPolicyEvent
|In scope:document shape (flat BSON under data.*)|OptFn query builders|FieldKey* constants|base.IOrm wiring|types↔BSON converters
|Out of scope:event recording orchestration (internal/backend/router)|proto conversion (pkg/proto)|business rules on events
|Structure:pkg/dao/mongo/configpolicy-event:{configpolicy_event.go,handler.go,tables.go,constants.go,options.go,handler_test.go}
|Where to look:dao+newDao+indexes:configpolicy_event.go
|Where to look:IHandler{List,Count,CreateMany}+IDistinctor{Distinct*}+tenant sync.Map+converters:handler.go
|Where to look:ConfigPolicyEvent document+UniqueFields:tables.go
|Where to look:FieldKey* BSON paths (flat data.* prefix):constants.go
|Where to look:OptFn filters {WithBizID,WithType,WithVersion,WithConfigPolicyType,WithConfigPolicyID,WithConfigPolicyName,WithOperator,WithOperateTimeRange}:options.go
|Related:base ORM+filters→pkg/dao/mongo/base|counter→pkg/dao/mongo/counter|peer→pkg/dao/mongo/configpolicy
|Conventions:context=contextx.IContext+CheckTenantID() on all entrypoints
|Conventions:filters=base.AliveFilter()→apply OptFn chain|same pattern as other pkg/dao/mongo/* handlers
|Conventions:field keys=FieldKey* in constants.go for all Mongo paths|must align with tables.go BSON tags
|Conventions:types boundary=IHandler public API uses *types.ConfigPolicyEvent|conversion in handler.go only
|Conventions:sequence=EventID auto-generated via counter.Generate() in CreateMany|callers must not set EventID
|Conventions:exported symbols=require English godoc comments
|Conventions:indexes=dao.GetIndexes() currently empty|add as query patterns grow
|Anti-patterns:no event orchestration or async pool logic here—belongs in internal/backend/{router,storage}
|Anti-patterns:no bypass of TableBroker/base.IOrm|no raw collection access
|Anti-patterns:no BSON tag changes on ConfigPolicyEvent without updating FieldKey* and consumers
|Anti-patterns:no service-specific or HTTP/proto types—pkg/types+BSON only
|Anti-patterns:no duplicate filter strings—use constants.go FieldKey*
|Anti-patterns:no caller-set EventID—auto-assigned by CreateMany via counter
