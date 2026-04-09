### Description

- API Version: v3.0.0+.
- Required Permission: package_view (View Package).
- Function: Get distinct lists of available OS types and CPU architectures for agent release packages.

### URL

POST /api/v3/package/release/agent/distinct

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 | Yes | Package generation (enum values: 2) |
| exact_include_conditions | object | No | Exact filter conditions |
| distinct_field | object | No | Specifies which fields to deduplicate |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| platform | object array | No | Platform filter conditions (os_type + cpu_arch) |
| version | string array | No | Version filter conditions |
| as_default | bool array | No | Filter by whether it is the default version |
| enabled | bool array | No | Filter by whether it is enabled |
| name | string array | No | Package name filter conditions |
| file_name | string array | No | Package file name filter conditions |

#### exact_include_conditions.platform[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| os_type | string | Yes | Operating system type (enum values: linux, windows, darwin) |
| cpu_arch | string | Yes | CPU architecture (enum values: 386, arm, arm64, amd64) |

#### distinct_field

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| os_type | bool | No | Whether to deduplicate OS types, default false |
| cpu_arch | bool | No | Whether to deduplicate CPU architectures, default false |

### Request Example

Get all available OS types and CPU architectures for generation 2 agent packages.

```json
{
  "generation": 2,
  "distinct_field": {
    "os_type": true,
    "cpu_arch": true
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
    "os_type": ["linux", "windows"],
    "cpu_arch": ["amd64", "arm64"]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 means success |
| message | string | Request message |
| request_id | string | Request ID |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| os_type | string array | Deduplicated OS type list (enum values: linux, windows, darwin) |
| cpu_arch | string array | Deduplicated CPU architecture list (enum values: 386, arm, arm64, amd64) |
