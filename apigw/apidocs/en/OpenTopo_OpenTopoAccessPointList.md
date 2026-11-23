### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access), networkunit_view (View Network Unit).
- Function: Query access points and endpoint addresses through the user-state API. Results are narrowed to access points associated with network units the user may view.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/accesspoint/list

### Input Parameters

| Parameter                | Type   | Required    | Description                                                                 |
| ------------------------ | ------ | ----------- | --------------------------------------------------------------------------- |
| page                     | object | Conditional | A valid `limit` is required unless `only_count=true`                        |
| only_count               | bool   | No          | Defaults to `false`; when `true`, returns only the total with empty `items` |
| exact_include_conditions | object | No          | Exact filters                                                               |

#### page

| Parameter | Type  | Required    | Description                                                                                                |
| --------- | ----- | ----------- | ---------------------------------------------------------------------------------------------------------- |
| offset    | int32 | No          | Record offset, defaults to `0`; must be nonnegative                                                        |
| limit     | int32 | Conditional | In `(0, 1000]` for list queries; may be omitted for count-only queries, but cannot be negative if provided |

#### exact_include_conditions

| Parameter         | Type        | Required | Description                                                                               |
| ----------------- | ----------- | -------- | ----------------------------------------------------------------------------------------- |
| bk_networkarea_id | int64 array | No       | Network area IDs                                                                          |
| accesspoint_id    | int64 array | No       | Access point IDs, intersected with access points associated with authorized network units |

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
        "tenant_id": "default",
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-prod",
        "bk_networkarea_id": 1,
        "endpoints": {
          "cluster": ["10.0.0.10:48668"],
          "file": ["10.0.0.10:48669"],
          "data": ["10.0.0.10:48670"]
        },
        "cloud_security_group_pool_id": ""
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

| Parameter | Type         | Description                                                 |
| --------- | ------------ | ----------------------------------------------------------- |
| total     | int64        | Total matching access points within the authorized scope    |
| items     | object array | Current page of access points; empty for count-only queries |

#### data.items[n]

| Parameter                    | Type   | Description                                                                                 |
| ---------------------------- | ------ | ------------------------------------------------------------------------------------------- |
| tenant_id                    | string | Tenant ID                                                                                   |
| accesspoint_id               | int64  | Access point ID                                                                             |
| accesspoint_name             | string | Access point name                                                                           |
| bk_networkarea_id            | int64  | Owning network area ID                                                                      |
| endpoints                    | object | Access point endpoint addresses                                                             |
| cloud_security_group_pool_id | string | Bound global security group pool ID; empty when cloud security group management is disabled |

#### data.items[n].endpoints

| Parameter | Type         | Description               |
| --------- | ------------ | ------------------------- |
| cluster   | string array | Cluster channel addresses |
| file      | string array | File channel addresses    |
| data      | string array | Data channel addresses    |

### Notes

- Returns an empty list when no access points match after authorization narrowing; authorization check failures return permission errors.
- Use `OpenTopo_OpenTopoAccessPointListBrief` for basic identifiers only. That endpoint omits endpoint addresses and has different visibility rules.
