|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/identifier
|Overview:Shared, business-agnostic runtime identifier generator; produces recognizable scene-tagged identifiers for logs and cross-service flows
|Boundary:pkg/identifier owns tag definitions and identifier formatting|callers own business semantics, persistence, validation, and lifecycle|internal/* owns service orchestration
|Structure:pkg/identifier:{identifier.go,identifier_test.go,README.md,AGENTS.md}
|Where to look:identifier format:identifier.go:{generateID,tagRequestID,tagMessageID,tagWorkflowID,tagTriggerID,tagOperationID,tagOperationInstanceID,tagActionInstanceID,tagServiceID,tagUploadID,tagExportID}
|Where to look:public generators:identifier.go:{GenRequestID,GenMessageID,GenWorkflowID,GenTriggerID,GenOperationID,GenOperationInstanceID,GenActionInstanceID,GenServiceID,GenUploadID,GenExportID}
|Where to look:package intent:README.md:{design intent,function boundary,design considerations}
|Where to look:generator coverage:identifier_test.go:{TestGenID}
|Conventions:all identifiers use tag:32-character hyphenless UUID|keep tags lowercase, concise, scene-distinct, and human-recognizable in logs
|Conventions:add an ID kind only after confirming no existing generator has the same domain meaning|add a dedicated Gen*ID public function with an English godoc comment|reuse generateID rather than formatting UUIDs at callers
|Conventions:keep generation stateless, deterministic in format, and free from context, persistence, network, or service-specific dependencies
|Anti-patterns:no ID-related business policy, validation, or lifecycle management in pkg/identifier|no direct UUID/tag concatenation outside this package|no overloaded tag for distinct business scenes|no speculative tags or generators without an actual flow
|Verification:go test ./pkg/identifier
