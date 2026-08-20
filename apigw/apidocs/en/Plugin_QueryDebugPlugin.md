### Description

- API version: v3.0.1-alpha.72+.
- Required permission: ⚠️ To be confirmed (handler not yet implemented).
- Description: Queries the aggregated status and execution results of a plugin debug workflow, including the life cycle and execution logs of each operation instance.

### URL

POST /api/v3/plugin/query_debug

### Request Parameters

| Parameter   | Type   | Required | Description                                                 |
| ----------- | ------ | -------- | ----------------------------------------------------------- |
| workflow_id | string | Yes      | Debug workflow ID to query, returned by the start debug API |

### Request Example

```json
{
  "workflow_id": "workflow-abc123def456"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "status": "running",
    "is_finish": false,
    "operations": [
      {
        "operation_id": "op-1001",
        "last_instance_id": "inst-5001",
        "oper_inst_logs": {
          "inst-5001": {
            "life_cycle": {
              "state": "running",
              "create_time": 1735000000,
              "start_time": 1735000010
            },
            "message": {
              "logs": [
                {
                  "time": 1735000010,
                  "level": "info",
                  "text_zh": "开始安装插件",
                  "text_en": "start installing plugin"
                }
              ]
            },
            "display_name_zh": "安装插件",
            "display_name_en": "Install Plugin"
          }
        },
        "extra_execution_logs": {
          "logs": []
        }
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                  |
| ---------- | ------ | ---------------------------- |
| code       | int32  | Status code, 0 means success |
| message    | string | Request message              |
| request_id | string | Request ID                   |
| error      | object | Error info, null on success  |
| data       | object | Response data                |

#### data

| Parameter  | Type   | Description                                                            |
| ---------- | ------ | ---------------------------------------------------------------------- |
| status     | string | Debug workflow status (enum: running, success, failed, partial_failed) |
| is_finish  | bool   | Whether the debug workflow has finished                                |
| operations | array  | List of operations in the debug workflow                               |

**status enum**:

- `running`: The workflow is running
- `success`: The workflow succeeded
- `failed`: The workflow failed
- `partial_failed`: Part of the workflow operations failed

#### data.operations[n]

| Parameter            | Type   | Description                                                 |
| -------------------- | ------ | ----------------------------------------------------------- |
| operation_id         | string | Operation ID                                                |
| last_instance_id     | string | ID of the latest executed operation instance                |
| oper_inst_logs       | object | Map from instance ID to operation data (WorkflowActionData) |
| extra_execution_logs | object | Extra execution logs (WorkflowActionMessage)                |

#### data.operations[n].oper_inst_logs.<key>

| Parameter         | Type   | Description                           |
| ----------------- | ------ | ------------------------------------- |
| life_cycle        | object | Life cycle info                       |
| message           | object | Operation log messages                |
| display_name_zh   | string | Chinese display name of the operation |
| display_name_en   | string | English display name of the operation |
| sub_workflow_refs | array  | List of sub-workflow references       |

#### data.operations[n].oper_inst_logs.<key>.life_cycle

| Parameter   | Type   | Description                  |
| ----------- | ------ | ---------------------------- |
| state       | string | Life cycle state             |
| create_time | int64  | Create time (Unix timestamp) |
| start_time  | int64  | Start time (Unix timestamp)  |
| end_time    | int64  | End time (Unix timestamp)    |
| stop_time   | int64  | Stop time (Unix timestamp)   |

#### data.operations[n].oper_inst_logs.<key>.message

| Parameter | Type  | Description          |
| --------- | ----- | -------------------- |
| logs      | array | List of log messages |

#### data.operations[n].oper_inst_logs.<key>.message.logs[n]

| Parameter | Type   | Description               |
| --------- | ------ | ------------------------- |
| time      | int64  | Log time (Unix timestamp) |
| level     | string | Log level                 |
| text_zh   | string | Chinese log content       |
| text_en   | string | English log content       |

#### error

| Parameter | Type   | Description             |
| --------- | ------ | ----------------------- |
| system    | string | Error system identifier |
| message   | string | Error message           |
| details   | array  | Error detail list       |

#### error.details[n]

| Parameter | Type   | Description   |
| --------- | ------ | ------------- |
| code      | string | Error code    |
| message   | string | Error message |
