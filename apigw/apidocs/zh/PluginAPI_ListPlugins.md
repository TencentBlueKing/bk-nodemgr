### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：。
- 该接口功能描述：查询插件列表，支持分页、仅统计总数，以及按精确或模糊条件过滤。

### URL

POST /api/v3/plugin/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页参数；当 `only_count=true` 时可省略 |
| only_count | bool | 否 | 是否只返回总数；为 `true` 时不返回列表项 |
| exact_include_conditions | object | 否 | 精确匹配条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| offset | int32 | 否 | 分页起始偏移量 |
| limit | int32 | 否 | 返回条数上限 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| name | string array | 否 | 按插件名称精确匹配 |
| group | string array | 否 | 按插件分组精确匹配 |

#### fuzzy_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| name | string array | 否 | 按插件名称模糊匹配 |
| pkg_name | string array | 否 | 按插件包名称模糊匹配 |

**参数说明**：
- `only_count=true` 时，接口只返回匹配总数，`items` 为空或不返回具体列表内容。
- `exact_include_conditions` 与 `fuzzy_include_conditions` 均为包含型条件；当前接口未暴露排除型条件。
- `group` 常见默认值可为 `default`，但接口本身不限制为固定枚举。

### 调用示例

查询名称包含 `monitor` 的插件列表，并返回前 20 条数据。

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "fuzzy_include_conditions": {
    "name": [
      "monitor"
    ]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000003",
  "error": null,
  "data": {
    "total": 2,
    "items": [
      {
        "tenant_id": "default",
        "name": "bk-monitor-agent",
        "group": "default",
        "pkg_name": "bkmonitoragent",
        "memo": "BlueKing monitor agent"
      },
      {
        "tenant_id": "default",
        "name": "bk-monitor-proxy",
        "group": "default",
        "pkg_name": "bkmonitorproxy",
        "memo": "BlueKing monitor proxy"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求ID |
| error | object | 错误信息，成功时为null |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 匹配条件的插件总数 |
| items | object array | 插件列表；当 `only_count=true` 时通常为空 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户ID |
| name | string | 插件名称 |
| group | string | 插件分组 |
| pkg_name | string | 插件包名称 |
| memo | string | 插件备注 |

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误系统标识 |
| message | string | 错误消息 |
| details | array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误代码 |
| message | string | 错误消息 |
