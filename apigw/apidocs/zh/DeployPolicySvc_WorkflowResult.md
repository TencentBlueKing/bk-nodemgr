### 描述

- 该接口提供版本：v3.0.1-alpha.82+
- 该接口所需权限：无新增 IAM 权限；保留现有身份认证和租户隔离。
- 该接口功能描述：查询一条部署策略某次执行的当前结果、发起状态和关联 node/plugin 子 workflow。

### URL

POST /api/v3/deploy_policy/workflow/result

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                                                                                               |
| ----------- | -------- | ---- | -------------------------------------------------------------------------------------------------- |
| workflow_id | string   | 是   | 非空的部署策略 workflow ID，由 execute 返回；不是策略 ID、trigger ID 或 node/plugin 子 workflow ID |

### 调用示例

```json
{
  "workflow_id": "workflow-abc123def456"
}
```

### 响应示例

发起过程已正常结束，两个子 workflow 中一个仍在执行：

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "workflow_id": "workflow-abc123def456",
    "deploy_policy_id": 1001,
    "status": "running",
    "total_count": 2,
    "finished_count": 1,
    "dispatch_status": "success",
    "dispatch_error": "",
    "children": [
      {
        "type": "node",
        "workflow_id": "node-workflow-example",
        "status": "success",
        "missing": false
      },
      {
        "type": "plugin",
        "workflow_id": "plugin-workflow-example",
        "status": "running",
        "missing": false
      }
    ],
    "unknown_reason": ""
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                                           |
| ---------- | -------- | ---------------------------------------------- |
| code       | int32    | API 状态码，0 表示本次查询成功，不表示部署成功 |
| message    | string   | API 请求信息，不用于业务状态判断               |
| request_id | string   | 请求 ID                                        |
| error      | object   | API 错误信息，成功时为空                       |
| permission | object   | 权限信息，无权限错误时为空                     |
| data       | object   | 查询成功时的当前聚合结果                       |

#### data

| 参数名称         | 参数类型     | 描述                                                                                                                    |
| ---------------- | ------------ | ----------------------------------------------------------------------------------------------------------------------- |
| workflow_id      | string       | 本次查询的部署策略 workflow ID                                                                                          |
| deploy_policy_id | int64        | 该执行记录所属的策略 ID                                                                                                 |
| status           | string       | `running`、`success`、`failed`、`partial_failed`、`unknown`；规则见下表                                                 |
| total_count      | int64        | 已确认实际创建的关联子 workflow 数，等于 `children` 长度；保留已创建但记录缺失的项，不包含未确认创建的意图              |
| finished_count   | int64        | 当前状态为 `success`、`failed` 或 `partial_failed` 的子 workflow 数；缺失项不计入                                       |
| dispatch_status  | string       | `running`：发起或自动重试尚未结束；`success`：正常结束；`failed`：最终失败且已结束；`unknown`：无法确认发起过程是否结束 |
| dispatch_error   | string       | 最近一次发起错误，无错误时为空；自动重试期间可与 `running` 共存，不代表子 workflow 错误或 API 错误                      |
| children         | object array | 完整关联集合，不分页、不按权限过滤为子集；无子流程时为 `[]`                                                             |
| unknown_reason   | string       | `status=unknown` 时非空的原因说明，其他状态为空；不是稳定错误码，调用方不要解析文本控制流程                             |

#### children[n]

| 参数名称    | 参数类型 | 描述                                                                                           |
| ----------- | -------- | ---------------------------------------------------------------------------------------------- |
| type        | string   | `node` 或 `plugin`                                                                             |
| workflow_id | string   | node/plugin 子 workflow ID；同一结果内按 `(type, workflow_id)` 去重                            |
| status      | string   | 子 workflow 当前状态：`running`、`success`、`failed`、`partial_failed`；无法判定时为 `unknown` |
| missing     | bool     | 曾确认创建的子记录是否缺失；为 true 时 `status=unknown`                                        |

### 聚合规则

按表格顺序判断。`dispatch_status=failed` 不会跳过对已创建子 workflow 的等待。

| 条件                                                                | data.status      | 当前已结束 |
| ------------------------------------------------------------------- | ---------------- | ---------- |
| 子记录缺失、创建结果未确认、发起是否结束无法确认，或其他证据不足    | `unknown`        | 无法判定   |
| 发起/自动重试尚未结束，或存在未结束的子 workflow                    | `running`        | 否         |
| 发起正常结束，且无需变更，没有子 workflow                           | `success`        | 是         |
| 发起正常结束，所有子 workflow 成功                                  | `success`        | 是         |
| 发起最终失败，所有实际创建的子 workflow 已结束，包括零子流程        | `failed`         | 是         |
| 发起正常结束，所有子 workflow 都是 failed                           | `failed`         | 是         |
| 发起正常结束，所有子 workflow 已结束，结果混合或包含 partial_failed | `partial_failed` | 是         |

- `total_count` 在发起期间可能增加，证据不足时不能视为最终完整数量。`total_count=finished_count`，甚至两者均为 0，都不能替代 `status` 判断。
- 每次查询重新读取关联集合和子 workflow 的业务状态，不冻结父记录终态。子 workflow 在原 ID 上重试后，后续查询可以重新返回 `running`。
- `dispatch_status=success` 仅表示发起结束，不保证子 workflow 成功。`dispatch_error` 非空也不能代替聚合状态判断。
- 多条策略执行记录可以共享同一个子 workflow；各策略计数相加可能重复，并且某条策略可能等待共享 workflow 中其他策略的任务。
- 不返回主机、operation 或日志详情；通过对应的 node/plugin workflow 接口进一步查询。

### 错误与调用边界

- `workflow_id` 为空、当前租户内不存在该部署策略执行记录或数据读取失败时，返回 API 错误，不返回伪造的空成功结果。
- 父执行记录存在，但关联子记录缺失或发起过程无法确认时，查询成功并返回 `status=unknown`，附带 `unknown_reason`。首次创建是否成功无法确认的项不计入实际子流程数。
- 已确认创建的子记录缺失时，仍在 `children` 中保留其 ID，以 `missing=true` 标识；不计入 `finished_count`。
- 超时或崩溃不保证发起过程已经退出；首版不承诺所有 `unknown` 都会自动恢复。
- 调用方自行设置轮询间隔、截止时间和后续业务策略。`unknown` 既不等于已结束，也不保证仍在运行，不能驱动下一策略。
- 当前已结束不等于全部成功。节点管理不自动触发下一策略，也不撤回调用方已经触发的后续动作。
- 结果不是跨多个 workflow 的事务快照，也不表示持续健康检查结果。

### 相关接口

- [执行部署策略](DeployPolicySvc_Execute.md)
