### Description

- API Version: v3.0.1-alpha.13+.
- Required Permission: None.
- Function: Query package operation events with pagination.

### URL

POST /api/v3/package/event/list

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| page | object | Yes | Pagination settings |
| only_count | bool | No | Whether to return only the event count; items is empty when true |
| exact_include_conditions | object | No | Exact-match conditions |
| fuzzy_include_conditions | object | No | Fuzzy-match conditions; no fields are currently available |
| operate_time_range | object | No | Operation time range; defaults to the latest 365 days and cannot exceed 365 days |

#### page

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| offset | int32 | Yes | Data offset, starting from 0 |
| limit | int32 | Yes | Number of items per page, from 1 to 1000 |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 array | No | Package generations (valid value: 2) |
| os_type | string array | No | Operating system types (valid values: linux, windows, darwin) |
| cpu_arch | string array | No | CPU architectures (valid values: 386, arm, arm64, amd64) |
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
  "page": {"offset": 0, "limit": 20},
  "only_count": false,
  "exact_include_conditions": {
    "release_type": ["agent"],
    "event_type": ["enable", "disable"]
  },
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
    "total": 1,
    "items": [
      {
        "name": "agent",
        "event_type": "enable",
        "generation": 2,
        "release_type": "agent",
        "os_type": "linux",
        "cpu_arch": "amd64",
        "version": "2.1.6-rc.5",
        "operate_time": 1750000000000,
        "operator": "admin"
      }
    ]
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
| data.total | int64 | Total number of events |
| data.items | object array | Event list; empty when only_count is true |
| data.items[].name | string | Package name |
| data.items[].event_type | string | Event type |
| data.items[].generation | int64 | Package generation |
| data.items[].release_type | string | Package type |
| data.items[].os_type | string | Operating system type |
| data.items[].cpu_arch | string | CPU architecture |
| data.items[].version | string | Version |
| data.items[].operate_time | int64 | Operation time as a Unix timestamp in milliseconds |
| data.items[].operator | string | Operator |
