# Execution Query Reference

Use only the selected domain. The MCP resources and CLI paths below refer to the same **user-state** APIs. Do not substitute the older `/api/v3/node/...` or `/api/v3/plugin/...` application-state routes because they have similar names.

## Resource Mapping

All nine resources are read-only POST requests. MCP tool registration can add a server-specific prefix; discover the actual tool name instead of guessing it.

| Domain | Query              | Resource / operationId                                         | Gateway entry path                                        |
| ------ | ------------------ | -------------------------------------------------------------- | --------------------------------------------------------- |
| node   | Workflow           | `OpenNodeWorkflow_OpenNodeWorkflowList`                        | `/api/v3/open/node/workflow/list`                         |
| node   | Operations         | `OpenNodeWorkflow_OpenNodeWorkflowOperationList`               | `/api/v3/open/node/workflow/operation/list`               |
| node   | Install parameters | `OpenNodeWorkflow_OpenNodeWorkflowOperationInstallParamsGet`   | `/api/v3/open/node/workflow/operation/install_params/get` |
| node   | Instances          | `OpenNodeWorkflow_OpenNodeWorkflowOperationInstanceList`       | `/api/v3/open/node/workflow/operation/instance/list`      |
| node   | Instance logs      | `OpenNodeWorkflow_OpenNodeWorkflowOperationInstanceLogGet`     | `/api/v3/open/node/workflow/operation/instance/log/get`   |
| plugin | Workflow           | `OpenPluginWorkflow_OpenPluginWorkflowList`                    | `/api/v3/open/plugin/workflow/list`                       |
| plugin | Operations         | `OpenPluginWorkflow_OpenPluginWorkflowOperationList`           | `/api/v3/open/plugin/workflow/operation/list`             |
| plugin | Instances          | `OpenPluginWorkflow_OpenPluginWorkflowOperationInstanceList`   | `/api/v3/open/plugin/workflow/operation/instance/list`    |
| plugin | Instance logs      | `OpenPluginWorkflow_OpenPluginWorkflowOperationInstanceLogGet` | `/api/v3/open/plugin/workflow/operation/instance/log/get` |

The runtime has a complete repository checkout. Resolve paths relative to that checkout, not this installed Skill's directory:

- `apigw/resources.yaml`: entry paths, methods, operation IDs, schemas, gateway identity requirements.
- `apigw/apidocs/zh/<operationId>.md` or `apigw/apidocs/en/<operationId>.md`: selected API's business meaning, field constraints, permissions, and response details.
- `apigw/mcp_servers.yaml`: built-in MCP resource membership; explicit stage definitions in `apigw/definition.yaml` can override it. Neither proves a deployed connection or publication.
- `pkg/identifier/identifier.go`: identifier prefixes and formatting.

Read only the relevant operations, not the entire catalog. If the actual tool schema, documentation, and observed response conflict, resolve the material difference or report it rather than inventing fields. Gateway app/resource permission and user IAM checks are separate. These user-state queries require `system_access` and the applicable node or plugin history permissions; obtain the exact target-dependent requirements from the selected API documentation.

## Minimal Bodies

The following examples use illustrative IDs, not real targets. Substitute the complete supplied or returned IDs. These bodies apply to both node and plugin unless noted; the transport wrapper may differ.

### Find an exact workflow

```json
{
  "page": { "offset": 0, "limit": 20 },
  "only_count": false,
  "exact_include_conditions": {
    "workflow_id": ["wf:7eb0e3fa0afd4543ac43561eb526e7ea"]
  }
}
```

Use a `workflow_id` **array inside `exact_include_conditions`**. Provide a valid page when fetching records and an explicit condition object. Do not add time or status filters that could hide the supplied workflow.

### List operations of that workflow

```json
{
  "page": { "offset": 0, "limit": 20 },
  "only_count": false,
  "workflow_id": "wf:7eb0e3fa0afd4543ac43561eb526e7ea"
}
```

Here `workflow_id` is a **top-level string**, not an array. A valid page is needed even for count-only requests. Current page limits are 1–500; 20 is a small initial page, not a completeness guarantee. Follow the returned total and payload field names from the current schema; do not assume all APIs return `items`.

