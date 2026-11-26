### Description

- API Version: v3.0.1-alpha.121+.
- Required Permission: plugin_view (View Plugin).
- Function: Download the dependency file content by UUID; a successful response is the binary file content.

### URL

GET /api/v3/plugin/requirement/download

### Input Parameters

| Parameter Name      | Parameter Type | Required | Description                                                     |
| ------------------- | -------------- | -------- | --------------------------------------------------------------- |
| requirement_file_id | string         | Yes      | UUID of the dependency file record, passed as a query parameter |

**Parameter Notes**:

- The interface checks the `plugin_view` permission for the plugin that owns the record.
- `requirement_file_id` must be a valid UUID, otherwise request validation fails.
- On success the binary file content is returned; on failure a standard JSON error response is returned.

### Request Example

```text
GET /api/v3/plugin/requirement/download?requirement_file_id=0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11
```

### Response Example

On success the response status is 200 and the body is the binary file content (not wrapped in JSON):

```text
<binary file content>
```

On failure a standard JSON error response is returned:

```json
{
  "code": 500,
  "message": "failed to download plugin requirement",
  "request_id": "req-20260324-000015",
  "error": {
    "system": "backend",
    "message": "failed to get requirement file",
    "details": []
  }
}
```

### Response Parameters

- Success: HTTP 200 with the binary file content as the body, not wrapped in JSON.
- Failure: a standard error response with the following fields.

| Parameter Name | Parameter Type | Description                        |
| -------------- | -------------- | ---------------------------------- |
| code           | int32          | Status code, 0 indicates success   |
| message        | string         | Request message                    |
| request_id     | string         | Request ID                         |
| error          | object         | Error information, null on success |

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
