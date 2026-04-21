### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：agent_operate（操作Agent）、proxy_operate（操作Proxy）。
  - 权限说明：根据任务流涉及的节点角色动态收敛。若任务流仅涉及 Agent 节点，则只需 agent_operate；若仅涉及 Proxy 节点，则只需
    proxy_operate；若涉及两者或角色未知，则需要两者。
- 该接口功能描述：根据操作实例 ID 获取节点任务操作实例日志，返回每个 action 的生命周期与日志内容，以及实例级额外执行日志。

### URL

POST /api/v3/node/workflow/operation/instance/log/get

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述        |
| ------------ | -------- | ---- | ----------- |
| oper_inst_id | string   | 是   | 操作实例 ID |

### 调用示例

查询操作实例 `oi-20260421-0001` 的执行日志。

```json
{
  "oper_inst_id": "oi-20260421-0001"
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

### 响应参数说明

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| code       | int32    | 状态码，`0` 表示成功     |
| message    | string   | 请求信息                 |
| request_id | string   | 请求 ID                  |
| error      | object   | 错误信息，成功时通常为空 |
| permission | object   | 权限信息                 |
| data       | object   | 响应数据                 |

#### error

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| system   | string   | 错误系统标识 |
| message  | string   | 错误消息     |
| details  | array    | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| code     | string   | 错误代码 |
| message  | string   | 错误消息 |

#### permission

| 参数名称    | 参数类型 | 描述         |
| ----------- | -------- | ------------ |
| system      | string   | 权限系统 ID  |
| system_name | string   | 权限系统名称 |
| apply_url   | string   | 权限申请链接 |
| actions     | array    | 关联动作列表 |

#### permission.actions[n]

| 参数名称               | 参数类型 | 描述         |
| ---------------------- | -------- | ------------ |
| id                     | string   | 动作 ID      |
| name                   | string   | 动作名称     |
| related_resource_types | array    | 关联资源类型 |

#### data

| 参数名称             | 参数类型 | 描述               |
| -------------------- | -------- | ------------------ |
| oper_inst_logs       | object   | action 日志映射    |
| extra_execution_logs | object   | 实例级额外执行日志 |

#### data.oper_inst_logs.{action_id}

`action_id` 为 action 标识，值为对应 action 的执行日志对象。

| 参数名称          | 参数类型 | 描述                |
| ----------------- | -------- | ------------------- |
| life_cycle        | object   | action 生命周期信息 |
| message           | object   | action 日志消息     |
| display_name_zh   | string   | action 中文展示名   |
| display_name_en   | string   | action 英文展示名   |
| sub_workflow_refs | array    | 关联子任务流引用    |

#### data.oper_inst_logs.{action_id}.life_cycle

| 参数名称    | 参数类型 | 描述                      |
| ----------- | -------- | ------------------------- |
| state       | string   | action 状态               |
| create_time | int64    | 创建时间，Unix 毫秒时间戳 |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳 |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳 |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳 |

#### data.oper_inst_logs.{action_id}.message

| 参数名称 | 参数类型 | 描述            |
| -------- | -------- | --------------- |
| logs     | array    | action 日志列表 |

#### data.oper_inst_logs.{action_id}.message.logs[n]

| 参数名称 | 参数类型 | 描述                      |
| -------- | -------- | ------------------------- |
| time     | int64    | 日志时间，Unix 毫秒时间戳 |
| level    | string   | 日志级别                  |
| text_zh  | string   | 中文日志内容              |
| text_en  | string   | 英文日志内容              |

#### data.oper_inst_logs.{action_id}.sub_workflow_refs[n]

| 参数名称        | 参数类型 | 描述        |
| --------------- | -------- | ----------- |
| workflow_id     | string   | 子任务流 ID |
| workflow_domain | string   | 子任务流域  |

#### data.extra_execution_logs

| 参数名称 | 参数类型 | 描述                   |
| -------- | -------- | ---------------------- |
| logs     | array    | 实例级额外执行日志列表 |

#### data.extra_execution_logs.logs[n]

| 参数名称 | 参数类型 | 描述                      |
| -------- | -------- | ------------------------- |
| time     | int64    | 日志时间，Unix 毫秒时间戳 |
| level    | string   | 日志级别                  |
| text_zh  | string   | 中文日志内容              |
| text_en  | string   | 英文日志内容              |
