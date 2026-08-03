|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/relayhandler
|Overview:Shared backend↔relay GSE message adapter; owns client/server messager interfaces, callback request/response, ack, push-to-client, processed-state guards
|Overview:Infrastructure boundary only; coordinate relay messages and protoRelay wire payloads, never service-specific policy or workflow decisions
|Structure:pkg/relayhandler:{relay.go interfaces/contracts,client.go agent-message client side,server.go server-api side,constant.go shared PluginName,handler_test.go push smoke test,README.md package stub,AGENTS.md scope rules}
|Where to look:interface contract:relay.go:{IClientMessager,ICallbackClient,IClientPush,IServerMessager,IPushServer,ICallbackServer,ServerReceivedData}
|Where to look:client-side flow:client.go|agentmessage.New→Launch→messageCallback→{CallbackResp,AckReq,ServerPushReq}|RequestCallback waits on context; ClientPushReq waits for ack tracker
|Where to look:server-side flow:server.go|serverapi.New→DecodeBaseRequest→Decode{Callback,Ack,ClientPush}Request→RespondCallback/PushToClient/SendAck|redis tracker owns ack/processed state
|Where to look:wiring:internal/relay/service/service.go:NewClientMessager|internal/backend/service/service.go:NewServerMessager|capabilities hold interfaces, not concrete structs
|Where to look:callers:internal/relay/{handler,router/api-v3/callback}|internal/backend/router/api-v3/proxy|internal/backend/manager/workflowdef/{node,plugin,pluginv2}
|Dependencies:external GSE SDK:{agent-message,server-api}|state:{internal/relay/messagetracker,redis}|context/logging:{pkg/contextx,pkg/logger}|wire:{pkg/proto/relay}|helpers:{pkg/runtime/retrier,pkg/runtime/conv,pkg/identifier}
|Related AGENTS:pkg/AGENTS.md|pkg/contextx/AGENTS.md|pkg/runtime/conv/AGENTS.md|pkg/thirdparty/gse/AGENTS.md
|Conventions:Keep interfaces narrow and caller-facing; extend IClientMessager/IServerMessager additively and preserve Start/Stop/Decode/Push/Callback/Ack return semantics
|Conventions:All long-running send/wait paths must honor contextx.IContext cancellation/deadline; timeout/ack loops should return ctx.Err when context ends
|Conventions:Preserve message state semantics:{TryMarkProcessed idempotency,MarkAcked ack contract,IsAcked polling,processed-before-dispatch guard}
|Conventions:Use protoRelay only as wire boundary; decode/encode inside relayhandler and pass upper layers ServerReceivedData or interface methods, not raw GSE callbacks
|Conventions:Use pkg/logger with message-id/original-message-id/event-type/agentIDs context; wrap returned errors with fmt.Errorf("...: %w", err) when adding operation context
|Conventions:Use retrier for dispatch attempts and bounded ack waits; do not add unbounded goroutines, sleeps, or polling loops without context cancellation
|Conventions:PluginName is shared workflow identity; update all workflowdef plugin/pluginv2 references if its meaning changes
|Anti-patterns:Do not put backend/relay route policy, workflow decisions, host/plugin business rules, or storage schema decisions in this package
|Anti-patterns:Do not leak GSE-native request/response structs to internal callers or make callers construct agentmessage/serverapi clients directly
|Anti-patterns:Do not bypass messagetracker for ack/processed state or duplicate parallel relay message trackers in callers
|Anti-patterns:Do not add new proto/front/API fields for relay messaging unless endpoint semantics require them and proto converters/callers are updated together
|Commands:go1.25.12 test ./pkg/relayhandler
