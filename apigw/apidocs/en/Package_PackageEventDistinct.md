### Description

- API Version: v3.0.1-alpha.13+.
- Required Permission: None.
- Function: Query distinct values of package operation event fields.

### URL

POST /api/v3/package/event/distinct

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| exact_include_conditions | object | No | Exact-match conditions |
| fuzzy_include_conditions | object | No | Fuzzy-match conditions; no fields are currently available |
| operate_time_range | object | No | Operation time range filter |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 array | No | Package generations (valid value: 2) |
| os_type | string array | No | Operating system types; each value must be a supported OS type |
| cpu_arch | string array | No | CPU architectures; each value must be a supported CPU architecture |
| release_type | string array | No | Package types (valid values: agent, proxy, cert, bintool, plugin_bintool, plugin) |
| operator | string array | No | Operators |
| event_type | string array | No | Event types (valid values: publish, delete, enable, disable, set_as_default, cancel_as_default, upload) |
| version | string array | No | Versions |

#### operate_time_range

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| start_timestamp_sec | int64 | Yes | Start time as a Unix timestamp in seconds |
| end_timestamp_sec | int64 | Yes | End time as a Unix timestamp in seconds |

### Request Example

```json
{
  "exact_include_conditions": {"generation": [2]},
  "fuzzy_include_conditions": {},
  "operate_time_range": {
    "start_timestamp_sec": 1748736000,
    "end_timestamp_sec": 1751327999
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "release_type": ["agent", "proxy", "plugin"],
    "event_type": ["publish", "enable", "disable"],
    "os_type": ["linux", "windows"],
    "cpu_arch": ["amd64", "arm64"],
    "version": ["2.1.6-rc.5"],
    "operator": ["admin"]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, where 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| data | object | Response data |
| data.release_type | string array | Distinct package types |
| data.event_type | string array | Distinct event types |
| data.os_type | string array | Distinct operating system types |
| data.cpu_arch | string array | Distinct CPU architectures |
| data.version | string array | Distinct versions |
| data.operator | string array | Distinct operators |
