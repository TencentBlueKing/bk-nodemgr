### 描述

- 该接口提供版本：v3.0.1-alpha.116+。
- 该接口所需权限：system_access（系统访问）和 platform_manage（平台管理），两者均须满足。平台管理资源类型为 `platform_resource`，资源 ID 为 `scheduler_workflow`（定时任务调度）。
- 该接口功能描述：按操作 ID 或操作实例 ID 查询当前租户定时任务执行链中的实例及生命周期，支持仅查询总数。

### URL

POST /api/v3/schedule/workflow/operation/instance/list

### 输入参数

| 参数名称     | 参数类型     | 必选     | 描述                                                                     |
| ------------ | ------------ | -------- | ------------------------------------------------------------------------ |
| only_count   | bool         | 否       | 是否只返回总数，不返回实例详情，默认为 `false`                           |
| operation_id | string array | 条件必选 | 操作 ID 列表；与 `oper_inst_id` 至少提供一类非空列表，每项均不可为空     |
| oper_inst_id | string array | 条件必选 | 操作实例 ID 列表；与 `operation_id` 至少提供一类非空列表，每项均不可为空 |

两类条件同时提供时按 AND 组合，仅返回同时匹配操作 ID 和实例 ID 的记录。本接口不接受 `page` 分页参数。

### 调用示例

```json
{
  "only_count": false,
  "operation_id": ["op-schedule-001"],
  "oper_inst_id": ["inst-schedule-001"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "oper_inst_data": [
      {
        "operation_id": "op-schedule-001",
        "oper_inst_id": "inst-schedule-001",
        "oper_inst_status": "success",
        "operation_def_name": "scheduled_dispatch",
        "parent_operation_id": "",
        "action_names": ["dispatch"],
        "life_cycle": {
          "state": "success",
          "create_time": 1719820000000,
          "start_time": 1719820001000,
          "end_time": 1719820002000,
          "stop_time": 0
        },
        "latest_action_inst_brief_data": { "name": "dispatch", "tags": [] }
      }
    ]
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

| 参数名称       | 参数类型 | 描述                                                                                                        |
| -------------- | -------- | ----------------------------------------------------------------------------------------------------------- |
| total          | int64    | 符合条件的实例总数                                                                                          |
| oper_inst_data | array    | 实例详情列表，使用与节点/插件任务流相同的 `WorflowOperationInstanceData` 结构；`only_count=true` 时为空数组 |

#### data.oper_inst_data[n]

| 参数名称                      | 参数类型     | 描述                                                                                  |
| ----------------------------- | ------------ | ------------------------------------------------------------------------------------- |
| operation_id                  | string       | 所属操作 ID                                                                           |
| oper_inst_id                  | string       | 操作实例 ID，可用于查询实例日志或作为父实例查询直接子操作                             |
| oper_inst_status              | string       | 实例状态：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated` |
| operation_def_name            | string       | 操作定义名称                                                                          |
| parent_operation_id           | string       | 父操作 ID；无父操作时为空                                                             |
| action_names                  | string array | 实例中的步骤名称列表                                                                  |
| life_cycle                    | object       | 实例生命周期                                                                          |
| latest_action_inst_brief_data | object       | 最近执行步骤摘要                                                                      |

#### data.oper_inst_data[n].life_cycle

| 参数名称    | 参数类型 | 描述                                    |
| ----------- | -------- | --------------------------------------- |
| state       | string   | 生命周期状态，取值同 `oper_inst_status` |
| create_time | int64    | 创建时间，Unix 毫秒时间戳               |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳               |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳               |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳               |

#### data.oper_inst_data[n].latest_action_inst_brief_data

| 参数名称 | 参数类型     | 描述             |
| -------- | ------------ | ---------------- |
| name     | string       | 最近执行步骤名称 |
| tags     | string array | 最近执行步骤标签 |

### 处理规则说明

- 租户和用户名来自服务端可信请求上下文，不通过请求体指定。
- 所有传入的操作 ID、实例 ID 及返回的实例都必须属于当前租户的定时任务执行链。子操作通过父操作和父实例关联逐级追溯至当前租户的调度根；不能仅凭已知 ID 查询无关任务或其他租户的数据，计数查询同样受此限制。
- 根操作关联以当前定时任务的 `trigger_id` 为准。调度重新初始化可能更换触发器，旧触发器下的记录不保证可查；本接口不是完整历史归档。根实例保留上限为 `100`，清理时会连带清理子任务。
- 派发实例成功不代表子工作流成功。使用实例 ID 作为[操作列表](ScheduleWorkflow_ScheduleWorkflowOperationList.md)的 `parent_oper_inst_id` 查看直接子操作，并通过[实例日志](ScheduleWorkflow_ScheduleWorkflowOperationInstanceLogGet.md)定位具体步骤。
