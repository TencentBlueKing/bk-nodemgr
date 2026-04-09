### Description

- API Version: v3.0.0+.
- Required Permission: package_view (View Package).
- Function: List plugin binary tool (PluginBinTool) release packages. Only packages authorized for the current user are returned.

### URL

POST /api/v3/package/release/plugin_bintool/list

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 | Yes | Package generation (enum values: 2) |

**Parameter Notes**:
- The API automatically filters results based on the user's `package_view` permission scope, returning only authorized plugin binary tool packages.

### Request Example

```json
{
  "generation": 2
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
        "release": {
          "name": "bkmonitorbeat",
          "generation": 2,
          "release_type": "plugin_bintool",
          "os_type": "linux",
          "cpu_arch": "amd64",
          "version": "3.20.2305",
          "file_name": "bkmonitorbeat-bintool-3.20.2305.tgz",
          "labels": [],
          "enabled": true,
          "as_default": true,
          "md5": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
          "updated_at": 1712000000,
          "operator": "admin"
        }
      }
    ]
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
| total | int64 | Total number of matching records |
| items | object array | Plugin BinTool package list |

#### data.items[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| release | object | Package base information |

#### data.items[n].release

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| name | string | Plugin name |
| generation | int64 | Package generation |
| release_type | string | Package type, fixed value: plugin_bintool |
| os_type | string | Operating system type (enum values: linux, windows, darwin) |
| cpu_arch | string | CPU architecture (enum values: 386, arm, arm64, amd64) |
| version | string | Version number |
| file_name | string | Package file name |
| labels | string array | Label list |
| enabled | bool | Whether enabled |
| as_default | bool | Whether this is the default version |
| md5 | string | Package MD5 checksum |
| updated_at | uint64 | Last update time (Unix timestamp in seconds) |
| operator | string | Last operator |
