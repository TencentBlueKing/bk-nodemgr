### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：按过滤条件查询拓扑事件中的去重字段值，返回事件类型、管控区域 ID、管控单元 ID、接入点 ID 和操作人候选集合。

### URL

POST /api/v3/topo/event/distinct

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |
| operate_time_range | object | 否 | 操作时间范围，单位为秒级 Unix 时间戳 |

#### exact_include_conditions

精确匹配包含条件，满足任一条件即可进入去重候选范围。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 array | 否 | 管控区域 ID 列表 |
| bk_networkunit_id | int64 array | 否 | 管控单元 ID 列表 |
| accesspoint_id | int64 array | 否 | 接入点 ID 列表 |
| type | string array | 否 | 事件类型列表，可选值：`networkarea-create`、`networkarea-update`、`networkarea-delete`、`networkunit-create`、`networkunit-update`、`networkunit-delete`、`accesspoint-create`、`accesspoint-update`、`accesspoint-delete` |
| operator | string array | 否 | 操作人用户名列表 |

#### fuzzy_include_conditions

模糊匹配包含条件，满足任一条件即可进入去重候选范围。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_name | string array | 否 | 管控区域名称列表，按名称模糊匹配 |
| bk_networkunit_name | string array | 否 | 管控单元名称列表，按名称模糊匹配 |

#### operate_time_range

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| start_timestamp_sec | int64 | 否 | 操作时间范围起始时间，秒级 Unix 时间戳 |
| end_timestamp_sec | int64 | 否 | 操作时间范围结束时间，秒级 Unix 时间戳 |

### 调用示例

查询 2026-01-01 当天、名称包含 `prod` 的拓扑事件，并返回候选去重值。

```json
{
  "exact_include_conditions": {
    "type": [
      "networkunit-create",
      "accesspoint-update"
    ],
    "operator": [
      "admin"
    ]
  },
  "fuzzy_include_conditions": {
    "bk_networkarea_name": [
      "prod"
    ]
  },
  "operate_time_range": {
    "start_timestamp_sec": 1767225600,
    "end_timestamp_sec": 1767311999
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
    "bk_networkarea_id": [
      1,
      2
    ],
    "bk_networkunit_id": [
      1001,
      1002
    ],
    "accesspoint_id": [
      10,
      11
    ],
    "type": [
      "networkunit-create",
      "accesspoint-update"
    ],
    "operator": [
      "admin",
      "ops"
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
| permission | object | 权限信息，当前接口通常为空 |
| data | object | 去重结果 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| bk_networkarea_id | int64 array | 命中条件的拓扑事件对应管控区域 ID 去重结果 |
| bk_networkunit_id | int64 array | 命中条件的拓扑事件对应管控单元 ID 去重结果 |
| accesspoint_id | int64 array | 命中条件的拓扑事件对应接入点 ID 去重结果 |
| type | string array | 命中条件的拓扑事件类型去重结果 |
| operator | string array | 命中条件的拓扑事件操作人去重结果 |

### 说明

- 该接口的契约以 `proto/backend/api/v3/topo.proto`、`pkg/proto/backend/api/v3/event.go` 和 `docs/api/swagger/backend/api/v3/topo.swagger.json` 为准。
- 当前 backend handler `internal/backend/router/api-v3/topo/event.go` 会将请求条件转换为 `types.TopoEventCondition`，再调用 `DistinctTopoEvent` 返回固定五组去重字段，不支持在请求中按字段开关选择返回列。
