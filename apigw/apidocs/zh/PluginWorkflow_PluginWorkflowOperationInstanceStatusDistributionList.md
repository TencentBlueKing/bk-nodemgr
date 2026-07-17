### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 版本变更：`v3.0.1-alpha.13+` 增加权限响应字段。
- 该接口所需权限：无。
- 该接口功能描述：按触发器 ID 查询插件任务流最新操作实例的状态分布。

### URL

POST /api/v3/plugin/workflow/operation/instance/status_distribution/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| trigger_id | string array | 是 | 触发器 ID 列表；至少包含一个 ID |

### 调用示例

```json
{
  "trigger_id": ["trigger-plugin-0001", "trigger-plugin-0002"]
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
      "trigger-plugin-0001": {
        "not_inited_count": 2,
        "state_counts": {
          "running": 3,
          "success": 8,
          "failed": 1
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
| items | object | 以触发器 ID 为键的状态分布映射 |

#### data.items.{trigger_id}

`trigger_id` 为请求中的触发器 ID，值为该触发器关联的最新操作实例状态分布。

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| not_inited_count | int64 | 尚未创建操作实例的操作数量 |
| state_counts | object | 按状态统计的操作实例数量映射 |

#### data.items.{trigger_id}.state_counts

键为操作实例状态，可选值：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated`；值为对应状态的实例数量。

### 说明

- 请求 Proto 未直接校验空列表，但存储层会拒绝空 `trigger_id`，因此调用时必须至少传入一个 ID。
- 当前 handler 未执行 IAM 权限检查。
