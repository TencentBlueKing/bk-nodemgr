### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Query the business list with pagination and filtering by business ID and business name.

### URL

POST /api/v3/topo/business/list

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| page | object | No | Pagination configuration. `offset` must be greater than or equal to `0`, and `limit` must be within `(0, 1000]` |
| only_count | bool | No | Whether to return only the total count |
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions |

#### page

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| offset | int32 | No | Record start offset, starting from `0` |
| limit | int32 | No | Maximum number of returned records. Valid range is `(0, 1000]` |

#### exact_include_conditions

Exact match include conditions. A record matches if its business ID matches any provided value.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_biz_id | int64 array | No | Business ID list |

#### fuzzy_include_conditions

Fuzzy match include conditions. A record matches if its business name matches any provided pattern.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_biz_name | string array | No | Business name list used for fuzzy matching |

### Request Example

Query businesses whose names contain `prod`.

```json
{
  "page": {
    "offset": 0,
    "limit": 10
  },
  "fuzzy_include_conditions": {
    "bk_biz_name": [
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
        "bk_biz_id": 2,
        "bk_biz_name": "prod-payment"
      },
      {
        "tenant_id": "default",
        "bk_biz_id": 7,
        "bk_biz_name": "prod-order"
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
| error | object | Error information, empty on success |
| permission | object | Permission information, typically empty for this API |
| data | object | Response data |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| total | int64 | Total record count matching the current criteria |
| items | array | Returned business records |

#### data.items[n]

| Parameter | Type | Description |
|---------|----------|------|
| tenant_id | string | Tenant ID |
| bk_biz_id | int64 | Business ID |
| bk_biz_name | string | Business name |

### Notes

- The request and response contract is defined by `proto/backend/api/v3/topo.proto` and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- The current backend permission semantics have been aligned so that this list API no longer requires pre-authorization. Use the detail or mutation APIs for permission-guarded follow-up actions.
