### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：agent_operate（操作Agent）、proxy_operate（操作Proxy）。
  - 权限说明：根据任务流涉及的节点角色动态收敛。若任务流仅涉及 Agent 节点，则只需 agent_operate；若仅涉及 Proxy 节点，则只需 proxy_operate；若涉及两者或角色未知，则需要两者。
- 该接口功能描述：按任务流 ID 查询节点任务流下 operation 最新实例状态（state）的去重结果。

### URL

POST /api/v3/node/workflow/operation/distinct

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                  |
| ----------- | -------- | ---- | --------------------- |
| workflow_id | string   | 是   | 任务流 ID。不能为空。 |
| selector    | object   | 否   | 去重字段选择器。      |

#### selector

| 参数名称 | 参数类型 | 必选 | 描述                                      |
| -------- | -------- | ---- | ----------------------------------------- |
| state    | bool     | 否   | 是否返回 `state` 去重结果。默认 `false`。 |

### 调用示例

查询任务流 `wf-20240601-0001` 下 operation 最新实例状态的去重结果。

```json
{
  "workflow_id": "wf-20240601-0001",
  "selector": {
    "state": true
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
    "state": ["running", "success", "failed"]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                         |
| ---------- | -------- | ---------------------------- |
| code       | int32    | 状态码，`0` 表示成功。       |
| message    | string   | 请求信息。                   |
| request_id | string   | 请求 ID。                    |
| error      | object   | 错误信息，成功时通常为空。   |
| permission | object   | 权限信息，当前接口通常为空。 |
| data       | object   | 响应数据。                   |

#### data

| 参数名称 | 参数类型     | 描述                                                                                                                    |
| -------- | ------------ | ----------------------------------------------------------------------------------------------------------------------- |
| state    | string array | operation 最新实例状态的去重结果。可选值：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated`。 |

### 说明

- `workflow_id` 会先转换为对应任务流的 `trigger_id`，再在该任务流范围内执行去重查询。
- 当 `selector.state` 为 `false` 或未传 `selector` 时，`data.state` 通常为空数组。
