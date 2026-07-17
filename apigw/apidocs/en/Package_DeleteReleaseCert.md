### Description

- API Version: v3.0.0+.
- Required Permission: package_manage (Manage Package).
- Function: Delete the certificate package for the specified generation.

### URL

POST /api/v3/package/release/cert/delete

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 | Yes | Package generation (enum value: 2) |

### Request Example

```json
{"generation": 2}
```

### Response Example

```json
{"code": 0, "message": "ok", "request_id": "req-1234567890", "data": {}}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, where 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| data | object | Empty object indicating a successful operation |
