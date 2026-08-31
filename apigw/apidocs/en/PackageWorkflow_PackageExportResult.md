### Description

- API Version: v3.0.1-alpha.78+
- Required Permission: None.
- Function: Query the execution result of a plugin package export workflow by workflow ID, including whether it has finished. When the workflow succeeds, a temporary download URL and its expiration time are returned.

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

### Response Example (Succeeded)

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123457",
  "data": {
    "status": "success",
    "is_finish": true,
    "download_url": "https://bk-nodemgr-file.example.com/api/v3/export/download/origin_plugin_package?token=AQ2hR9pK7vN4mT6xQ8sW0yZaBcDeFgHiJkLmNoPqRsTuVw",
    "download_url_expired_at": 1798761600000
  }
}
```

### Response Example (Running)

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123458",
  "data": {
    "status": "running",
    "is_finish": false
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
| is_finish               | bool   | Yes      | Whether the workflow has finished; `true` for `success`, `failed`, and `partial_failed`                        |
| download_url            | string | No       | Temporary download URL for the exported plugin package; only returned when `status` is `success`              |
| download_url_expired_at | int64  | No       | Expiration timestamp of the temporary download URL (Unix milliseconds); returned together with `download_url` |

**Status values**:

- `running`: The workflow is still running; query again later.
- `success`: The workflow completed successfully; the package can be downloaded from `download_url`.
- `failed`: The workflow failed; no download URL is returned.
- `partial_failed`: Some workflow operations failed; no download URL is returned.

> When `is_finish` is `false`, continue querying the workflow result. `download_url` already contains the download token and can be used directly; it cannot be used after it expires.
