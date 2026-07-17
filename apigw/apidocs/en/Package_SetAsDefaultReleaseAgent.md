### Description

- API Version: v3.0.1-alpha.14+.
- Required Permission: package_manage (Manage Package).
- Function: Set the specified agent package version as the default version.

### URL

POST /api/v3/package/release/agent/set_as_default

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 | Yes | Package generation (enum value: 2) |
| release_type | string | Yes | Package type, fixed to agent for this API |
| platform | object | Yes | Target platform |
| version | string | Yes | Package version |

#### platform

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| os_type | string | Yes | Operating system type (forms a supported platform combination with cpu_arch) |
| cpu_arch | string | Yes | CPU architecture (forms a supported platform combination with os_type) |

### Request Example

```json
{
  "generation": 2,
  "release_type": "agent",
  "platform": {"os_type": "linux", "cpu_arch": "amd64"},
  "version": "2.1.6-rc.5"
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
