|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow .agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md from repo root (pipe-index format, concise, no prose/code blocks)
|Scope:internal/backend/storage/node
|Overview:Backend storage for node deployments, node workflows and archived node operations|owns DAO adaptation, input guards, metrics and workflow monitoring
|Parent rules:internal/backend/storage/AGENTS.md|DAO history rules:pkg/dao/mongo/node-ops-history/AGENTS.md
|Structure:{iface.go,storage.go,dao_node_deloyment.go,dao_node_workflow.go,dao_node_ops_history.go}|dao_node_deloyment.go is the existing filename; do not rename incidentally
|Where to look:capability contracts:iface.go:{IStorage,IDaoNodeDeployment,IDaoNodeWorkflow,IDaoNodeOpsHistory}|inject the matching narrow interface into callers
|Where to look:initialization+DAO wiring+health+exported wrappers:storage.go:{NewStorage,initDao,check}|metricOperation* constants and WrapFn instrument public operations
|Where to look:deployment configuration+queries:dao_node_deloyment.go|workflow filters+distinct+host-IP lookup:dao_node_workflow.go|history CRUD+projection forwarding:dao_node_ops_history.go
|Where to look:workflow monitoring:storage.go:{registerScheduler,monitoredWorkflows,updateNodeWorkflowResult}|preserve scheduler lifecycle and synchronization when touching monitoring
|Where to look:shared models+selection:pkg/types/{node_deployment.go,node_workflow.go,node_ops_history.go}|query contracts:pkg/types/condition.go|persistence:pkg/dao/mongo/{node-deployment,node-workflow,node-ops-history}
|Conventions:follow existing node/workflow storage signatures, data flow, guards and errors|extend iface.go→storage.go wrapper→private dao_xxx.go implementation→DAO; do not establish a parallel storage design
|Conventions:NodeOpsHistory is single-collection access under IDaoNodeOpsHistory and dao_node_ops_history.go|do not introduce a domain interface/file merely to carry permission filtering
|Conventions:ListNodeOpsHistory/GetNodeOpsHistory accept caller-supplied NodeOpsHistoryFieldSelection and forward it unchanged to DAO ListWithFields/GetWithFields|nil or all-false selection returns complete records
|Conventions:history list returns []*types.NodeOpsHistory, not only Info|preserve selected TenantID/Info/Snapshots; API response shaping belongs to router/proto|do not hardcode selection in storage or add redundant DAO Get/List wrappers
|Conventions:history List/Count share convertNodeOpsHistoryConditionsToOptions|exact include/exclude use ordinary {BizID,OperateID,Operator,NodeRole} fields; OperateTimeRange uses the DAO time option|Count has no selection or pagination
|Boundary:callers own authorization using archived NodeOpsHistoryInfo.BizIDs, not current-host business membership|storage consumes data conditions, not authorization results or policies
|Boundary:do not reintroduce NodeOpsHistoryScope, HostScopes, WithHostScopes or equivalent renamed/map-based permission contracts|no IAM calls, AllBusinesses interpretation or permission-to-query orchestration in storage/DAO
|Boundary:empty exact-value slices leave DAO filters unrestricted; they do not represent deny-all|callers must stop before querying when authorized IDs are empty|matching an operation by BizID returns its complete snapshot array
|Conventions:history creation accepts supplied snapshots without reading current hosts|empty snapshots are a storage no-op|validate tenant match, operation ID/time, supported role and positive unique snapshot HostIDs before persistence
|Conventions:preserve wrapped DAO errors including mongo.ErrNoDocuments; the caller owns missing-record response semantics|return errors without logging them again in storage
|Verification:go test ./internal/backend/storage/node ./pkg/dao/mongo/node-ops-history ./pkg/types|include affected router packages for interface changes; use root Makefile protobuf linker flags when mixed-service proto registration is involved|use root lint entrypoint
