### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：networkunit_view（查看管控单元）。
- 该接口功能描述：查询管控单元列表，支持分页和条件过滤。

### URL

POST /api/v3/topo/networkunit/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页配置 |
| only_count | bool | 否 | 是否只返回总数，不返回详情 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| offset | int32 | 否 | 偏移量，起始值为 0 |
| limit | int32 | 否 | 每页限制条数 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkunit_id | int64 array | 否 | 管控单元 ID |
| bk_networkarea_id | int64 array | 否 | 管控区域 ID |
| is_direct | bool array | 否 | 是否直连 |
| generation | int64 array | 否 | 代际 |

### 调用示例

```json
{
  "page": {
    "offset": 1,
    "limit": 1
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_networkunit_id": [
      1
    ],
    "bk_networkarea_id": [
      1
    ],
    "is_direct": [
      false
    ],
    "generation": [
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
  "error": null,
  "permission": null,
  "data": {
    "total": 1,
    "items": [
      {
        "tenant_id": "id-001",
        "bk_networkunit_id": 1,
        "bk_networkunit_name": "default",
        "bk_networkarea_id": 1,
        "accesspoints": [
          1
        ],
        "links": {
          "cluster": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1,
            "accesspoint_id": 1
          },
          "file": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1,
            "accesspoint_id": 1
          },
          "data": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1,
            "accesspoint_id": 1
          }
        },
        "is_direct": false,
        "direct_endpoints": {
          "cluster": [
            "string"
          ],
          "file": [
            "string"
          ],
          "data": [
            "string"
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
| error | object | 错误信息，成功时为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 总记录数 |
| items | object array | 数据列表 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户 ID |
| bk_networkunit_id | int64 | 管控单元 ID |
| bk_networkunit_name | string | 管控单元名称 |
| bk_networkarea_id | int64 | 管控区域 ID |
| accesspoints | int64 array | 接入点列表 |
| links | object | 链路配置 |
| is_direct | bool | 是否直连 |
| direct_endpoints | object | 直连端点配置 |
| generation | int64 | 代际 |
| custom_deploy_config | object | 自定义部署配置 |
