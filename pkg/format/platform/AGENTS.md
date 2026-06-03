|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/format/platform
|Overview:Shared platform normalization and validation utilities; owns cross-system OS/CPUArch standardization, alias mapping, Platform value object, and binary suffix formatting
|Parent AGENTS:pkg/AGENTS.md
|Structure:pkg/format/platform:{platform.go}
|Where to look:OS normalization:platform.go:{StandardOSMap,NormalizeOS}:aliases→pkg/runtime/criteria.OSType
|Where to look:arch normalization:platform.go:{StandardArchMap,NormalizeArch,parseArmVersion}:aliases/armv*→pkg/runtime/criteria.CPUArch
|Where to look:platform value:platform.go:{Platform,UnknownPlatform,NewPlatform,Normalize,String,Validate,ValidPlatforms}:normalized pair plus supported-combination check
|Where to look:binary naming:platform.go:{FormatBinaryFileName}:Windows adds .exe; other OS keeps original name
|Dependencies:pkg/runtime/criteria is the canonical enum source for OS and CPU arch values
|Contract:normalization absorbs naming differences between systems; do not remove aliases or change mappings without checking generated artifact compatibility
|Conventions:normalize raw strings at package boundary; downstream formatters should consume Platform or criteria enums, not duplicate alias maps
|Conventions:add aliases in StandardOSMap/StandardArchMap only when they map unambiguously to criteria values; keep ValidPlatforms aligned with supported GOOS/GOARCH pairs
|Conventions:return explicit errors for empty/unknown OS or arch; preserve partial Normalize result with combined error semantics
|Anti-patterns:no service-specific platform aliases|no duplicated OS/arch maps in sibling packages|no silent unknown fallback except UnknownPlatform constructor
|Verification:go test ./pkg/format/platform
