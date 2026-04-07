### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：。
- 该接口功能描述：查询管控区域列表，支持分页与按管控区域 ID、名称、云区域进行过滤。

### URL

POST /api/v3/topo/networkarea/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页配置；当 `limit` 为 `0` 时，服务端按不分页处理 |
| only_count | bool | 否 | 是否只返回总数，不返回详情 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| count | bool | 是 | 是否返回总记录条数 |
| start | uint32 | 否 | 记录开始位置，起始值为 0 |
| limit | uint32 | 否 | 每页限制条数；`0` 表示不分页 |
| sort | string | 否 | 排序字段 |
| order | string | 否 | 排序顺序（ASC、DESC） |

#### exact_include_conditions

精确匹配包含条件，满足任一条件即匹配。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 array | 否 | 管控区域 ID 列表 |
| cloud_vendor | string array | 否 | 云区域标识列表 |

#### fuzzy_include_conditions

模糊匹配包含条件，满足任一条件即匹配。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_name | string array | 否 | 管控区域名称列表（模糊匹配） |

### 调用示例

查询名称包含 `prod` 的管控区域列表，并返回总数。

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "bk_networkarea_id",
    "order": "ASC"
  },
  "fuzzy_include_conditions": {
    "bk_networkarea_name": [
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
        "bk_networkarea_id": 1,
        "bk_networkarea_name": "prod-default",
        "cloud_vendor": "tencent"
      },
      {
        "tenant_id": "default",
        "bk_networkarea_id": 2,
        "bk_networkarea_name": "prod-overseas",
        "cloud_vendor": "aws"
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
| error | object | 错误信息（成功时为空） |
| permission | object | 权限信息（当前接口通常为空） |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 当前规则能匹配到的总记录条数 |
| items | array | 查询返回的数据 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户 ID |
| bk_networkarea_id | int64 | 管控区域 ID |
| bk_networkarea_name | string | 管控区域名称 |
| cloud_vendor | string | 云区域标识 |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/application/api/v3/topo.proto` 与 `docs/api/swagger/application/api/v3/topo.swagger.json`。
- 当前后端权限语义已调整为：列表接口本身无需预鉴权；如需详情或后续操作，请使用对应详情/编辑/删除接口并按接口权限校验。
