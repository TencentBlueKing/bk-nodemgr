### 描述

- 该接口提供版本：v3.0.1-alpha.107+。
- 该接口所需权限：system_access（系统访问）。
- 该接口功能描述：通过用户态入口分页查询当前租户的业务列表，支持业务 ID 精确过滤和业务名称模糊过滤；不按用户的 Agent、Proxy 查看权限过滤业务列表。
- 用户身份及租户信息由 API 网关传递，无需在请求体中提供用户名或租户 ID。

### URL

POST /api/v3/open/topo/business/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                                                             |
| ------------------------ | -------- | ---- | ---------------------------------------------------------------- |
| page                     | object   | 是   | 分页配置，必须提供有效的 `limit`                                 |
| only_count               | bool     | 否   | 默认 `false`；当前接口不按此字段切换为仅计数模式，仍返回分页列表 |
| exact_include_conditions | object   | 否   | 精确匹配包含条件                                                 |
| fuzzy_include_conditions | object   | 否   | 模糊匹配包含条件                                                 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                              |
| -------- | -------- | ---- | ------------------------------------------------- |
| offset   | int32    | 否   | 偏移量，默认 `0`，必须大于或等于 `0`              |
| limit    | int32    | 是   | 每页条数，范围 `1`～`1000`；不支持 `0` 或省略分页 |

#### exact_include_conditions

| 参数名称  | 参数类型    | 必选 | 描述         |
| --------- | ----------- | ---- | ------------ |
| bk_biz_id | int64 array | 否   | 业务 ID 列表 |

#### fuzzy_include_conditions

| 参数名称    | 参数类型     | 必选 | 描述                 |
| ----------- | ------------ | ---- | -------------------- |
| bk_biz_name | string array | 否   | 业务名称模糊匹配列表 |

### 调用示例

```json
{
  "page": { "offset": 0, "limit": 10 },
  "exact_include_conditions": { "bk_biz_id": [2] }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      { "tenant_id": "default", "bk_biz_id": 2, "bk_biz_name": "示例业务" }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                                                                        |
| ---------- | -------- | --------------------------------------------------------------------------- |
| code       | int32    | 状态码，`0` 表示成功                                                        |
| message    | string   | 请求结果信息                                                                |
| request_id | string   | 请求 ID                                                                     |
| error      | object   | 错误信息，包含 `system`、`message` 和 `details`（错误码及说明）；成功时为空 |
| permission | object   | 权限申请信息，权限不足时可能返回                                            |
| data       | object   | 响应数据                                                                    |

#### data

| 参数名称 | 参数类型     | 描述                   |
| -------- | ------------ | ---------------------- |
| total    | int64        | 符合查询条件的总记录数 |
| items    | object array | 当前页业务列表         |

#### data.items[n]

| 参数名称    | 参数类型 | 描述     |
| ----------- | -------- | -------- |
| tenant_id   | string   | 租户 ID  |
| bk_biz_id   | int64    | 业务 ID  |
| bk_biz_name | string   | 业务名称 |
