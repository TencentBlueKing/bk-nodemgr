### Description

- API Version: v3.0.0+.
- Required Permission: package_view (View Package).
- Function: List plugin release packages with support for pagination and condition filtering. Only packages authorized for the current user are returned.

### URL

POST /api/v3/package/release/plugin/list

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| generation | int64 | Yes | Package generation (enum values: 2) |
| page | object | No | Pagination configuration; returns all data if not specified |
| only_count | bool | No | Return only the total count without list data, default false |
| exact_include_conditions | object | No | Exact filter conditions |

#### page

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| offset | int32 | No | Offset, starting from 0 |
| limit | int32 | No | Number of records per page |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| platform | object array | No | Platform filter conditions (os_type + cpu_arch) |
| version | string array | No | Version filter conditions |
| as_default | bool array | No | Filter by whether it is the default version |
| enabled | bool array | No | Filter by whether it is enabled |
| name | string array | No | Plugin name filter conditions; if not specified, returns all authorized plugins |
| file_name | string array | No | Package file name filter conditions |

#### exact_include_conditions.platform[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| os_type | string | Yes | Operating system type (enum values: linux, windows, darwin) |
| cpu_arch | string | Yes | CPU architecture (enum values: 386, arm, arm64, amd64) |

**Parameter Notes**:
- `exact_include_conditions.name`: The API intersects the requested plugin names with authorized plugin names for the current user. Plugin packages for which the user has no permission are filtered out.

### Request Example

List enabled plugin packages for Linux amd64 in generation 2.

```json
{
  "generation": 2,
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "platform": [
      {
        "os_type": "linux",
        "cpu_arch": "amd64"
      }
    ],
    "enabled": [true]
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
        "release": {
          "name": "bkmonitorbeat",
          "generation": 2,
          "release_type": "plugin",
          "os_type": "linux",
          "cpu_arch": "amd64",
          "version": "3.20.2305",
          "file_name": "bkmonitorbeat-3.20.2305.tgz",
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
| items | object array | Plugin package list |

#### data.items[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| release | object | Package base information |

#### data.items[n].release

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| name | string | Plugin name |
| generation | int64 | Package generation |
| release_type | string | Package type, fixed value: plugin |
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
