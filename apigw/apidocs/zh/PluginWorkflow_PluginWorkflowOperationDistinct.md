### 描述

- 该接口提供版本：v3.0.1-alpha.4+。
- 版本变更：`v3.0.1-alpha.13+` 增加权限响应字段。
- 该接口所需权限：无。
- 该接口功能描述：获取指定插件任务流下操作状态的去重候选值。

### URL

POST /api/v3/plugin/workflow/operation/distinct

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| workflow_id | string | 是 | 插件任务流 ID，不可为空 |
| selector | object | 否 | 去重字段选择器 |

#### selector

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| state | bool | 否 | 是否返回操作状态去重结果；为 `true` 时返回 `data.state` |

### 调用示例

```json
{
  "workflow_id": "wf-plugin-0001",
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

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| state | string array | 操作状态去重结果，可选值：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated` |

### 说明

- 未传 `selector` 或 `selector.state=false` 时，不选择状态列，`data.state` 为空数组。
- 当前 handler 未执行 IAM 权限检查。
