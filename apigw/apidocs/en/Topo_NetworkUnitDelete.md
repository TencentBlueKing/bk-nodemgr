### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: networkunit_delete (Delete Network Unit).
- Function: Delete a network unit by network unit ID.

### URL

POST /api/v3/topo/networkunit/delete

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkunit_id | int64 | No | Network unit ID |

### Request Example

```json
{
  "bk_networkunit_id": 1
}
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
    "bk_networkunit_id": 1
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
| bk_networkunit_id | int64 | Network unit ID |
