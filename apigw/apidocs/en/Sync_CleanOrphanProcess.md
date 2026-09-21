### Description

- API Version: v3.0.1-alpha.86+.
- Required Permission: None.
- Function: Trigger a one-off workflow that cleans the orphan process records whose host no longer exists.

### URL

POST /api/v3/sync/process/clean_orphan

### Input Parameters

No business parameters; submit an empty JSON object `{}`. The tenant ID and operator username are taken from the authenticated request context, must both be non-empty, and are not passed in the request body.

### Usage Notes

- No extra IAM action permission is required; API authentication still applies.
- Candidates must have status `unknown` and a last process-info sync time older than 48 hours. Missing or null sync times are also treated as stale. No trusteeship filter is applied. The cleanup task first checks local host records, then queries CMDB only for hosts whose local records are soft deleted or missing. A process is deleted only when its host is also absent from CMDB. Processes are retained if the host exists locally or in CMDB. If either query fails, the current cleanup task fails without deleting any process records.
- The generator loads the matching candidates and creates cleanup tasks in batches of at most 500 processes, then activates the trigger after all tasks have been created. The tasks check host existence and delete orphan records individually through the existing process deletion path; they never ask GSE to stop or uninstall a process.
- A successful response only means the workflow was created and activated, not that the cleanup finished; check the workflow execution record for the result.

### Request Example

```json
{}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "trigger_id": "trig:6c0379fe82b44920a8c49508fb744d72"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 means the workflow was created and activated |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| trigger_id | string | Trigger ID of the one-off workflow, not the execution result |
