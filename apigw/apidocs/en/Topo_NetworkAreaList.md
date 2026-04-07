### Description

- API Version: v3.0.1+.
- Required Permission: .
- Function: Query the network area list with pagination and filtering by network area ID, name, and cloud vendor.

### URL

POST /api/v3/topo/networkarea/list

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| page | object | No | Pagination configuration. When `limit` is `0`, the server treats the request as unpaginated |
| only_count | bool | No | Whether to return only the total count without details |
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions |

#### page

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| count | bool | Yes | Whether to return the total record count |
| start | uint32 | No | Record start position, starting from 0 |
| limit | uint32 | No | Records per page; `0` means no pagination |
| sort | string | No | Sort field |
| order | string | No | Sort order (ASC, DESC) |

#### exact_include_conditions

Exact match include conditions. Matches if any condition is met.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkarea_id | int64 array | No | Network area ID list |
| cloud_vendor | string array | No | Cloud vendor identifier list |

#### fuzzy_include_conditions

Fuzzy match include conditions. Matches if any condition is met.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkarea_name | string array | No | Network area name list (fuzzy match) |

### Request Example

Query network areas whose names contain `prod` and return the total count.

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "bk_networkarea_id",
    "order": "ASC"
  },
  "fuzzy_include_conditions": {
    "bk_networkarea_name": [
      "prod"
    ]
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 2,
    "items": [
      {
        "tenant_id": "default",
        "bk_networkarea_id": 1,
        "bk_networkarea_name": "prod-default",
        "cloud_vendor": "tencent"
      },
      {
        "tenant_id": "default",
        "bk_networkarea_id": 2,
        "bk_networkarea_name": "prod-overseas",
        "cloud_vendor": "aws"
      }
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
| error | object | Error information (empty on success) |
| permission | object | Permission information (typically empty for this API) |
| data | object | Response data |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| total | int64 | Total record count matching the criteria |
| items | array | Returned records |

#### data.items[n]

| Parameter | Type | Description |
|---------|----------|------|
| tenant_id | string | Tenant ID |
| bk_networkarea_id | int64 | Network area ID |
| bk_networkarea_name | string | Network area name |
| cloud_vendor | string | Cloud vendor identifier |

### Notes

- The request and response contract is defined by `proto/application/api/v3/topo.proto` and `docs/api/swagger/application/api/v3/topo.swagger.json`.
- The current permission semantics have been aligned so that this list API no longer requires pre-authorization. Use the detail or mutation APIs for permission-guarded follow-up actions.
