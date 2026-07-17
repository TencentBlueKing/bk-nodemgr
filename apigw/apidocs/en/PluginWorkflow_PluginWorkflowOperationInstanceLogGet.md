### Description

- API Version: v3.0.1-alpha.1+.
- Version Changes: `v3.0.1-alpha.13+` added instance-level extra execution logs and the permission response field; `v3.0.1-alpha.19+` added `plugin_operate` authorization.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Get plugin workflow operation instance logs by operation instance ID, including per-action lifecycle, log messages, sub-workflow references, and extra execution logs at the instance level.

### URL

POST /api/v3/plugin/workflow/operation/instance/log/get

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| oper_inst_id | string | Yes | Operation instance ID; cannot be empty |

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
      "act_upgrade_plugin": {
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
              "level": "info",
              "text_zh": "开始执行插件升级",
              "text_en": "Start upgrading the plugin"
            }
          ]
        },
        "display_name_zh": "升级插件",
        "display_name_en": "Upgrade Plugin",
        "sub_workflow_refs": [
          {
            "workflow_id": "wf-sub-0001",
            "workflow_domain": "plugin"
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

| Parameter | Type | Description |
| --- | --- | --- |
| code | int32 | Status code; `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter | Type | Description |
| --- | --- | --- |
| oper_inst_logs | object | Map from action ID to action execution logs |
| extra_execution_logs | object | Extra execution logs at the operation instance level |

#### data.oper_inst_logs.{action_id}

`action_id` is a dynamic action identifier, and each value contains execution data for that action.

| Parameter | Type | Description |
| --- | --- | --- |
| life_cycle | object | Action lifecycle information |
| message | object | Action log messages |
| display_name_zh | string | Chinese display name of the action |
| display_name_en | string | English display name of the action |
| sub_workflow_refs | array | Related sub-workflow references |

#### data.oper_inst_logs.{action_id}.life_cycle

| Parameter | Type | Description |
| --- | --- | --- |
| state | string | Action state |
| create_time | int64 | Creation time as a Unix timestamp in milliseconds |
| start_time | int64 | Start time as a Unix timestamp in milliseconds |
| end_time | int64 | End time as a Unix timestamp in milliseconds |
| stop_time | int64 | Stop time as a Unix timestamp in milliseconds; declared in the protocol but not populated by the current backend converter |

#### data.oper_inst_logs.{action_id}.message

| Parameter | Type | Description |
| --- | --- | --- |
| logs | array | Action log entries |

#### data.oper_inst_logs.{action_id}.message.logs[n]

| Parameter | Type | Description |
| --- | --- | --- |
| time | int64 | Log time as a Unix timestamp in milliseconds |
| level | string | Log level |
| text_zh | string | Chinese log content |
| text_en | string | English log content |

#### data.oper_inst_logs.{action_id}.sub_workflow_refs[n]

| Parameter | Type | Description |
| --- | --- | --- |
| workflow_id | string | Sub-workflow ID |
| workflow_domain | string | Sub-workflow domain |

#### data.extra_execution_logs

| Parameter | Type | Description |
| --- | --- | --- |
| logs | array | Extra execution log entries at the operation instance level |

#### data.extra_execution_logs.logs[n]

| Parameter | Type | Description |
| --- | --- | --- |
| time | int64 | Log time as a Unix timestamp in milliseconds |
| level | string | Log level |
| text_zh | string | Chinese log content |
| text_en | string | English log content |

### Processing Rules

- The backend first locates the operation instance and its owning workflow by `oper_inst_id`, then checks `plugin_operate` permission.
- Keys in `oper_inst_logs` are dynamic action IDs; an empty object is returned when there are no action logs.
- `extra_execution_logs.logs` is an empty array when there are no extra logs.
- `life_cycle.stop_time` for an action is declared in the shared Proto but is not set by the current response converter, so it is omitted from the success example.
