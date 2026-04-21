### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: package_view (View Package).
- API Function: Query Agent package brief list with only core fields, supports pagination and filtering.

### URL

POST /api/v3/package/release/agent/list/brief

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| generation | int64 | Yes | Package generation (enum: 2) |
| page | object | No | Pagination config, returns all data if not provided |
| only_count | bool | No | Only return total count without list data, default false |
| exact_include_conditions | object | No | Exact filter conditions |

#### page

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| offset | int32 | No | Offset, starts from 0 |
| limit | int32 | No | Items per page |

#### exact_include_conditions

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| platform | object array | No | Platform (OS + CPU arch) filter |
| version | string array | No | Version filter |
| as_default | bool array | No | Default version filter |
| enabled | bool array | No | Enabled status filter |
| name | string array | No | Package name filter |
| file_name | string array | No | Package file name filter |

#### exact_include_conditions.platform[n]

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| os_type | string | Yes | OS type (enum: linux, windows, darwin) |
| cpu_arch | string | Yes | CPU architecture (enum: 386, arm, arm64, amd64) |

### Request Example

Query enabled Agent packages for Linux amd64 platform with generation 2.

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
        "generation": 2,
        "os_type": "linux",
        "cpu_arch": "amd64",
        "version": "2.1.6-rc.5",
        "enabled": true,
        "as_default": true
      }
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| code | int32 | Status code, 0 for success |
| message | string | Response message |
| request_id | string | Request ID |
| data | object | Response data |

#### data

| Parameter | Type | Description |
|-----------|------|-------------|
| total | int64 | Total count of matching records |
| items | object array | Agent package brief list |

#### data.items[n]

| Parameter | Type | Description |
|-----------|------|-------------|
| generation | int64 | Package generation |
| os_type | string | OS type (enum: linux, windows, darwin) |
| cpu_arch | string | CPU architecture (enum: 386, arm, arm64, amd64) |
| version | string | Version number |
| enabled | bool | Enabled status |
| as_default | bool | Default version flag |
