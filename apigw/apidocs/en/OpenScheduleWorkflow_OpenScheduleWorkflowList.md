### Description

- API Version: v3.0.1-alpha.116+.
- Required Permissions: Both system_access (System Access) and platform_manage (Platform Management). The platform resource type is `platform_resource`, and the resource ID is `scheduler_workflow` (Scheduled Workflow Management).
- Function: Query scheduled workflows in the current tenant as an authenticated user with pagination, exact conditions, and operation time filtering, including enabled state and scheduling interval.
- Calls require API Gateway application authorization and user authentication. Username and tenant come from the trusted request context established by gateway authentication, not the request body.

### URL

POST /api/v3/open/schedule/workflow/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                                                              |
| ------------------------ | ------ | -------- | -------------------------------------------------------------------------------------------------------- |
| page                     | object | No       | Pagination; valid pagination is required when requesting details                                         |
| only_count               | bool   | No       | Return only the count without workflow details; defaults to `false`                                      |
| exact_include_conditions | object | No       | Exact-match inclusion conditions                                                                         |
| fuzzy_include_conditions | object | No       | No fields are currently supported; omit or pass `{}`                                                     |
| operate_time_range       | object | No       | Workflow operation time range using Unix timestamps in seconds; applies to detail and count-only queries |

#### page

| Parameter | Type  | Required | Description                           |
| --------- | ----- | -------- | ------------------------------------- |
| offset    | int32 | No       | Zero-based offset; cannot be negative |
| limit     | int32 | Yes      | Page size in `(0, 500]`               |

#### exact_include_conditions

| Parameter     | Type         | Required | Description    |
| ------------- | ------------ | -------- | -------------- |
| workflow_id   | string array | No       | Workflow IDs   |
| workflow_name | string array | No       | Workflow names |
| operator      | string array | No       | Operators      |

#### operate_time_range

| Parameter           | Type  | Required | Description                                                                                                            |
| ------------------- | ----- | -------- | ---------------------------------------------------------------------------------------------------------------------- |
| start_timestamp_sec | int64 | No       | Nonnegative start Unix timestamp in seconds; defaults to `0` when omitted                                              |
| end_timestamp_sec   | int64 | No       | Nonnegative end Unix timestamp in seconds; defaults to `0` when omitted and must be greater than or equal to the start |

Omitting `operate_time_range` disables operation time filtering. When supplied, the range is validated and applied to the query, including both boundaries. Omitted bounds are not open-ended: supplying a positive start without an end fails validation because the end defaults to `0`.

### Request Example

```json
{
  "page": { "offset": 0, "limit": 20 },
  "only_count": false,
  "exact_include_conditions": { "workflow_id": ["wf-schedule-001"] },
  "operate_time_range": {
    "start_timestamp_sec": 1717171200,
    "end_timestamp_sec": 1719859599
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      {
        "workflow_id": "wf-schedule-001",
        "workflow_name": "scheduled_task",
        "trigger_id": "trigger-schedule-001",
        "operator": "admin",
        "operate_time": 1719820000000,
        "enabled": true,
        "interval": "1m"
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                 |
| ---------- | ------ | ------------------------------------------- |
| code       | int32  | Status code; `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

#### data

| Parameter | Type  | Description                                                 |
| --------- | ----- | ----------------------------------------------------------- |
| total     | int64 | Total matching workflows                                    |
| items     | array | Workflows on the current page; empty when `only_count=true` |

#### data.items[n]

| Parameter     | Type   | Description                                                                                      |
| ------------- | ------ | ------------------------------------------------------------------------------------------------ |
| workflow_id   | string | Scheduled workflow ID                                                                            |
| workflow_name | string | Workflow name                                                                                    |
| trigger_id    | string | Currently associated schedule trigger ID                                                         |
| operator      | string | Operator                                                                                         |
| operate_time  | int64  | Operation time as a Unix timestamp in milliseconds, unlike the seconds used in the request range |
| enabled       | bool   | Whether periodic scheduling is enabled; does not indicate the latest execution succeeded         |
| interval      | string | Scheduling interval string, such as `1m`                                                         |

### Processing Rules

- Tenant and username come from the trusted request context established by gateway authentication. The request body does not accept tenant or username selectors.
- This API returns scheduling configuration summaries, not raw initialization content, execution content, or private data.
- Use a returned `workflow_id` to query the [operation list](OpenScheduleWorkflow_OpenScheduleWorkflowOperationList.md), then retrieve instances and logs.
- Operation history is associated through the current `trigger_id`. Scheduler reinitialization may replace a trigger, so execution queries are not a complete historical archive. Root operation instances have a retention limit of `100`, and cleanup also removes their child tasks.
- Enabled scheduling, successful dispatch, and successful child workflows are different concepts. Inspect child operation instance states to determine actual task results.
