### Description

- API Version: v3.0.1-alpha.85+.
- Required Permission: None.
- Function: Start a one-time workflow to create missing default plugins from enabled plugin releases in the current tenant.

### URL

POST /api/v3/sync/plugin/ensure_default

### Input Parameters

No business parameters. Submit an empty JSON object `{}`. The tenant ID and operator username are taken from the authenticated request context and must both be non-empty; do not pass them in the request body.

### Usage Notes

- No additional IAM action permission is required. API identity authentication is still required.
- The workflow checks for an existing plugin by package name and creates a default plugin only when none exists. It does not deploy or start plugin processes on hosts.
- A successful response means the workflow has been created and activated; default plugin reconciliation may still be running.

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
    "workflow_id": "trig:6c0379fe82b44920a8c49508fb744d72"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code; 0 means the workflow was created and activated successfully |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| workflow_id | string | Trigger ID of the one-time workflow; this is not an execution result |
