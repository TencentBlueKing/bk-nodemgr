### Description

- API Version: v3.0.1+.
- Required Permission: .
- Function: Set the memo for a specified plugin.

### URL

POST /api/v3/plugin/set_memo

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| plugin_name | string | Yes | Plugin name |
| memo | string | No | Plugin memo content; may be an empty string |

**Parameter Notes**:
- `plugin_name` must not be empty, otherwise request validation fails.
- `memo` currently has no length or format restriction and is written directly to the plugin memo field.

### Request Example

```json
{
  "plugin_name": "bk-monitor-agent",
  "memo": "Used for host monitoring collection"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000004",
  "error": null
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, null on success |

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
