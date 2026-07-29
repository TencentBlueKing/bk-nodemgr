|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/business
|OVERVIEW:Tenant-scoped Mongo DAO for Business records|owns Count/List/UpsertMany, BSON persistence, filters, and pkg/types.Business conversion
|In scope:collection naming|base.IOrm wiring|Business document shape|OptFn query builders|bulk upsert|integration tests
|Out of scope:business workflow rules|CMDB synchronization orchestration|cross-entity joins|HTTP/proto conversion|cache policy
|Structure:pkg/dao/mongo/business:{handler.go,business.go,table.go,constants.go,options.go,handler_test.go,README.md}
|WHERE TO LOOK:public API+tenant routing+types conversion:handler.go:{IHandler,Handler,New,tenantDao,Count,List,UpsertMany}
|WHERE TO LOOK:per-tenant dao+base.IOrm+BulkWrite params:business.go:{newDao,dao,buildUpsertParams,buildUpsertManyParams,upsertMany}
|WHERE TO LOOK:collection name+document contract+unique key:table.go:{TableName,Business,UniqueFields,UniqueKey,TableBusiness}
|WHERE TO LOOK:BSON field keys+filters:{constants.go,options.go}:{FieldKeyBizID,WithBizID,WithoutBizID,WithFuzzyBizName,WithoutFuzzyBizName}
|WHERE TO LOOK:integration behavior:handler_test.go|test environment helpers:testsuite/support|shared Mongo rules:pkg/dao/mongo/base/AGENTS.md
|CONVENTIONS:collection=`business_<tenantID>` via TableName(tenantID)|resolve tenant only from contextx.IContext after CheckTenantID()
|CONVENTIONS:daoMap is owned by tenantDao|do not load/store tenant dao entries elsewhere|index bootstrap stays in tenantDao through EnsureIndexes()
|CONVENTIONS:document envelope=base.TableBroker[*Business] with basic.*+data.*|Business implements base.IData|BizID is the unique key
|CONVENTIONS:handler boundary accepts/returns pkg/types.Business|Mongo Business structs and BSON details stay package-private
|CONVENTIONS:read filters=base.AliveFilter()→apply OptFn chain|List computes total with Count then applies base.ParsePage(page)
|CONVENTIONS:UpsertMany rejects empty/nil items and requires every Business.TenantID to match nCtx.TenantID()
|CONVENTIONS:Mongo field paths must align with table.go BSON tags|reuse FieldKey* constants for new filters instead of duplicating paths
|CONVENTIONS:exported Go symbols require English godoc comments|errors preserve base sentinel semantics and wrapping
|Tests:real Mongo tests require //go:build integration|use support.RequireMongoDatabase(t)|fixture writes and assertions share one isolated handler/database
|Commands:default=GOTOOLCHAIN=local go test ./pkg/dao/mongo/business -count=1|compile=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/business -run '^$' -count=1
|ANTI-PATTERNS:no direct mongo.Connect, package-local .env loading, or bare MONGO_* keys in tests|use testsuite/support
|ANTI-PATTERNS:no package-global clients, sync.Once fixture data, fixed test databases, or test-order dependence
|ANTI-PATTERNS:no bypass of IHandler/tenantDao to access arbitrary collections|tenant isolation and table naming must remain centralized
|ANTI-PATTERNS:no proto/HTTP structs or workflow orchestration in this DAO|convert at transport/storage boundaries and use pkg/types
|ANTI-PATTERNS:no change to BizID uniqueness, TableBroker envelope, BSON tags, or filter paths without checking indexes and all consumers
