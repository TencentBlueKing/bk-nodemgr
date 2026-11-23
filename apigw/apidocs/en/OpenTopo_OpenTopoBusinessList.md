### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access).
- Function: Query a paginated business list in the current tenant through the user-state API. Supports exact business ID filters and fuzzy business name filters. The business list is not filtered by the user's Agent or Proxy viewing permissions.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/business/list

### Input Parameters

| Parameter                | Type   | Required | Description                                                                                                                  |
| ------------------------ | ------ | -------- | ---------------------------------------------------------------------------------------------------------------------------- |
| page                     | object | Yes      | Pagination settings; a valid `limit` is required                                                                             |
| only_count               | bool   | No       | Defaults to `false`. This endpoint does not switch to count-only mode based on this field and still returns a paginated list |
| exact_include_conditions | object | No       | Exact-match inclusion conditions                                                                                             |
| fuzzy_include_conditions | object | No       | Fuzzy-match inclusion conditions                                                                                             |

#### page

| Parameter | Type  | Required | Description                                                               |
| --------- | ----- | -------- | ------------------------------------------------------------------------- |
| offset    | int32 | No       | Offset; defaults to `0` and must be nonnegative                           |
| limit     | int32 | Yes      | Page size, from `1` to `1000`; `0` or omitted pagination is not supported |

#### exact_include_conditions

| Parameter | Type        | Required | Description  |
| --------- | ----------- | -------- | ------------ |
| bk_biz_id | int64 array | No       | Business IDs |

#### fuzzy_include_conditions

| Parameter   | Type         | Required | Description                       |
| ----------- | ------------ | -------- | --------------------------------- |
| bk_biz_name | string array | No       | Business names for fuzzy matching |

### Request Example

```json
{
  "page": { "offset": 0, "limit": 10 },
  "exact_include_conditions": { "bk_biz_id": [2] }
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
        "bk_biz_id": 2,
        "bk_biz_name": "Example Business"
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                                                                  |
| ---------- | ------ | ------------------------------------------------------------------------------------------------------------ |
| code       | int32  | Status code; `0` indicates success                                                                           |
| message    | string | Request result message                                                                                       |
| request_id | string | Request ID                                                                                                   |
| error      | object | Error information, including `system`, `message`, and `details` (error codes and messages); empty on success |
| permission | object | Permission application information that may be returned when access is denied                                |
| data       | object | Response data                                                                                                |

#### data

| Parameter | Type         | Description                                     |
| --------- | ------------ | ----------------------------------------------- |
| total     | int64        | Total number of records matching the conditions |
| items     | object array | Businesses on the current page                  |

#### data.items[n]

| Parameter   | Type   | Description   |
| ----------- | ------ | ------------- |
| tenant_id   | string | Tenant ID     |
| bk_biz_id   | int64  | Business ID   |
| bk_biz_name | string | Business name |
