### 描述

- 该接口提供版本：v3.0.1-alpha.114+。
- 该接口所需权限：system_access（系统访问）、plugin_history_view（查看插件历史）。
- 该接口功能描述：根据操作实例 ID 查询插件任务流操作实例日志，包括各动作的生命周期、日志消息、子任务流引用及实例级额外执行日志。

### URL

POST /api/v3/open/plugin/workflow/operation/instance/log/get

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述                  |
| ------------ | -------- | ---- | --------------------- |
| oper_inst_id | string   | 是   | 操作实例 ID，不可为空 |

### 调用示例

```json
{
  "oper_inst_id": "inst-plugin-0001"
}
```

### 响应示例

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

### 响应参数说明

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| code       | int32    | 状态码，`0` 表示成功     |
| message    | string   | 请求信息                 |
| request_id | string   | 请求 ID                  |
| error      | object   | 错误信息，成功时通常为空 |
| permission | object   | 权限信息                 |
| data       | object   | 响应数据                 |

#### data

| 参数名称             | 参数类型 | 描述                         |
| -------------------- | -------- | ---------------------------- |
| oper_inst_logs       | object   | 动作名称到动作执行日志的映射 |
| extra_execution_logs | object   | 操作实例级额外执行日志       |

#### data.oper_inst_logs.{action_name}

`action_name` 为动态动作名称，例如 `upgrade_plugin`，对应操作实例列表中的 `action_names`；每个值包含该动作的执行数据。

| 参数名称          | 参数类型 | 描述                                   |
| ----------------- | -------- | -------------------------------------- |
| life_cycle        | object   | 动作生命周期信息                       |
| message           | object   | 动作日志消息                           |
| display_name_zh   | string   | 动作中文显示名                         |
| display_name_en   | string   | 动作英文显示名                         |
| sub_workflow_refs | array    | 关联的子任务流引用；没有引用时为空数组 |

#### data.oper_inst_logs.{action_name}.life_cycle

| 参数名称    | 参数类型 | 描述                                                                                                       |
| ----------- | -------- | ---------------------------------------------------------------------------------------------------------- |
| state       | string   | 动作状态，可选值：`pending`、`running`、`success`、`failed`、`timeout`、`skipped`、`terminated`、`unknown` |
| create_time | int64    | 创建时间，Unix 毫秒时间戳                                                                                  |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳；未记录时为 `0`                                                                  |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳；未记录时为 `0`                                                                  |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳；当前动作日志不提供此字段                                                        |

#### data.oper_inst_logs.{action_name}.message

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| logs     | array    | 动作日志列表 |

#### data.oper_inst_logs.{action_name}.message.logs[n]

| 参数名称 | 参数类型 | 描述                                   |
| -------- | -------- | -------------------------------------- |
| time     | int64    | 日志时间，Unix 毫秒时间戳              |
| level    | string   | 日志级别，例如 `INFO`、`WARN`、`ERROR` |
| text_zh  | string   | 中文日志内容                           |
| text_en  | string   | 英文日志内容                           |

#### data.oper_inst_logs.{action_name}.sub_workflow_refs[n]

| 参数名称        | 参数类型 | 描述                                   |
| --------------- | -------- | -------------------------------------- |
| workflow_id     | string   | 子任务流 ID                            |
| workflow_domain | string   | 子任务流领域，可选值：`node`、`plugin` |

#### data.extra_execution_logs

| 参数名称 | 参数类型 | 描述                       |
| -------- | -------- | -------------------------- |
| logs     | array    | 操作实例级额外执行日志列表 |

#### data.extra_execution_logs.logs[n]

| 参数名称 | 参数类型 | 描述                                   |
| -------- | -------- | -------------------------------------- |
| time     | int64    | 日志时间，Unix 毫秒时间戳              |
| level    | string   | 日志级别，例如 `INFO`、`WARN`、`ERROR` |
| text_zh  | string   | 中文日志内容                           |
| text_en  | string   | 英文日志内容                           |

### 处理规则说明

- 当前用户和租户由网关认证上下文确定，不能通过请求参数指定或切换。
- 根据 `oper_inst_id` 定位操作实例及所属任务流，并校验关联业务的 `plugin_history_view` 权限。
- `oper_inst_logs` 的键为动作名称而非动作实例 ID；没有动作日志时返回空对象 `{}`。对象键没有顺序保证，可按操作实例列表中的 `action_names` 展示动作日志。
- 没有额外执行日志时，`extra_execution_logs.logs` 为空数组。
- 本接口没有分页、日志级别过滤、时间范围过滤或排序参数；不要将示例中的动作名称视为固定响应字段。
