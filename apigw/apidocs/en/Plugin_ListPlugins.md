### Description

- API Version: v3.0.1+.
- Required Permission: plugin_view (View Plugin).
- Function: Query the plugin list with pagination, count-only mode, and exact or fuzzy filters.
- Since v3.0.1-alpha.117+, `visible_biz_ids` is applied throughout the shared query path: the plugin list and count are filtered by the narrowed business visibility scope. Exact-array matching for multiple business IDs is unchanged.

### URL

POST /api/v3/plugin/list

### Input Parameters

| Parameter Name           | Parameter Type | Required | Description                                                                                     |
| ------------------------ | -------------- | -------- | ----------------------------------------------------------------------------------------------- |
| page                     | object         | No       | Pagination parameters; required for non-count queries and can be omitted when `only_count=true` |
| only_count               | bool           | No       | Whether to return only the total count; when `true`, no plugin items are returned               |
| exact_include_conditions | object         | No       | Exact-match filter conditions                                                                   |
| fuzzy_include_conditions | object         | No       | Fuzzy-match filter conditions                                                                   |

#### page

| Parameter Name | Parameter Type | Required | Description                                                   |
| -------------- | -------------- | -------- | ------------------------------------------------------------- |
| offset         | int32          | No       | Pagination offset; defaults to `0`                            |
| limit          | int32          | Yes      | Maximum number of returned records; value range is `(0, 500]` |

#### exact_include_conditions

| Parameter Name  | Parameter Type | Required | Description                                                                                                  |
| --------------- | -------------- | -------- | ------------------------------------------------------------------------------------------------------------ |
| name            | string array   | No       | Exact match on plugin names                                                                                  |
| group           | string array   | No       | Exact match on plugin groups                                                                                 |
| visible_biz_ids | int64 array    | No       | Plugin visibility business scope, narrowed by the caller's authorization; omission uses the authorized scope |

#### fuzzy_include_conditions

| Parameter Name | Parameter Type | Required | Description                         |
| -------------- | -------------- | -------- | ----------------------------------- |
| name           | string array   | No       | Fuzzy match on plugin names         |
| pkg_name       | string array   | No       | Fuzzy match on plugin package names |

**Parameter Notes**:

- When `only_count=true`, the API returns only the matched total count, and `items` is empty or omitted in practice.
- Both `exact_include_conditions` and `fuzzy_include_conditions` are inclusive filters; exclude filters are not exposed by this API.
- `group` commonly uses `default`, but the API itself does not enforce a fixed enum.
- When `only_count=false` or omitted, provide a valid `page.limit` in `(0, 500]`.
- The API checks the caller's `plugin_view` permission and narrows `visible_biz_ids` to the authorized business scope. A restricted caller with an empty effective scope receives an empty list or zero count; a failed permission check returns a permission error.
- For a nonempty effective business scope, results include plugins with an empty, null, or missing visibility field, and plugins whose visible-business-ID array exactly equals the effective business-scope array. This is not an any-overlap membership filter. A caller with unrestricted business permissions and no business condition has no additional visible-business filter.
- When loading group options through `POST /api/v3/plugin/distinct`, supply the same `visible_biz_ids` as the list and refresh the options when the selected business changes.

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
    "name": ["monitor"]
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

| Parameter Name | Parameter Type | Description                        |
| -------------- | -------------- | ---------------------------------- |
| code           | int32          | Status code, 0 indicates success   |
| message        | string         | Request message                    |
| request_id     | string         | Request ID                         |
| error          | object         | Error information, null on success |
| data           | object         | Response data                      |

#### data

| Parameter Name | Parameter Type | Description                                         |
| -------------- | -------------- | --------------------------------------------------- |
| total          | int64          | Total number of plugins matching the conditions     |
| items          | object array   | Plugin list; typically empty when `only_count=true` |

#### data.items[n]

| Parameter Name | Parameter Type | Description         |
| -------------- | -------------- | ------------------- |
| tenant_id      | string         | Tenant ID           |
| name           | string         | Plugin name         |
| group          | string         | Plugin group        |
| pkg_name       | string         | Plugin package name |
| memo           | string         | Plugin memo         |

#### error

| Parameter Name | Parameter Type | Description             |
| -------------- | -------------- | ----------------------- |
| system         | string         | Error system identifier |
| message        | string         | Error message           |
| details        | array          | Error detail list       |

#### error.details[n]

| Parameter Name | Parameter Type | Description   |
| -------------- | -------------- | ------------- |
| code           | string         | Error code    |
| message        | string         | Error message |
