### Description

- API Version: v3.0.1-alpha.84+
- Required Permission: None. Existing authentication and tenant isolation remain; no additional IAM action is required.
- Function: List deploy policy executions with pagination, exact filters, and an operation time range. Parent status describes dispatch, not child deployment results.

### URL

POST /api/v3/deploy_policy/workflow/list

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| page | object | No | Required with a valid value when `only_count=false` or omitted |
| only_count | bool | No | Return only the matching count; `items` is empty |
| exact_include_conditions | object | No | Exact include filters; values within one field are alternatives, different fields are combined |
| operate_time_range | object | No | Filter operation start time using Unix timestamps in seconds |

#### page

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| offset | int32 | No | Start position, starting from `0` |
| limit | int32 | Yes | Records per page, in `(0, 500]` |

#### exact_include_conditions

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| workflow_id | string array | No | Parent workflow IDs returned by execute |
| deploy_policy_id | int64 array | No | Policy IDs, including executions discovered through related policies or schedules |
| status | string array | No | `running`, `success`, `failed`, `partial_failed` |
| operator | string array | No | Operators |

#### operate_time_range

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| start_timestamp_sec | int64 | No | Start timestamp in seconds |
| end_timestamp_sec | int64 | No | End timestamp in seconds |

### Request Example

```json
{
  "page": {"offset": 0, "limit": 20},
  "only_count": false,
  "exact_include_conditions": {
    "deploy_policy_id": [1001],
    "status": ["running", "success"]
  },
  "operate_time_range": {
    "start_timestamp_sec": 1717171200,
    "end_timestamp_sec": 1719859599
  }
}
```

To poll one execution, use `"exact_include_conditions": {"workflow_id": ["workflow-example"]}` without a status filter. To count matching records, set `only_count=true`; `page` can be omitted.

### Response Example

The parent dispatch succeeded. The linked plugin workflow may still be running.

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      {
        "workflow_id": "workflow-example",
        "trigger_id": "trigger-example",
        "deploy_policy_id": 1001,
        "operator": "admin",
        "operate_time": 1717171200000,
        "finish_time": 1717171205000,
        "status": "success",
        "children": [{"type": "plugin", "workflow_id": "plugin-workflow-example"}]
      }
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| code | int32 | API status code; `0` means the query succeeded, not deployment success |
| message | string | API message |
| request_id | string | Request ID |
| error | object | API error information; empty on success |
| permission | object | Permission information; empty without a permission error |
| data | object | Matching parent execution records |

#### data

| Parameter | Type | Description |
| --- | --- | --- |
| total | int64 | Matching parent count before pagination, not a child count |
| items | object array | Parent records; empty for count-only queries or no matches |

#### items[n]

| Parameter | Type | Description |
| --- | --- | --- |
| workflow_id | string | This policy execution's parent workflow ID |
| trigger_id | string | Trigger association; not an alias for `workflow_id` and not returned by execute |
| deploy_policy_id | int64 | Policy associated with the execution |
| operator | string | Operator |
| operate_time | int64 | Unix timestamp in milliseconds |
| finish_time | int64 | Unix timestamp in milliseconds; `0` while unfinished |
| status | string | `running`, `success`, `failed`, or `partial_failed`; describes only the parent's dispatch operation |
| children | object array | Successfully launched and linked child workflows; no child status or progress aggregation |

#### children[n]

| Parameter | Type | Description |
| --- | --- | --- |
| type | string | `node` or `plugin` |
| workflow_id | string | Child workflow ID; identify a child by `(type, workflow_id)` |

### Status and Limits

| Parent operation state | status | Meaning |
| --- | --- | --- |
| Unfinished | `running` | Dispatch has not finished |
| Successful | `success` | Dispatch finished successfully, including when no changes were needed |
| Failed or timed out | `failed` | Dispatch failed; already launched children may still run |
| Terminated | `partial_failed` | Dispatch ended without normal completion; not a count of failed children |

- The parent tracks one dispatch operation. It never waits for or aggregates child status, and child retry does not reopen the parent.
- Status is synchronized in the background, not recalculated from children during list queries. Legacy records without status follow the same synchronization. A missing operation is allowed a one-minute grace period before an unfinished record becomes `failed`; an already terminal record keeps its result.
- `children` contains only successful launches whose links were recorded. Missing or unreadable child records do not change the parent status. Failed dispatch can leave already launched children running.
- Use existing [node](NodeWorkflow_NodeWorkflowList.md) or [plugin](PluginWorkflow_PluginWorkflowList.md) workflow APIs for child results. Their authentication and IAM permissions are unchanged; a parent record does not grant child access.
- No matching parent in the authenticated tenant returns an empty list, not proof of successful completion. Invalid requests and data read failures return API errors.
- There is no automatic dispatch action retry and no deploy-policy workflow retry, terminate, distinct, or statistics API. A new execute request creates a new execution and is not an idempotent retry.
- Callers choose polling deadlines and continuation rules. Parent success is neither child completion nor a machine-health guarantee.

### Related APIs

- [Execute a deploy policy](DeployPolicySvc_Execute.md)
