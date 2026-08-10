|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:docs/integration/deploy_policy
|Overview:deploy_policy third-party integration guide|cohesive function folder|curl-first|desired-state oriented
|Structure:docs/integration/deploy_policy:{README.md,specify_plugin.md,specify_plugin_pkg.md,specify_plugin_sub_config.md}
|Source of truth:spec semantics=docs/concepts/deploy_policy/spec.md|scope semantics=docs/concepts/deploy_policy/scope.md|API schema=docs/api/swagger/backend/api/v3/deploy_policy.swagger.json and proto/backend/api/v3/deploy_policy.proto
|Conventions:README.md is the flow entrypoint|organize by integration flow|put spec modes under Declare desired state|link mode pages for details
|Conventions:mode page sections={Purpose and applicability,Input,Minimal payload and curl template,System interpretation,Immediate output,Eventual or machine-visible artifact,Repeat behavior,Failure cases and limits,Contract references}
|Conventions:each mode page must include a curl template|use placeholders only|state that auth/tenant headers are deployment-specific|do not include secrets/real tenant/business/host/plugin identifiers
|Conventions:separate API accepted/created IDs from eventual machine-visible effects|mark public-contract gaps as 待确认 or not specified
|specify_plugin_sub_config:config-only for an already installed plugin|does not install/upgrade plugin|do not promise automatic config push through deploy_policy until execution chain is confirmed
|Anti-patterns:no full Swagger duplication|no internal call chains|no undocumented operation/status endpoint|no invented machine paths/reload/health/idempotency/retry/timing guarantees
