### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Query the brief network unit list with pagination and exact filtering by network unit ID, network area ID, direct mode, and generation.

### URL

POST /api/v3/topo/networkunit/list/brief

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| page | object | No | Pagination configuration. In the application layer, when `limit` is `0`, the server treats the request as unpaginated |
| only_count | bool | No | Whether to return only the total count without details |
| exact_include_conditions | object | No | Exact match include conditions |

#### page

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| count | bool | Yes | Whether to return the total record count |
| start | uint32 | No | Record start position, starting from 0 |
| limit | uint32 | No | Records per page. In the application layer, `0` means no pagination |
| sort | string | No | Sort field |
| order | string | No | Sort order (ASC, DESC) |

#### exact_include_conditions

Exact match include conditions.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkunit_id | int64 array | No | Network unit ID list |
| bk_networkarea_id | int64 array | No | Network area ID list |
| is_direct | bool array | No | Whether the network unit is direct |
| generation | int64 array | No | Node generation list |

### Request Example

Query the brief list of direct network units under the specified network area and return the total count.

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "bk_networkunit_id",
    "order": "ASC"
  },
  "exact_include_conditions": {
    "bk_networkarea_id": [
      1
    ],
    "is_direct": [
      true
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
        "bk_networkunit_id": 1001,
        "bk_networkunit_name": "direct-unit-shenzhen",
        "bk_networkarea_id": 1
      },
      {
        "tenant_id": "default",
        "bk_networkunit_id": 1002,
        "bk_networkunit_name": "direct-unit-shanghai",
        "bk_networkarea_id": 1
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
| total | int64 | Total record count matching the criteria |
| items | array | Returned records. This field is typically empty when `only_count` is `true` |

#### data.items[n]

| Parameter | Type | Description |
|---------|----------|------|
| tenant_id | string | Tenant ID |
| bk_networkunit_id | int64 | Network unit ID |
| bk_networkunit_name | string | Network unit name |
| bk_networkarea_id | int64 | Owning network area ID |

### Notes

- The request and response contract is defined by `proto/application/api/v3/topo.proto` and `docs/api/swagger/application/api/v3/topo.swagger.json`.
- The application router implementation is in `internal/application/router/api-v3/topo/networkunit.go`, and it proxies the request to the backend brief list API with the same capability.
- Although the API reuses the `NetworkUnitBrief` protocol type, the current implementation returns only 4 fields: `tenant_id`, `bk_networkunit_id`, `bk_networkunit_name`, and `bk_networkarea_id`. It does not return access points, links, direct endpoints, or custom deploy config in this brief response.
- The current request supports only `exact_include_conditions`. Fuzzy filters and exclude conditions are not supported.
