### Description

- API Version: v3.0.1-alpha.1+.
- Required Permission: None.
- Function: Count processes for each plugin name under the specified filter conditions.

### URL

POST /api/v3/process/get_distribution_by_plugin_name

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| exact_include_conditions | object | No | Exact-match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy-match include conditions |

#### exact_include_conditions

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_host_id | int64 array | No | Host ID list |
| bk_biz_id | int64 array | No | Business ID list |
| plugin_group | string array | No | Plugin group list; `default` is common but is not a fixed enum |
| generation | int64 array | No | Process generation list |
| platform_os | string array | No | Operating system type list |
| platform_arch | string array | No | CPU architecture list |
| status | string array | No | Process status list. Available values: `init`, `running`, `stopped`, `unregister`, `unknown` |
| agent_id | string array | No | Agent ID list |
| version | string array | No | Process version list |
| plugin_name | string array | No | Plugin name list |
| plugin_pkg_name | string array | No | Plugin package name list |

#### fuzzy_include_conditions

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| name | string array | No | Process binary name list, matched fuzzily |
| plugin_pkg_name | string array | No | Plugin package name list, matched fuzzily |

**Parameter Notes**:

- Different fields in one condition object are combined with AND; values in one array form the candidate set for that field.
- Both `exact_include_conditions` and `fuzzy_include_conditions` are inclusive filters. This API does not expose exclude filters.

### Request Example

Count running processes by plugin name in business `2`.

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "status": ["running"]
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260717-000003",
  "error": null,
  "permission": null,
  "data": {
    "bk-monitor-agent": 15,
    "bk-log-collector": 8
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission application information |
| data | object | Process distribution by plugin name, keyed by plugin name with matched process counts as values |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| {plugin_name} | int64 | Dynamic field keyed by plugin name |
| value of {plugin_name} | int64 | Number of processes under the plugin name that match the filter conditions |

#### error

| Parameter | Type | Description |
|---------|----------|------|
| system | string | Error system identifier |
| message | string | Error message |
| details | object array | Error detail list |

#### error.details[n]

| Parameter | Type | Description |
|---------|----------|------|
| code | string | Error code |
| message | string | Error message |

### Notes

- `data` is defined as `map<string, int64>` in Proto, so JSON object keys are plugin names.
- The current handler does not apply permission narrowing for business visibility.
