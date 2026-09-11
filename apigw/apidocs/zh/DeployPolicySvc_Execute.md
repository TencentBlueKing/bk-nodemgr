### 描述

- 该接口提供版本：v3.0.1+
- 该接口所需权限：无新增 IAM 权限；保留现有身份认证和租户隔离。
- 该接口功能描述：异步执行部署策略，返回被请求策略本次执行的 workflow ID。

### URL

POST /api/v3/deploy_policy/execute

### 输入参数

| 参数名称         | 参数类型 | 必选 | 描述       |
| ---------------- | -------- | ---- | ---------- |
| deploy_policy_id | int64    | 是   | 部署策略ID |

### 调用示例

```json
{
  "deploy_policy_id": 1001
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "workflow_id": "workflow-abc123def456"
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                         |
| ---------- | -------- | ---------------------------- |
| code       | int32    | 状态码，0表示成功            |
| message    | string   | 请求信息                     |
| request_id | string   | 请求ID                       |
| error      | object   | 错误信息（成功时为空）       |
| permission | object   | 权限信息（无权限错误时为空） |
| data       | object   | 响应数据                     |

#### data

| 参数名称    | 参数类型 | 描述                                                                                        |
| ----------- | -------- | ------------------------------------------------------------------------------------------- |
| workflow_id | string   | 被请求策略本次执行的业务 workflow ID，用于调用 `POST /api/v3/deploy_policy/workflow/result` |

### 执行语义与限制

- 每次 execute 都产生新的执行身份；同一策略的连续或并发执行不共用 workflow ID。
- discovery 仍可能关联执行其他策略。每个参与策略建立自己的执行记录；本接口只返回被请求策略的 workflow ID。
- 同一次执行中的策略可以共享 node/plugin 子 workflow，因此相关策略之间不保证严格串行。
- 自动重试沿用本次执行身份，实际新增的子 workflow 全部计入；本接口不提供幂等键或全面防重复发起保证。
- 参数无效、策略不存在、策略未启用或启动失败仍返回接口错误。成功响应只确认已发起，不表示子流程已结束或机器已收敛。
- 调用方通过[执行结果查询](DeployPolicySvc_WorkflowResult.md)轮询，自行决定何时执行下一策略。节点管理不自动触发下一策略。
