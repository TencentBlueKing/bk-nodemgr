### Description

- API Version: v3.0.1-alpha.56+.
- Required Permission: None.
- Function: Query operation instance status distribution by trigger IDs.

**Permission Notes**: The current backend handler does not perform a direct permission check.

### URL

POST /api/v3/node/workflow/operation/instance/status_distribution/list

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| trigger_id | string array | No | Workflow trigger ID list. When omitted, backend returns data according to its query condition |

### Request Example

```json
{
  "trigger_id": [
    "trigger-9f2b"
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": {
      "trigger-9f2b": {
        "not_inited_count": 1,
        "state_counts": {
          "running": 2,
          "success": 3
        }
      }
    }
  }
}
```

### Response Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| code | int32 | Status code, `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter | Type | Description |
| --- | --- | --- |
| items | object | Status distribution map. The key is trigger_id and the value is the distribution |

#### data.items.<trigger_id>

| Parameter | Type | Description |
| --- | --- | --- |
| not_inited_count | int64 | Number of operation instances that are not initialized |
| state_counts | object | Operation instance state counts. The key is state and the value is count |
