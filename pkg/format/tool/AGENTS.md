|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/format/tool
|Overview:Shared formatter for generated Node Manager tool artifacts; owns stable installer archive naming and OS-aware path joining
|Parent AGENTS:pkg/AGENTS.md
|Structure:pkg/format/tool:{tool.go}
|Where to look:installer naming:tool.go:{NamePrefixInstaller,FormatInstallerName}:installer_<os>_<arch> plus Windows .exe suffix
|Where to look:path joining:tool.go:{JoinPath}:criteria.OSWindows→pkg/runtime/winpath.Join; other OS→filepath.Join
|Dependencies:pkg/runtime/criteria for validated OSType/CPUArch|pkg/runtime/winpath for Windows path semantics
|Contract:installer filename is an archive code format consumed outside this package; do not change prefix/order/separators/suffix without explicit compatibility review
|Conventions:validate OS and CPU arch before producing artifact names; return fmt.Errorf with %w for validation errors
|Conventions:keep Windows suffix/path behavior centralized here; callers should pass criteria types, not raw strings
|Anti-patterns:no caller-specific naming branches|no hardcoded slash joins for Windows|no unvalidated os/arch in artifact names
|Verification:go test ./pkg/format/tool
