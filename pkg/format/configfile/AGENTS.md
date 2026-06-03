|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/format/configfile
|Overview:Shared formatter for generated config file bytes; owns deterministic readable JSON marshaling for configfile content
|Parent AGENTS:pkg/AGENTS.md
|Structure:pkg/format/configfile:{configfile.go}
|Where to look:JSON formatting:configfile.go:{FormatConfigFileJson,configFileJsonPrefix,configFileJsonIndent}:stable indentation for config file output
|Boundary:input must implement json.Marshaler|this package formats bytes only; config schema/validation belongs in pkg/config or domain packages
|Contract:generated config files must remain human-readable; indentation/prefix choices are output format compatibility and should not be changed casually
|Conventions:preserve deterministic output shape; wrap or return marshal errors without masking root cause
|Conventions:exported functions need English godoc comments ending with period; keep package comment aligned with package role
|Anti-patterns:no config business validation|no file IO side effects|no service-specific config branches|no ad-hoc string-built JSON
|Verification:go test ./pkg/format/configfile
