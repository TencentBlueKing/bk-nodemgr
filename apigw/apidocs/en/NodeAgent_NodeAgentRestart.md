### Description

- API Version: v3.0.0+.
- Required Permission: agent_operate (Operate Agent).
- Function: Batch restart node agents with support for forced restart and graceful restart timeout configuration.

### URL

POST /api/v3/node/agent/restart

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| host | object array | Yes | List of hosts for agent restart |

#### host[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 | Yes | Host ID |
| force | bool | No | Whether to force restart, default is false. When set to true, skips graceful restart and forces an immediate restart |
| graceful_restart_timeout_sec | int64 | No | Graceful restart timeout in seconds. The agent will be forcefully restarted after this timeout. Default 0 uses the system default timeout |

### Request Example

Batch restart agents on two hosts, with one configured with a 60-second graceful restart timeout.

```json
{
  "host": [
    {
      "bk_host_id": 1001,
      "force": false,
      "graceful_restart_timeout_sec": 60
    },
    {
      "bk_host_id": 1002,
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
| workflow_id | string | Workflow ID, can be used to query restart task status |

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
