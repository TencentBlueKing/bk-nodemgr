### Description

- API Version: v3.0.1-alpha.72+
- Required Permission: None.
- Function: Import a plugin V3 package from a download URL. The interface launches an asynchronous import workflow (download and upload the package, publish the package, update package visibility) and returns a workflow ID, which can be used to query the import result via the PackageImportResult API.

### URL

POST /api/v3/package/workflow/import/v3/plugin

### Request Parameters

| Parameter    | Type   | Required | Description                      |
| ------------ | ------ | -------- | -------------------------------- |
| filename     | string | Yes      | Package file name                |
| download_url | string | Yes      | Package download URL             |
| md5          | string | Yes      | MD5 checksum of the package file |

### Request Example

```json
{
  "filename": "bkmonitorbeat-2.1.3.tgz",
  "download_url": "https://example.com/packages/bkmonitorbeat-2.1.3.tgz",
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
