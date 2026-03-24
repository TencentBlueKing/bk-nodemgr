### Description

- API Version: v3.0.1+.
- Required Permission: .
- Function: Batch uninstall plugins from specified hosts.

### URL

POST /api/v3/plugin/uninstall

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| plugin | object array | Yes | Plugin uninstall targets; at least one item is required |

#### plugin[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 | No | Target host ID; defaults to -1 when omitted, but must be a positive integer in practice |
| plugin_name | string | Yes | Plugin name |

**Parameter Notes**:
- `plugin`: At least one item must be provided, otherwise request validation fails.
- `bk_host_id`: Although optional in proto, omitting it causes the server to auto-fill `-1`, which then fails validation. In practice, provide a valid host ID.
- `plugin_name`: Must not be empty.

### Request Example

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent"
    },
    {
      "bk_host_id": 1002,
      "plugin_name": "bk-log-collector"
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000005",
  "error": null,
  "data": {
    "workflow_id": "plugin-uninstall-workflow-123456"
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
| workflow_id | string | Workflow ID for tracking the plugin uninstall task |

#### error

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| system | string | Error system identifier |
| message | string | Error message |
| details | array | Error detail list |

#### error.details[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | string | Error code |
| message | string | Error message |
