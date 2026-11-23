### 描述

- 该接口提供版本：v3.0.1-alpha.107+。
- 该接口所需权限：system_access（系统访问）。
- 该接口功能描述：通过用户态入口查询管控区域列表，支持分页、精确条件和名称模糊过滤。
- 用户身份及租户信息由 API 网关传递，无需在请求体中提供用户名或租户 ID。列表不额外按 `networkarea_view` 收窄；查询详情仍需对应区域的查看权限。

### URL

POST /api/v3/open/topo/networkarea/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                                                      |
| ------------------------ | -------- | ---- | --------------------------------------------------------- |
| page                     | object   | 是   | 分页配置；即使 `only_count=true` 也必须提供有效的 `limit` |
| only_count               | bool     | 否   | 当前接口不单独处理此字段，仍返回分页列表和总数            |
| exact_include_conditions | object   | 否   | 精确匹配条件                                              |
| fuzzy_include_conditions | object   | 否   | 模糊匹配条件                                              |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                                      |
| -------- | -------- | ---- | --------------------------------------------------------- |
| offset   | int32    | 否   | 记录起始偏移量，默认 `0`，必须大于等于 `0`                |
| limit    | int32    | 是   | 返回记录上限，范围为 `(0, 1000]`；不支持用 `0` 表示不分页 |

#### exact_include_conditions

| 参数名称          | 参数类型     | 必选 | 描述                                 |
| ----------------- | ------------ | ---- | ------------------------------------ |
| bk_networkarea_id | int64 array  | 否   | 管控区域 ID 列表                     |
| cloud_vendor      | string array | 否   | CMDB 云厂商标识列表，不是管控区域 ID |

#### fuzzy_include_conditions

| 参数名称            | 参数类型     | 必选 | 描述                     |
| ------------------- | ------------ | ---- | ------------------------ |
| bk_networkarea_name | string array | 否   | 管控区域名称模糊匹配列表 |

### 调用示例

```json
{
  "page": { "offset": 0, "limit": 10 },
  "fuzzy_include_conditions": { "bk_networkarea_name": ["prod"] }
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
      {
        "tenant_id": "default",
        "bk_networkarea_id": 1,
        "bk_networkarea_name": "prod-default",
        "cloud_vendor": "1"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                             |
| ---------- | -------- | -------------------------------- |
| code       | int32    | 状态码，`0` 表示成功             |
| message    | string   | 请求结果信息                     |
| request_id | string   | 请求 ID                          |
| error      | object   | 错误信息，成功时为空             |
| permission | object   | 权限申请信息，权限不足时可能返回 |
| data       | object   | 响应数据                         |

#### data

| 参数名称 | 参数类型     | 描述                   |
| -------- | ------------ | ---------------------- |
| total    | int64        | 满足查询条件的总记录数 |
| items    | object array | 当前页管控区域列表     |

#### data.items[n]

| 参数名称            | 参数类型 | 描述            |
| ------------------- | -------- | --------------- |
| tenant_id           | string   | 租户 ID         |
| bk_networkarea_id   | int64    | 管控区域 ID     |
| bk_networkarea_name | string   | 管控区域名称    |
| cloud_vendor        | string   | CMDB 云厂商标识 |
