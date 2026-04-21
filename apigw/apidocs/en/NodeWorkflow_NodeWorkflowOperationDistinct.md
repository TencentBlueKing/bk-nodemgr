### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: agent_operate (Operate Agent), proxy_operate (Operate Proxy).
  - Permission Note: Dynamically narrowed based on node roles in the workflow. If the workflow only involves Agent nodes, only agent_operate is required; if only Proxy nodes, only proxy_operate is required; if both or unknown roles, both are required.
- Function: Query distinct values of latest operation instance state for a node workflow by workflow ID.

### URL

POST /api/v3/node/workflow/operation/distinct

### Request Parameters

| Parameter   | Type   | Required | Description                     |
| ----------- | ------ | -------- | ------------------------------- |
| workflow_id | string | Yes      | Workflow ID. Must not be empty. |
| selector    | object | No       | Distinct field selector.        |

#### selector

| Parameter | Type | Required | Description                                                        |
| --------- | ---- | -------- | ------------------------------------------------------------------ |
| state     | bool | No       | Whether to return distinct result for `state`. Default is `false`. |

### Request Example

Query distinct latest operation instance states for workflow `wf-20240601-0001`.

```json
{
  "workflow_id": "wf-20240601-0001",
  "selector": {
    "state": true
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "state": ["running", "success", "failed"]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                           |
| ---------- | ------ | ----------------------------------------------------- |
| code       | int32  | Status code, `0` means success.                       |
| message    | string | Response message.                                     |
| request_id | string | Request ID.                                           |
| error      | object | Error information, usually empty on success.          |
| permission | object | Permission information, typically empty for this API. |
| data       | object | Response data.                                        |

#### data

| Parameter | Type         | Description                                                                                                                           |
| --------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| state     | string array | Distinct latest operation instance states. Valid values: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated`. |

### Notes

- `workflow_id` is first resolved to its `trigger_id`, then the distinct query is executed within that workflow scope.
- When `selector.state` is `false`, or `selector` is omitted, `data.state` is usually an empty array.
