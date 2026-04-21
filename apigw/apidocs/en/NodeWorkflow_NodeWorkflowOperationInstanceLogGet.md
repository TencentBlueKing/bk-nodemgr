### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: agent_operate (Operate Agent), proxy_operate (Operate Proxy).
  - Permission Note: Dynamically narrowed based on node roles in the workflow. If the workflow only involves Agent nodes, only agent_operate is required; if only Proxy nodes, only proxy_operate is required; if both or unknown roles, both are required.
- Function: Get node workflow operation instance logs by operation instance ID, including per-action lifecycle and log messages, and extra execution logs at instance level.

### URL

POST /api/v3/node/workflow/operation/instance/log/get

### Request Parameters

| Parameter    | Type   | Required | Description           |
| ------------ | ------ | -------- | --------------------- |
| oper_inst_id | string | Yes      | Operation instance ID |

### Request Example

Query logs for operation instance `oi-20260421-0001`.

```json
{
  "oper_inst_id": "oi-20260421-0001"
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
      "act_install_agent": {
        "life_cycle": {
          "state": "success",
          "create_time": 1719820000000,
          "start_time": 1719820005000,
          "end_time": 1719820030000,
          "stop_time": 0
        },
        "message": {
          "logs": [
            {
              "time": 1719820010000,
              "level": "info",
              "text_zh": "开始执行安装任务",
              "text_en": "Start install task"
            }
          ]
        },
        "display_name_zh": "安装 Agent",
        "display_name_en": "Install Agent",
        "sub_workflow_refs": [
          {
            "workflow_id": "wf-sub-001",
            "workflow_domain": "node"
          }
        ]
      }
    },
    "extra_execution_logs": {
      "logs": [
        {
          "time": 1719820040000,
          "level": "warning",
          "text_zh": "检测到重试操作",
          "text_en": "Retry operation detected"
        }
      ]
    }
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                 |
| ---------- | ------ | ------------------------------------------- |
| code       | int32  | Status code, `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

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

#### permission

| Parameter   | Type   | Description            |
| ----------- | ------ | ---------------------- |
| system      | string | Permission system ID   |
| system_name | string | Permission system name |
| apply_url   | string | Permission apply URL   |
| actions     | array  | Related action list    |

#### permission.actions[n]

| Parameter              | Type   | Description            |
| ---------------------- | ------ | ---------------------- |
| id                     | string | Action ID              |
| name                   | string | Action name            |
| related_resource_types | array  | Related resource types |

#### data

| Parameter            | Type   | Description                      |
| -------------------- | ------ | -------------------------------- |
| oper_inst_logs       | object | Action log map                   |
| extra_execution_logs | object | Extra execution logs of instance |

#### data.oper_inst_logs.{action_id}

`action_id` is the action identifier, and each value is a log object for that action.

| Parameter         | Type   | Description                     |
| ----------------- | ------ | ------------------------------- |
| life_cycle        | object | Action lifecycle information    |
| message           | object | Action log messages             |
| display_name_zh   | string | Chinese display name of action  |
| display_name_en   | string | English display name of action  |
| sub_workflow_refs | array  | Related sub-workflow references |

#### data.oper_inst_logs.{action_id}.life_cycle

| Parameter   | Type   | Description                                 |
| ----------- | ------ | ------------------------------------------- |
| state       | string | Action status                               |
| create_time | int64  | Create time, Unix timestamp in milliseconds |
| start_time  | int64  | Start time, Unix timestamp in milliseconds  |
| end_time    | int64  | End time, Unix timestamp in milliseconds    |
| stop_time   | int64  | Stop time, Unix timestamp in milliseconds   |

#### data.oper_inst_logs.{action_id}.message

| Parameter | Type  | Description        |
| --------- | ----- | ------------------ |
| logs      | array | Action log entries |

#### data.oper_inst_logs.{action_id}.message.logs[n]

| Parameter | Type   | Description                              |
| --------- | ------ | ---------------------------------------- |
| time      | int64  | Log time, Unix timestamp in milliseconds |
| level     | string | Log level                                |
| text_zh   | string | Chinese log content                      |
| text_en   | string | English log content                      |

#### data.oper_inst_logs.{action_id}.sub_workflow_refs[n]

| Parameter       | Type   | Description         |
| --------------- | ------ | ------------------- |
| workflow_id     | string | Sub-workflow ID     |
| workflow_domain | string | Sub-workflow domain |

#### data.extra_execution_logs

| Parameter | Type  | Description                             |
| --------- | ----- | --------------------------------------- |
| logs      | array | Extra execution log entries of instance |

#### data.extra_execution_logs.logs[n]

| Parameter | Type   | Description                              |
| --------- | ------ | ---------------------------------------- |
| time      | int64  | Log time, Unix timestamp in milliseconds |
| level     | string | Log level                                |
| text_zh   | string | Chinese log content                      |
| text_en   | string | English log content                      |
