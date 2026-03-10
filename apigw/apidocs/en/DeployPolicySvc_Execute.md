### Description

- API Version: v3.0.1+
- Required Permission:
- Function: Execute a deploy policy.

### URL

POST /api/v3/deploy_policy/execute

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| deploy_policy_id | int64 | Yes | Deploy policy ID |

### Request Example

```json
{
  "deploy_policy_id": 1001
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "trigger_id": "trigger-abc123def456"
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, 0 for success |
| message | string | Response message |
| request_id | string | Request ID |
| error | object | Error information (empty on success) |
| data | object | Response data |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| trigger_id | string | Trigger ID, can be used to query execution status |
