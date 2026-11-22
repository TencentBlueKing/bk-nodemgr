### 描述

- 该接口提供版本：v3.0.1-alpha.116+。
- 该接口所需权限：system_access（系统访问）和 platform_manage（平台管理），两者均须满足。平台管理资源类型为 `platform_resource`，资源 ID 为 `scheduler_workflow`（定时任务调度）。
- 该接口功能描述：查询定时任务调度的操作列表及最近一次执行摘要，支持分页和按父操作实例查询直接子操作。

### URL

POST /api/v3/schedule/workflow/operation/list

### 输入参数

| 参数名称            | 参数类型 | 必选 | 描述                                                                                                      |
| ------------------- | -------- | ---- | --------------------------------------------------------------------------------------------------------- |
| only_count          | bool     | 否   | 是否只返回总数，不返回操作详情，默认为 `false`                                                            |
| page                | object   | 否   | 分页配置；查询详情时须提供有效分页                                                                        |
| workflow_id         | string   | 是   | 定时任务调度工作流 ID，不可为空                                                                           |
| parent_oper_inst_id | string   | 否   | 父操作实例 ID；非空时查询该实例的直接子操作，不递归返回后代；不传或为空时查询该工作流当前触发器关联的操作 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                  |
| -------- | -------- | ---- | ------------------------------------- |
| offset   | int32    | 否   | 分页起始位置，从 `0` 开始，不可为负数 |
| limit    | int32    | 是   | 每页条数，取值范围为 `(0, 500]`       |

### 调用示例

查询定时任务调度工作流的操作记录。查询某次执行直接派发的子操作时，在此基础上指定 `parent_oper_inst_id`。

```json
{
  "only_count": false,
  "page": { "offset": 0, "limit": 20 },
  "workflow_id": "wf-schedule-001"
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
    "operations": [
      {
        "operation_id": "op-schedule-001",
        "instance_ids": ["inst-schedule-001"],
        "operator": "admin",
        "create_time": 1719820000000,
        "parent_operation_id": "",
        "parent_oper_inst_id": "",
        "latest_oper_inst_brief_data": {
          "life_cycle": {
            "state": "success",
            "create_time": 1719820000000,
            "start_time": 1719820001000,
            "end_time": 1719820002000,
            "stop_time": 0
          },
          "latest_action_inst_brief_data": { "name": "dispatch", "tags": [] }
        }
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

| 参数名称   | 参数类型 | 描述                                           |
| ---------- | -------- | ---------------------------------------------- |
| total      | int64    | 符合条件的操作总数                             |
| operations | array    | 当前页的操作列表；`only_count=true` 时为空数组 |

#### data.operations[n]

| 参数名称                    | 参数类型     | 描述                                   |
| --------------------------- | ------------ | -------------------------------------- |
| operation_id                | string       | 操作 ID，可用于查询操作实例            |
| instance_ids                | string array | 该操作关联的执行实例 ID 列表           |
| operator                    | string       | 操作初始化时记录的操作人；未记录时为空 |
| create_time                 | int64        | 操作创建时间，Unix 毫秒时间戳          |
| latest_oper_inst_brief_data | object       | 最近一次执行实例摘要                   |
| parent_operation_id         | string       | 父操作 ID；无父操作时为空              |
| parent_oper_inst_id         | string       | 父操作实例 ID；无父实例时为空          |

#### data.operations[n].latest_oper_inst_brief_data

| 参数名称                      | 参数类型 | 描述                                                                                 |
| ----------------------------- | -------- | ------------------------------------------------------------------------------------ |
| life_cycle                    | object   | 最近一次执行实例的生命周期                                                           |
| latest_action_inst_brief_data | object   | 最近执行步骤摘要，包含 `name`（string，步骤名称）和 `tags`（string array，步骤标签） |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| 参数名称    | 参数类型 | 描述                                                                                  |
| ----------- | -------- | ------------------------------------------------------------------------------------- |
| state       | string   | 实例状态：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated` |
| create_time | int64    | 创建时间，Unix 毫秒时间戳                                                             |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳                                                             |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳                                                             |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳                                                             |

### 处理规则说明

- 租户和用户名来自服务端可信请求上下文，不通过请求体指定；所有查询均受当前租户和平台权限约束。
- 非空的 `parent_oper_inst_id` 必须属于指定工作流的执行链，不能用于查询其他租户或无关任务的操作。
- 根操作通过工作流当前的 `trigger_id` 关联。调度重新初始化时可能更换触发器，因此本接口不是完整历史归档；没有当前触发器时返回空结果。
- 定时任务根操作实例的保留上限为 `100`，清理时会连带清理子任务；已清理记录不再可查询。
- 调度派发实例的 `success` 仅表示派发步骤成功，不代表子工作流成功。可按父实例查询直接子操作，再通过[实例列表](ScheduleWorkflow_ScheduleWorkflowOperationInstanceList.md)和[实例日志](ScheduleWorkflow_ScheduleWorkflowOperationInstanceLogGet.md)查看实际执行结果。
