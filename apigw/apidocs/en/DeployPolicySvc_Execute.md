### Description

- API Version: v3.0.1+
- Required Permission: No additional IAM permission; existing authentication and tenant isolation remain.
- Function: Execute a deploy policy asynchronously and return the requested policy's workflow ID for this execution.

### URL

POST /api/v3/deploy_policy/execute

### Request Parameters

| Parameter        | Type  | Required | Description      |
| ---------------- | ----- | -------- | ---------------- |
| deploy_policy_id | int64 | Yes      | Deploy policy ID |

### Request Example

```json
{
  "deploy_policy_id": 1001
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

| Parameter  | Type   | Description                                               |
| ---------- | ------ | --------------------------------------------------------- |
| code       | int32  | Status code, 0 for success                                |
| message    | string | Response message                                          |
| request_id | string | Request ID                                                |
| error      | object | Error information (empty on success)                      |
| permission | object | Permission information (empty without a permission error) |
| data       | object | Response data                                             |

#### data

| Parameter   | Type   | Description                                                                                                               |
| ----------- | ------ | ------------------------------------------------------------------------------------------------------------------------- |
| workflow_id | string | Business workflow ID for this execution of the requested policy; use it with `POST /api/v3/deploy_policy/workflow/result` |

### Execution Semantics and Limits

- Each execute call creates a new execution identity. Repeated or concurrent calls for the same policy do not share a workflow ID.
- Discovery may execute related policies. Each participating policy has its own execution record; this endpoint returns only the requested policy's workflow ID.
- Policies in the same execution may share node/plugin child workflows. Strict sequencing between related policies is not guaranteed.
- Automatic retries retain this execution identity and include every newly created child workflow. There is no idempotency key or comprehensive duplicate-launch prevention guarantee.
- Invalid parameters, missing or disabled policies, and launch failures remain API errors. A successful response confirms launch, not child completion or machine convergence.
- Poll the [workflow result endpoint](DeployPolicySvc_WorkflowResult.md) and decide when to execute the next policy. Node Manager does not trigger the next policy automatically.
