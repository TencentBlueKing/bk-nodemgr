### 描述

- 该接口提供版本：v3.0.1-alpha.116+。
- 该接口所需权限：system_access（系统访问）和 platform_manage（平台管理），两者均须满足。平台管理资源类型为 `platform_resource`，资源 ID 为 `scheduler_workflow`（定时任务调度）。
- 该接口功能描述：以用户身份查询当前租户的定时任务调度工作流，支持分页、精确条件和操作时间范围过滤，返回启用状态和调度间隔。
- 调用须具备 API 网关应用授权并通过用户身份认证；用户名和租户来自网关认证后的可信请求上下文，不通过请求体指定。

### URL

POST /api/v3/open/schedule/workflow/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                                                             |
| ------------------------ | -------- | ---- | ---------------------------------------------------------------- |
| page                     | object   | 否   | 分页配置；查询详情时须提供有效分页                               |
| only_count               | bool     | 否   | 是否只返回总数，不返回工作流详情，默认为 `false`                 |
| exact_include_conditions | object   | 否   | 精确匹配包含条件                                                 |
| fuzzy_include_conditions | object   | 否   | 当前无可用字段，可不传或传 `{}`                                  |
| operate_time_range       | object   | 否   | 工作流操作时间范围，使用 Unix 秒时间戳；适用于详情查询和计数查询 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                  |
| -------- | -------- | ---- | ------------------------------------- |
| offset   | int32    | 否   | 分页起始位置，从 `0` 开始，不可为负数 |
| limit    | int32    | 是   | 每页条数，取值范围为 `(0, 500]`       |

#### exact_include_conditions

| 参数名称      | 参数类型     | 必选 | 描述           |
| ------------- | ------------ | ---- | -------------- |
| workflow_id   | string array | 否   | 工作流 ID 列表 |
| workflow_name | string array | 否   | 工作流名称列表 |
| operator      | string array | 否   | 操作人列表     |

#### operate_time_range

| 参数名称            | 参数类型 | 必选 | 描述                                                                 |
| ------------------- | -------- | ---- | -------------------------------------------------------------------- |
| start_timestamp_sec | int64    | 否   | 起始 Unix 秒时间戳，不可为负数，省略时为 `0`                         |
| end_timestamp_sec   | int64    | 否   | 结束 Unix 秒时间戳，不可为负数，省略时为 `0`，且须大于或等于起始时间 |

不传 `operate_time_range` 时不按操作时间过滤；传入时会校验并实际应用该范围，包含起止边界。省略边界不表示无限范围：例如只传正数起始时间而省略结束时间，会因结束值为 `0` 而校验失败。

### 调用示例

```json
{
  "page": { "offset": 0, "limit": 20 },
  "only_count": false,
  "exact_include_conditions": { "workflow_id": ["wf-schedule-001"] },
  "operate_time_range": {
    "start_timestamp_sec": 1717171200,
    "end_timestamp_sec": 1719859599
  }
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
    "items": [
      {
        "workflow_id": "wf-schedule-001",
        "workflow_name": "scheduled_task",
        "trigger_id": "trigger-schedule-001",
        "operator": "admin",
        "operate_time": 1719820000000,
        "enabled": true,
        "interval": "1m"
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

| 参数名称 | 参数类型 | 描述                                             |
| -------- | -------- | ------------------------------------------------ |
| total    | int64    | 符合条件的工作流总数                             |
| items    | array    | 当前页的工作流列表；`only_count=true` 时为空数组 |

#### data.items[n]

| 参数名称      | 参数类型 | 描述                                                |
| ------------- | -------- | --------------------------------------------------- |
| workflow_id   | string   | 定时任务调度工作流 ID                               |
| workflow_name | string   | 工作流名称                                          |
| trigger_id    | string   | 当前关联的调度触发器 ID                             |
| operator      | string   | 操作人                                              |
| operate_time  | int64    | 操作时间，Unix 毫秒时间戳；与请求范围中的秒单位不同 |
| enabled       | bool     | 是否启用周期调度，不代表最近一次任务执行成功        |
| interval      | string   | 调度间隔字符串，例如 `1m`                           |

### 处理规则说明

- 租户和用户名来自网关认证后的可信请求上下文，请求体不接受租户或用户名选择参数。
- 本接口返回调度配置摘要，不返回原始初始化内容、执行内容或私有数据。
- 使用返回的 `workflow_id` 查询[操作列表](OpenScheduleWorkflow_OpenScheduleWorkflowOperationList.md)，再查询执行实例和日志。
- 操作历史关联以当前 `trigger_id` 为准；调度重新初始化可能更换触发器，因此执行查询不是完整历史归档。根操作实例的保留上限为 `100`，清理时会连带清理子任务。
- 调度启用状态、派发步骤成功和子工作流成功是不同概念，需结合子操作实例状态判断实际任务结果。
