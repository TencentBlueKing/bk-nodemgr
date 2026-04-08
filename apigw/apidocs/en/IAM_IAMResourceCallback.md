### Description

- API Version: v3.0.1+.
- Required Permission: none.
- Function: Provide Node Manager resource data to BlueKing IAM callbacks, dispatching requests by resource type and callback method.

### URL

POST /api/v3/iam/v3/resource

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| type | string | Yes | Resource type. Resource providers currently registered on this route include `networkarea`, `networkunit`, `package_type`, and `package`. |
| method | string | Yes | IAM callback method identifier that determines how the request is processed. |
| filter | object | No | Query conditions. The structure varies by `method` and follows the IAM callback contract. |
| page | object | No | Pagination parameters used by list-style callback methods. |

#### method enum values

`list_attr`, `list_attr_value`, `list_instance`, `fetch_instance_info`, `list_instance_by_policy`, `search_instance`, `fetch_instance_list`, `fetch_resource_type_schema`

#### page

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| offset | int32 | No | Starting offset for pagination. |
| limit | int32 | No | Maximum number of returned items. |

### Request Example

This is a callback-style endpoint and is typically called by IAM with Basic Auth credentials.

```json
{
  "type": "networkunit",
  "method": "list_instance",
  "filter": {
    "parent": {
      "id": "2"
    }
  },
  "page": {
    "offset": 0,
    "limit": 100
  }
}
```

### Response Example

The `data` structure varies by callback method. The following example shows a typical `list_instance` response.

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 1,
    "results": [
      {
        "id": "1001",
        "display_name": "default-network-unit"
      }
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| code | int32 | Response code, `0` means success. Business errors are usually still returned with HTTP 200 and reflected in this field. |
| message | string | Response message. |
| data | object/null/array/string | Callback result payload. The exact structure depends on `method`. |

#### Common `data` shapes

| Scenario | Type | Description |
|----------|------|-------------|
| `list_instance` | object | Returns a resource list, typically with `count` and `results`. |
| `fetch_instance_info` | array | Returns resource detail entries. |
| `list_attr` / `list_attr_value` | array | Returns resource attributes or attribute values. |
| `fetch_resource_type_schema` | object | Returns the resource type schema. |

#### data.results[n]

| Parameter | Type | Description |
|-----------|------|-------------|
| id | string | Unique resource instance identifier. |
| display_name | string | Display name of the resource instance. |

### Notes

- This endpoint is an IAM callback entry and is not intended for direct use by normal business users.
- The route configuration sets `resourcePermissionRequired: false`, so no business resource permission check is performed.
- Authentication relies on dedicated Basic Auth middleware. The username must be `bk_iam`, and the password must match the current IAM system token.
