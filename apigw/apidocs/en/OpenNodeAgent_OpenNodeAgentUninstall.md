### Description

- API Version: v3.0.1-alpha.105+
- Required Permissions: system_access (System Access), agent_operate (Operate Agent).
- Function: Uninstall Agents on existing hosts in batches as the authenticated user, optionally unbinding Agent IDs.
- The API Gateway authenticates and forwards the user identity. Do not include a username or tenant parameter in the request body. `agent_operate` is checked against the target hosts' businesses.

### URL

POST /api/v3/open/node/agent/uninstall

### Input Parameters

| Parameter Name  | Parameter Type | Required | Description                                         |
| --------------- | -------------- | -------- | --------------------------------------------------- |
| host            | object array   | Yes      | Nonempty uninstall host list; use distinct host IDs |
| unbind_agent_id | bool           | No       | Whether to unbind Agent IDs, default `false`        |

#### host[n]

| Parameter Name | Parameter Type | Required | Description                                    |
| -------------- | -------------- | -------- | ---------------------------------------------- |
| bk_host_id     | int64          | Yes      | Existing host ID, greater than or equal to `0` |

Target hosts must have an assigned network unit ID greater than or equal to `0`. Uninstallation affects host management capabilities that depend on the Agent. A returned workflow ID does not mean uninstallation has completed.

### Request Example

```json
{
  "host": [{ "bk_host_id": 1001 }],
  "unbind_agent_id": false
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": { "workflow_id": "workflow-abc123" }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                                                                         |
| -------------- | -------------- | ----------------------------------------------------------------------------------- |
| code           | int32          | Status code; `0` means the request succeeded, not that uninstallation has completed |
| message        | string         | Request message                                                                     |
| request_id     | string         | Request ID                                                                          |
| error          | object         | Error information                                                                   |
| permission     | object         | Permission information                                                              |
| data           | object         | Response data                                                                       |

#### data

| Parameter Name | Parameter Type | Description                                                              |
| -------------- | -------------- | ------------------------------------------------------------------------ |
| workflow_id    | string         | Workflow ID for tracking asynchronous uninstallation, not a `trigger_id` |

From v3.0.1-alpha.114, use `OpenNodeWorkflow_OpenNodeWorkflowList` with `exact_include_conditions.workflow_id` to query execution status. The query additionally requires the corresponding history-view permission.
