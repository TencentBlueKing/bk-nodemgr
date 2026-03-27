|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/proto/backend/api/v3
|Overview:Backend API v3 protocol boundary; owns generated wire models|handwritten validation/conversion/defaulting; excludes business flow and storage orchestration
|Structure:pkg/proto/backend/api/v3:{*.pb.go generated,*.go handwritten supplements,README.md}
|Where to look:contracts:proto/backend/api/v3/*.proto:source of truth for request/response fields and RPC contracts
|Where to look:generated files:pkg/proto/backend/api/v3/*.pb.go:read-only generated artifacts
|Where to look:shared helpers:pkg/proto/backend/api/v3:{common.go,constant.go}
|Where to look:domain helpers:pkg/proto/backend/api/v3:{iam.go,globalsettings.go,deploy_policy.go,schedule_workflow.go,sync.go,node_agent.go,node_proxy.go,node_workflow.go,plugin.go,plugin_workflow.go,process.go,cipher.go,configpolicy.go}
|Where to look:usage guide:pkg/proto/backend/api/v3/README.md:proto lifecycle confinement and package intent
|Conventions:edit handwritten .go files only unless contract changes require proto edit+regen
|Conventions:proto lifecycle stays here=proto→pkg/types before business logic|pkg/types→proto for responses
|Conventions:before adding converters/helpers, search this directory for analogous implementations and extend existing message methods over parallel logic
|Conventions:prefer generated-message methods/helpers over brand-new mirror types
|Conventions:semantic defaults belong in AutoConvert or conversion helpers; normalize nil vs zero explicitly
|How to change:contract change=edit proto/backend/api/v3/*.proto→cd proto && make clean && make all→update handwritten converters/validators here
|Testing:go test ./pkg/proto/backend/api/v3|focused=go test ./pkg/proto/backend/api/v3 -run 'TestDeployPolicy'
|Anti-patterns:no hand edits to *.pb.go|no treating generated files as business logic|no using proto structs directly in upper business layers|no duplicate conversion/validation outside this package|no mirror types without strong reason|no new parallel converters before checking existing files in this directory
