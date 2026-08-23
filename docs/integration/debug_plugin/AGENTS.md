|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format|concise|no prose sections|no code blocks
|Scope:docs/integration/debug_plugin
|Overview:plugin debug third-party integration guide|cohesive function folder|curl-first|workflow-poll oriented
|Structure:docs/integration/debug_plugin:{AGENTS.md,README.md,start_debug.md,stop_debug.md,read_debug_logs.md}
|Source of truth:start_debug/stop_debug contracts=docs/api/swagger/backend/api/v3/plugin.swagger.json|debug request/response variants=proto/backend/api/v3/plugin.proto|workflow read contracts=docs/api/swagger/backend/api/v3/plugin_workflow.swagger.json|domain terms=docs/concepts/glossary.md and docs/concepts/plugin/README.md
|Conventions:README.md is the flow entrypoint|organize by integration flow|put start/stop as separate mode pages|put the workflow read sequence in a dedicated read mode page
|Conventions:mode page sections={Purpose and applicability,Input,Minimal payload and curl template,Immediate output,Eventual or machine-visible artifact,Repeat behavior,Failure cases and limits,Contract references}
|Conventions:each executable mode page must include a curl template|initialize shell variables|extract returned IDs when reused|do not invent auth/tenant headers|do not include secrets/real identifiers
|Conventions:separate API accepted/created IDs from eventual machine-visible effects|mark public-contract gaps as 待确认 or not specified
|start_debug:ScopeInstance is the public scope shape shared with deploy_policy|debug_info fields are defined by the DebugInfo proto variant|returns workflow_id as the integration handle for downstream read and stop calls
|stop_debug:non-blocking|only requires workflow_id|signals the running debug session to terminate|does not wait for the running debug command to exit
|read_debug_logs:composes 4 existing public workflow endpoints (workflow/list -> operation/list -> instance/list -> log/get)|debug streaming output is read via the same log endpoint the front-end task history uses|polling cadence is integrator's responsibility
|Anti-patterns:no full Swagger duplication|no internal call chains|no undocumented operation/status endpoint|no invented machine paths/health/idempotency/retry/timing/timeout guarantees|no state-machine enums beyond the response shape
