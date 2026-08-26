### Description

- API Version: v3.0.1-alpha.74+
- Required Permission: None.
- API Name: PackageExportPlugin.
- Function: Start an asynchronous plugin package export workflow by the name and version of a published plugin package. The workflow packages the target plugin into a downloadable artifact and returns a workflow ID. Use the PackageExportResult API to query the export progress and the temporary download URL.

### URL

POST /api/v3/package/workflow/export/plugin

### Request Parameters

| Parameter          | Type   | Required | Description                   |
| ------------------ | ------ | -------- | ----------------------------- |
| plugin_pkg_name    | string | Yes      | Target plugin package name    |
| plugin_pkg_version | string | Yes      | Target plugin package version |

### Request Example

```json
{
  "plugin_pkg_name": "bkmonitorbeat",
  "plugin_pkg_version": "2.1.3"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "workflow_id": "workflow-export-abc123def456"
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

| Parameter   | Type   | Description                                                          |
| ----------- | ------ | -------------------------------------------------------------------- |
| workflow_id | string | Export workflow ID, used to query the result via PackageExportResult |
