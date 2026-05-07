|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/types
|Overview:Shared business-layer domain contracts for bk-nodemgr services; owns structs/enums/query contracts/operation params that flow handler→service/manager→storage/workflow→proto response
|Boundary:pkg/proto subpackages convert wire models↔pkg/types at transport edge|pkg/dao/mongo consumes condition/distinct/page contracts|internal service packages own business orchestration and policy decisions
|Structure:pkg/types:{host.go,topo.go,scope.go,release.go,plugin.go,config_policy.go,deploy_policy.go,node_deployment.go,plugin_deployment.go,node_workflow.go,plugin_workflow.go,scheduled_workflow.go,node_agent.go,node_proxy.go,manager.go,gse.go,cmdb.go,process.go,network_policy.go,condition.go,distinct.go,page.go,event files,README.md}
|Where to look:core inventory:{host.go,topo.go,scope.go,tenant.go,service_instance.go}:host/topology/scope/tenant references
|Where to look:package+plugin:{release.go,plugin.go,upload.go,package_event.go}:release metadata, plugin groups, upload package categorization, package lifecycle events
|Where to look:deployment:{deploy_policy.go,node_deployment.go,plugin_deployment.go,config_policy.go,network_policy.go}:policy/spec/state/config/network domain models
|Where to look:workflow:{node_workflow.go,plugin_workflow.go,scheduled_workflow.go,graph.go,operinst_private_data.go}:operation workflow models and visualization/private-data keys
|Where to look:operation params:{node_agent.go,node_proxy.go,manager.go}:manager input contracts for install/upgrade/restart/reconfig/update/uninstall/execute policy flows
|Where to look:query contracts:{condition.go,distinct.go,page.go}:DAO filtering, aggregation selectors, pagination/sort helpers
|Where to look:external projections:{gse.go,cmdb.go,notice.go,iam.go,relay.go}:internal projections of third-party concepts; raw SDK/API shapes stay in pkg/thirdparty
|Conventions:domain type first=business code should depend on pkg/types instead of pb structs or raw third-party structs
|Conventions:validation scope=structural/value correctness only; service-specific policy and cross-entity business rules stay in internal/<service>
|Conventions:enum pattern=named string/int64 type→const block→Validate() when externally selectable→list/string conversion helpers when API serialization needs them
|Conventions:union type pattern=exactly one branch set; use NewScopeWithX and NewDeploySpecWithX constructors plus Type()/GetX accessors; never set union fields directly
|Conventions:condition pattern=EntityExactFields+EntityFuzzyFields+EntityCondition with TimeRange/Page as needed; map to pkg/dao/mongo OptFn chains
|Conventions:distinct pattern=EntityDistinctRequest/Selector bool fields + EntityDistinctResult slice fields + NewXAllSet constructor for all-column aggregation
|Conventions:page pattern=Page carries sort/limit/offset; use UnlimitedPage/SingleItemPage helpers instead of magic limits
|Conventions:naming alignment=same concept name across proto,pkg/types,DAO,front docs; grep existing terms before adding synonyms
|How to add type:find analogous entity in this package→extend domain struct/enum/query contracts→update pkg/proto converters if transport-visible→update DAO/service callers only through existing layer boundaries
|Anti-patterns:no service-specific business logic|no direct proto structs as business types|no raw CMDB/GSE/IAM shapes in service code|no duplicate query/distinct helpers|no direct Scope/DeploySpec field construction|no speculative pkg/common abstractions
|Verification:go test ./pkg/types|if converter touched:go test ./pkg/proto/...|if DAO query contract touched:go test ./pkg/dao/mongo/...
