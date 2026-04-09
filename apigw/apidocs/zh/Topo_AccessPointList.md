### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询接入点列表，支持分页与按管控区域 ID、接入点 ID 进行精确过滤。

### URL

POST /api/v3/topo/accesspoint/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页配置；当 `limit` 为 `0` 时，服务端按不分页处理 |
| only_count | bool | 否 | 是否只返回总数，不返回详情 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| count | bool | 是 | 是否返回总记录条数 |
| start | uint32 | 否 | 记录开始位置，起始值为 0 |
| limit | uint32 | 否 | 每页限制条数；`0` 表示不分页 |
| sort | string | 否 | 排序字段 |
| order | string | 否 | 排序顺序（ASC、DESC） |

#### exact_include_conditions

精确匹配包含条件。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 array | 否 | 管控区域 ID 列表 |
| accesspoint_id | int64 array | 否 | 接入点 ID 列表 |

### 调用示例

查询指定管控区域下的接入点列表，并返回总数。

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "accesspoint_id",
    "order": "ASC"
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
        "tenant_id": "default",
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-shenzhen-prod",
        "bk_networkarea_id": 1,
        "endpoints": {
          "cluster": [
            "https://bcs-cluster.example.com"
          ],
          "file": [
            "https://bcs-file.example.com"
          ],
          "data": [
            "https://bcs-data.example.com"
          ]
        }
      },
      {
        "tenant_id": "default",
        "accesspoint_id": 1002,
        "accesspoint_name": "ap-shanghai-prod",
        "bk_networkarea_id": 1,
        "endpoints": {
          "cluster": [
            "https://bcs-cluster-sh.example.com"
          ],
          "file": [
            "https://bcs-file-sh.example.com"
          ],
          "data": [
            "https://bcs-data-sh.example.com"
          ]
        }
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
| permission | object | 权限信息；当请求的接入点范围超出当前有权查看的管控单元范围时，可能返回权限相关信息 |
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
| accesspoint_id | int64 | 接入点 ID |
| accesspoint_name | string | 接入点名称 |
| bk_networkarea_id | int64 | 所属管控区域 ID |
| endpoints | object | 接入点地址配置 |

#### data.items[n].endpoints

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| cluster | string array | 集群访问地址列表 |
| file | string array | 文件访问地址列表 |
| data | string array | 数据访问地址列表 |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto` 与 `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- 路由实现位于 `internal/backend/router/api-v3/topo/accesspoint.go`，服务端会根据当前用户有权查看的管控单元范围自动收窄可返回的接入点集合。
