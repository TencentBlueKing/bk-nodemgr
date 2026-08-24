|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (scoping)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format|concise|no prose sections|no code blocks
|Scope:docs/integration/import_package
|Overview:plugin package import third-party integration guide|cohesive function folder|curl-first|result-oriented
|Structure:docs/integration/import_package:{AGENTS.md,README.md,import_v3_plugin.md,import_v2_plugin.md,import_v2_external_plugin.md,import_result.md}
|Source of truth:import contract=docs/api/swagger/backend/api/v3/pkg_workflow.swagger.json|import request/response variants=proto/backend/api/v3/pkg_workflow.proto|workflow action/log types=proto/backend/api/v3/workflow.proto
|Conventions:README.md is the flow entrypoint|organize by integration flow|put the import action and the import result as separate mode pages
|Conventions:mode page sections={Purpose and applicability,Input,Minimal payload and curl template,Immediate output,Eventual or machine-visible artifact,Repeat behavior,Failure cases and limits,Contract references}
|Conventions:each executable mode page must include a curl template|initialize shell variables|extract returned IDs when reused|do not invent auth/tenant headers|do not include secrets/real identifiers
|Conventions:separate API accepted/created IDs from eventual machine-visible effects|mark public-contract gaps as 待确认 or not specified
|Conventions:import variant mapping|import_v3_plugin -> operation_id package_plugin_v3_import|import_v2_plugin -> operation_id package_plugin_v2_import|import_v2_external_plugin -> operation_id package_external_plugin_v2_import|all three share the same request shape (filename/download_url/md5) and the same import_result read path
|import_v3_plugin:import endpoint for v3 plugin package|returns workflow_id as the integration handle for the result read
|import_v2_plugin:import endpoint for official v2 plugin package|returns workflow_id as the integration handle for the result read
|import_v2_external_plugin:import endpoint for external v2 plugin package|returns workflow_id as the integration handle for the result read
|import_result:import workflow has no per-host scope, so a dedicated result endpoint is required|response carries operation source info via data.operations[] (operation_id, last_instance_id, oper_inst_logs, extra_execution_logs)|applies to all three import variants; do not redirect to generic workflow list endpoints (those are scoped to host-bound workflows)
|Anti-patterns:no full Swagger duplication|no internal call chains|no undocumented operation/status endpoint|no invented machine paths/health/idempotency/retry/timing/timeout guarantees|no state-machine enums beyond the response shape
