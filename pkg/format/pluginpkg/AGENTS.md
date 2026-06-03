|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/format/pluginpkg
|Overview:Shared formatter for generated plugin package artifact names; owns stable archive naming contract for release-type/generation/platform/version
|Parent AGENTS:pkg/AGENTS.md
|Structure:pkg/format/pluginpkg:{plugin.go}
|Where to look:package naming:plugin.go:{PkgExtension,FormatPkgFileName}:bk-nodemgr_<releaseType>_<gen>_<plugin>-<version>-<platform>.tgz
|Dependencies:pkg/types:{ReleaseType,Generation}|pkg/format/platform:{Platform}
|Contract:package filename is an archive code format consumed outside this package; do not change segments/order/separators/extension without explicit compatibility review
|Conventions:validate release type before formatting; require non-empty version; validate platform only for platform-specific plugin packages
|Conventions:origin plugin packages are all-platform artifacts and must keep "-all.tgz" naming without OS/arch
|Conventions:error context prefix stays "format plugin pkg name failed"; wrap validation errors with %w where available
|Anti-patterns:no service-specific package naming|no duplicated platform validation|no speculative extension formats|no raw OS/arch strings in formatter API
|Verification:go test ./pkg/format/pluginpkg
