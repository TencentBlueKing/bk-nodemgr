### Description

- API Version: v3.0.1+.
- Required Permission: networkarea_view (View network area details).
- Function: Query a single network area by network area ID.

### URL

POST /api/v3/topo/networkarea/get

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkarea_id | int64 | Yes | Network area ID |

### Request Example

```json
{
  "bk_networkarea_id": 1
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "tenant_id": "default",
    "bk_networkarea_id": 1,
    "bk_networkarea_name": "prod-default",
    "cloud_vendor": "tencent"
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, `0` means success |
| message | string | Response message |
| request_id | string | Request ID |
| data | object | Network area details |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| tenant_id | string | Tenant ID |
| bk_networkarea_id | int64 | Network area ID |
| bk_networkarea_name | string | Network area name |
| cloud_vendor | string | Cloud vendor identifier |

### Notes

- The contract of this API is defined by `proto/application/api/v3/topo.proto`, `pkg/proto/application/api/v3/networkarea.go`, and `docs/api/swagger/application/api/v3/topo.swagger.json`.
- There is a current implementation gap in `internal/application/router/api-v3/topo/networkarea.go`: the application handler currently returns a create-style payload containing only `bk_networkarea_id`. This document follows the proto/swagger contract as the intended standard response. Please verify runtime behavior before depending on the full object fields.
