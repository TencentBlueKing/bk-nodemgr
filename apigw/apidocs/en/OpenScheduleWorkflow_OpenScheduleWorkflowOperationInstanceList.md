### Description

- API Version: v3.0.1-alpha.116+.
- Required Permissions: Both system_access (System Access) and platform_manage (Platform Management). The platform resource type is `platform_resource`, and the resource ID is `scheduler_workflow` (Scheduled Workflow Management).
- Function: Query instances and lifecycle data in the current tenant's scheduled workflow execution chains by operation IDs or instance IDs as an authenticated user, with count-only support.
- Calls require API Gateway application authorization and user authentication. Username and tenant come from the trusted request context established by gateway authentication, not the request body.

### URL

POST /api/v3/open/schedule/workflow/operation/instance/list

### Request Parameters

| Parameter    | Type         | Required    | Description                                                                                                                             |
| ------------ | ------------ | ----------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| only_count   | bool         | No          | Return only the count without instance details; defaults to `false`                                                                     |
| operation_id | string array | Conditional | Operation IDs; provide at least one nonempty list of `operation_id` or `oper_inst_id`. Every supplied element must be nonempty          |
| oper_inst_id | string array | Conditional | Operation instance IDs; provide at least one nonempty list of `operation_id` or `oper_inst_id`. Every supplied element must be nonempty |

When both lists are supplied, they are combined with AND: a returned instance must match both its operation ID and its instance ID. This API does not accept a `page` pagination parameter.

### Request Example

```json
{
  "only_count": false,
  "operation_id": ["op-schedule-001"],
  "oper_inst_id": ["inst-schedule-001"]
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
    "oper_inst_data": [
      {
        "operation_id": "op-schedule-001",
        "oper_inst_id": "inst-schedule-001",
        "oper_inst_status": "success",
        "operation_def_name": "scheduled_dispatch",
        "parent_operation_id": "",
        "action_names": ["dispatch"],
        "life_cycle": {
          "state": "success",
          "create_time": 1719820000000,
          "start_time": 1719820001000,
          "end_time": 1719820002000,
          "stop_time": 0
        },
        "latest_action_inst_brief_data": { "name": "dispatch", "tags": [] }
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

| Parameter      | Type  | Description                                                                                                                     |
| -------------- | ----- | ------------------------------------------------------------------------------------------------------------------------------- |
| total          | int64 | Total matching instances                                                                                                        |
| oper_inst_data | array | Instance details using the same `WorflowOperationInstanceData` structure as node/plugin workflows; empty when `only_count=true` |

#### data.oper_inst_data[n]

| Parameter                     | Type         | Description                                                                                 |
| ----------------------------- | ------------ | ------------------------------------------------------------------------------------------- |
| operation_id                  | string       | Owning operation ID                                                                         |
| oper_inst_id                  | string       | Instance ID, usable to query logs or direct child operations by parent instance             |
| oper_inst_status              | string       | Instance state: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |
| operation_def_name            | string       | Operation definition name                                                                   |
| parent_operation_id           | string       | Parent operation ID; empty when there is no parent operation                                |
| action_names                  | string array | Action names in the instance                                                                |
| life_cycle                    | object       | Instance lifecycle                                                                          |
| latest_action_inst_brief_data | object       | Latest action summary                                                                       |

#### data.oper_inst_data[n].life_cycle

| Parameter   | Type   | Description                                                |
| ----------- | ------ | ---------------------------------------------------------- |
| state       | string | Lifecycle state; values are the same as `oper_inst_status` |
| create_time | int64  | Creation time as a Unix timestamp in milliseconds          |
| start_time  | int64  | Start time as a Unix timestamp in milliseconds             |
| end_time    | int64  | End time as a Unix timestamp in milliseconds               |
| stop_time   | int64  | Stop time as a Unix timestamp in milliseconds              |

#### data.oper_inst_data[n].latest_action_inst_brief_data

| Parameter | Type         | Description        |
| --------- | ------------ | ------------------ |
| name      | string       | Latest action name |
| tags      | string array | Latest action tags |

### Processing Rules

- Tenant and username come from the trusted request context established by gateway authentication, not the request body.
- Every supplied operation ID, instance ID, and returned instance must belong to a scheduled workflow execution chain of the current tenant. Child operations are traced through their parent operation and parent instance links to a current scheduled root. Knowing an ID does not grant access to unrelated tasks or another tenant's data. Count-only queries have the same restrictions.
- Root association uses current scheduled workflow `trigger_id` values. Scheduler reinitialization may replace a trigger, so records under an older trigger are not guaranteed to remain queryable. This API is not a complete historical archive. Root instances have a retention limit of `100`, and cleanup also removes their child tasks.
- A successful dispatch instance does not imply successful child workflows. Use the instance ID as `parent_oper_inst_id` in the [operation list](OpenScheduleWorkflow_OpenScheduleWorkflowOperationList.md) to query direct child operations, and inspect [instance logs](OpenScheduleWorkflow_OpenScheduleWorkflowOperationInstanceLogGet.md) for action details.
