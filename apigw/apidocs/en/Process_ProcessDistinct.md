### Description

- API Version: v3.0.1-alpha.1+.
- Required Permission: None.
- Function: Get distinct candidate values for selected process fields under the specified filter conditions.

### URL

POST /api/v3/process/distinct

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| selector | object | No | Selects fields for distinct queries; fields set to `true` are queried |
| exact_include_conditions | object | No | Exact-match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy-match include conditions |

#### selector

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| os_type | bool | No | Whether to return distinct operating system types |
| cpu_arch | bool | No | Whether to return distinct CPU architectures |
| version | bool | No | Whether to return distinct process versions |
| status | bool | No | Whether to return distinct process statuses |
| plugin_name | bool | No | Whether to return distinct plugin names |
| plugin_group | bool | No | Whether to return distinct plugin groups |
| plugin_pkg_name | bool | No | Whether to return distinct plugin package names |

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
- Only fields set to `true` in `selector` are queried. Unselected response fields are returned as empty arrays.

### Request Example

Query operating system type, CPU architecture, and plugin name candidates for running processes in business `2`.

```json
{
  "selector": {
    "os_type": true,
    "cpu_arch": true,
    "plugin_name": true
  },
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
  "request_id": "req-20260717-000004",
  "error": null,
  "permission": null,
  "data": {
    "os_type": ["linux", "windows"],
    "cpu_arch": ["amd64", "arm64"],
    "version": [],
    "status": [],
    "plugin_name": ["bk-monitor-agent", "bk-log-collector"],
    "plugin_group": [],
    "plugin_pkg_name": []
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
| data | object | Distinct query results |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| os_type | string array | Distinct operating system types |
| cpu_arch | string array | Distinct CPU architectures |
| version | string array | Distinct process versions |
| status | string array | Distinct process statuses. Available values: `init`, `running`, `stopped`, `unregister`, `unknown` |
| plugin_name | string array | Distinct plugin names |
| plugin_group | string array | Distinct plugin groups |
| plugin_pkg_name | string array | Distinct plugin package names |

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

- The response always includes all seven fields under `data`. Fields not selected in `selector` are returned as empty arrays.
- The current handler does not apply permission narrowing for business visibility.
