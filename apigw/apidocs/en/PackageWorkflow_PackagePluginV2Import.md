### Description

- API Version: v3.0.1-alpha.74+
- Required Permission: None.
- Function: Import an official plugin V2 package from a download URL. The interface launches an asynchronous import workflow (download and upload the V2 package, publish the package, update package visibility) and returns a workflow ID, which can be used to query the import result via the PackageImportResult API.

### URL

POST /api/v3/package/workflow/import/v2/plugin

### Request Parameters

| Parameter    | Type   | Required | Description                                |
| ------------ | ------ | -------- | ------------------------------------------ |
| filename     | string | Yes      | Plugin V2 package file name                |
| download_url | string | Yes      | Plugin V2 package download URL             |
| md5          | string | Yes      | MD5 checksum of the plugin V2 package file |

### Request Example

```json
{
  "filename": "gse_agent-2.0.0.tgz",
  "download_url": "https://example.com/packages/gse_agent-2.0.0.tgz",
  "md5": "5d41402abc4b2a76b9719d911017c592"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "workflow_id": "workflow-abc123def456"
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
| data       | object | Response data                        |

#### data

| Parameter   | Type   | Description                                                                         |
| ----------- | ------ | ----------------------------------------------------------------------------------- |
| workflow_id | string | Workflow ID, can be used to query the import result via the PackageImportResult API |

> Whether the import is finally complete needs to be determined by reading the workflow status via the PackageImportResult API; through the `workflow_id`, the execution of the `package_plugin_v2_import` operation can be observed in the response `data.operations[]`, which contains the execution logs of four actions: `package_import_plugin_v2_pkg_fetch_and_upload` / `package_publish_plugin_v2_pkg` / `package_release_plugin_enable` / `package_release_plugin_hidden`.
