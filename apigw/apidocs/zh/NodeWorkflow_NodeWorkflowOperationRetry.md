### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：agent_operate（操作Agent）、proxy_operate（操作Proxy）。
  - 权限说明：根据任务流涉及的节点角色动态收敛。若任务流仅涉及 Agent 节点，则只需 agent_operate；若仅涉及 Proxy 节点，则只需 proxy_operate；若涉及两者或角色未知，则需要两者。
- 该接口功能描述：对指定节点任务流中的操作执行重试，支持全量重试或按操作 ID 部分重试。

### URL

POST /api/v3/node/workflow/operation/retry

### 输入参数

| 参数名称      | 参数类型     | 必选 | 描述                                                                     |
| ------------- | ------------ | ---- | ------------------------------------------------------------------------ |
| workflow_id   | string       | 是   | 任务流 ID                                                                |
| retry_mod     | string       | 是   | 重试模式，可选值：`ALL`（重试全部失败操作）、`PARTIAL`（仅重试指定操作） |
| operation_ids | string array | 是   | 操作 ID 列表；当 `retry_mod=PARTIAL` 时用于指定需要重试的操作            |

### 调用示例

仅重试部分失败操作。

```json
{
  "workflow_id": "wf-20240601-0001",
  "retry_mod": "PARTIAL",
  "operation_ids": ["op-001", "op-003"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {}
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
| data       | object   | 响应数据，成功时为空对象 |

#### error

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| system   | string   | 错误来源系统 |
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

当前对象为空，返回 `{}`。
