### Description

- API Version: v3.0.1-alpha.1+.
- Version Changes: `v3.0.1-alpha.13+` added the permission response field; `v3.0.1-alpha.18+` added `plugin_operate` authorization.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Retry operations in a plugin workflow, either all eligible operations or selected operation IDs.

### URL

POST /api/v3/plugin/workflow/operation/retry

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| workflow_id | string | Yes | Plugin workflow ID; must not be empty |
| retry_mod | string | Yes | Retry mode. Available values: `ALL` (retry all eligible operations), `PARTIAL` (retry selected operations) |
| operation_ids | string array | Yes | Operation ID list; must contain at least one non-empty ID |

### Request Example

```json
{
  "workflow_id": "wf-plugin-0001",
  "retry_mod": "PARTIAL",
  "operation_ids": ["op-0001", "op-0003"]
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

- `retry_mod` is case-sensitive. The execution layer currently accepts only `ALL` and `PARTIAL`.
- Every element in `operation_ids` must be non-empty.
