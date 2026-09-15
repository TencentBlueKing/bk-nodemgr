### 描述

- 该接口提供版本：v3.0.1-alpha.84+
- 该接口所需权限：无。保留现有身份认证和租户隔离，不增加 IAM action。
- 该接口功能描述：分页查询部署策略执行记录，支持精确条件和操作时间范围过滤。父流程状态表示 dispatch 状态，不表示子流程部署结果。

### URL

POST /api/v3/deploy_policy/workflow/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| page | object | 否 | `only_count=false` 或未提供时，必须传入有效分页参数 |
| only_count | bool | 否 | 只返回匹配记录数，`items` 为空 |
| exact_include_conditions | object | 否 | 精确包含条件；同字段多个值取并集，不同字段组合匹配 |
| operate_time_range | object | 否 | 按操作开始时间过滤，Unix 时间戳单位为秒 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| offset | int32 | 否 | 分页起始位置，从 `0` 开始 |
| limit | int32 | 是 | 每页记录数，取值范围 `(0, 500]` |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| workflow_id | string array | 否 | execute 返回的父 workflow ID 列表 |
| deploy_policy_id | int64 array | 否 | 策略 ID 列表，可查询关联策略和定时执行产生的记录 |
| status | string array | 否 | `running`、`success`、`failed`、`partial_failed` |
| operator | string array | 否 | 操作人列表 |

#### operate_time_range

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| start_timestamp_sec | int64 | 否 | 开始时间戳，单位为秒 |
| end_timestamp_sec | int64 | 否 | 结束时间戳，单位为秒 |

### 调用示例

```json
{
  "page": {"offset": 0, "limit": 20},
  "only_count": false,
  "exact_include_conditions": {
    "deploy_policy_id": [1001],
    "status": ["running", "success"]
  },
  "operate_time_range": {
    "start_timestamp_sec": 1717171200,
    "end_timestamp_sec": 1719859599
  }
}
```

轮询单次执行时使用 `"exact_include_conditions": {"workflow_id": ["workflow-example"]}`，不加状态过滤。只统计数量时设置 `only_count=true`，可省略 `page`。

### 响应示例

父流程 dispatch 已成功；关联的 plugin workflow 仍可能运行中。

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      {
        "workflow_id": "workflow-example",
        "trigger_id": "trigger-example",
        "deploy_policy_id": 1001,
        "operator": "admin",
        "operate_time": 1717171200000,
        "finish_time": 1717171205000,
        "status": "success",
        "children": [{"type": "plugin", "workflow_id": "plugin-workflow-example"}]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 接口状态码；`0` 表示查询成功，不表示部署成功 |
| message | string | 接口消息 |
| request_id | string | 请求 ID |
| error | object | 接口错误信息，成功时为空 |
| permission | object | 权限信息，无权限错误时为空 |
| data | object | 匹配的父流程执行记录 |

#### data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| total | int64 | 分页前匹配的父流程总数，不是子流程数量 |
| items | object array | 父流程列表；仅统计或无匹配记录时为空 |

#### items[n]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| workflow_id | string | 该策略本次执行的父 workflow ID |
| trigger_id | string | 关联的 trigger ID，不是 `workflow_id` 的别名，也不是 execute 返回值 |
| deploy_policy_id | int64 | 本次执行关联的策略 ID |
| operator | string | 操作人 |
| operate_time | int64 | Unix 时间戳，单位为毫秒 |
| finish_time | int64 | Unix 时间戳，单位为毫秒；未完成时为 `0` |
| status | string | `running`、`success`、`failed` 或 `partial_failed`，只描述父流程 dispatch operation |
| children | object array | 启动成功且已记录关联的子流程，不包含子流程状态或进度聚合 |

#### children[n]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| type | string | `node` 或 `plugin` |
| workflow_id | string | 子 workflow ID；通过 `(type, workflow_id)` 标识子流程 |

### 状态与限制

| 父 operation 状态 | status | 含义 |
| --- | --- | --- |
| 未完成 | `running` | dispatch 尚未完成 |
| 成功 | `success` | dispatch 成功完成，包括无需变更的情况 |
| 失败或超时 | `failed` | dispatch 失败，已发起的子流程仍可能运行 |
| 被终止 | `partial_failed` | dispatch 非正常结束，不表示失败子流程的比例 |

- 父流程只跟踪一个 dispatch operation，不等待或聚合子流程状态。重试子流程不会重新打开父流程。
- 状态由后台同步，列表查询不会从子流程重新计算。缺少状态的旧记录经过同一同步流程；未完成记录的 operation 缺失时给予一分钟宽限期，之后标记 `failed`；已有终态记录保留原结果。
- `children` 只记录启动成功且关联已写入的子流程。子记录缺失或不可读不改变父状态；dispatch 失败也不会终止已发起的子流程。
- 子流程结果通过现有 [node](NodeWorkflow_NodeWorkflowList.md) 或 [plugin](PluginWorkflow_PluginWorkflowList.md) workflow API 查询，其认证与 IAM 权限不变；父流程可见不代表拥有子流程访问权限。
- 当前认证租户内没有匹配父记录时返回空列表，不代表执行成功。无效请求和读取失败返回接口错误。
- dispatch action 不自动重试，不提供部署策略 workflow 的 retry、terminate、distinct 或 statistics API。再次 execute 是新的执行，不是幂等重试。
- 调用方自行设定轮询期限和继续条件。父流程成功不表示子流程完成，也不是机器健康保证。

### 相关接口

- [执行部署策略](DeployPolicySvc_Execute.md)
