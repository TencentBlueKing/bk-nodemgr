|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/operinstdata
|Overview:Mongo DAO for workflow operation-instance records|owns oper_inst_data schema, action/operation CRUD, db↔workflow type conversion
|Structure:pkg/dao/mongo/operinstdata:{handler.go,table.go,constants.go,options.go,conver_type.go,action_inst.go,operation_inst.go,*_test.go}
|Where to look:entry+indexes:handler.go:New() creates dao and EnsureIndexes()
|Where to look:table schema:table.go:{ActionInstData,OperInstData,LifeCycle,TableName}:collection=oper_inst_data
|Where to look:field paths:constants.go:{FieldKeyOperInstID,FieldKeyActionInstData,FieldKeyActInstPrivateData,...}:Mongo update/query keys
|Where to look:query builders:options.go:{WithOperInstID,WithOperationID,WithTriggerID,WithActionName,...}:compose base.AliveFilter()
|Where to look:action data DAO:action_inst.go:{GetActionInstData,UpdateActionInstData,GetActInstPrivateData,PushActInstPrivateData,UpdateActionInstContent,UpdateActionInstStatus}
|Where to look:operation data DAO:operation_inst.go:{Upsert,FindOne,ListFullData,ListWithoutActInst,UpdateLifeCycle,UpdateLatestActionInstBriefData,Delete*}
|Where to look:db↔workflow conversion:conver_type.go:{ConvActionInstDataToDB,ConvOpInstanceDataToDB,ConvAOperaInstDataWithoutActionFromDB}:bridge pkg/workflow/{action,operation}
|Type Flow:workflow action/operation structs→conver_type.go→table.go BSON models→Mongo|read path mirrors this in reverse
|Conventions:Content and InitContent are stored as JSON strings, not nested maps; marshal on write and unmarshal on read
|Conventions:PrivateData is persisted as map[string]any; key-specific value encoding must be decided by callers and preserved verbatim here
|Conventions:current sub_workflow_refs convention=JSON string of []types.SubWorkflowRef; DAO must not reinterpret or reshape it
|Conventions:tenant safety comes from upper filters/collection naming; keep queries scoped by opts + base.AliveFilter()
|Conventions:when updating one action field, reuse FieldKeyActionInst* helpers instead of handwritten dotted paths
|Conventions:nil/empty guards are part of DAO contract here; preserve current error style and avoid silent coercion
|Dependencies:pkg/dao/mongo/base for broker/filter/page helpers|pkg/workflow/{action,operation,common} for runtime structs|pkg/types for page and shared DTOs
|Anti-patterns:no business workflow rules, status policy, retry semantics, or permission logic in this package
|Anti-patterns:no proto imports or API-shaping logic here; this layer stores and reconstructs workflow runtime data only
|Anti-patterns:no key-specific compatibility hacks in DAO for PrivateData values; if a value needs serialization, callers must serialize before write
|Anti-patterns:no ad-hoc JSON/BSON shape changes for existing fields without updating conver_type.go and affected tests together
|Tests:focus=go test ./pkg/dao/mongo/operinstdata -run 'Test_handler_|Test_Conv'
