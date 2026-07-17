### Description

- API Version: v3.0.0+.
- Required Permission: None.
- Function: Set labels in bulk for agent packages that match the conditions.

### URL

POST /api/v3/package/release/agent/set_labels_many

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 | Yes | Package generation (enum value: 2) |
| exact_include_conditions | object | Yes | Exact-match conditions |
| labels | string array | Yes | Labels to set |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| platform | object array | No | Platform list |
| version | string array | No | Version list |
| as_default | bool array | No | Whether the package is a default version |
| enabled | bool array | No | Whether the package is enabled |
| name | string array | No | Package name list |
| file_name | string array | No | File name list |

#### platform

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| os_type | string | Yes | Operating system type (enum values: linux, windows, darwin) |
| cpu_arch | string | Yes | CPU architecture (enum values: 386, arm, arm64, amd64) |

### Request Example

```json
{
  "generation": 2,
  "exact_include_conditions": {
    "platform": [{"os_type": "linux", "cpu_arch": "amd64"}],
    "version": ["2.1.6-rc.5"]
  },
  "labels": ["stable"]
}
```

### Response Example

```json
{"code": 0, "message": "ok", "request_id": "req-1234567890", "data": {}}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, where 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| data | object | Empty object indicating a successful operation |
