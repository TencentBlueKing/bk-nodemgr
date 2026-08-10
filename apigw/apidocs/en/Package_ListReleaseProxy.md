### Description

- API Version: v3.0.1-alpha.67+.
- Required Permission: package_view (View Package).
- Function: List proxy release packages with support for pagination and condition filtering.

### URL

POST /api/v3/package/release/proxy/list

### Input Parameters

| Parameter Name           | Parameter Type | Required | Description                                                                |
| ------------------------ | -------------- | -------- | -------------------------------------------------------------------------- |
| generation               | int64          | Yes      | Package generation (enum values: 2)                                        |
| page                     | object         | Yes      | Required for normal list queries; may be omitted only when only_count=true |
| only_count               | bool           | No       | Return only the total count without list data, default false               |
| exact_include_conditions | object         | No       | Exact filter conditions                                                    |

#### page

| Parameter Name | Parameter Type | Required | Description                                |
| -------------- | -------------- | -------- | ------------------------------------------ |
| offset         | int32          | No       | Offset, starting from 0                    |
| limit          | int32          | Yes      | Number of records per page, from 1 to 1000 |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description                                     |
| -------------- | -------------- | -------- | ----------------------------------------------- |
| platform       | object array   | No       | Platform filter conditions (os_type + cpu_arch) |
| version        | string array   | No       | Version filter conditions                       |
| as_default     | bool array     | No       | Filter by whether it is the default version     |
| enabled        | bool array     | No       | Filter by whether it is enabled                 |
| name           | string array   | No       | Package name filter conditions                  |
| file_name      | string array   | No       | Package file name filter conditions             |
| is_hidden      | bool array     | No       | Hidden state filter                             |

#### exact_include_conditions.platform[n]

| Parameter Name | Parameter Type | Required | Description                                                                  |
| -------------- | -------------- | -------- | ---------------------------------------------------------------------------- |
| os_type        | string         | Yes      | Operating system type (forms a supported platform combination with cpu_arch) |
| cpu_arch       | string         | Yes      | CPU architecture (forms a supported platform combination with os_type)       |

### Request Example

List enabled proxy packages for Linux amd64 in generation 2.

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
          "name": "gse_proxy",
          "generation": 2,
          "release_type": "proxy",
          "os_type": "linux",
          "cpu_arch": "amd64",
          "version": "2.1.6-rc.5",
          "file_name": "gse_proxy-linux-x86_64.tgz",
          "labels": [],
          "enabled": true,
          "as_default": true,
          "is_hidden": false,
          "md5": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
          "updated_at": 1712000000000,
          "operator": "admin"
        },
        "change_log_zh": "修复已知问题",
        "change_log_en": "Bug fixes"
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                  |
| -------------- | -------------- | ---------------------------- |
| code           | int32          | Status code, 0 means success |
| message        | string         | Request message              |
| request_id     | string         | Request ID                   |
| data           | object         | Response data                |

#### data

| Parameter Name | Parameter Type | Description                      |
| -------------- | -------------- | -------------------------------- |
| total          | int64          | Total number of matching records |
| items          | object array   | Proxy package list               |

#### data.items[n]

| Parameter Name | Parameter Type | Description              |
| -------------- | -------------- | ------------------------ |
| release        | object         | Package base information |
| change_log_zh  | string         | Chinese change log       |
| change_log_en  | string         | English change log       |

#### data.items[n].release

| Parameter Name | Parameter Type | Description                                                                  |
| -------------- | -------------- | ---------------------------------------------------------------------------- |
| name           | string         | Package name                                                                 |
| generation     | int64          | Package generation                                                           |
| release_type   | string         | Package type, fixed value: proxy                                             |
| os_type        | string         | Operating system type (forms a supported platform combination with cpu_arch) |
| cpu_arch       | string         | CPU architecture (forms a supported platform combination with os_type)       |
| version        | string         | Version number                                                               |
| file_name      | string         | Package file name                                                            |
| labels         | string array   | Label list                                                                   |
| enabled        | bool           | Whether enabled                                                              |
| as_default     | bool           | Whether this is the default version                                          |
| is_hidden      | bool           | Whether hidden                                                               |
| md5            | string         | Package MD5 checksum                                                         |
| updated_at     | uint64         | Last update time (Unix timestamp in milliseconds)                            |
| operator       | string         | Last operator                                                                |
