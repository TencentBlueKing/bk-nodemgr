### Description

- API Version: v3.0.1-alpha.114+.
- Required Permissions: system_access (System Access), agent_history_view (View Agent History), proxy_history_view (View Proxy History). History permissions are checked against the instance's workflow's stored businesses and node roles: `blank` and `agent` require Agent history permission; `proxy` requires Proxy history permission. Both categories, missing roles, or unknown roles require both permissions.
- Function: Query step states, logs, child workflow references, and extra instance-level execution logs by node operation instance ID.
- The API Gateway authenticates and forwards the user identity. Do not supply username or tenant parameters in the request body.

### URL

POST /api/v3/open/node/workflow/operation/instance/log/get

### Request Parameters

| Parameter    | Type   | Required | Description                                                                        |
| ------------ | ------ | -------- | ---------------------------------------------------------------------------------- |
| oper_inst_id | string | Yes      | Non-empty operation instance ID, obtained from `oper_inst_id` in the instance list |

### Request Example

```json
{
  "oper_inst_id": "inst-001"
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
      "InstallAgent": {
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
        "sub_workflow_refs": []
      }
    },
    "extra_execution_logs": {
      "logs": []
    }
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                        |
| ---------- | ------ | ---------------------------------- |
| code       | int32  | Status code; `0` indicates success |
| message    | string | Request message                    |
| request_id | string | Request ID                         |
| error      | object | Error information                  |
| permission | object | Permission information             |
| data       | object | Response data                      |

#### data

| Parameter            | Type   | Description                                                               |
| -------------------- | ------ | ------------------------------------------------------------------------- |
| oper_inst_logs       | object | Step logs keyed by action name (`action_name`), not by action instance ID |
| extra_execution_logs | object | Extra instance-level execution logs                                       |

#### data.oper_inst_logs.{action_name}

`action_name` is an action name from the operation definition and corresponds to `action_names` in the instance list. Map keys have no ordering semantics; do not infer step execution order from them. This API has no log pagination parameters.

| Parameter         | Type   | Description                                                                 |
| ----------------- | ------ | --------------------------------------------------------------------------- |
| life_cycle        | object | Action lifecycle                                                            |
| message           | object | Action log messages                                                         |
| display_name_zh   | string | Chinese action display name                                                 |
| display_name_en   | string | English action display name                                                 |
| sub_workflow_refs | array  | Child workflow references associated with the action; empty when none exist |

#### data.oper_inst_logs.{action_name}.life_cycle

| Parameter   | Type   | Description                                                                                              |
| ----------- | ------ | -------------------------------------------------------------------------------------------------------- |
| state       | string | Action state, such as `running`, `success`, or `failed`                                                  |
| create_time | int64  | Creation time in Unix milliseconds                                                                       |
| start_time  | int64  | Start time in Unix milliseconds                                                                          |
| end_time    | int64  | End time in Unix milliseconds                                                                            |
| stop_time   | int64  | Not populated in the current step log response; `0` or omitted. Do not use it to determine the stop time |

Creation, start, and end timestamps that have not been recorded are `0`.

#### data.oper_inst_logs.{action_name}.message

| Parameter | Type  | Description        |
| --------- | ----- | ------------------ |
| logs      | array | Action log entries |

#### data.oper_inst_logs.{action_name}.message.logs[n]

| Parameter | Type   | Description                        |
| --------- | ------ | ---------------------------------- |
| time      | int64  | Log timestamp in Unix milliseconds |
| level     | string | Log level                          |
| text_zh   | string | Chinese log content                |
| text_en   | string | English log content                |

#### data.oper_inst_logs.{action_name}.sub_workflow_refs[n]

| Parameter       | Type   | Description           |
| --------------- | ------ | --------------------- |
| workflow_id     | string | Child workflow ID     |
| workflow_domain | string | Child workflow domain |

#### data.extra_execution_logs

| Parameter | Type  | Description                                                                                                                                                 |
| --------- | ----- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| logs      | array | Extra instance-level execution logs. Each entry contains `time`, `level`, `text_zh`, and `text_en`, with the same types and time unit as action log entries |
