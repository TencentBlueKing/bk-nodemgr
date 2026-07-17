### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: package_view (View Package).
- API Function: Query Proxy package brief list with core fields including changelogs, supports pagination and filtering.

### URL

POST /api/v3/package/release/proxy/list/brief

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| generation | int64 | Yes | Package generation (enum: 2) |
| page | object | Yes | Required for normal list queries; may be omitted only when only_count=true |
| only_count | bool | No | Only return total count without list data, default false |
| exact_include_conditions | object | No | Exact filter conditions |

#### page

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| offset | int32 | No | Offset, starts from 0 |
| limit | int32 | Yes | Items per page, from 1 to 1000 |

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
| os_type | string | Yes | OS type (forms a supported platform combination with cpu_arch) |
| cpu_arch | string | Yes | CPU architecture (forms a supported platform combination with os_type) |

### Request Example

Query enabled Proxy packages for Linux amd64 platform with generation 2.

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
        "as_default": true,
        "change_log_en": "Bug fixes and performance improvements",
        "change_log_zh": "修复已知问题并提升性能"
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
| items | object array | Proxy package brief list |

#### data.items[n]

| Parameter | Type | Description |
|-----------|------|-------------|
| generation | int64 | Package generation |
| os_type | string | OS type (forms a supported platform combination with cpu_arch) |
| cpu_arch | string | CPU architecture (forms a supported platform combination with os_type) |
| version | string | Version number |
| enabled | bool | Enabled status |
| as_default | bool | Default version flag |
| change_log_en | string | English changelog |
| change_log_zh | string | Chinese changelog |
