|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:中文主述|English technical terms/code identifiers/paths/API names
|Compression Rule:pipe-index format|concise|no prose sections|no code blocks
|Scope:docs/integration
|Overview:第三方平台接入指南|API-first|优先描述 observable behavior|不是内部开发手册
|Audience:理解 HTTP APIs/JSON/curl 的外部平台开发者|不是 bk-nodemgr 维护者|不是新手教程
|Where to look:API schema:docs/api/swagger/**:request/response shape 的事实源
|Where to look:domain concepts:docs/concepts/**:公开概念语义的事实源
|Where to look:internal development docs:docs/developer/**:仅必要时链接|不要复制内部开发内容
|Conventions:一个 cohesive integration capability 一个目录|从 end-to-end flow 开始|contract 支撑时提供 curl template|链接 canonical proto/swagger/concepts 而不是复制完整 reference
|Conventions:分离 immediate API response 与 eventual system effect|明确 observable outputs 和 limitations
|Conventions:evidence hierarchy=concept docs + proto/swagger > directly inspected public behavior > explicit unknown/待确认
|Anti-patterns:不写 router/service/storage/workflow/DAO 调用链|不把内部实现写成用户 contract|不发明 file paths/reload behavior/health checks/status APIs/timing/idempotency/retry guarantees
|Child AGENTS:docs/integration/deploy_policy/AGENTS.md
