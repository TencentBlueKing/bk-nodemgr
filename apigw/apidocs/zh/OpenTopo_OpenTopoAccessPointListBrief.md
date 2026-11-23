### 描述

- 该接口提供版本：v3.0.1-alpha.107+。
- 该接口所需权限：system_access（系统访问）。
- 该接口功能描述：通过用户态入口查询接入点简表，仅返回基础标识，不返回端点地址或安全组池配置。
- 用户身份及租户信息由 API 网关传递，无需在请求体中提供用户名或租户 ID。该简表不按 `networkunit_view` 额外收窄，与完整接入点列表的可见范围不同。

### URL

POST /api/v3/open/topo/accesspoint/list/brief

### 输入参数

| 参数名称                 | 参数类型 | 必选     | 描述                                               |
| ------------------------ | -------- | -------- | -------------------------------------------------- |
| page                     | object   | 条件必选 | `only_count` 不为 `true` 时必须提供有效的 `limit`  |
| only_count               | bool     | 否       | 默认 `false`；为 `true` 时只返回总数，`items` 为空 |
| exact_include_conditions | object   | 否       | 精确匹配条件                                       |

#### page

| 参数名称 | 参数类型 | 必选     | 描述                                                         |
| -------- | -------- | -------- | ------------------------------------------------------------ |
| offset   | int32    | 否       | 起始偏移量，默认 `0`，必须大于等于 `0`                       |
| limit    | int32    | 条件必选 | 查询列表时范围为 `(0, 1000]`；仅计数时可省略，传入值不能为负 |

#### exact_include_conditions

| 参数名称          | 参数类型    | 必选 | 描述             |
| ----------------- | ----------- | ---- | ---------------- |
| bk_networkarea_id | int64 array | 否   | 管控区域 ID 列表 |
| accesspoint_id    | int64 array | 否   | 接入点 ID 列表   |

### 调用示例

```json
{
  "page": { "offset": 0, "limit": 10 },
  "exact_include_conditions": { "bk_networkarea_id": [1] }
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
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-prod",
        "bk_networkarea_id": 1
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

| 参数名称 | 参数类型     | 描述                         |
| -------- | ------------ | ---------------------------- |
| total    | int64        | 满足条件的接入点总数         |
| items    | object array | 当前页基础信息；仅计数时为空 |

#### data.items[n]

| 参数名称          | 参数类型 | 描述            |
| ----------------- | -------- | --------------- |
| accesspoint_id    | int64    | 接入点 ID       |
| accesspoint_name  | string   | 接入点名称      |
| bk_networkarea_id | int64    | 所属管控区域 ID |

### 说明

如需端点地址，请调用 `OpenTopo_OpenTopoAccessPointList`，并具备对应管控单元的查看权限。
