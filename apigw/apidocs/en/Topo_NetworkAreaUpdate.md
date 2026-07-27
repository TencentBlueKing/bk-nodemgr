### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: networkarea_edit (Edit Network Area).
- Function: Update a network area name and cloud vendor by network area ID.

### URL

POST /api/v3/topo/networkarea/update

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkarea_id | int64 | No | Network area ID |
| bk_networkarea_name | string | Yes | Network area name |
| cloud_vendor | string | Yes | Cloud vendor identifier |

### Request Example

```json
{
  "bk_networkarea_id": 1,
  "bk_networkarea_name": "default",
  "cloud_vendor": "string"
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
    "bk_networkarea_id": 1
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
| bk_networkarea_id | int64 | Network area ID |
