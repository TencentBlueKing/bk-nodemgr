### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 版本变更：`v3.0.1-alpha.13+` 增加权限响应字段；`v3.0.1-alpha.18+` 增加 `plugin_operate` 鉴权。
- 该接口所需权限：plugin_operate（操作插件）。
- 该接口功能描述：重试指定插件任务流中的操作，支持全部重试或按操作 ID 部分重试。

### URL

POST /api/v3/plugin/workflow/operation/retry

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| workflow_id | string | 是 | 插件任务流 ID，不可为空 |
| retry_mod | string | 是 | 重试模式，可选值：`ALL`（重试全部符合条件的操作）、`PARTIAL`（重试指定操作） |
| operation_ids | string array | 是 | 操作 ID 列表，至少包含一个非空 ID |

### 调用示例

```json
{
  "workflow_id": "wf-plugin-0001",
  "retry_mod": "PARTIAL",
  "operation_ids": ["op-0001", "op-0003"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": null
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息 |
| data | null | 成功时无业务响应字段，当前序列化结果为 `null` |

### 说明

- `retry_mod` 区分大小写；当前执行层仅接受 `ALL` 和 `PARTIAL`。
- 每个 `operation_ids` 元素都不可为空。
