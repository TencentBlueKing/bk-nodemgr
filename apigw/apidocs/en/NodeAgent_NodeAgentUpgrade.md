### Description

- API Version: v3.0.0+.
- Required Permission: agent_operate (Operate Agent).
- Function: Batch upgrade node agents to a specified version, with support for forced restart and graceful restart timeout configuration.

### URL

POST /api/v3/node/agent/upgrade

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| host | object array | Yes | List of hosts for agent upgrade |

#### host[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 | Yes | Host ID |
| target_version | string | Yes | Target agent version number |
| force | bool | No | Whether to force restart, default is false. When set to true, skips graceful restart and forces an immediate restart |
| graceful_restart_timeout_sec | int64 | No | Graceful restart timeout in seconds. The agent will be forcefully restarted after this timeout. Default 0 uses the system default timeout |

### Request Example

Batch upgrade agents on two hosts to version 2.1.0, with one using forced restart.

```json
{
  "host": [
    {
      "bk_host_id": 1001,
      "target_version": "2.1.0",
      "force": false,
      "graceful_restart_timeout_sec": 60
    },
    {
      "bk_host_id": 1002,
      "target_version": "2.1.0",
      "force": true
    }
  ]
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
| workflow_id | string | Workflow ID, can be used to query upgrade task status |

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
