### Description

- API Version: v3.0.1-alpha.4+.
- Version Changes: `v3.0.1-alpha.13+` added the permission response field.
- Required Permission: None.
- Function: Get distinct operation state candidates under the specified plugin workflow.

### URL

POST /api/v3/plugin/workflow/operation/distinct

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| workflow_id | string | Yes | Plugin workflow ID; must not be empty |
| selector | object | No | Distinct field selector |

#### selector

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| state | bool | No | Whether to return distinct operation states in `data.state` |

### Request Example

```json
{
  "workflow_id": "wf-plugin-0001",
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

| Parameter | Type | Description |
| --- | --- | --- |
| code | int32 | Status code; `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter | Type | Description |
| --- | --- | --- |
| state | string array | Distinct operation states. Available values: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |

### Notes

- If `selector` is omitted or `selector.state=false`, the state column is not selected and `data.state` is an empty array.
- The current handler does not perform an IAM permission check.
