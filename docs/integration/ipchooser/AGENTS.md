|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/comments/docs/tech discussions
|Compression Rule:pipe-index format|concise|no prose sections|no code blocks
|Scope:docs/topo
|Overview:Four public topo API endpoints composed as a backend invocation chain|read-only query composition|third-party integration oriented
|Structure:docs/integration/ipchooser:{AGENTS.md,README.md,list_business.md,get_business_host_count.md,get_business_inst_topo.md,list_host.md}
|Source of truth:4 endpoint contracts=docs/api/swagger/backend/api/v3/topo.swagger.json|4 endpoint request/response variants=proto/backend/api/v3/topo.proto|shared Business/Host/HostInfo/HostState/Page=proto/backend/api/v3/common.proto
|Conventions:README.md is the flow entrypoint|one mode page per endpoint
|Conventions:mode page sections={Purpose and applicability,Input,Minimal payload and curl template,System interpretation,Immediate output,Eventual or machine-visible artifact,Repeat behavior,Failure cases and limits,Contract references}
|Conventions:each executable mode page must include a curl template|initialize shell variables|extract returned IDs/host counts when reused|do not invent auth/tenant headers|do not include secrets/real identifiers
|Conventions:describe only request/response shape and field semantics|mark public-contract gaps as 待确认 or not specified
|four endpoints:ListBusiness (POST /api/v3/topo/business/list) | GetBusinessHostCount (POST /api/v3/topo/business/host_count/get) | GetBusinessInstTopo (POST /api/v3/topo/business/inst_topo/get) | ListHost (POST /api/v3/topo/host/list)
|composition:four endpoints in sequence (1) business list (2) per-business host count (3) per-business instance topology with aggregated host counts (4) host list under any topology node|read-only query composition
|ListBusiness:returns Business items (tenant_id, bk_biz_id, bk_biz_name)|pageable|supports exact bk_biz_id filter and fuzzy bk_biz_name filter
|GetBusinessHostCount:returns BusinessHostCount items (bk_biz_id, host_count)|takes repeated bk_biz_id
|GetBusinessInstTopo:returns single TopoNodeInfo root|each node carries topo_obj_id, topo_inst_id, topo_inst_name, host_count, children|host_count is pre-aggregated per node by server
|ListHost:returns Host items (tenant_id, bk_host_id, info, state, create_at, updated_at)|pageable|serves host list under a topology node via exact_include_conditions + fuzzy_include_conditions
|Anti-patterns:no full Swagger duplication|no internal call chains (no router/storage/manager reference)|no assumed client UI (no sidebar/tree/table/dialog/widget language)|no invented machine paths/health/idempotency/retry/timing guarantees|no state-machine enums beyond the response shape|no fake "ip-selector" aggregate endpoint
|Anti-patterns:do not document topo/host/distinct or host/scenario/* endpoints here (out of the 4-endpoint composition scope); do not document networkarea/networkunit endpoints here
|Anti-patterns:do not reference internal narrowAuthorizedBizIDsForHostList or storage handler names; these are not part of the public API contract
|Anti-patterns:do not assume any ordering of inputs other than the documented server-side composition; do not promise an aggregate single-call substitute for the four-call composition
