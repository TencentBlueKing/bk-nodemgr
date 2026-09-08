|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/proto/relay
|Overview:Hand-written JSON wire contract for backend↔relay GSE messages|NOT protobuf-generated, unlike sibling pkg/proto/*/api/v3:{no .proto source,no protoc regen,edit relay.go directly}
|Overview:Owns the transport envelope plus every server→relay push payload|pure data types only, zero behaviour
|Overview:Direction boundary:server→relay push lives here|relay→backend report payloads live in pkg/proto/backend/callback
|Structure:pkg/proto/relay:{relay.go all types+consts,README.md stub,AGENTS.md scope rules}
|Where to look:transport envelope:relay.go:{MessageType consts,Base,CallbackReq,CallbackResp,ServerPushReq,ClientPushReq,AckReq}
|Where to look:push event registry:relay.go:{ServerPushEventType consts:check_pkg_state,notify_receive,detect_info_by_{ssh,wmi,windows_ssh},install_by_{ssh,windows_ssh,wmi},install_proxy_by_ssh}
|Where to look:push payloads:relay.go:{CheckPkgStateReq,NotifyReceiveReq,FileInfo,DetectInfoBy{SSH,WMI,WindowsSSH}Req,InstallPagentBy{SSH,WindowsSSH,WMI}Req,InstallProxyBySSHReq}
|Where to look:envelope encode/decode:pkg/relayhandler:{client.go client side,server.go server side}
|Where to look:event→handler registration:internal/relay/service/service.go:173-181|dispatch:internal/relay/manager/dispatcher.go
|Where to look:senders:internal/backend/manager/workflowdef/node:{action_ensure_pkg_to_relay.go,action_pagent_detect_info_by_*.go,action_install_pagent_by_*.go,action_install_proxy_by_ssh.go}
|Where to look:receivers:internal/relay/handler:{ensure_pkg.go,detect_info_by_*.go,install_by_*.go,install_proxy_by_ssh.go}
|Where to look:reverse direction payloads:pkg/proto/backend/callback/nodeinstall.go:{ReportDetectResultReq,ReportStorageResultReq}|relay-side senders:internal/relay/handler/report.go
|Flow:backend action→PushToClient(eventType,payload)→GSE→relay messageCallback→dispatcher.Dispatch→handler→ClientPushReq→backend proxy router→callback router
|Conventions:Every push payload carries ActionName+OperInstID; they are the only correlation keys the relay has and the relay derives per-instance staging paths from OperInstID
|Conventions:Adding an event needs four edits together:{ServerPushEventType const,payload struct,IHandler method,RegisterHandler in service.go}
|Conventions:Extend additively and keep json tags stable; relay and backend are separately deployed binaries so a removed/renamed tag silently breaks the peer
|Conventions:Deprecate instead of delete when a field is superseded, keep it decodable for older peers|reference:NotifyReceiveReq.PkgName→FileList
|Conventions:Payloads that carry FileInfo.FileMD5 are the relay's only integrity signal; the relay is fail-closed and refuses to cache a package when the md5 is absent or malformed
|Conventions:Keep types flat and json-tagged; conversion, defaulting and validation belong to pkg/relayhandler or the owning internal/* layer
|Anti-patterns:Do not treat this package as generated code|no protoc, no `cd proto && make all`, no .pb.go conventions
|Anti-patterns:Do not add methods, validation, defaulting or conversion helpers here; it must stay a dependency-free data package
|Anti-patterns:Do not log payload structs verbatim; DetectInfo*/InstallPagent*/InstallProxy* all carry a plaintext Password field
|Anti-patterns:Do not trust CheckPkgStateReq.FileStorageTmpDir; it is dead (never written, never read) and the staging dir actually flows the other way via reportRelayFileState.StorageTmpDir
|Anti-patterns:Do not assume event const names match handler names; ServerPushEventTypeInstallBy{SSH,WindowsSSH,WMI} map to InstallPagentBy* handlers
|Related AGENTS:pkg/AGENTS.md|pkg/relayhandler/AGENTS.md
|Commands:go1.25.12 build ./pkg/proto/relay
