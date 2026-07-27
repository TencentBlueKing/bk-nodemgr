### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: None.
- Function: Trigger synchronization of GSE Agent information for alive hosts.

### URL

POST /api/v3/sync/gse/agent/info/alive

### Input Parameters

None

### Request Example

```json
{}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "trigger_id": "trigger-001"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| trigger_id | string | Trigger ID |
