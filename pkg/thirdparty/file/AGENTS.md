|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Use concise pipe-index rows
|Scope:pkg/thirdparty/file
|Overview:File service client and handler adapters
|Handler assembly:`handler.go` owns `IHandler`, the concrete `handler`, and `New`|`IHandler` directly composes every exported capability interface
|Handler capabilities:Place each exported capability interface and its `*handler` implementations in the corresponding domain file|examples:{upload.go,release.go,plugin_requirement.go,sync.go}|business modules call the File service through handler interfaces
|Client layer:`file.go` owns client setup and all private `*cli` methods|keep File service HTTP and wire protocol details inside this package
|Signature wrapping:Keep function signatures on one line when they fit|otherwise break before the return list at its first opening parenthesis|if still too long, break after the parameter list's first opening parenthesis|match existing File adapter style
