### Description

- API Version: v3.0.0+.
- Required Permission: package_manage (Manage Package).
- Function: Enable the specified plugin package version.

### URL

POST /api/v3/package/release/plugin/enable

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 | Yes | Package generation (enum value: 2) |
| name | string | Yes | Plugin name |
| platform | object | Yes | Target platform |
| version | string | Yes | Package version |

#### platform

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| os_type | string | Yes | Operating system type (enum values: linux, windows, darwin) |
| cpu_arch | string | Yes | CPU architecture (enum values: 386, arm, arm64, amd64) |

### Request Example

```json
{
  "generation": 2,
  "name": "bkmonitorbeat",
  "platform": {"os_type": "linux", "cpu_arch": "amd64"},
  "version": "3.6.0"
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
