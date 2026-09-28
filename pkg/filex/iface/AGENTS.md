|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow .agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/filex/iface
|Parent rules:pkg/filex/AGENTS.md owns shared stream ownership and close-once conventions
|Overview:backend-neutral file and directory contracts, file metadata, and path-component conversion|filesystem/network implementations and service policy stay outside this package
|Where to look:contracts:iface.go:{FileGroup,File,FileContent,FileInfo,FileObject}|path helper and tests:{iface.go,iface_test.go}:ConvertAbsPathToAbsDirs
|Where to look:implementations:pkg/filex/local/{local_file.go,local_file_group.go},pkg/thirdparty/bkrepo/file_group.go|streaming consumer:pkg/filex/transfer/copy.go
|FileGroup:Name is initialization-time identity; SubGroups and AllFiles enumerate immediate children with actual I/O|AbsDirs exposes path components, not a traversal API
|Directory paths:{IsDir,GetSubGroup,EnsureSubGroup} accept group-relative slash paths; "." identifies the group itself|backend implementations own validation
|Directory errors:IsDir and GetSubGroup report missing nodes with errors matching fs.ErrNotExist; unsupported node types must not be reported as ordinary files|GetSubGroup rejects files; EnsureSubGroup creates missing parents and rejects a file at the requested path
|Copy:business callers use FileGroup.Copy; backend implementations own native dispatch and copy policy|cross-backend paths use group-relative slashes; native copies retain backend syntax; source identifies a node and destination may be "."
|Ownership:Store borrows io.ReadCloser and never closes it; caller closes after use|Content transfers a newly opened stream to its caller, which immediately defers Close
|Metadata:File.Info is cached when the File is obtained; reacquire the File for fresh metadata|FileInfo carries {Name,Size,MD5,ModTime,Description,ExtendFields}; keep backend-specific metadata in adapter handling
|FileObject:LocalFile="local",RemoteFile="remote" identify local versus remote files; preserve existing values
|Path helper:ConvertAbsPathToAbsDirs replaces backslashes with slashes, splits, and drops empty components|preserves dot/dot-dot and volume text; does not validate, resolve, or enforce containment; leading root separators are not retained
|Compatibility:before changing shared contracts inspect local and BKRepo implementations and transfer consumers|preserve public signatures, stream ownership, missing-node error matching, and cached metadata semantics
|Anti-patterns:backend SDK imports or I/O implementation here; service-specific policy in shared interfaces; treating path splitting as sanitization; assuming all backends share native path syntax
|Commands:focused helper tests:GOTOOLCHAIN=go1.25.12 go test ./pkg/filex/iface
