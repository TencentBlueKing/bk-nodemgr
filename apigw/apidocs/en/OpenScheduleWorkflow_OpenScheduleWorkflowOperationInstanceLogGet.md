### Description

- API Version: v3.0.1-alpha.116+.
- Required Permissions: Both system_access (System Access) and platform_manage (Platform Management). The platform resource type is `platform_resource`, and the resource ID is `scheduler_workflow` (Scheduled Workflow Management).
- Function: Get action lifecycles, bilingual logs, and sub-workflow references for one operation instance in the current tenant's scheduled workflow execution chains as an authenticated user.
- Calls require API Gateway application authorization and user authentication. Username and tenant come from the trusted request context established by gateway authentication, not the request body.

### URL

POST /api/v3/open/schedule/workflow/operation/instance/log/get

### Request Parameters

| Parameter    | Type   | Required | Description                            |
| ------------ | ------ | -------- | -------------------------------------- |
| oper_inst_id | string | Yes      | Operation instance ID; cannot be empty |

### Request Example

```json
{
  "oper_inst_id": "inst-schedule-001"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "oper_inst_logs": {
      "action-dispatch-001": {
        "life_cycle": {
          "state": "success",
          "create_time": 1719820000000,
          "start_time": 1719820001000,
          "end_time": 1719820002000,
          "stop_time": 0
        },
        "message": {
          "logs": [
            {
              "time": 1719820002000,
              "level": "info",
              "text_zh": "派发完成",
              "text_en": "Dispatch completed"
            }
          ]
        },
        "display_name_zh": "派发任务",
        "display_name_en": "Dispatch tasks",
        "sub_workflow_refs": []
      }
    },
    "extra_execution_logs": { "logs": [] }
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

| Parameter            | Type   | Description                                                                     |
| -------------------- | ------ | ------------------------------------------------------------------------------- |
| oper_inst_logs       | object | Map of action IDs to `WorkflowActionData` log objects                           |
| extra_execution_logs | object | Instance-level extra execution logs using the `WorkflowActionMessage` structure |

#### data.oper_inst_logs.{action_id}

| Parameter         | Type   | Description                                                                 |
| ----------------- | ------ | --------------------------------------------------------------------------- |
| life_cycle        | object | Action lifecycle                                                            |
| message           | object | Bilingual action messages containing a `logs` array                         |
| display_name_zh   | string | Chinese action display name                                                 |
| display_name_en   | string | English action display name                                                 |
| sub_workflow_refs | array  | Allowlisted sub-workflow references; empty when no references are available |

#### data.oper_inst_logs.{action_id}.life_cycle

| Parameter   | Type   | Description                                       |
| ----------- | ------ | ------------------------------------------------- |
| state       | string | Action state                                      |
| create_time | int64  | Creation time as a Unix timestamp in milliseconds |
| start_time  | int64  | Start time as a Unix timestamp in milliseconds    |
| end_time    | int64  | End time as a Unix timestamp in milliseconds      |
| stop_time   | int64  | Stop time as a Unix timestamp in milliseconds     |

#### data.oper_inst_logs.{action_id}.message.logs[n]

| Parameter | Type   | Description                                  |
| --------- | ------ | -------------------------------------------- |
| time      | int64  | Log time as a Unix timestamp in milliseconds |
| level     | string | Log level                                    |
| text_zh   | string | Chinese log content                          |
| text_en   | string | English log content                          |

#### data.oper_inst_logs.{action_id}.sub_workflow_refs[n]

| Parameter       | Type   | Description         |
| --------------- | ------ | ------------------- |
| workflow_id     | string | Sub-workflow ID     |
| workflow_domain | string | Sub-workflow domain |

#### data.extra_execution_logs

| Parameter | Type  | Description                                                                                                                               |
| --------- | ----- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| logs      | array | Extra execution logs; each entry has the same structure as action `message.logs[n]`, containing `time`, `level`, `text_zh`, and `text_en` |

### Processing Rules

- Tenant and username come from the trusted request context established by gateway authentication, not the request body. The instance must be associated through its operation and parent operation/instance chain with a scheduled root of the current tenant. An instance ID cannot bypass tenant isolation or select an unrelated task.
- Association uses current scheduled workflow `trigger_id` values. Scheduler reinitialization may replace a trigger, so this API is not a complete historical archive. Root instances have a retention limit of `100`; cleanup also removes child tasks and their logs.
- This API returns action state, bilingual logs, and allowlisted `sub_workflow_refs` only. It does not return raw `init_content`, `content`, or private data.
- `sub_workflow_refs` are navigation references, not proof of child task success or authorization to access the referenced workflows. Successful dispatch does not imply successful child workflows. Use the current instance ID as `parent_oper_inst_id` in the [operation list](OpenScheduleWorkflow_OpenScheduleWorkflowOperationList.md) to query direct child operations, then inspect their [instance state](OpenScheduleWorkflow_OpenScheduleWorkflowOperationInstanceList.md).
