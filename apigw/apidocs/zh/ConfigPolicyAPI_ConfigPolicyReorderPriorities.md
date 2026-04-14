### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：config_policy_manage（管理配置策略）。
- 该接口功能描述：按指定顺序重排同一业务和同一策略类型下的已启用策略优先级。

### URL

POST /api/v3/policy/config/reorder_priorities

### 输入参数

| 参数名称                    | 参数类型        | 必选 | 描述                                                                            |
|-------------------------|-------------|----|-------------------------------------------------------------------------------|
| bk_biz_id               | int64       | 是  | 业务 ID                                                                         |
| configpolicy_type       | string      | 是  | 配置策略类型，可选值：`config_policy_agent`、`config_policy_proxy`、`config_policy_plugin` |
| ordered_configpolicy_id | int64 array | 否  | 期望的策略 ID 顺序列表，列表中不能包含重复 ID                                                    |

### 调用示例

```json
{
  "bk_biz_id": 2,
  "configpolicy_type": "config_policy_agent",
  "ordered_configpolicy_id": [
    10003,
    10001,
    10002
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123463",
  "data": {}
}
```

### 响应参数说明

| 参数名称       | 参数类型   | 描述           |
|------------|--------|--------------|
| code       | int32  | 状态码，`0` 表示成功 |
| message    | string | 请求信息         |
| request_id | string | 请求 ID        |
| error      | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息         |
| data       | object | 响应数据，成功时为空对象 |

### 说明

- 协议语义为，`ordered_configpolicy_id` 中的策略按 `1..N` 分配优先级。
- 同范围内未在列表中的已启用策略会保持原相对顺序，并从 `N+1` 开始顺延。
