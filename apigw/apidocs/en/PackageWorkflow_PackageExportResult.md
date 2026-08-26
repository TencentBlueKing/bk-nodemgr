### Description

- API Version: v3.0.1-alpha.74+
- Required Permission: None.
- Function: Query the execution result of a plugin package export workflow by workflow ID. When the workflow succeeds, a temporary download URL and its expiration time are returned.

### URL

POST /api/v3/package/workflow/export_result

### Request Parameters

| Parameter   | Type   | Required | Description                                         |
| ----------- | ------ | -------- | --------------------------------------------------- |
| workflow_id | string | Yes      | Workflow ID returned by the PackageExportPlugin API |

### Request Example

```json
{
  "workflow_id": "workflow-export-abc123def456"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123457",
  "data": {
    "status": "success",
    "error_message": "",
    "download_url": "https://example.com/download/plugin-export-abc123.tgz",
    "download_url_expired_at": 1767225600000
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                          |
| ---------- | ------ | ------------------------------------ |
| code       | int32  | Status code, 0 for success           |
| message    | string | Response message                     |
| request_id | string | Request ID                           |
| error      | object | Error information (empty on success) |
| permission | object | Permission information               |
| data       | object | Response data                        |

#### data

| Parameter               | Type   | Required | Description                                                                                                   |
| ----------------------- | ------ | -------- | ------------------------------------------------------------------------------------------------------------- |
| status                  | string | Yes      | Workflow status (enum: `running`, `success`, `failed`, `partial_failed`)                                      |
| error_message           | string | No       | Error message when the export fails; only returned when `status` is `failed` or `partial_failed`              |
| download_url            | string | No       | Temporary download URL for the exported plugin package; only returned when `status` is `success`              |
| download_url_expired_at | int64  | No       | Expiration timestamp of the temporary download URL (Unix milliseconds); returned together with `download_url` |

**Status values**:

- `running`: The workflow is still running; query again later.
- `success`: The workflow completed successfully; the package can be downloaded from `download_url`.
- `failed`: The workflow failed; check `error_message` for details.
- `partial_failed`: Some workflow operations failed; check `error_message` for details.

> When `status` is `running`, `error_message`, `download_url`, and `download_url_expired_at` are not returned. The download URL cannot be used after it expires.
