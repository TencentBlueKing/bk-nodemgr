### Description

- API Version: v3.0.1-alpha.90+.
- Required Permission: None.
- Function: Asynchronously check plugin process status expiry for all hosts in the current tenant and mark expired statuses as `unknown`.

### Behavior

- Using the time when the correction task executes, mark a process as `unknown` if its last synchronization time is more than 48 hours old. Records exactly at the cutoff are excluded.
- Records with a missing, `null`, or zero last synchronization time are also processed.
- Update only the process status. Preserve the last synchronization time, automatic supervision flag, PID, version, Agent ID, and other process information.
- This API corrects stored records based on their synchronization times. It does not query GSE for actual process status or start, stop, or remove processes from supervision.
- A successful response means the asynchronous task has been created and activated, not that all process statuses have been updated.
- The full check processes the current tenant's host inventory in batches, including hosts without an Agent ID.

### URL

POST /api/v3/sync/gse/plugin/process/correct_unknown_status/all

### Input Parameters

None. Send `{}` as the request body.

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
    "trigger_id": "trigger-001"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code; 0 means the task was successfully triggered |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| trigger_id | string | Trigger ID of this asynchronous correction task |
