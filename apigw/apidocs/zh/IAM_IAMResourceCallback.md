### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：供 BlueKing IAM 回调查询节点管理资源信息，按请求中的资源类型与回调方法分发到对应资源提供方。

### URL

POST /api/v3/iam/v3/resource

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| type | string | 是 | 资源类型。当前路由已接入的资源类型包括 `networkarea`、`networkunit`、`package_type`、`package`。 |
| method | string | 是 | IAM 回调方法标识，决定本次查询的处理逻辑。 |
| filter | object | 否 | 查询条件，结构随 `method` 变化，由 IAM 按回调协议传入。 |
| page | object | 否 | 分页参数，部分列表类回调方法使用。 |

#### method 可选值

`list_attr`、`list_attr_value`、`list_instance`、`fetch_instance_info`、`list_instance_by_policy`、`search_instance`、`fetch_instance_list`、`fetch_resource_type_schema`

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| offset | int32 | 否 | 分页起始偏移量。 |
| limit | int32 | 否 | 分页返回上限。 |

### 调用示例

该接口为回调式接口，通常由 IAM 系统携带 Basic Auth 凭据发起调用。

```json
{
  "type": "networkunit",
  "method": "list_instance",
  "filter": {
    "parent": {
      "id": "2"
    }
  },
  "page": {
    "offset": 0,
    "limit": 100
  }
}
```

### 响应示例

不同回调方法的 `data` 结构不同，下面示例展示 `list_instance` 场景下的典型返回。

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 1,
    "results": [
      {
        "id": "1001",
        "display_name": "default-network-unit"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 响应码，0 表示成功。业务错误通常仍以 HTTP 200 返回，并在该字段体现具体结果。 |
| message | string | 响应信息。 |
| data | object/null/array/string | 回调结果数据，具体结构随 `method` 变化。 |

#### data 常见结构

| 场景 | 参数类型 | 描述 |
|------|----------|------|
| `list_instance` | object | 返回资源列表，通常包含 `count` 和 `results`。 |
| `fetch_instance_info` | array | 返回资源详情列表。 |
| `list_attr` / `list_attr_value` | array | 返回资源属性或属性值列表。 |
| `fetch_resource_type_schema` | object | 返回资源类型 schema。 |

#### data.results[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| id | string | 资源实例唯一标识。 |
| display_name | string | 资源实例展示名称。 |

### 说明

- 该接口是 IAM 侧回调入口，不面向普通业务用户直接调用。
- 路由配置声明 `resourcePermissionRequired: false`，因此不会进行业务资源权限校验。
- 接口鉴权依赖专用 Basic Auth 中间件，用户名固定为 `bk_iam`，密码需与当前 IAM system token 匹配。