The current `latest_oper_inst_brief_data` contains lifecycle and latest-action summaries, not an explicit latest `oper_inst_id`. The action summary's `name` is not an action-instance ID either. The operation-list `instance_ids` is different from the instance-history response: current handlers return the runtime's append-ordered list unchanged, and the runtime uses its final element as the latest instance ID (`GetLastInstanceID` in `pkg/workflow/operation/operation.go`). When this version/response contract applies, use that association and reconcile it with instance history; an empty list means no associated instance is available. Do not transfer that ordering rule to `oper_inst_data`, sort IDs, or invent an ID from a summary.

### List instances of one operation

```json
{
  "only_count": false,
  "operation_id": ["oper:00000000000000000000000000000001"]
}
```

Use one operation in the array. Do not batch operations from different workflows. Read the returned `oper_inst_data` and `total`. No page or sorting parameters are supported. Node optionally accepts `oper_inst_id` as an additional array filter within this operation; plugin does not.

### Read a selected instance's steps and logs

```json
{
  "oper_inst_id": "oper-inst:00000000000000000000000000000002"
}
```

Here `oper_inst_id` is a **string**. `oper_inst_logs` is keyed by action name, not action-instance ID; map order is not execution order. Use the instance's `action_names` and lifecycle timestamps. Inspect `message.logs`, `sub_workflow_refs`, and relevant `extra_execution_logs` according to the response schema. Log entries use `text_zh` and `text_en`, not the former `text` field; this also applies to `extra_execution_logs.logs`. Prefer the user's language and fall back to the other available text. An empty language field or missing `.text` does not mean there are no logs. Do not add undocumented action, paging, or time-window parameters to this endpoint.

### Read historical installation parameters

For `node` workflows of type `install_agent` or `install_proxy` only:

```json
{
  "workflow_id": "wf:7eb0e3fa0afd4543ac43561eb526e7ea",
  "operation_ids": ["oper:00000000000000000000000000000001"]
}
```

Both fields are required. Here `operation_ids` is plural, unlike the instance-list filter. Select nonempty, unique operation IDs belonging to the same workflow; use the operation list if selection is needed. Do not invent a workflow ID when only an operation ID is known. No pagination or instance ID is accepted.

Inspect `data.items` by `operation_id`, including `host`, `target_version`, `is_manual`, and `is_offline`. `missing_fields` identifies inputs that could not be recovered, and `unavailable_operation_ids` identifies operations without usable deployment records. Do not interpret unavailable or missing values as previous choices. Historical records are not a complete request snapshot, a current configuration, or evidence of a particular attempt's actual runtime settings. Credentials and deployment tokens are excluded; do not seek them in other sources. This query never requires a failed operation and does not install or retry anything.

## One CLI Example

After confirming context, stage, and tenant, preview the exact plugin workflow query:

```bash
bk-cli api bk-nodemgr POST /api/v3/open/plugin/workflow/list \
  --context "${CONTEXT:?confirmed context}" \
  --stage "${STAGE:?confirmed stage}" \
  --header "X-Bk-Tenant-Id:${TENANT_ID:?confirmed tenant}" \
  --body '{"page":{"offset":0,"limit":20},"only_count":false,"exact_include_conditions":{"workflow_id":["wf:7eb0e3fa0afd4543ac43561eb526e7ea"]}}' \
  --dry-run
```

This is a preview with an illustrative ID. For the real scoped query, supply the actual ID and remove `--dry-run` after inspecting construction. If the confirmed tenant is already supplied by the CLI context, omit the override instead of inventing another tenant. Authentication is supplied by the configured runtime, not a body username. Do not use documentation's `operator` filter to impersonate a user.

A CLI result can contain both the CLI envelope and the Node Manager envelope. For ordinary JSON queries:

| Check                                          | Meaning                                      |
| ---------------------------------------------- | -------------------------------------------- |
| Exit status, outer `ok` and `status`           | CLI/HTTP outcome                             |
| `.data.code`                                   | Node Manager business result, success is `0` |
| `.data.data`                                   | Business payload                             |
| `.data.request_id` and error/permission fields | Correlation and failure evidence             |

The MCP may use another wrapper. Inspect its actual schema/response rather than applying these CLI JSON paths blindly.
