|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format|concise|no prose sections|no code blocks
|Scope:docs/integration/deploy_policy
|Overview:deploy_policy third-party integration guide|cohesive function folder|curl-first|desired-state oriented
|Structure:docs/integration/deploy_policy:{README.md,specify_plugin.md,specify_plugin_pkg.md,specify_plugin_sub_config.md}
|Source of truth:spec semantics=docs/concepts/deploy_policy/spec.md|scope semantics=docs/concepts/deploy_policy/scope.md|API endpoints/envelopes=docs/api/swagger/backend/api/v3/deploy_policy.swagger.json|variant fields=proto/backend/api/v3/deploy_policy.proto
|Conventions:README.md is the flow entrypoint|organize by integration flow|put spec modes under Declare desired state|link mode pages for details
|Conventions:mode page sections={Purpose and applicability,Input,Minimal payload and curl template,System interpretation,Immediate output,Eventual or machine-visible artifact,Repeat behavior,Failure cases and limits,Contract references}
|Conventions:each executable mode page must include a curl template|initialize shell variables|extract returned IDs when reused|do not invent auth/tenant headers|do not include secrets/real identifiers
|Conventions:separate API accepted/created IDs from eventual machine-visible effects|mark public-contract gaps as 待确认 or not specified
|specify_plugin_sub_config:config-only for an already installed plugin|does not install/upgrade plugin|execute converges declared non-main config files|include execute curl|separate trigger-id from final config effect
|Unsupported mode:specify_plugin_pkg_sub_config appears in concept docs but is absent from current proto/swagger/types|document as unavailable
|Anti-patterns:no full Swagger duplication|no internal call chains|no undocumented operation/status endpoint|no invented machine paths/reload/health/idempotency/retry/timing guarantees
