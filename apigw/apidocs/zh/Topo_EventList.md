### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询拓扑变更事件列表，支持分页、按对象 ID 或操作人精确过滤、按对象名称模糊过滤，以及按操作时间范围过滤。

### URL

POST /api/v3/topo/event/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页配置；单页最大 `1000` 条 |
| only_count | bool | 否 | 是否只返回总数，不返回详情 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |
| operate_time_range | object | 否 | 操作时间范围；未提供时服务端默认查询最近 `365` 天 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| count | bool | 是 | 是否返回总记录条数 |
| start | uint32 | 否 | 记录开始位置，起始值为 `0` |
| limit | uint32 | 否 | 每页限制条数，最大 `1000` |
| sort | string | 否 | 排序字段 |
| order | string | 否 | 排序顺序（`ASC`、`DESC`） |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 array | 否 | 管控区域 ID 列表 |
| bk_networkunit_id | int64 array | 否 | 管控单元 ID 列表 |
| accesspoint_id | int64 array | 否 | 接入点 ID 列表 |
| type | string array | 否 | 事件类型列表。可选值：`networkarea-create`、`networkarea-update`、`networkarea-delete`、`networkunit-create`、`networkunit-update`、`networkunit-delete`、`accesspoint-create`、`accesspoint-update`、`accesspoint-delete` |
| operator | string array | 否 | 操作人列表 |

#### fuzzy_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_name | string array | 否 | 管控区域名称列表，模糊匹配 |
| bk_networkunit_name | string array | 否 | 管控单元名称列表，模糊匹配 |

#### operate_time_range

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| start_timestamp_sec | int64 | 否 | 起始时间戳，单位为秒 |
| end_timestamp_sec | int64 | 否 | 结束时间戳，单位为秒 |

### 调用示例

查询最近一段时间内某个管控区域下，由指定操作人执行的拓扑事件，并返回总数与明细。

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 20,
    "sort": "operate_time",
    "order": "DESC"
  },
  "exact_include_conditions": {
    "bk_networkarea_id": [
      1
    ],
    "type": [
      "networkunit-create",
      "accesspoint-update"
    ],
    "operator": [
      "admin"
    ]
  },
  "fuzzy_include_conditions": {
    "bk_networkunit_name": [
      "prod"
    ]
  },
  "operate_time_range": {
    "start_timestamp_sec": 1735689600,
    "end_timestamp_sec": 1735776000
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
        "type": "networkunit-create",
        "bk_networkarea_id": 1,
        "bk_networkarea_name": "prod-default",
        "bk_networkunit_id": 101,
        "bk_networkunit_name": "prod-unit-a",
        "accesspoint_id": 0,
        "accesspoint_name": "",
        "operate_time": 1735722000000,
        "operator": "admin"
      },
      {
        "tenant_id": "default",
        "type": "accesspoint-update",
        "bk_networkarea_id": 1,
        "bk_networkarea_name": "prod-default",
        "bk_networkunit_id": 101,
        "bk_networkunit_name": "prod-unit-a",
        "accesspoint_id": 2001,
        "accesspoint_name": "ap-prod-1",
        "operate_time": 1735725600000,
        "operator": "admin"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息，当前接口通常为空 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 当前过滤条件命中的总记录条数 |
| items | array | 事件明细列表；当 `only_count` 为 `true` 时通常为空数组 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户 ID |
| type | string | 事件类型 |
| bk_networkarea_id | int64 | 管控区域 ID |
| bk_networkarea_name | string | 管控区域名称 |
| bk_networkunit_id | int64 | 管控单元 ID |
| bk_networkunit_name | string | 管控单元名称 |
| accesspoint_id | int64 | 接入点 ID；非接入点事件时可能为 `0` |
| accesspoint_name | string | 接入点名称；非接入点事件时可能为空字符串 |
| operate_time | int64 | 操作时间，Unix 毫秒时间戳 |
| operator | string | 操作人 |

### 说明

- 该接口的请求与响应契约以 `proto/backend/api/v3/topo.proto`、`pkg/proto/backend/api/v3/event.go` 和 `docs/api/swagger/backend/api/v3/topo.swagger.json` 为准。
- Router 实现位于 `internal/backend/router/api-v3/topo/event.go`，请求中的 `operate_time_range` 未提供时会由协议层默认补齐为最近 `365` 天。
