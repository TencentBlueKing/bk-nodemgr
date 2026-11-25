### Description

- API Version: v3.0.1-alpha.119+.
- Required Permission: plugin_view (View Plugin).
- Function: List the current tenant's dependency files by metadata with pagination, for file record management.

### URL

POST /api/v3/plugin/requirement/list

### Input Parameters

| Parameter Name | Parameter Type | Required | Description            |
| -------------- | -------------- | -------- | ---------------------- |
| page           | object         | Yes      | Pagination parameters  |
| conditions     | object         | No       | Exact match conditions |

#### page

| Parameter Name | Parameter Type | Required | Description                            |
| -------------- | -------------- | -------- | -------------------------------------- |
| offset         | int64          | No       | Page offset starting from 0, default 0 |
| limit          | int64          | Yes      | Page size, must be greater than 0      |
| sort           | string         | No       | Optional sort expression               |

#### conditions.exact_include

| Parameter Name      | Parameter Type | Required | Description                                       |
| ------------------- | -------------- | -------- | ------------------------------------------------- |
| requirement_file_id | string array   | No       | Exact match by record UUID                        |
| name                | string array   | No       | Exact match by dependency business name           |
| os                  | string array   | No       | Exact match by applicable operating systems       |
| arch                | string array   | No       | Exact match by applicable CPU architectures       |
| version             | string array   | No       | Exact match by applicable plugin package versions |
| plugin_name         | string array   | No       | Exact match by owning plugin                      |
| target_file_name    | string array   | No       | Exact match by delivery filename                  |

**Parameter Notes**:

- `page` is required; `offset` must not be negative and `limit` must be greater than 0.
- Records are filtered by values stored on the record; values within one field use OR and fields use AND.
- Records with an empty applicability scope are not expanded into concrete operating systems, architectures or versions, so they do not match concrete `os`/`arch`/`version` conditions.
- The interface checks the caller's `plugin_view` permission and narrows results to the plugins visible to the caller: only records whose owning plugin is visible are returned; if no plugin is visible within the authorized scope, an empty list with total 0 is returned.

### Request Example

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "conditions": {
    "exact_include": {
      "plugin_name": ["bk-monitor-agent"],
      "os": ["linux"]
    }
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000014",
  "error": null,
  "data": {
    "total": 1,
    "items": [
      {
        "requirement_file_id": "0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11",
        "name": "agent-config-v2",
        "file_name": "9f3c2a5e-1b7d-4e2f-8a6c-0d1e2f3a4b5c",
        "md5": "6f1ed002ab5595859014ebf0951522d9",
        "size": 18,
        "updated_at": 1774339320,
        "operator": "admin",
        "os": ["linux", "windows"],
        "arch": ["amd64"],
        "version": ["1.2.0", "1.3.0"],
        "plugin_name": "bk-monitor-agent",
        "target_file_name": "agent.conf"
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                        |
| -------------- | -------------- | ---------------------------------- |
| code           | int32          | Status code, 0 indicates success   |
| message        | string         | Request message                    |
| request_id     | string         | Request ID                         |
| error          | object         | Error information, null on success |
| data           | object         | Response data                      |

#### data

| Parameter Name | Parameter Type | Description                          |
| -------------- | -------------- | ------------------------------------ |
| total          | int64          | Total records matching conditions    |
| items          | object array   | Dependency file records on this page |

#### data.items[n]

| Parameter Name      | Parameter Type | Description                                             |
| ------------------- | -------------- | ------------------------------------------------------- |
| requirement_file_id | string         | Stable UUID of the dependency file record               |
| name                | string         | Dependency business name, unique per tenant and plugin  |
| file_name           | string         | FileGroup object name in storage (UUID), distinct from the delivery filename                      |
| md5                 | string         | MD5 of the file content                                 |
| size                | int64          | Decoded file size in bytes                              |
| updated_at          | int64          | Last update time, Unix timestamp in seconds             |
| operator            | string         | Last operator                                           |
| os                  | string array   | Applicable operating systems; empty means any           |
| arch                | string array   | Applicable CPU architectures; empty means any           |
| version             | string array   | Applicable plugin package versions; empty means any     |
| plugin_name         | string         | Name of the plugin that owns the dependency file        |
| target_file_name    | string         | Delivery filename on the target host, without directory |

#### error

| Parameter Name | Parameter Type | Description             |
| -------------- | -------------- | ----------------------- |
| system         | string         | Error system identifier |
| message        | string         | Error message           |
| details        | array          | Error detail list       |

#### error.details[n]

| Parameter Name | Parameter Type | Description   |
| -------------- | -------------- | ------------- |
| code           | string         | Error code    |
| message        | string         | Error message |
