### Description

- API Version: v3.0.1-alpha.119+.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Replace the dependency file's applicability scope entirely, preserving the record ID and content.

### URL

POST /api/v3/plugin/requirement/update_scope

### Input Parameters

| Parameter Name      | Parameter Type | Required | Description                                          |
| ------------------- | -------------- | -------- | ---------------------------------------------------- |
| requirement_file_id | string         | Yes      | UUID of the dependency file record                   |
| os                  | string array   | No       | Replacement operating systems; empty means any       |
| arch                | string array   | No       | Replacement CPU architectures; empty means any       |
| version             | string array   | No       | Replacement plugin package versions; empty means any |

**Parameter Notes**:

- The interface checks the `plugin_operate` permission for the plugin that owns the record.
- `requirement_file_id` must be a valid UUID.
- This interface replaces the entire applicability scope; the record ID and file content remain unchanged.
- Dimensions are unrestricted when empty or omitted; values within one dimension use OR and dimensions use AND; values must be concrete (not `unknown`, an empty string or `all`).

### Request Example

```json
{
  "requirement_file_id": "0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11",
  "os": ["linux", "windows"],
  "arch": ["amd64"],
  "version": ["1.2.0", "1.3.0"]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000012",
  "error": null,
  "data": {
    "file": {
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

| Parameter Name | Parameter Type | Description            |
| -------------- | -------------- | ---------------------- |
| file           | object         | Dependency file record |

#### data.file

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
