### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：无。
- 该接口功能描述：查询接入点简要列表，支持按管控区域 ID、接入点 ID 进行精确过滤；`brief` 响应仅返回接入点基础标识信息，不返回 `endpoints` 等敏感地址配置。

### URL

POST /api/v3/topo/accesspoint/list/brief

### 输入参数

| 参数名称                     | 参数类型   | 必选 | 描述                                                                              |
|--------------------------|--------|----|---------------------------------------------------------------------------------|
| page                     | object | 否  | 分页配置；当 `only_count` 不为 `true` 时，`offset` 必须大于等于 `0`，`limit` 必须在 `(0, 1000]` 范围内 |
| only_count               | bool   | 否  | 是否只返回总数，不返回详情                                                                   |
| exact_include_conditions | object | 否  | 精确匹配包含条件                                                                        |

#### page

| 参数名称   | 参数类型  | 必选 | 描述                       |
|--------|-------|----|--------------------------|
| offset | int32 | 否  | 记录起始偏移量，起始值为 `0`         |
| limit  | int32 | 否  | 返回记录上限，取值范围为 `(0, 1000]` |

#### exact_include_conditions

精确匹配包含条件，满足任一接入点 ID 或任一管控区域 ID 即可参与查询范围收敛。

| 参数名称              | 参数类型        | 必选 | 描述         |
|-------------------|-------------|----|------------|
| bk_networkarea_id | int64 array | 否  | 管控区域 ID 列表 |
| accesspoint_id    | int64 array | 否  | 接入点 ID 列表  |

### 调用示例

查询指定管控区域下的接入点简要列表，并返回总数。

```json
{
  "page": {
    "offset": 0,
    "limit": 10
  },
  "exact_include_conditions": {
    "bk_networkarea_id": [
      1
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
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-shenzhen-prod",
        "bk_networkarea_id": 1
      },
      {
        "accesspoint_id": 1002,
        "accesspoint_name": "ap-shanghai-prod",
        "bk_networkarea_id": 1
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称       | 参数类型   | 描述                                    |
|------------|--------|---------------------------------------|
| code       | int32  | 状态码，0 表示成功                            |
| message    | string | 请求信息                                  |
| request_id | string | 请求 ID                                 |
| error      | object | 错误信息（成功时为空）                           |
| permission | object | 权限信息；当请求范围超出当前有权查看的管控单元范围时，可能返回权限相关信息 |
| data       | object | 响应数据                                  |

#### data

| 参数名称  | 参数类型  | 描述                                    |
|-------|-------|---------------------------------------|
| total | int64 | 当前规则能匹配到的总记录条数                        |
| items | array | 查询返回的数据；当 `only_count` 为 `true` 时通常为空 |

#### data.items[n]

| 参数名称              | 参数类型   | 描述        |
|-------------------|--------|-----------|
| accesspoint_id    | int64  | 接入点 ID    |
| accesspoint_name  | string | 接入点名称     |
| bk_networkarea_id | int64  | 所属管控区域 ID |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto`、`proto/backend/api/v3/common.proto` 与
  `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- 路由实现位于 `internal/backend/router/api-v3/topo/accesspoint.go`。该 `brief` 列表接口当前不做额外权限收窄，按请求条件直接执行列表查询。
- `brief` 响应不会返回 `endpoints` 等接入点地址配置；如需完整接入点地址信息，请使用完整的 `accesspoint/list` 接口。
