|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/package-workflow
|Overview:Tenant-scoped Mongo DAO for package import workflow lifecycle records|owns CRUD-like atomic access, status/finish-time updates, filters, indexes, and pkg/types.PackageWorkflow conversion
|In scope:collection naming|base.IOrm wiring|Data BSON contract|OptFn query builders|tenant dao cache|integration behavior
|Out of scope:package import orchestration|workflow status derivation from operations|scheduler monitoring policy|HTTP/proto conversion|cross-collection joins
|Structure:pkg/dao/mongo/package-workflow:{package_workflow.go,handler.go,table.go,constant.go,options.go,handler_test.go}
|Where to look:public API+tenant routing+types conversion:handler.go:{IHandler,Handler,New,tenantDao,Count,List,Create,Get,GetStatus,UpdateStatus,UpdateFinishTime}
|Where to look:per-tenant dao+base.IOrm+indexes:package_workflow.go:{newDao,dao,GetClient,GetTableName,GetIndexes}
|Where to look:collection name+document contract+unique key:table.go:{TableName,Data,UniqueFields,UniqueKey,Table}
|Where to look:BSON field keys+filters:{constant.go,options.go}:{FieldKeyWorkflowID,FieldKeyTriggerID,FieldKeyStatus,FieldKeyType,FieldKeyOperator,FieldKeyOperateTime,FieldKeyFinishTime,WithWorkflowID,WithTriggerID,WithStatus,WithType,WithOperator,WithOperateTimeRange}
|Where to look:domain enums+finished statuses:pkg/types/package_workflow.go:{PackageWorkflow,PackageWorkflowType,PackageWorkflowStatus,GetFinishedPackageWorkflowStatus}
|Where to look:upstream storage orchestration:internal/backend/storage/workflow:{dao_package_workflow.go,monitor_package_workflow.go,storage.go}
|Where to look:integration behavior:handler_test.go|shared Mongo rules:pkg/dao/mongo/base/AGENTS.md|peer refs:pkg/dao/mongo/{business,node-workflow,plugin-workflow}/AGENTS.md
|Conventions:collection=`package_workflow_<tenantID>` via TableName(tenantID)|resolve tenant only from contextx.IContext after CheckTenantID()
|Conventions:tenantDao owns daoMap and EnsureIndexes() bootstrap|do not cache tenant dao entries outside Handler
|Conventions:document envelope=base.TableBroker with Data payload under basic+data|Data implements base.IData|WorkflowID is the unique key
|Conventions:handler boundary accepts/returns pkg/types.PackageWorkflow|Mongo Data structs, BSON paths, and base.IOrm details stay package-private
|Conventions:read filters=base.AliveFilter()→apply OptFn chain|List computes total with Count then applies base.ParsePage(page)
|Conventions:Create rejects nil workflow, tenant mismatch, empty WorkflowID, and empty TriggerID before convertPackageWorkflowFromTypes
|Conventions:UpdateStatus validates types.PackageWorkflowStatus via Validate()|UpdateFinishTime rejects zero finishTime
|Conventions:status/type semantics live in pkg/types/package_workflow.go and workflow storage|DAO only persists/query atomic fields
|Conventions:Mongo field paths must align with table.go BSON tags|reuse FieldKey constants for new filters/indexes instead of duplicating strings
|Conventions:new indexes use mongo.IndexModel in dao.GetIndexes() and must match actual query filters|current explicit index=FieldKeyTriggerID
|Conventions:exported Go symbols require English godoc comments|errors preserve base sentinel semantics and wrapping
|Tests:real Mongo tests require //go:build integration|use testsuite/support helpers when adding behavior coverage|fixtures must isolate tenant/database and avoid order dependence
|Commands:default=GOTOOLCHAIN=local go test ./pkg/dao/mongo/package-workflow -count=1|compile=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/package-workflow -run '^$' -count=1
|Anti-patterns:no business workflow orchestration, status aggregation, scheduler logic, proto/HTTP structs, or package import validation in this DAO
|Anti-patterns:no cross-tenant collection access, cross-collection joins, direct mongo.Connect, package-local .env loading, package-global clients, or fixed test databases
|Anti-patterns:no bypass of Handler/tenantDao/base.IOrm to access arbitrary collections|tenant isolation and TableBroker envelope must remain centralized
|Anti-patterns:no change to WorkflowID uniqueness, BSON tags, TableName format, or FieldKey* paths without checking indexes and all storage callers
