### Description

- API Version: v3.0.0+.
- Version Changes: `v3.0.1-alpha.96+` added the `unbind_agent_id` parameter.
- Required Permission: agent_operate (Operate Agent).
- Function: Batch uninstall node agents.

### URL

POST /api/v3/node/agent/uninstall

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| host | object array | Yes | List of hosts for agent uninstallation |
| unbind_agent_id | bool | No | Whether to unbind the host AgentID relation after successful uninstallation. Default false |

#### host[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 | Yes | Host ID |

### Request Example

Batch uninstall agents on two hosts and unbind the host AgentID relation after successful uninstallation.

```json
{
  "host": [
    {
      "bk_host_id": 1001
    },
    {
      "bk_host_id": 1002
    }
  ],
  "unbind_agent_id": true
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "workflow_id": "workflow-abc123def456"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, null on success |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| workflow_id | string | Workflow ID, can be used to query uninstallation task status |

#### error

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| system | string | Error system identifier |
| message | string | Error message |
| details | array | Error details list |

#### error.details[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | string | Error code |
| message | string | Error message |
