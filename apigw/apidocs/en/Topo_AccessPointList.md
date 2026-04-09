### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Query the access point list with pagination and exact filtering by network area ID and access point ID.

### URL

POST /api/v3/topo/accesspoint/list

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
| bk_networkarea_id | int64 array | No | Network area ID list |
| accesspoint_id | int64 array | No | Access point ID list |

### Request Example

Query access points under the specified network area and return the total count.

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "accesspoint_id",
    "order": "ASC"
  },
  "exact_include_conditions": {
    "bk_networkarea_id": [
      1
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
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-shenzhen-prod",
        "bk_networkarea_id": 1,
        "endpoints": {
          "cluster": [
            "https://bcs-cluster.example.com"
          ],
          "file": [
            "https://bcs-file.example.com"
          ],
          "data": [
            "https://bcs-data.example.com"
          ]
        }
      },
      {
        "tenant_id": "default",
        "accesspoint_id": 1002,
        "accesspoint_name": "ap-shanghai-prod",
        "bk_networkarea_id": 1,
        "endpoints": {
          "cluster": [
            "https://bcs-cluster-sh.example.com"
          ],
          "file": [
            "https://bcs-file-sh.example.com"
          ],
          "data": [
            "https://bcs-data-sh.example.com"
          ]
        }
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
| permission | object | Permission information. The API may return permission-related details when the requested access point scope exceeds the network units visible to the current user |
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
| accesspoint_id | int64 | Access point ID |
| accesspoint_name | string | Access point name |
| bk_networkarea_id | int64 | Owning network area ID |
| endpoints | object | Access point endpoint configuration |

#### data.items[n].endpoints

| Parameter | Type | Description |
|---------|----------|------|
| cluster | string array | Cluster endpoint list |
| file | string array | File endpoint list |
| data | string array | Data endpoint list |

### Notes

- The request and response contract is defined by `proto/backend/api/v3/topo.proto` and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- The router implementation is in `internal/backend/router/api-v3/topo/accesspoint.go`. The server automatically narrows the returned access points to those under network units the current user is allowed to view.
