### Description

- API Version: v3.0.1-alpha.66+.
- Required Permission: None.
- Function: Get distinct lists of available OS types, CPU architectures, plugin package names, and versions for plugin packages.

### URL

POST /api/v3/package/release/plugin/distinct

### Input Parameters

| Parameter Name           | Parameter Type | Required | Description                                |
| ------------------------ | -------------- | -------- | ------------------------------------------ |
| generation               | int64          | Yes      | Plugin package generation (enum values: 2) |
| exact_include_conditions | object         | No       | Exact filter conditions                    |
| distinct_field           | object         | No       | Specifies which fields to deduplicate      |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description                                     |
| -------------- | -------------- | -------- | ----------------------------------------------- |
| platform       | object array   | No       | Platform filter conditions (os_type + cpu_arch) |
| version        | string array   | No       | Version filter conditions                       |
| as_default     | bool array     | No       | Filter by whether it is the default version     |
| enabled        | bool array     | No       | Filter by whether it is enabled                 |
| name           | string array   | No       | Plugin package name filter conditions           |
| file_name      | string array   | No       | Plugin package file name filter conditions      |
| is_visible     | bool array     | No       | Filter by whether it is visible in the frontend |

#### exact_include_conditions.platform[n]

| Parameter Name | Parameter Type | Required | Description                                                                  |
| -------------- | -------------- | -------- | ---------------------------------------------------------------------------- |
| os_type        | string         | Yes      | Operating system type (forms a supported platform combination with cpu_arch) |
| cpu_arch       | string         | Yes      | CPU architecture (forms a supported platform combination with os_type)       |

#### distinct_field

| Parameter Name | Parameter Type | Required | Description                                                |
| -------------- | -------------- | -------- | ---------------------------------------------------------- |
| os_type        | bool           | No       | Whether to deduplicate OS types, default false             |
| cpu_arch       | bool           | No       | Whether to deduplicate CPU architectures, default false    |
| name           | bool           | No       | Whether to deduplicate plugin package names, default false |
| version        | bool           | No       | Whether to deduplicate versions, default false             |

### Request Example

Get all available OS types, CPU architectures, plugin package names, and versions for generation 2 plugin packages.

```json
{
  "generation": 2,
  "distinct_field": {
    "os_type": true,
    "cpu_arch": true,
    "name": true,
    "version": true
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
    "cpu_arch": ["amd64", "arm64"],
    "name": ["bkmonitorbeat"],
    "version": ["3.3.1", "3.4.0"]
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

| Parameter Name | Parameter Type | Description                                                      |
| -------------- | -------------- | ---------------------------------------------------------------- |
| os_type        | string array   | Deduplicated OS type list derived from matching data             |
| cpu_arch       | string array   | Deduplicated CPU architecture list derived from matching data    |
| name           | string array   | Deduplicated plugin package name list derived from matching data |
| version        | string array   | Deduplicated version list derived from matching data             |
