|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format|concise|no prose sections|no code blocks
|Scope:docs/integration/plugin_migration
|Overview:plugin V2 to V3 process migration third-party integration guide|cohesive function folder|curl-first|workflow-poll oriented
|Structure:docs/integration/plugin_migration:{AGENTS.md,README.md,list_v2_processes.md,list_v3_processes.md,sync_v2_alive_process_info.md,sync_v2_stopped_process_info.md,install_v3_no_start.md,stop_v2.md,start_plugin.md}
|Source of truth:V2/V3 process list=docs/api/swagger/backend/api/v3/process.swagger.json|V2 process sync=docs/api/swagger/backend/api/v3/sync.swagger.json|install_no_start/stop_v2/start contracts=docs/api/swagger/backend/api/v3/plugin.swagger.json|workflow read contracts=docs/api/swagger/backend/api/v3/plugin_workflow.swagger.json|offline process data migration=docs/operation/migrate/nodeman_process_to_nodemgr/README.md
|Conventions:README.md is the only end-to-end flow entrypoint|put data migration, interface ordering, workflow reads, and final verification only in README|interface pages describe only their own endpoint
|Conventions:mode page sections={Purpose and applicability,Input,Minimal payload and curl template,System interpretation,Immediate output,Eventual or machine-visible artifact,Repeat behavior,Failure cases and limits,Contract references}
|Conventions:each executable mode page must include a curl template|initialize shell variables|extract returned IDs when reused|do not invent auth/tenant headers|do not include secrets/real identifiers
|Conventions:separate API accepted/created IDs from eventual machine-visible effects|mark public-contract gaps as 待确认 or not specified
|Conventions:all migration interfaces are bk-nodemgr endpoints; use process/list_v2 + process/list + sync/gse/plugin/process_v2/info/alive + sync/gse/plugin/process_v2/info/stopped + plugin/install_no_start + plugin/stop_v2 + plugin/start
|list_v2_processes:uses process/list_v2|paginated V2 compatible process list|supports plugin_name/status/biz/host filters|read-only|do not frame it as a migration-specific inventory step
|list_v3_processes:uses process/list|paginated V3 process list|supports plugin_name/status/biz/host filters|read-only
|sync_v2_alive_process_info:uses sync/gse/plugin/process_v2/info/alive|host-scoped V2 process sync|syncs running and unknown+auto_start records|returns trigger_id
|sync_v2_stopped_process_info:uses sync/gse/plugin/process_v2/info/stopped|host-scoped V2 process sync|syncs stopped records|returns trigger_id
|install_v3_no_start:uses plugin/install_no_start|delivers V3 package and config without starting|stops a running V3 process first|task ends with V3 process stopped
|stop_v2:uses plugin/stop_v2|async V2 stop workflow|does not uninstall V2 files
|start_plugin:uses plugin/start|starts the installed V3 process|waits until the process is running and auto-start is true
|Conventions:interface pages must not link sibling pages or describe migration ordering; contract references are limited to external Swagger/proto/apigw docs
|Anti-patterns:no full Swagger duplication|no internal call chains|no migration-flow references inside interface pages|no sibling-doc links|no migrate_from_v2/install substitution for this flow|no invented batch limits/health/idempotency/retry/timing guarantees
