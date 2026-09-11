### Description

- API Version: v3.0.1-alpha.82+
- Required Permission: No additional IAM permission; existing authentication and tenant isolation remain.
- Function: Query the current result, dispatch status, and associated node/plugin child workflows of one deploy policy execution.

### URL

POST /api/v3/deploy_policy/workflow/result

### Request Parameters

| Parameter   | Type   | Required | Description                                                                                                            |
| ----------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------------------- |
| workflow_id | string | Yes      | Non-empty deploy policy workflow ID returned by execute; not a policy ID, trigger ID, or node/plugin child workflow ID |

### Request Example

```json
{
  "workflow_id": "workflow-abc123def456"
}
```

### Response Example

Dispatch has completed normally; one of the two children is still running:

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "workflow_id": "workflow-abc123def456",
    "deploy_policy_id": 1001,
    "status": "running",
    "total_count": 2,
    "finished_count": 1,
    "dispatch_status": "success",
    "dispatch_error": "",
    "children": [
      {
        "type": "node",
        "workflow_id": "node-workflow-example",
        "status": "success",
        "missing": false
      },
      {
        "type": "plugin",
        "workflow_id": "plugin-workflow-example",
        "status": "running",
        "missing": false
      }
    ],
    "unknown_reason": ""
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                          |
| ---------- | ------ | -------------------------------------------------------------------- |
| code       | int32  | API status code; 0 means the query succeeded, not deployment success |
| message    | string | API message, not a business status indicator                         |
| request_id | string | Request ID                                                           |
| error      | object | API error information; empty on success                              |
| permission | object | Permission information; empty without a permission error             |
| data       | object | Current aggregate result when the query succeeds                     |

#### data

| Parameter        | Type         | Description                                                                                                                                                            |
| ---------------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| workflow_id      | string       | Deploy policy workflow ID being queried                                                                                                                                |
| deploy_policy_id | int64        | Policy associated with this execution record                                                                                                                           |
| status           | string       | `running`, `success`, `failed`, `partial_failed`, or `unknown`; see the rules below                                                                                    |
| total_count      | int64        | Known actually created child workflows; equals the length of `children`. Includes previously created but missing records, excludes unconfirmed creation intents        |
| finished_count   | int64        | Children currently `success`, `failed`, or `partial_failed`; missing children are excluded                                                                             |
| dispatch_status  | string       | `running`: dispatch or automatic retries remain open; `success`: normally closed; `failed`: finally failed and closed; `unknown`: dispatch closure cannot be confirmed |
| dispatch_error   | string       | Latest dispatch error, empty when none. May coexist with `running` during retries; not a child workflow error or API error                                             |
| children         | object array | Complete associated set, without pagination or permission-filtered subsets; `[]` when there are no children                                                            |
| unknown_reason   | string       | Non-empty explanation when `status=unknown`, empty otherwise. Not a stable error code; do not parse it for control flow                                                |

#### children[n]

| Parameter   | Type   | Description                                                                                           |
| ----------- | ------ | ----------------------------------------------------------------------------------------------------- |
| type        | string | `node` or `plugin`                                                                                    |
| workflow_id | string | Child business workflow ID; deduplicated by `(type, workflow_id)` within the result                   |
| status      | string | Current child status: `running`, `success`, `failed`, `partial_failed`, or `unknown` when unavailable |
| missing     | bool   | Whether a previously created child record is missing; true implies `status=unknown`                   |

### Aggregation Rules

Apply the following rules in order. `dispatch_status=failed` does not bypass waiting for already created children.

| Condition                                                                                              | data.status      | Currently finished |
| ------------------------------------------------------------------------------------------------------ | ---------------- | ------------------ |
| Missing child record, unconfirmed creation, uncertain dispatch closure, or other incomplete evidence   | `unknown`        | Cannot determine   |
| Dispatch/automatic retries remain open, or any child is unfinished                                     | `running`        | No                 |
| Dispatch completed normally with no changes needed and zero children                                   | `success`        | Yes                |
| Dispatch completed normally and every child succeeded                                                  | `success`        | Yes                |
| Dispatch finally failed and every actually created child has finished, including zero children         | `failed`         | Yes                |
| Dispatch completed normally and every child is failed                                                  | `failed`         | Yes                |
| Dispatch completed normally, every child has finished, and results are mixed or include partial_failed | `partial_failed` | Yes                |

- `total_count` may grow during dispatch and is not a complete final total when evidence is incomplete. `total_count=finished_count`, even when both are 0, does not replace checking `status`.
- Every query rereads associations and child business statuses. The parent result is not frozen. Retrying a child under the same ID can make a later query return `running` again.
- `dispatch_status=success` confirms dispatch closure, not child success. A non-empty `dispatch_error` is not a substitute for the aggregate status either.
- Multiple policy execution records may share a child workflow. Summing their counts may count a child more than once; a policy may wait for other policies' tasks within a shared child.
- Host, operation, and log details are not included. Use the corresponding node/plugin workflow endpoints for further inspection.

### Errors and Caller Responsibilities

- Empty `workflow_id`, no matching deploy policy execution record in the current tenant, or data read failures produce API errors, not fabricated empty success results.
- If the parent record exists but a child is missing or dispatch cannot be confirmed, the query succeeds with `status=unknown` and `unknown_reason`. Unconfirmed initial creation is not counted as an actual child.
- A previously created but missing child's ID remains in `children` with `missing=true` and is excluded from `finished_count`.
- Timeout or crash does not guarantee dispatch has stopped. The first version does not promise automatic recovery for every `unknown` result.
- Callers choose polling intervals, deadlines, and business continuation rules. `unknown` means neither confirmed completion nor confirmed running; do not use it to advance to the next policy.
- Currently finished does not mean entirely successful. Node Manager neither triggers the next policy nor retracts subsequent actions already initiated by the caller.
- Results are not a transactionally consistent snapshot across workflows or a continuous health check.

### Related APIs

- [Execute a deploy policy](DeployPolicySvc_Execute.md)
