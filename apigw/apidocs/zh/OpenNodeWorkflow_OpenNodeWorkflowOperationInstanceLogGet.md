### 描述

- 该接口提供版本：v3.0.1-alpha.114+。
- 该接口所需权限：system_access（系统访问）、agent_history_view（查看 Agent 历史）、proxy_history_view（查看 Proxy 历史）。历史权限按实例所属任务流记录的业务和节点角色校验：`blank`、`agent` 需要 Agent 历史权限，`proxy` 需要 Proxy 历史权限；涉及两者、角色缺失或未知时需要两者。
- 该接口功能描述：按节点操作实例 ID 查询各步骤的状态、日志和子任务流引用，以及实例级额外执行日志。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。

### URL

POST /api/v3/open/node/workflow/operation/instance/log/get

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述                                                |
| ------------ | -------- | ---- | --------------------------------------------------- |
| oper_inst_id | string   | 是   | 非空操作实例 ID，可从实例列表的 `oper_inst_id` 获取 |

### 调用示例

```json
{
  "oper_inst_id": "inst-001"
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

### 响应参数说明

| 参数名称   | 参数类型 | 描述                 |
| ---------- | -------- | -------------------- |
| code       | int32    | 状态码，`0` 表示成功 |
| message    | string   | 请求信息             |
| request_id | string   | 请求 ID              |
| error      | object   | 错误信息             |
| permission | object   | 权限信息             |
| data       | object   | 响应数据             |

#### data

| 参数名称             | 参数类型 | 描述                                                                  |
| -------------------- | -------- | --------------------------------------------------------------------- |
| oper_inst_logs       | object   | 以动作名称 `action_name` 为动态键的步骤日志映射，不是动作实例 ID 映射 |
| extra_execution_logs | object   | 实例级额外执行日志                                                    |

#### data.oper_inst_logs.{action_name}

`action_name` 是操作定义中的动作名称，与实例列表的 `action_names` 对应。映射键没有顺序含义，不能据此推断步骤执行顺序。接口没有日志分页参数。

| 参数名称          | 参数类型 | 描述                                     |
| ----------------- | -------- | ---------------------------------------- |
| life_cycle        | object   | 动作生命周期                             |
| message           | object   | 动作日志消息                             |
| display_name_zh   | string   | 动作中文展示名                           |
| display_name_en   | string   | 动作英文展示名                           |
| sub_workflow_refs | array    | 此动作关联的子任务流引用；没有引用时为空 |

#### data.oper_inst_logs.{action_name}.life_cycle

| 参数名称    | 参数类型 | 描述                                                              |
| ----------- | -------- | ----------------------------------------------------------------- |
| state       | string   | 动作状态，例如 `running`、`success`、`failed`                     |
| create_time | int64    | 创建时间，Unix 毫秒时间戳                                         |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳                                         |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳                                         |
| stop_time   | int64    | 当前步骤日志响应未填充此字段，为 `0` 或省略；不能据此判断停止时间 |

尚未记录的创建、开始和结束时间为 `0`。

#### data.oper_inst_logs.{action_name}.message

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| logs     | array    | 动作日志列表 |

#### data.oper_inst_logs.{action_name}.message.logs[n]

| 参数名称 | 参数类型 | 描述                      |
| -------- | -------- | ------------------------- |
| time     | int64    | 日志时间，Unix 毫秒时间戳 |
| level    | string   | 日志级别                  |
| text_zh  | string   | 中文日志内容              |
| text_en  | string   | 英文日志内容              |

#### data.oper_inst_logs.{action_name}.sub_workflow_refs[n]

| 参数名称        | 参数类型 | 描述           |
| --------------- | -------- | -------------- |
| workflow_id     | string   | 子任务流 ID    |
| workflow_domain | string   | 子任务流所属域 |

#### data.extra_execution_logs

| 参数名称 | 参数类型 | 描述                                                                                                 |
| -------- | -------- | ---------------------------------------------------------------------------------------------------- |
| logs     | array    | 实例级额外执行日志列表；每条记录包含 `time`、`level`、`text_zh`、`text_en`，类型和时间单位同动作日志 |
