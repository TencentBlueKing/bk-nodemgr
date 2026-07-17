### Description

- API Version: v3.0.1-alpha.1+.
- Version Changes: `v3.0.1-alpha.13+` added the permission response field.
- Required Permission: None.
- Function: Query the status distribution of the latest plugin workflow operation instances by trigger ID.

### URL

POST /api/v3/plugin/workflow/operation/instance/status_distribution/list

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| trigger_id | string array | Yes | Trigger ID list; must contain at least one ID |

### Request Example

```json
{
  "trigger_id": ["trigger-plugin-0001", "trigger-plugin-0002"]
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
      "trigger-plugin-0001": {
        "not_inited_count": 2,
        "state_counts": {
          "running": 3,
          "success": 8,
          "failed": 1
        }
      }
    }
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
| items | object | Status distribution map keyed by trigger ID |

#### data.items.{trigger_id}

`trigger_id` is a trigger ID from the request. Its value is the latest operation instance status distribution associated with that trigger.

| Parameter | Type | Description |
| --- | --- | --- |
| not_inited_count | int64 | Number of operations for which no operation instance has been created |
| state_counts | object | Operation instance count map keyed by state |

#### data.items.{trigger_id}.state_counts

Keys are operation instance states. Available values: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated`. Each value is the number of instances in that state.

### Notes

- The request Proto does not directly reject an empty list, but the storage layer does. At least one `trigger_id` must therefore be provided.
- The current handler does not perform an IAM permission check.
