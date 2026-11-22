### 描述

- 该接口提供版本：v3.0.1-alpha.116+。
- 该接口所需权限：system_access（系统访问）和 platform_manage（平台管理），两者均须满足。平台管理资源类型为 `platform_resource`，资源 ID 为 `scheduler_workflow`（定时任务调度）。
- 该接口功能描述：查询当前租户定时任务执行链中一个操作实例的步骤生命周期、双语日志和子工作流引用。

### URL

POST /api/v3/schedule/workflow/operation/instance/log/get

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述                  |
| ------------ | -------- | ---- | --------------------- |
| oper_inst_id | string   | 是   | 操作实例 ID，不可为空 |

### 调用示例

```json
{
  "oper_inst_id": "inst-schedule-001"
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

| 参数名称             | 参数类型 | 描述                                                  |
| -------------------- | -------- | ----------------------------------------------------- |
| oper_inst_logs       | object   | 以步骤 ID 为键、`WorkflowActionData` 为值的日志映射   |
| extra_execution_logs | object   | 实例级额外执行日志，使用 `WorkflowActionMessage` 结构 |

#### data.oper_inst_logs.{action_id}

| 参数名称          | 参数类型 | 描述                                       |
| ----------------- | -------- | ------------------------------------------ |
| life_cycle        | object   | 步骤生命周期                               |
| message           | object   | 步骤双语日志，包含 `logs` 数组             |
| display_name_zh   | string   | 步骤中文展示名称                           |
| display_name_en   | string   | 步骤英文展示名称                           |
| sub_workflow_refs | array    | 允许公开的子工作流引用；没有引用时为空数组 |

#### data.oper_inst_logs.{action_id}.life_cycle

| 参数名称    | 参数类型 | 描述                      |
| ----------- | -------- | ------------------------- |
| state       | string   | 步骤状态                  |
| create_time | int64    | 创建时间，Unix 毫秒时间戳 |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳 |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳 |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳 |

#### data.oper_inst_logs.{action_id}.message.logs[n]

| 参数名称 | 参数类型 | 描述                      |
| -------- | -------- | ------------------------- |
| time     | int64    | 日志时间，Unix 毫秒时间戳 |
| level    | string   | 日志级别                  |
| text_zh  | string   | 中文日志内容              |
| text_en  | string   | 英文日志内容              |

#### data.oper_inst_logs.{action_id}.sub_workflow_refs[n]

| 参数名称        | 参数类型 | 描述             |
| --------------- | -------- | ---------------- |
| workflow_id     | string   | 子工作流 ID      |
| workflow_domain | string   | 子工作流所属领域 |

#### data.extra_execution_logs

| 参数名称 | 参数类型 | 描述                                                                                                        |
| -------- | -------- | ----------------------------------------------------------------------------------------------------------- |
| logs     | array    | 实例级额外执行日志，每项结构与步骤日志的 `message.logs[n]` 相同，包含 `time`、`level`、`text_zh`、`text_en` |

### 处理规则说明

- 租户和用户名来自服务端可信请求上下文，不通过请求体指定。实例必须通过所属操作及父操作/父实例链关联至当前租户的定时任务调度根，不能通过实例 ID 跨租户或查询无关任务。
- 关联以当前定时任务的 `trigger_id` 为准。调度重新初始化可能更换触发器，本接口不是完整历史归档；根实例保留上限为 `100`，连带清理的子任务及其日志也不再可查。
- 本接口仅返回步骤状态、双语日志和白名单中的 `sub_workflow_refs`；不返回原始 `init_content`、`content` 或私有数据。
- `sub_workflow_refs` 是导航引用，不是子任务的成功证明，也不授予访问目标工作流的权限。调度派发步骤成功不代表子工作流成功；可使用当前实例 ID 作为[操作列表](ScheduleWorkflow_ScheduleWorkflowOperationList.md)的 `parent_oper_inst_id` 查询直接子操作，再查询其[实例状态](ScheduleWorkflow_ScheduleWorkflowOperationInstanceList.md)。
