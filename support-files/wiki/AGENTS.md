|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format, concise, no prose/code blocks|Guide:.agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md
|Scope:support-files/wiki
|Overview:product user Wiki delivered with application and readable on GitHub; not developer or deployment-maintainer documentation
|Where to look:embedded resources:support-files/wiki/embed.go|localized entrypoints:support-files/wiki/{zh,en}/README.md
|Where to look:document constants/link construction:pkg/wiki/|HTTP access:internal/application/router/wiki/
|Conventions:embed only localized zh/en trees; synchronize localized files, README links and pkg/wiki document constants
|Conventions:published paths are stable; use relative Markdown links and local images for offline access; keep instructions short, actionable and aimed at computer-literate product users
|Anti-patterns:no copying all docs/ into the product Wiki; no external runtime resources, credentials, internal-only diagnostics or explanatory filler; no embedding root Go/AGENTS files
