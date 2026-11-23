### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access).
- Function: Query access point summaries through the user-state API. Returns basic identifiers only, without endpoint addresses or security group pool configuration.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body. This summary list is not additionally narrowed by `networkunit_view` and has different visibility rules from the full access point list.

### URL

POST /api/v3/open/topo/accesspoint/list/brief

### Input Parameters

| Parameter                | Type   | Required    | Description                                                                 |
| ------------------------ | ------ | ----------- | --------------------------------------------------------------------------- |
| page                     | object | Conditional | A valid `limit` is required unless `only_count=true`                        |
| only_count               | bool   | No          | Defaults to `false`; when `true`, returns only the total with empty `items` |
| exact_include_conditions | object | No          | Exact filters                                                               |

#### page

| Parameter | Type  | Required    | Description                                                                                                |
| --------- | ----- | ----------- | ---------------------------------------------------------------------------------------------------------- |
| offset    | int32 | No          | Starting offset, defaults to `0`; must be nonnegative                                                      |
| limit     | int32 | Conditional | In `(0, 1000]` for list queries; may be omitted for count-only queries, but cannot be negative if provided |

#### exact_include_conditions

| Parameter         | Type        | Required | Description      |
| ----------------- | ----------- | -------- | ---------------- |
| bk_networkarea_id | int64 array | No       | Network area IDs |
| accesspoint_id    | int64 array | No       | Access point IDs |

### Request Example

```json
{
  "page": { "offset": 0, "limit": 10 },
  "exact_include_conditions": { "bk_networkarea_id": [1] }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      {
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-prod",
        "bk_networkarea_id": 1
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                                   |
| ---------- | ------ | ----------------------------------------------------------------------------- |
| code       | int32  | Status code; `0` indicates success                                            |
| message    | string | Request result message                                                        |
| request_id | string | Request ID                                                                    |
| error      | object | Error information; empty on success                                           |
| permission | object | Permission application information that may be returned when access is denied |
| data       | object | Response data                                                                 |

#### data

| Parameter | Type         | Description                                                     |
| --------- | ------------ | --------------------------------------------------------------- |
| total     | int64        | Total matching access points                                    |
| items     | object array | Current page of basic information; empty for count-only queries |

#### data.items[n]

| Parameter         | Type   | Description            |
| ----------------- | ------ | ---------------------- |
| accesspoint_id    | int64  | Access point ID        |
| accesspoint_name  | string | Access point name      |
| bk_networkarea_id | int64  | Owning network area ID |

### Notes

Use `OpenTopo_OpenTopoAccessPointList` for endpoint addresses; viewing permission for the corresponding network units is required.
