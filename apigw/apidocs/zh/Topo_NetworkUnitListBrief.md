### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：networkunit_view（查看管控单元）。
- 该接口功能描述：查询管控单元简要列表，支持分页与按管控单元 ID、管控区域 ID、是否直连、代际进行精确过滤；`brief` 返回不会暴露 `direct_endpoints` 与
  `custom_deploy_config` 等敏感字段。

### URL

POST /api/v3/topo/networkunit/list/brief

### 输入参数

| 参数名称                     | 参数类型   | 必选 | 描述                               |
|--------------------------|--------|----|----------------------------------|
| page                     | object | 否  | 分页配置；当 `limit` 为 `0` 时，服务端按不分页处理 |
| only_count               | bool   | 否  | 是否只返回总数，不返回详情                    |
| exact_include_conditions | object | 否  | 精确匹配包含条件                         |

#### page

| 参数名称  | 参数类型   | 必选 | 描述               |
|-------|--------|----|------------------|
| count | bool   | 是  | 是否返回总记录条数        |
| start | uint32 | 否  | 记录开始位置，起始值为 0    |
| limit | uint32 | 否  | 每页限制条数；`0` 表示不分页 |
| sort  | string | 否  | 排序字段             |
| order | string | 否  | 排序顺序（ASC、DESC）   |

#### exact_include_conditions

精确匹配包含条件。

| 参数名称              | 参数类型        | 必选 | 描述         |
|-------------------|-------------|----|------------|
| bk_networkunit_id | int64 array | 否  | 管控单元 ID 列表 |
| bk_networkarea_id | int64 array | 否  | 管控区域 ID 列表 |
| is_direct         | bool array  | 否  | 是否直连       |
| generation        | int64 array | 否  | 管控单元代际     |

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
        "bk_networkarea_id": 1,
        "accesspoints": [
          2001,
          2002
        ],
        "links": {
          "cluster": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1001,
            "accesspoint_id": 2001
          },
          "file": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1001,
            "accesspoint_id": 2002
          },
          "data": {
            "bk_networkarea_id": -1,
            "bk_networkunit_id": -1,
            "accesspoint_id": -1
          }
        },
        "is_direct": true,
        "generation": 2
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

| 参数名称                | 参数类型        | 描述            |
|---------------------|-------------|---------------|
| tenant_id           | string      | 租户 ID         |
| bk_networkunit_id   | int64       | 管控单元 ID       |
| bk_networkunit_name | string      | 管控单元名称        |
| bk_networkarea_id   | int64       | 所属管控区域 ID     |
| accesspoints        | int64 array | 关联的接入点 ID 列表  |
| links               | object      | 当前管控单元的上游链路关系 |
| is_direct           | bool        | 是否直连          |
| generation          | int64       | 管控单元代际        |

#### data.items[n].links

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| cluster | object | 集群链路 |
| file    | object | 文件链路 |
| data    | object | 数据链路 |

#### data.items[n].links.{cluster\|file\|data}

| 参数名称              | 参数类型  | 描述                    |
|-------------------|-------|-----------------------|
| bk_networkarea_id | int64 | 目标管控区域 ID；无值时通常为 `-1` |
| bk_networkunit_id | int64 | 目标管控单元 ID；无值时通常为 `-1` |
| accesspoint_id    | int64 | 目标接入点 ID；无值时通常为 `-1`  |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto`、`proto/backend/api/v3/common.proto` 以及
  `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- 路由实现位于 `internal/backend/router/api-v3/topo/networkunit.go`，服务端会根据当前用户有权查看的管控单元范围自动收窄返回结果。
- `brief` 响应不会返回 `direct_endpoints`、`custom_deploy_config` 等敏感字段；如需完整配置，请使用完整的 `networkunit` 详情接口。
