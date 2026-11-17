|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format, concise, no prose/code blocks|Guide:.agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md
|Scope:internal/application/router/wiki
|Overview:anonymous, read-only HTTP access to embedded product Wiki; no business API, authentication decision or remote content retrieval
|Where to look:registration/file handling:internal/application/router/wiki/wiki.go:{Load,serveFile}|basic server wiring:internal/application/service/service.go:{registerBasicServer,withWiki}
|Where to look:content access:pkg/wiki/content.go:ReadFile|embedded source:support-files/wiki/embed.go|link contracts:pkg/wiki/{wiki.go,prefix.go}
|Conventions:GET/HEAD /wiki/<language>/<document>; delegate language/path validation and content reads to pkg/wiki.ReadFile
|Conventions:serve Markdown as UTF-8 plain text with nosniff; preserve relative document/resource URLs; directory and unknown paths return 404
|Conventions:serve only content returned by pkg/wiki; keep HTTP status, headers and response rendering here; keep route URLs stable if adding a reading view later
|Anti-patterns:no OS filesystem reads, directory listings, developer docs/ exposure, GitHub fetches, business capability dependencies or proto endpoints; no silent path cleaning that accepts traversal
