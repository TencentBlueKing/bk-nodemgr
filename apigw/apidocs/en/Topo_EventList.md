### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Query topology event records with pagination, exact filtering by object IDs or operator, fuzzy filtering by object names, and filtering by operation time range.

### URL

POST /api/v3/topo/event/list

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| page | object | No | Pagination configuration, with a maximum page size of `1000` |
| only_count | bool | No | Whether to return only the total count without item details |
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions |
| operate_time_range | object | No | Operation time range. If omitted, the server defaults to the most recent `365` days |

#### page

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| count | bool | Yes | Whether to return the total record count |
| start | uint32 | No | Record start position, starting from `0` |
| limit | uint32 | No | Records per page, up to `1000` |
| sort | string | No | Sort field |
| order | string | No | Sort order (`ASC`, `DESC`) |

#### exact_include_conditions

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkarea_id | int64 array | No | Network area ID list |
| bk_networkunit_id | int64 array | No | Network unit ID list |
| accesspoint_id | int64 array | No | Access point ID list |
| type | string array | No | Event type list. Available values: `networkarea-create`, `networkarea-update`, `networkarea-delete`, `networkunit-create`, `networkunit-update`, `networkunit-delete`, `accesspoint-create`, `accesspoint-update`, `accesspoint-delete` |
| operator | string array | No | Operator list |

#### fuzzy_include_conditions

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkarea_name | string array | No | Network area name list for fuzzy matching |
| bk_networkunit_name | string array | No | Network unit name list for fuzzy matching |

#### operate_time_range

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| start_timestamp_sec | int64 | No | Start timestamp in seconds |
| end_timestamp_sec | int64 | No | End timestamp in seconds |

### Request Example

Query topology events within a recent time range for one network area, filtered by operator, and return both the total count and item details.

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 20,
    "sort": "operate_time",
    "order": "DESC"
  },
  "exact_include_conditions": {
    "bk_networkarea_id": [
      1
    ],
    "type": [
      "networkunit-create",
      "accesspoint-update"
    ],
    "operator": [
      "admin"
    ]
  },
  "fuzzy_include_conditions": {
    "bk_networkunit_name": [
      "prod"
    ]
  },
  "operate_time_range": {
    "start_timestamp_sec": 1735689600,
    "end_timestamp_sec": 1735776000
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
        "type": "networkunit-create",
        "bk_networkarea_id": 1,
        "bk_networkarea_name": "prod-default",
        "bk_networkunit_id": 101,
        "bk_networkunit_name": "prod-unit-a",
        "accesspoint_id": 0,
        "accesspoint_name": "",
        "operate_time": 1735722000000,
        "operator": "admin"
      },
      {
        "tenant_id": "default",
        "type": "accesspoint-update",
        "bk_networkarea_id": 1,
        "bk_networkarea_name": "prod-default",
        "bk_networkunit_id": 101,
        "bk_networkunit_name": "prod-unit-a",
        "accesspoint_id": 2001,
        "accesspoint_name": "ap-prod-1",
        "operate_time": 1735725600000,
        "operator": "admin"
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
| error | object | Error information, usually empty on success |
| permission | object | Permission information, typically empty for this API |
| data | object | Response data |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| total | int64 | Total number of records matching the filter criteria |
| items | array | Event detail list. It is typically an empty array when `only_count` is `true` |

#### data.items[n]

| Parameter | Type | Description |
|---------|----------|------|
| tenant_id | string | Tenant ID |
| type | string | Event type |
| bk_networkarea_id | int64 | Network area ID |
| bk_networkarea_name | string | Network area name |
| bk_networkunit_id | int64 | Network unit ID |
| bk_networkunit_name | string | Network unit name |
| accesspoint_id | int64 | Access point ID. It can be `0` for non access-point events |
| accesspoint_name | string | Access point name. It can be an empty string for non access-point events |
| operate_time | int64 | Operation time as a Unix timestamp in milliseconds |
| operator | string | Operator |

### Notes

- The request and response contract of this API is defined by `proto/backend/api/v3/topo.proto`, `pkg/proto/backend/api/v3/event.go`, and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- The router implementation is in `internal/backend/router/api-v3/topo/event.go`. When `operate_time_range` is omitted, the protocol layer fills it with the most recent `365` days by default.
