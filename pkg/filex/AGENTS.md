|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format; concise; no prose/code blocks|Reference:.agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md
|Scope:pkg/filex
|Overview:shared file interfaces, local file access, and close-once stream wrappers|service-specific upload/publish policy belongs in internal/<service>
|Structure:pkg/filex:{close_once.go,iface,local,filelock}
|Where to look:ownership contracts:pkg/filex/iface/iface.go:{FileGroup.Store,FileContent.Content}
|Where to look:close-once wrappers:pkg/filex/close_once.go:{OnceReadCloser,OnceReadWriteCloser}
|Where to look:local stream acquisition:pkg/filex/local/local_file.go:LocalFile.Content
|Where to look:local storage and destination cleanup:pkg/filex/local/local_file_group.go:LocalDir.Store
|Ownership:the opener owns the stream and registers defer immediately after successful acquisition; cover early returns and panic unwinding
|Borrowing:passing a stream as an argument does not transfer ownership; callees must not close borrowed streams, including on errors
|Store:borrows its input stream; caller closes input after use|implementation owns and defers closure of destination files it opens
|Content:returns a newly opened stream with ownership transferred to the caller; caller immediately registers defer|never close a successfully returned stream before its caller can consume it
|Interfaces:prefer io.Reader/io.Writer for new borrowing APIs; io.ReadCloser in an existing signature alone does not imply ownership transfer|keep public signatures stable
|Close-once:wrap a non-nil stream once at acquisition or the ownership-return boundary; all consumers share that wrapper|reuse OnceReadCloser/OnceReadWriteCloser instead of parallel wrappers
|Close-once semantics:underlying Close executes at most once; later calls return the cached close error|does not prevent premature Close or make concurrent Read/Write safe
|Cleanup:close-once is a safeguard, not permission for borrowers to close or a substitute for owner defer|do not bypass the wrapper or separately wrap aliases of the same underlying resource
|Errors:preserve operation and relevant Close errors with errors.Join in deferred cleanup targeting the returned error|avoid shadowed errors; do not return a live stream alongside a cleanup failure
|Integration:inspect callee ownership contracts before passing streams to external adapters; isolate borrowed inputs from APIs that close them when safe|do not silently migrate existing ownership-transfer APIs outside this scope
|Anti-patterns:Store closing its input; closing only in explicit error branches; hidden ownership transfer; per-layer Close calls justified by close-once; duplicate close-once implementations
