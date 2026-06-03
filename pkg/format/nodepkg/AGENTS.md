|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/format/nodepkg
|Overview:Shared formatter for generated node/agent package artifact names; owns stable archive naming contract for generation/release/platform/version
|Parent AGENTS:pkg/AGENTS.md
|Structure:pkg/format/nodepkg:{nodepkg.go}
|Where to look:package naming:nodepkg.go:{PkgExtension,FormatPkgFileName}:gse_<releaseType>-<generation>-<version>-<platform>.tgz
|Dependencies:pkg/types:{Generation,ReleaseType}|pkg/format/platform:{Platform}
|Contract:package filename is an archive code format consumed outside this package; do not change segments/order/separators/extension without explicit compatibility review
|Conventions:validate generation then release type; require non-empty version; validate platform for non-origin agent artifacts
|Conventions:origin agent packages are all-platform artifacts and must keep "-all.tgz" naming without OS/arch
|Conventions:error context prefix stays "format node pkg name failed"; wrap validation errors with %w where available
|Anti-patterns:no service-specific package naming|no duplicated platform validation|no speculative extension formats|no raw OS/arch strings in formatter API
|Verification:go test ./pkg/format/nodepkg
