|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/proto/file/api/v3
|Overview:File API v3 protocol boundary; owns generated wire models|handwritten validation/conversion/defaulting; excludes business flow and storage orchestration
|Structure:pkg/proto/file/api/v3:{_.pb.go generated,_.go handwritten supplements,README.md}
|Where to look:contracts:proto/file/api/v3/_.proto:source of truth for request/response fields and RPC contracts
|Where to look:generated files:pkg/proto/file/api/v3/_.pb.go:read-only generated artifacts
|Where to look:shared helpers:pkg/proto/file/api/v3:{common.go}
|Where to look:domain helpers:pkg/proto/file/api/v3:{download.go,upload.go,publish.go,transfer.go}|info/export wire models live only in _.pb.go (no handwritten domain helper)
|Where to look:usage guide:pkg/proto/file/api/v3/README.md:proto lifecycle confinement and package intent
|Conventions:edit handwritten .go files only unless contract changes require proto edit+regen
|Conventions:proto lifecycle stays here=proto→pkg/types before business logic|pkg/types→proto for responses
|Conventions:pkg/proto/_* only handles transport conversion (proto <-> pkg/types); business decisions/result construction must stay in internal/<service> layers
|Conventions:before adding converters/helpers, search this directory for analogous implementations and extend existing message methods over parallel logic
|Conventions:prefer generated-message methods/helpers over brand-new mirror types
|Conventions:semantic defaults belong in AutoConvert or conversion helpers; normalize nil vs zero explicitly (e.g., id=-1 when zero is meaningful)
|Conventions:platform helper shared across domains:ConvertPlatformToTypes|ConvertPlatformFromTypes (common.go:28/36) — reuse instead of redefining
|Conventions:transfer-domain uses ISimpleTransferHandler contract from pkg/types for ConvertResult responses (transfer.go:68/124/152/180)
|How to change:contract change=edit proto/file/api/v3/_.proto→cd proto && make clean && make all→update handwritten converters/validators here
|Testing:go test ./pkg/proto/file/api/v3
|Anti-patterns:no hand edits to *.pb.go|no treating generated files as business logic|no using proto structs directly in upper business layers|no business-result derivation in pkg/proto/*_ (e.g., request-derived authorized results)|no duplicate conversion/validation outside this package|no mirror types without strong reason|no new parallel converters before checking existing files in this directory
