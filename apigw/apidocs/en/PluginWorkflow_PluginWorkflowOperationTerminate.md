### Description

- API Version: v3.0.1-alpha.1+.
- Version Changes: `v3.0.1-alpha.13+` added the permission response field; `v3.0.1-alpha.18+` added `plugin_operate` authorization.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Terminate the latest execution instance of one or more operations in a plugin workflow.

### URL

POST /api/v3/plugin/workflow/operation/terminate

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| workflow_id | string | Yes | Plugin workflow ID; must not be empty |
| operation_ids | string array | Yes | Operation IDs to terminate; must contain at least one element |

### Request Example

```json
{
  "workflow_id": "wf-plugin-0001",
  "operation_ids": ["op-0001", "op-0002"]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": null
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
| data | null | No business response fields on success; currently serialized as `null` |

### Notes

- The backend terminates the latest operation instance for each specified operation.
- Request validation requires a non-empty `operation_ids` array; it currently does not reject empty strings element by element.
