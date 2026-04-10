### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询管控单元简要列表，支持分页与按管控单元 ID、管控区域 ID、是否直连、节点代次进行精确过滤。

### URL

POST /api/v3/topo/networkunit/list/brief

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页配置；在 application 层，当 `limit` 为 `0` 时，服务端按不分页处理 |
| only_count | bool | 否 | 是否只返回总数，不返回详情 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| count | bool | 是 | 是否返回总记录条数 |
| start | uint32 | 否 | 记录开始位置，起始值为 0 |
| limit | uint32 | 否 | 每页限制条数；在 application 层，`0` 表示不分页 |
| sort | string | 否 | 排序字段 |
| order | string | 否 | 排序顺序（ASC、DESC） |

#### exact_include_conditions

精确匹配包含条件。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkunit_id | int64 array | 否 | 管控单元 ID 列表 |
| bk_networkarea_id | int64 array | 否 | 管控区域 ID 列表 |
| is_direct | bool array | 否 | 是否为直连管控单元 |
| generation | int64 array | 否 | 节点代次列表 |

### 调用示例

查询指定管控区域下的直连管控单元简要列表，并返回总数。

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "bk_networkunit_id",
    "order": "ASC"
  },
  "exact_include_conditions": {
    "bk_networkarea_id": [
      1
    ],
    "is_direct": [
      true
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
        "bk_networkunit_id": 1001,
        "bk_networkunit_name": "direct-unit-shenzhen",
        "bk_networkarea_id": 1
      },
      {
        "tenant_id": "default",
        "bk_networkunit_id": 1002,
        "bk_networkunit_name": "direct-unit-shanghai",
        "bk_networkarea_id": 1
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
| items | array | 查询返回的数据；当 `only_count` 为 `true` 时通常为空 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户 ID |
| bk_networkunit_id | int64 | 管控单元 ID |
| bk_networkunit_name | string | 管控单元名称 |
| bk_networkarea_id | int64 | 所属管控区域 ID |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/application/api/v3/topo.proto` 与 `docs/api/swagger/application/api/v3/topo.swagger.json`。
- application 路由实现位于 `internal/application/router/api-v3/topo/networkunit.go`，其会代理到 backend 的同名 brief 列表接口。
- 当前实现虽然复用了 `NetworkUnitBrief` 协议类型，但实际仅返回 `tenant_id`、`bk_networkunit_id`、`bk_networkunit_name`、`bk_networkarea_id` 这 4 个字段，不返回接入点、链路、直连地址或自定义部署配置等扩展信息。
- 当前请求仅支持 `exact_include_conditions`，不支持模糊过滤或排除条件。
