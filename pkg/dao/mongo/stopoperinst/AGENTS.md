|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/stopoperinst
|Overview:Mongo DAO for transient operation-instance stop markers|collection=stopping_operation_inst|production consumers use indexed polling
|Structure:pkg/dao/mongo/stopoperinst:{handler.go,stopoperinst.go,table.go,constant.go,handler_test.go,README.md}
|Where to look:handler API+TTL:handler.go:{IHandler,New,Upsert,FindAll,FindByIDs,stopOperInstTTL}
|Where to look:base ORM+indexes+persistence:stopoperinst.go:{newDao,dao,GetClient,GetTableName,GetIndexes,upsert,buildUpsertParams}
|Where to look:schema+unique contract:table.go:{StopOperInst,UniqueFields,UniqueKey,TableStopOperInst}:unique=data.oper_inst_id
|Where to look:Mongo field paths:constant.go:{FieldKeyOperInstID,FieldKeyExpireAt}
|Where to look:production orchestration:internal/backend/storage/workflow/domain_stop_oper_inst.go:{syncSubscribedStopOperInsts,syncStopOperInsts,upsertNeedStopOperInst}
|Data Model:base.TableBroker[*StopOperInst] envelope|data={oper_inst_id,expire_at}|marker is transient state, not workflow history
|Index Contract:GetIndexes declares TTL on data.expire_at with expireAfterSeconds=0|base.Orm.EnsureIndexes additionally creates unique data.oper_inst_id and basic.is_deleted indexes
|Runtime Contract:Upsert refreshes expire_at to now+30s|same oper_inst_id updates one marker|MongoDB TTL monitor removes expired markers asynchronously
|Polling Contract:FindByIDs is the hot path|build filter from base.AliveFilter() then base.WithValues(FieldKeyOperInstID, ids...)|empty IDs return nil without querying MongoDB
|Polling Contract:workflow storage polls only locally subscribed oper_inst_ids every 1s|same-backend stop writes local state and notifies immediately|10s FindAll sync remains fallback
|Interface Contract:IHandler exposes only {Upsert,FindAll,FindByIDs}|do not add transport, scheduler, subscription, or Change Stream methods
|Conventions:follow base.IOrm[*StopOperInst,StopOperInst] and base.IDao methods|initialize with base.NewOrm|ensure indexes in New
|Conventions:declare package-specific indexes only in GetIndexes|do not add separate buildIndex/ensureIndexes helpers
|Conventions:use FieldKey* constants for dotted Mongo paths|use base.AliveFilter/base.WithValues for queries|use pkg/runtime/conv for map/slice conversion
|Conventions:keep DAO storage-oriented|stop notification scheduling, subscription maps, retries, and workflow lifecycle decisions belong in internal/backend/storage/workflow
|Conventions:preserve stable Upsert error semantics|Mongo write failure must be returned to the caller|index creation failure is logged during construction
|Anti-patterns:no MongoDB Change Stream for stop propagation|high-write deployments cause each backend watcher to scan large volumes of unrelated oplog entries
|Anti-patterns:no full collection polling for the 1s hot path|query only active subscribed IDs|FindAll is reserved for the slower reconciliation task
|Anti-patterns:no TTL shorter than the polling/reconciliation recovery window|coordinate TTL changes with the 1s poll and 10s fallback intervals
|Anti-patterns:no handwritten cursor decoding/list metrics when base.Orm.List provides the required behavior
|Anti-patterns:no duplicate constructor variants or parallel handler interfaces|use New and IHandler
|Tests:handler_test.go uses //go:build integration + testsuite/support.RequireMongoDatabase for isolated MongoDB state
|Tests:default=GOTOOLCHAIN=local go test ./pkg/dao/mongo/stopoperinst|integration=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/stopoperinst -count=1
