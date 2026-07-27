### 描述

- 该接口提供版本：v3.0.1-alpha.56+。
- 该接口所需权限：无。
- 该接口功能描述：按触发 ID 查询操作实例状态分布。

**权限说明**：当前后端 handler 未执行直接鉴权。

### URL

POST /api/v3/node/workflow/operation/instance/status_distribution/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| trigger_id | string array | 否 | 任务流触发 ID 列表；不传时按后端查询条件返回 |

### 调用示例

```json
{
  "trigger_id": [
    "trigger-9f2b"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": {
      "trigger-9f2b": {
        "not_inited_count": 1,
        "state_counts": {
          "running": 2,
          "success": 3
        }
      }
    }
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
| items | object | 状态分布映射，键为 trigger_id，值为状态分布 |

#### data.items.<trigger_id>

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| not_inited_count | int64 | 未初始化的操作实例数量 |
| state_counts | object | 各操作实例状态数量，键为状态，值为数量 |
