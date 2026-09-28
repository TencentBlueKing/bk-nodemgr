|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A and README|English for code and agent guidance
|Compression Rule:Follow .agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/thirdparty/bkrepo
|Overview:BKRepo client, file group and scenario-oriented adapter|keep protocol requests and response mapping inside this package
|Where to look:client and requests:bkrepo.go|handler operations:handler.go|file metadata/content:file.go|group methods:file_group.go|wire types:types.go|headers:header.go
|Copy:FileGroup.Copy uses CopyNode for two BKRepo groups only with the same Handler, project and repository|other supported backends use pkg/filex/transfer.CopyStream stream copy
|Paths:directoryPath reuses resolveNodePath for group-relative slash paths; reject empty, absolute, backslash, NUL and escaping paths|native CopyNode keeps resolveNodePath behavior|source must name a node; destination may be "." for group root
|Directories:GetSubGroup(".") queries live state|EnsureSubGroup queries the target first and creates parents only for errNodeNotFound; propagate other errors|Handler.EnsureFileGroup requeries once after mkdir failure and accepts only an existing directory
|Errors:IsDir/GetSubGroup join fs.ErrNotExist with errNodeNotFound so callers can match either|UploadFile checks business response failure before returning data
|Streams:DownloadFile returns a caller-owned stream; Store borrows reader|UploadFile wraps borrowed input with io.NopCloser because HTTP transport closes request bodies
|Conventions:expose scenario APIs via Handler; keep BKRepo protocol models inside this package
|Anti-patterns:do not depend on BKRepo native CopyNode across projects, repositories, handlers, or non-BKRepo backends|do not expose protocol models to upstream modules
|Tests:bkrepo_test.go covers native copy requests, directory methods and local/BKRepo streaming through httptest
