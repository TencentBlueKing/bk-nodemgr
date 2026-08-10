|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format|concise|no prose sections|no code blocks
|Scope:docs/integration
|Overview:third-party platform integration guides|API-first|observable behavior first|not internal developer manuals
|Audience:external platform developers who understand HTTP APIs/JSON/curl|not bk-nodemgr maintainers|not beginner tutorials
|Where to look:API schema:docs/api/swagger/**:source of request/response shapes
|Where to look:domain concepts:docs/concepts/**:source of public concept semantics
|Where to look:internal development docs:docs/developer/**:link only when needed; do not duplicate
|Conventions:one cohesive integration capability per folder|start with end-to-end flow|include curl templates when contract supports them|link canonical proto/swagger/concepts instead of copying full references
|Conventions:separate immediate API response from eventual system effect|state observable outputs and limitations explicitly
|Conventions:evidence hierarchy=concept docs + proto/swagger > directly inspected public behavior > explicit unknown/待确认
|Anti-patterns:no router/service/storage/workflow/DAO call-chain docs|no internal implementation as user contract|no invented file paths/reload behavior/health checks/status APIs/timing/idempotency/retry guarantees
|Child AGENTS:docs/integration/deploy_policy/AGENTS.md
