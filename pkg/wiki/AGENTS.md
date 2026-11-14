|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format, concise, no prose/code blocks|Guide:.agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md
|Scope:pkg/wiki
|Overview:shared, version-aware product Wiki links; owns document paths and source-specific URL construction, not document hosting or service orchestration
|Where to look:documents/languages/public API:pkg/wiki/wiki.go:{Document,IURL,UnixSSHConnectionFailed,Language}
|Where to look:configuration/initialization:pkg/wiki/wiki.go:{Config,IHandler,New,SetConfig,URL}
|Where to look:GitHub validation/version-derived prefix:pkg/wiki/github.go:{GitHubConfig,githubHandler}|version parsing:pkg/version/{version.go,release.go}
|Where to look:product user content:support-files/wiki/{zh,en}/|runtime YAML/defaults:pkg/config/{config.go,backend.go}|startup wiring:internal/backend/service/service.go
|Conventions:Document values are paths relative to support-files/wiki/<language>/; keep published paths stable and synchronize constants, localized files and Wiki indexes
|Conventions:callers use document constants with .Zh()/.En(); keep mode, version and URL assembly inside this package; name platform-specific documents explicitly
|Conventions:adding a language requires IURL methods, Document implementation, compile-time interface check and matching localized content
|Conventions:source-specific inputs stay in their config block; Config validates and New selects an IHandler implementation; currently only github is supported
|Conventions:GitHub repository URL excludes tag/document path; reuse version.ParseRelease to derive the release tag without Mark; non-release builds use master
|Conventions:SetConfig initializes once before requests/workers; URL returns an empty string before initialization; preserve these semantics
|Anti-patterns:no runtime document downloads, GitHub API calls, persistence or service-specific workflow logic; no unimplemented application mode scaffolding
|Anti-patterns:no hardcoded full links or document paths in callers; no separately configured GitHub tag or duplicate version parser; do not treat developer docs/ as the product Wiki
