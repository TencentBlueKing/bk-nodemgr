### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 版本变更：`v3.0.1-alpha.13+` 增加实例级额外执行日志与权限响应字段；`v3.0.1-alpha.19+` 增加 `plugin_operate` 鉴权。
- 该接口所需权限：plugin_operate（操作插件）。
- 该接口功能描述：根据操作实例 ID 获取插件任务流操作实例日志，包括各 action 的生命周期、日志内容、子任务流引用及实例级额外执行日志。

### URL

POST /api/v3/plugin/workflow/operation/instance/log/get

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| oper_inst_id | string | 是 | 操作实例 ID，不可为空 |

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

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| oper_inst_logs | object | action ID 到 action 执行日志的映射 |
| extra_execution_logs | object | 操作实例级额外执行日志 |

#### data.oper_inst_logs.{action_id}

`action_id` 为动态的 action 标识，值为对应 action 的执行数据。

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| life_cycle | object | action 生命周期信息 |
| message | object | action 日志消息 |
| display_name_zh | string | action 中文展示名 |
| display_name_en | string | action 英文展示名 |
| sub_workflow_refs | array | 关联子任务流引用 |

#### data.oper_inst_logs.{action_id}.life_cycle

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| state | string | action 状态 |
| create_time | int64 | 创建时间，Unix 毫秒时间戳 |
| start_time | int64 | 开始时间，Unix 毫秒时间戳 |
| end_time | int64 | 结束时间，Unix 毫秒时间戳 |
| stop_time | int64 | 停止时间，Unix 毫秒时间戳；协议已声明，但当前后端转换器不填充 |

#### data.oper_inst_logs.{action_id}.message

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| logs | array | action 日志列表 |

#### data.oper_inst_logs.{action_id}.message.logs[n]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| time | int64 | 日志时间，Unix 毫秒时间戳 |
| level | string | 日志级别 |
| text_zh | string | 中文日志内容 |
| text_en | string | 英文日志内容 |

#### data.oper_inst_logs.{action_id}.sub_workflow_refs[n]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| workflow_id | string | 子任务流 ID |
| workflow_domain | string | 子任务流域 |

#### data.extra_execution_logs

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| logs | array | 操作实例级额外执行日志列表 |

#### data.extra_execution_logs.logs[n]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| time | int64 | 日志时间，Unix 毫秒时间戳 |
| level | string | 日志级别 |
| text_zh | string | 中文日志内容 |
| text_en | string | 英文日志内容 |

### 处理规则说明

- 后端先根据 `oper_inst_id` 定位操作实例及所属任务流，再校验 `plugin_operate` 权限。
- `oper_inst_logs` 的键为动态 action ID；无 action 日志时返回空对象。
- `extra_execution_logs.logs` 无额外日志时返回空数组。
- action 的 `life_cycle.stop_time` 已在共享 Proto 中声明，但当前响应转换器未设置该字段，因此成功示例不包含它。
