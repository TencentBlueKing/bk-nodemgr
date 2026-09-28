|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A and README|English for code and agent guidance
|Compression Rule:Follow .agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/filex/transfer
|Overview:low-level streaming copy between FileGroup instances via CopyStream; backend Copy methods own dispatch
|Where to look:copy contract:copy.go:{CopyStream,copyDir,copyFile,storeFile,validateChildName}|stream cleanup tests:copy_test.go
|Conventions:cross-backend paths are group-relative slash paths; backend directory methods own path validation
|Conventions:pass original paths to IsDir before path.Clean so normalization cannot hide invalid input|CopyStream rejects root as source and permits it as destination
|Conventions:existing destination directory receives source basename; directory copy into a file fails; directories merge recursively
|Conventions:overwrite passes to target Store for each file; directory merge retains unrelated entries|metadata behavior is backend-specific
|FileGroup:IsDir reports missing paths with fs.ErrNotExist|GetSubGroup gets an existing directory|EnsureSubGroup creates missing parents
|Traversal:SubGroups and AllFiles list immediate children; validateChildName rejects empty, dot, dot-dot, slash, backslash and NUL names|source IsDir still checks enumerated files because AllFiles may include unsupported copy nodes
|Destination:only fs.ErrNotExist permits a missing target; other query errors stop copy|storeFile checks the final file target after directory/basename resolution and rejects directories before opening content
|Entry points:business callers use FileGroup.Copy; adapters use CopyStream after native dispatch|CopyStream does not check source/target identity or directory overlap and does not apply tenant policy; callers own these preconditions
|Ownership:CopyStream owns source Content streams; defer Close immediately and preserve operation and close errors with errors.Join|Store borrows input
|Failure:stop at first error and return it; partial destination writes are not rolled back
|Anti-patterns:do not introduce parallel ReadDir/GetFile/Store endpoint APIs, use AbsDirs for traversal, or bypass native dispatch
