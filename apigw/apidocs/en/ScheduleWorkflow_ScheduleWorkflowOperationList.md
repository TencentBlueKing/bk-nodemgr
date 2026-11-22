### Description

- API Version: v3.0.1-alpha.116+.
- Required Permissions: Both system_access (System Access) and platform_manage (Platform Management). The platform resource type is `platform_resource`, and the resource ID is `scheduler_workflow` (Scheduled Workflow Management).
- Function: Query scheduled workflow operations and their latest execution summaries, with pagination and direct-child lookup by parent operation instance.

### URL

POST /api/v3/schedule/workflow/operation/list

### Request Parameters

| Parameter           | Type   | Required | Description                                                                                                                                                                                  |
| ------------------- | ------ | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| only_count          | bool   | No       | Return only the count without operation details; defaults to `false`                                                                                                                         |
| page                | object | No       | Pagination; valid pagination is required when requesting details                                                                                                                             |
| workflow_id         | string | Yes      | Scheduled workflow ID; cannot be empty                                                                                                                                                       |
| parent_oper_inst_id | string | No       | Parent operation instance ID; when nonempty, return its direct child operations, not all descendants. Omit or leave empty to query operations associated with the workflow's current trigger |

#### page

| Parameter | Type  | Required | Description                           |
| --------- | ----- | -------- | ------------------------------------- |
| offset    | int32 | No       | Zero-based offset; cannot be negative |
| limit     | int32 | Yes      | Page size in `(0, 500]`               |

### Request Example

Query operations of a scheduled workflow. Add `parent_oper_inst_id` to query child operations dispatched directly by a particular execution.

```json
{
  "only_count": false,
  "page": { "offset": 0, "limit": 20 },
  "workflow_id": "wf-schedule-001"
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
    "operations": [
      {
        "operation_id": "op-schedule-001",
        "instance_ids": ["inst-schedule-001"],
        "operator": "admin",
        "create_time": 1719820000000,
        "parent_operation_id": "",
        "parent_oper_inst_id": "",
        "latest_oper_inst_brief_data": {
          "life_cycle": {
            "state": "success",
            "create_time": 1719820000000,
            "start_time": 1719820001000,
            "end_time": 1719820002000,
            "stop_time": 0
          },
          "latest_action_inst_brief_data": { "name": "dispatch", "tags": [] }
        }
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

| Parameter  | Type  | Description                                                  |
| ---------- | ----- | ------------------------------------------------------------ |
| total      | int64 | Total matching operations                                    |
| operations | array | Operations on the current page; empty when `only_count=true` |

#### data.operations[n]

| Parameter                   | Type         | Description                                                                   |
| --------------------------- | ------------ | ----------------------------------------------------------------------------- |
| operation_id                | string       | Operation ID, usable to query operation instances                             |
| instance_ids                | string array | Execution instance IDs associated with this operation                         |
| operator                    | string       | Operator recorded when the operation was initialized; empty when not recorded |
| create_time                 | int64        | Operation creation time as a Unix timestamp in milliseconds                   |
| latest_oper_inst_brief_data | object       | Latest execution instance summary                                             |
| parent_operation_id         | string       | Parent operation ID; empty when there is no parent operation                  |
| parent_oper_inst_id         | string       | Parent operation instance ID; empty when there is no parent instance          |

#### data.operations[n].latest_oper_inst_brief_data

| Parameter                     | Type   | Description                                                                                           |
| ----------------------------- | ------ | ----------------------------------------------------------------------------------------------------- |
| life_cycle                    | object | Lifecycle of the latest execution instance                                                            |
| latest_action_inst_brief_data | object | Latest action summary, containing `name` (string, action name) and `tags` (string array, action tags) |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| Parameter   | Type   | Description                                                                                 |
| ----------- | ------ | ------------------------------------------------------------------------------------------- |
| state       | string | Instance state: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |
| create_time | int64  | Creation time as a Unix timestamp in milliseconds                                           |
| start_time  | int64  | Start time as a Unix timestamp in milliseconds                                              |
| end_time    | int64  | End time as a Unix timestamp in milliseconds                                                |
| stop_time   | int64  | Stop time as a Unix timestamp in milliseconds                                               |

### Processing Rules

- Tenant and username come from the server's trusted request context, not the request body. Every query is scoped to the current tenant and platform permissions.
- A nonempty `parent_oper_inst_id` must belong to the specified workflow's execution chain; it cannot select operations of another tenant or unrelated task.
- Root operations are associated through the workflow's current `trigger_id`. Scheduler reinitialization may replace the trigger, so this API is not a complete historical archive. A workflow without a current trigger returns no results.
- Scheduled root operation instances have a retention limit of `100`; cleanup also removes their child tasks. Cleaned-up records are no longer queryable.
- A scheduler dispatch instance reporting `success` means only that dispatch succeeded, not that its child workflows succeeded. Query direct child operations by parent instance, then use the [instance list](ScheduleWorkflow_ScheduleWorkflowOperationInstanceList.md) and [instance logs](ScheduleWorkflow_ScheduleWorkflowOperationInstanceLogGet.md) to inspect execution results.
