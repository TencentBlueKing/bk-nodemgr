### Description

- API Version: v3.0.1-alpha.121+.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Replace the dependency file's name and content, preserving the record ID, owning plugin and applicability scope.

### URL

POST /api/v3/plugin/requirement/update

### Input Parameters

| Parameter Name      | Parameter Type | Required | Description                                        |
| ------------------- | -------------- | -------- | -------------------------------------------------- |
| requirement_file_id | string         | Yes      | UUID of the dependency file record                 |
| name                | string         | Yes      | Replacement dependency business name               |
| content_base64      | string         | Yes      | Replacement file content, standard Base64 encoding |
| target_file_name    | string         | Yes      | Replacement delivery filename                      |

**Parameter Notes**:

- The interface checks the `plugin_operate` permission for the plugin that owns the record.
- `requirement_file_id` must be a valid UUID.
- `name` must not be empty or contain control characters; it is unique per tenant and plugin.
- `target_file_name` must not be empty, must not be `.` or `..`, must not contain characters such as `/ \ : < > " | ? *`, must not end with a space or `.`, and must not be a reserved device name.
- This interface only replaces the name, content and delivery filename; the owning plugin and applicability scope remain unchanged. Use update_scope to adjust the scope.

### Request Example

```json
{
  "requirement_file_id": "0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11",
  "name": "agent-config-v2",
  "content_base64": "IyBhZ2VudCBjb25maWcgdjIK",
  "target_file_name": "agent.conf"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000011",
  "error": null,
  "data": {
    "file": {
      "requirement_file_id": "0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11",
      "name": "agent-config-v2",
      "file_name": "9f3c2a5e-1b7d-4e2f-8a6c-0d1e2f3a4b5c",
      "md5": "6f1ed002ab5595859014ebf0951522d9",
      "size": 18,
      "updated_at": 1774339260,
      "operator": "admin",
      "os": ["linux"],
      "arch": ["amd64", "arm64"],
      "version": [],
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
