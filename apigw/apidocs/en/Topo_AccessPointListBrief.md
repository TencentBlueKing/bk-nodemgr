### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: None.
- Function: Query the access point brief list with exact filtering by network area ID and access point ID. The `brief` response returns only the basic
  access point identity fields and does not expose sensitive endpoint configuration such as `endpoints`.

### URL

POST /api/v3/topo/accesspoint/list/brief

### Request Parameters

| Parameter                | Type   | Required | Description                                                                                                                                      |
|--------------------------|--------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| page                     | object | No       | Pagination configuration. When `only_count` is not `true`, `offset` must be greater than or equal to `0`, and `limit` must be within `(0, 1000]` |
| only_count               | bool   | No       | Whether to return only the total count without details                                                                                           |
| exact_include_conditions | object | No       | Exact match include conditions                                                                                                                   |

#### page

| Parameter | Type  | Required | Description                                                    |
|-----------|-------|----------|----------------------------------------------------------------|
| offset    | int32 | No       | Record start offset, starting from `0`                         |
| limit     | int32 | No       | Maximum number of returned records. Valid range is `(0, 1000]` |

#### exact_include_conditions

Exact match include conditions. A record can participate in the narrowed query scope if any access point ID or any network area ID matches.

| Parameter         | Type        | Required | Description          |
|-------------------|-------------|----------|----------------------|
| bk_networkarea_id | int64 array | No       | Network area ID list |
| accesspoint_id    | int64 array | No       | Access point ID list |

### Request Example

Query the brief list of access points under the specified network area and return the total count.

```json
{
  "page": {
    "offset": 0,
    "limit": 10
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
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-shenzhen-prod",
        "bk_networkarea_id": 1
      },
      {
        "accesspoint_id": 1002,
        "accesspoint_name": "ap-shanghai-prod",
        "bk_networkarea_id": 1
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                                                                                                          |
|------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| code       | int32  | Status code, `0` means success                                                                                                                       |
| message    | string | Response message                                                                                                                                     |
| request_id | string | Request ID                                                                                                                                           |
| error      | object | Error information, empty on success                                                                                                                  |
| permission | object | Permission information. The API may return permission-related details when the requested scope exceeds the network units visible to the current user |
| data       | object | Response data                                                                                                                                        |

#### data

| Parameter | Type  | Description                                                                 |
|-----------|-------|-----------------------------------------------------------------------------|
| total     | int64 | Total record count matching the current criteria                            |
| items     | array | Returned records. This field is typically empty when `only_count` is `true` |

#### data.items[n]

| Parameter         | Type   | Description            |
|-------------------|--------|------------------------|
| accesspoint_id    | int64  | Access point ID        |
| accesspoint_name  | string | Access point name      |
| bk_networkarea_id | int64  | Owning network area ID |

### Notes

- The request and response contract is defined by `proto/backend/api/v3/topo.proto`, `proto/backend/api/v3/common.proto`, and
  `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- The router implementation is in `internal/backend/router/api-v3/topo/accesspoint.go`. This `brief` list API currently performs the list query directly from the request conditions without additional permission-based narrowing.
- The `brief` response does not include access point endpoint configuration such as `endpoints`. Use the full `accesspoint/list` API when complete
  endpoint details are required.
