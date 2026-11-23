### Description

- API Version: v3.0.1-alpha.105+
- Required Permissions: system_access (System Access), agent_operate (Operate Agent).
- Function: Redeploy Agent configuration and restart Agents in batches as the authenticated user. This API does not accept configuration file content.
- The API Gateway authenticates and forwards the user identity. Do not include a username or tenant parameter in the request body. `agent_operate` is checked against the target hosts' businesses.

### URL

POST /api/v3/open/node/agent/reconfig

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                               |
| -------------- | -------------- | -------- | --------------------------------------------------------- |
| host           | object array   | Yes      | Nonempty reconfiguration host list; use distinct host IDs |

#### host[n]

| Parameter Name               | Parameter Type | Required | Description                                                                              |
| ---------------------------- | -------------- | -------- | ---------------------------------------------------------------------------------------- |
| bk_host_id                   | int64          | Yes      | Existing host ID, greater than or equal to `0`                                           |
| force                        | bool           | No       | Whether to force restart, default `false`; forwarded to the GSE restart operation        |
| graceful_restart_timeout_sec | int64          | No       | Graceful restart timeout in seconds, default `0`; forwarded to the GSE restart operation |

Target hosts must have an assigned network unit ID greater than or equal to `0`. Reconfiguration includes a restart, during which Agent service may be temporarily unavailable.

### Request Example

```json
{
  "host": [
    { "bk_host_id": 1001, "force": false, "graceful_restart_timeout_sec": 60 }
  ]
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

| Parameter Name | Parameter Type | Description                                                                          |
| -------------- | -------------- | ------------------------------------------------------------------------------------ |
| code           | int32          | Status code; `0` means the request succeeded, not that reconfiguration has completed |
| message        | string         | Request message                                                                      |
| request_id     | string         | Request ID                                                                           |
| error          | object         | Error information                                                                    |
| permission     | object         | Permission information                                                               |
| data           | object         | Response data                                                                        |

#### data

| Parameter Name | Parameter Type | Description                                                                                    |
| -------------- | -------------- | ---------------------------------------------------------------------------------------------- |
| workflow_id    | string         | Workflow ID for tracking asynchronous configuration deployment and restart, not a `trigger_id` |

From v3.0.1-alpha.114, use `OpenNodeWorkflow_OpenNodeWorkflowList` with `exact_include_conditions.workflow_id` to query execution status. The query additionally requires the corresponding history-view permission.
