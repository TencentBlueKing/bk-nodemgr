### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: config_policy_manage (Manage Config Policy).
- Function: Enable config policies in batch.

### URL

POST /api/v3/policy/config/enable

### Request Parameters

| Parameter       | Type        | Required | Description                     |
|-----------------|-------------|----------|---------------------------------|
| configpolicy_id | int64 array | Yes      | Config policy ID list to enable |

### Request Example

```json
{
  "configpolicy_id": [
    10001,
    10002
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123460",
  "data": {}
}
```

### Response Parameters

| Parameter  | Type   | Description                                 |
|------------|--------|---------------------------------------------|
| code       | int32  | Status code, `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data, empty object on success      |
