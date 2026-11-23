### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access).
- Function: Query network areas through the user-state API, with pagination, exact filters and fuzzy name filters.
- User identity and tenant information are supplied by API Gateway; do not include a username or tenant ID in the request body. The list is not additionally narrowed by `networkarea_view`. Viewing an area's details still requires the corresponding viewing permission.

### URL

POST /api/v3/open/topo/networkarea/list

### Input Parameters

| Parameter                | Type   | Required | Description                                                                          |
| ------------------------ | ------ | -------- | ------------------------------------------------------------------------------------ |
| page                     | object | Yes      | Pagination; a valid `limit` is required even when `only_count=true`                  |
| only_count               | bool   | No       | Not handled separately by this endpoint; it still returns a paginated list and total |
| exact_include_conditions | object | No       | Exact filters                                                                        |
| fuzzy_include_conditions | object | No       | Fuzzy filters                                                                        |

#### page

| Parameter | Type  | Required | Description                                                                |
| --------- | ----- | -------- | -------------------------------------------------------------------------- |
| offset    | int32 | No       | Starting record offset, defaults to `0`; must be nonnegative               |
| limit     | int32 | Yes      | Maximum records to return, in `(0, 1000]`; `0` does not disable pagination |

#### exact_include_conditions

| Parameter         | Type         | Required | Description                                         |
| ----------------- | ------------ | -------- | --------------------------------------------------- |
| bk_networkarea_id | int64 array  | No       | Network area IDs                                    |
| cloud_vendor      | string array | No       | CMDB cloud vendor identifiers, not network area IDs |

#### fuzzy_include_conditions

| Parameter           | Type         | Required | Description                |
| ------------------- | ------------ | -------- | -------------------------- |
| bk_networkarea_name | string array | No       | Network area name patterns |

### Request Example

```json
{
  "page": { "offset": 0, "limit": 10 },
  "fuzzy_include_conditions": { "bk_networkarea_name": ["prod"] }
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
        "bk_networkarea_id": 1,
        "bk_networkarea_name": "prod-default",
        "cloud_vendor": "1"
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

| Parameter | Type         | Description                       |
| --------- | ------------ | --------------------------------- |
| total     | int64        | Total number of matching records  |
| items     | object array | Network areas on the current page |

#### data.items[n]

| Parameter           | Type   | Description                  |
| ------------------- | ------ | ---------------------------- |
| tenant_id           | string | Tenant ID                    |
| bk_networkarea_id   | int64  | Network area ID              |
| bk_networkarea_name | string | Network area name            |
| cloud_vendor        | string | CMDB cloud vendor identifier |
