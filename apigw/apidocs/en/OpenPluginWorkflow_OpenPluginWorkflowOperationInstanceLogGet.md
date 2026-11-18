### Description

- API Version: v3.0.1-alpha.114+.
- Required Permission: system_access (System Access), plugin_history_view (View Plugin History).
- Function: Get plugin workflow operation instance logs by operation instance ID, including per-action lifecycle, log messages, sub-workflow references, and extra execution logs at the instance level.

### URL

POST /api/v3/open/plugin/workflow/operation/instance/log/get

### Request Parameters

| Parameter    | Type   | Required | Description                            |
| ------------ | ------ | -------- | -------------------------------------- |
| oper_inst_id | string | Yes      | Operation instance ID; cannot be empty |

### Request Example

```json
{
  "oper_inst_id": "inst-plugin-0001"
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
      "upgrade_plugin": {
        "life_cycle": {
          "state": "success",
          "create_time": 1719820000000,
          "start_time": 1719820005000,
          "end_time": 1719820030000
        },
        "message": {
          "logs": [
            {
              "time": 1719820010000,
              "level": "INFO",
              "text_zh": "开始执行插件升级",
              "text_en": "Start upgrading the plugin"
            }
          ]
        },
        "display_name_zh": "升级插件",
        "display_name_en": "Upgrade Plugin",
        "sub_workflow_refs": []
      }
    },
    "extra_execution_logs": {
      "logs": [
        {
          "time": 1719820040000,
          "level": "WARN",
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
| code       | int32  | Status code; `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

#### data

| Parameter            | Type   | Description                                          |
| -------------------- | ------ | ---------------------------------------------------- |
| oper_inst_logs       | object | Map from action names to action execution logs       |
| extra_execution_logs | object | Extra execution logs at the operation instance level |

#### data.oper_inst_logs.{action_name}

`action_name` is a dynamic action name, such as `upgrade_plugin`, corresponding to `action_names` in the operation instance list. Each value contains execution data for that action.

| Parameter         | Type   | Description                                                         |
| ----------------- | ------ | ------------------------------------------------------------------- |
| life_cycle        | object | Action lifecycle information                                        |
| message           | object | Action log messages                                                 |
| display_name_zh   | string | Chinese display name of the action                                  |
| display_name_en   | string | English display name of the action                                  |
| sub_workflow_refs | array  | Related sub-workflow references; an empty array when there are none |

#### data.oper_inst_logs.{action_name}.life_cycle

| Parameter   | Type   | Description                                                                                            |
| ----------- | ------ | ------------------------------------------------------------------------------------------------------ |
| state       | string | Action state: `pending`, `running`, `success`, `failed`, `timeout`, `skipped`, `terminated`, `unknown` |
| create_time | int64  | Creation time as a Unix timestamp in milliseconds                                                      |
| start_time  | int64  | Start time as a Unix timestamp in milliseconds; `0` when not recorded                                  |
| end_time    | int64  | End time as a Unix timestamp in milliseconds; `0` when not recorded                                    |
| stop_time   | int64  | Stop time as a Unix timestamp in milliseconds; currently not provided in action logs                   |

#### data.oper_inst_logs.{action_name}.message

| Parameter | Type  | Description        |
| --------- | ----- | ------------------ |
| logs      | array | Action log entries |

#### data.oper_inst_logs.{action_name}.message.logs[n]

| Parameter | Type   | Description                                   |
| --------- | ------ | --------------------------------------------- |
| time      | int64  | Log time as a Unix timestamp in milliseconds  |
| level     | string | Log level, such as `INFO`, `WARN`, or `ERROR` |
| text_zh   | string | Chinese log content                           |
| text_en   | string | English log content                           |

#### data.oper_inst_logs.{action_name}.sub_workflow_refs[n]

| Parameter       | Type   | Description                             |
| --------------- | ------ | --------------------------------------- |
| workflow_id     | string | Sub-workflow ID                         |
| workflow_domain | string | Sub-workflow domain: `node` or `plugin` |

#### data.extra_execution_logs

| Parameter | Type  | Description                                                 |
| --------- | ----- | ----------------------------------------------------------- |
| logs      | array | Extra execution log entries at the operation instance level |

#### data.extra_execution_logs.logs[n]

| Parameter | Type   | Description                                   |
| --------- | ------ | --------------------------------------------- |
| time      | int64  | Log time as a Unix timestamp in milliseconds  |
| level     | string | Log level, such as `INFO`, `WARN`, or `ERROR` |
| text_zh   | string | Chinese log content                           |
| text_en   | string | English log content                           |

### Processing Rules

- The current user and tenant come from the gateway authentication context. Request parameters cannot select or switch them.
- The operation instance and its owning workflow are located by `oper_inst_id`, and `plugin_history_view` permission is checked for the associated businesses.
- Keys in `oper_inst_logs` are action names, not action instance IDs. An empty object `{}` is returned when there are no action logs. Object keys have no guaranteed order; use `action_names` from the operation instance list to order action logs for display.
- `extra_execution_logs.logs` is an empty array when there are no extra logs.
- This API has no pagination, log-level filtering, time-range filtering, or sorting parameters. Do not treat the action name in the example as a fixed response field.
