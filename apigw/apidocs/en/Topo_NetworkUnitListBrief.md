### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: networkunit_view (View Network Unit).
- Function: Query the network unit brief list with pagination and exact filtering by network unit ID, network area ID, direct-connect flag, and generation. The `brief` response does not expose sensitive fields such as `direct_endpoints` and `custom_deploy_config`.

### URL

POST /api/v3/topo/networkunit/list/brief

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| page | object | No | Pagination configuration. When `limit` is `0`, the server treats the request as unpaginated |
| only_count | bool | No | Whether to return only the total count without details |
| exact_include_conditions | object | No | Exact match include conditions |

#### page

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| count | bool | Yes | Whether to return the total record count |
| start | uint32 | No | Record start position, starting from 0 |
| limit | uint32 | No | Records per page; `0` means no pagination |
| sort | string | No | Sort field |
| order | string | No | Sort order (ASC, DESC) |

#### exact_include_conditions

Exact match include conditions.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkunit_id | int64 array | No | Network unit ID list |
| bk_networkarea_id | int64 array | No | Network area ID list |
| is_direct | bool array | No | Whether the network unit is direct-connected |
| generation | int64 array | No | Network unit generation |

### Request Example

Query the brief list of direct-connected network units under the specified network area and return the total count.

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
        "bk_networkarea_id": 1,
        "accesspoints": [
          2001,
          2002
        ],
        "links": {
          "cluster": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1001,
            "accesspoint_id": 2001
          },
          "file": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1001,
            "accesspoint_id": 2002
          },
          "data": {
            "bk_networkarea_id": -1,
            "bk_networkunit_id": -1,
            "accesspoint_id": -1
          }
        },
        "is_direct": true,
        "generation": 2
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
| permission | object | Permission information. The API may return permission-related details when the requested scope exceeds the network units visible to the current user |
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
| accesspoints | int64 array | Associated access point ID list |
| links | object | Upstream link relationships of the current network unit |
| is_direct | bool | Whether the network unit is direct-connected |
| generation | int64 | Network unit generation |

#### data.items[n].links

| Parameter | Type | Description |
|---------|----------|------|
| cluster | object | Cluster link |
| file | object | File link |
| data | object | Data link |

#### data.items[n].links.{cluster\|file\|data}

| Parameter | Type | Description |
|---------|----------|------|
| bk_networkarea_id | int64 | Target network area ID. Usually `-1` when unset |
| bk_networkunit_id | int64 | Target network unit ID. Usually `-1` when unset |
| accesspoint_id | int64 | Target access point ID. Usually `-1` when unset |

### Notes

- The request and response contract is defined by `proto/backend/api/v3/topo.proto`, `proto/backend/api/v3/common.proto`, and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- The router implementation is in `internal/backend/router/api-v3/topo/networkunit.go`. The server automatically narrows the returned results to the network units visible to the current user.
- The `brief` response does not include sensitive fields such as `direct_endpoints` and `custom_deploy_config`. Use the full `networkunit` detail API when complete configuration data is required.
