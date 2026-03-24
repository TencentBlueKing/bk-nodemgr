### Description

- API Version: v3.0.1+.
- Required Permission: .
- Function: Query the plugin list with pagination, count-only mode, and exact or fuzzy filters.

### URL

POST /api/v3/plugin/list

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| page | object | No | Pagination parameters; can be omitted when `only_count=true` |
| only_count | bool | No | Whether to return only the total count; when `true`, no plugin items are returned |
| exact_include_conditions | object | No | Exact-match filter conditions |
| fuzzy_include_conditions | object | No | Fuzzy-match filter conditions |

#### page

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| offset | int32 | No | Pagination offset |
| limit | int32 | No | Maximum number of returned records |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| name | string array | No | Exact match on plugin names |
| group | string array | No | Exact match on plugin groups |

#### fuzzy_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| name | string array | No | Fuzzy match on plugin names |
| pkg_name | string array | No | Fuzzy match on plugin package names |

**Parameter Notes**:
- When `only_count=true`, the API returns only the matched total count, and `items` is empty or omitted in practice.
- Both `exact_include_conditions` and `fuzzy_include_conditions` are inclusive filters; exclude filters are not exposed by this API.
- `group` commonly uses `default`, but the API itself does not enforce a fixed enum.

### Request Example

Query plugins whose names contain `monitor`, returning the first 20 records.

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "fuzzy_include_conditions": {
    "name": [
      "monitor"
    ]
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000003",
  "error": null,
  "data": {
    "total": 2,
    "items": [
      {
        "tenant_id": "default",
        "name": "bk-monitor-agent",
        "group": "default",
        "pkg_name": "bkmonitoragent",
        "memo": "BlueKing monitor agent"
      },
      {
        "tenant_id": "default",
        "name": "bk-monitor-proxy",
        "group": "default",
        "pkg_name": "bkmonitorproxy",
        "memo": "BlueKing monitor proxy"
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, null on success |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| total | int64 | Total number of plugins matching the conditions |
| items | object array | Plugin list; typically empty when `only_count=true` |

#### data.items[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| tenant_id | string | Tenant ID |
| name | string | Plugin name |
| group | string | Plugin group |
| pkg_name | string | Plugin package name |
| memo | string | Plugin memo |

#### error

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| system | string | Error system identifier |
| message | string | Error message |
| details | array | Error detail list |

#### error.details[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | string | Error code |
| message | string | Error message |
