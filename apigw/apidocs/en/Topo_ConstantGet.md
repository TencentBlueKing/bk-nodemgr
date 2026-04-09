### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Retrieve topology-related constant values on demand. The API currently supports the cloud vendor list and the OS type list.

### URL

POST /api/v3/topo/constant/get

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| cloud_vendor | bool | No | Whether to return the cloud vendor list. When `true`, the API returns `data.cloud_vendor`. When `false` or omitted, it returns an empty array |
| os_type | bool | No | Whether to return the OS type list. When `true`, the API returns `data.os_type`. When `false` or omitted, it returns an empty array |

### Request Example

```json
{
  "cloud_vendor": true,
  "os_type": true
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "cloud_vendor": [
      "tencent",
      "aws"
    ],
    "os_type": [
      "LINUX",
      "WINDOWS"
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, `0` means success |
| message | string | Response message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information, typically empty for this API |
| data | object | Constant data |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| cloud_vendor | string array | Cloud vendor list |
| os_type | string array | OS type list |

### Notes

- The request and response contract is defined by `proto/backend/api/v3/topo.proto` and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- The runtime behavior is implemented in `internal/backend/router/api-v3/topo/constant.go`: the API only returns the constant categories explicitly enabled in the request, and returns empty arrays for categories that are not enabled.
