### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询业务列表，支持分页与按业务 ID、业务名称进行过滤。

### URL

POST /api/v3/topo/business/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页配置；`offset` 必须大于等于 `0`，`limit` 必须在 `(0, 1000]` 范围内 |
| only_count | bool | 否 | 是否只返回总数 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| offset | int32 | 否 | 记录起始偏移量，起始值为 `0` |
| limit | int32 | 否 | 返回记录上限，取值范围为 `(0, 1000]` |

#### exact_include_conditions

精确匹配包含条件，满足任一业务 ID 即匹配。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_biz_id | int64 array | 否 | 业务 ID 列表 |

#### fuzzy_include_conditions

模糊匹配包含条件，满足任一业务名称模式即匹配。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_biz_name | string array | 否 | 业务名称列表，按名称做模糊匹配 |

### 调用示例

查询业务名称包含 `prod` 的业务列表。

```json
{
  "page": {
    "offset": 0,
    "limit": 10
  },
  "fuzzy_include_conditions": {
    "bk_biz_name": [
      "prod"
    ]
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
    "total": 2,
    "items": [
      {
        "tenant_id": "default",
        "bk_biz_id": 2,
        "bk_biz_name": "prod-payment"
      },
      {
        "tenant_id": "default",
        "bk_biz_id": 7,
        "bk_biz_name": "prod-order"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时为空 |
| permission | object | 权限信息，当前接口通常为空 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 当前规则能匹配到的总记录条数 |
| items | array | 查询返回的业务列表 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户 ID |
| bk_biz_id | int64 | 业务 ID |
| bk_biz_name | string | 业务名称 |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto` 与 `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- 当前后端权限语义已调整为：列表接口本身无需预鉴权；如需详情或后续操作，请使用对应详情或变更接口并按接口权限校验。
