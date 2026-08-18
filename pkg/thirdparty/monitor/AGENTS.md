|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/thirdparty/monitor
|Overview:Monitor API Gateway provider for get-or-create Agent event data_id by bk_biz_id|exposes IHandler and explicit disabled no-op behavior
|In scope:APIGW client wiring|provider config|wire response model|response validation|disabled semantics
|Out of scope:BizEventDataIDConf persistence|agentBaseAlarmEventDataID resolution|global defaults|workflow policy
|Structure:pkg/thirdparty/monitor:{handler.go,monitor.go,types.go,noop.go,README.md,AGENTS.md}
|Where to look:caller contract+response validation:handler.go:{IHandler,Handler,New,GetOrCreateAgentEventDataID}
|Where to look:APIGW client+endpoint request:monitor.go:{cli,newClient,getOrCreateAgentEventDataID}
|Where to look:provider config+wire broker:types.go:{Config,BaseBroker,GetOrCreateAgentEventDataIDResp,Validate,IsFailed}
|Where to look:disabled mode:noop.go:{NoOpHandler,NewNoOpHandler,GetOrCreateAgentEventDataID}
|Where to look:integration contract+fallback ownership:README.md|handler initialization:internal/backend/service/service.go:newMonitorHandler
|Type flow:bk_biz_id→IHandler→Monitor APIGW→bk_data_id→backend resolver|raw BaseBroker remains package-local
|Conventions:callers depend on IHandler only|low-level cli and APIGW wire types never escape this package
|Conventions:newClient validates Config before use|endpoint path and query parameter assembly stay in monitor.go
|Conventions:GetOrCreateAgentEventDataID requires bk_biz_id>0 and returned bk_data_id>0
|Conventions:real handler returns (dataID,true,nil)|no-op returns (0,false,nil) without outbound requests
|Conventions:found=false means caller-owned fallback|this provider never supplies or persists runtime defaults
|Conventions:transport/broker failures preserve wrapped causes|do not log credentials,auth headers,or access tokens
|Dependencies:pkg/thirdparty/apigw/client:authenticated APIGW client|pkg/rest/client:request transport|pkg/contextx:IContext
|Parent rules:pkg/thirdparty/AGENTS.md|pkg/AGENTS.md
|Anti-patterns:no storage reads/writes or BizEventDataIDConf construction in this provider
|Anti-patterns:no merge of agentBaseAlarmEventDataID,taskProcEventDataID,or global defaults
|Anti-patterns:no outbound request when disabled|no raw API Gateway response exposed to internal/backend callers
|Anti-patterns:no endpoint literals outside monitor.go|no provider client construction outside backend service wiring
|Commands:go test ./pkg/thirdparty/monitor|compile=go test ./pkg/thirdparty/monitor -run '^$'
