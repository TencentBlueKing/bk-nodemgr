### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：config_policy_manage（管理配置策略）。
- 该接口功能描述：批量启用配置策略。

### URL

POST /api/v3/policy/config/enable

### 输入参数

| 参数名称            | 参数类型        | 必选 | 描述             |
|-----------------|-------------|----|----------------|
| configpolicy_id | int64 array | 是  | 待启用的配置策略 ID 列表 |

### 调用示例

```json
{
  "configpolicy_id": [
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
  "request_id": "req-123460",
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
