### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：config_policy_history_view（查看配置策略历史）。
- 该接口功能描述：查询配置策略事件列表，支持分页、条件过滤和操作时间范围过滤。

### URL

POST /api/v3/policy/config/event/list

### 输入参数

| 参数名称                     | 参数类型   | 必选 | 描述            |
|--------------------------|--------|----|---------------|
| page                     | object | 否  | 分页配置          |
| only_count               | bool   | 否  | 是否只返回总数，不返回详情 |
| exact_include_conditions | object | 否  | 精确匹配包含条件      |
| fuzzy_include_conditions | object | 否  | 模糊匹配包含条件      |
| operate_time_range       | object | 否  | 操作时间范围        |

#### page

| 参数名称   | 参数类型  | 必选 | 描述                  |
|--------|-------|----|---------------------|
| offset | int32 | 否  | 分页起始位置，起始值为 `0`     |
| limit  | int32 | 否  | 每页条数，协议层限制最大 `1000` |

#### exact_include_conditions

| 参数名称              | 参数类型         | 必选 | 描述                                                                              |
|-------------------|--------------|----|---------------------------------------------------------------------------------|
| configpolicy_id   | int64 array  | 否  | 配置策略 ID 列表                                                                      |
| configpolicy_type | string array | 否  | 配置策略类型列表，可选值：`config_policy_agent`、`config_policy_proxy`、`config_policy_plugin` |
| version           | int64 array  | 否  | 配置策略版本列表                                                                        |
| type              | string array | 否  | 事件类型列表，可选值：`create`、`update`、`delete`、`enable`、`disable`、`reorder_priorities`   |
| bk_biz_id         | int64 array  | 否  | 业务 ID 列表                                                                        |

#### fuzzy_include_conditions

| 参数名称              | 参数类型         | 必选 | 描述            |
|-------------------|--------------|----|---------------|
| configpolicy_name | string array | 否  | 配置策略名称列表，模糊匹配 |
| operator          | string array | 否  | 操作人列表，模糊匹配    |

#### operate_time_range

| 参数名称                | 参数类型  | 必选 | 描述        |
|---------------------|-------|----|-----------|
| start_timestamp_sec | int64 | 否  | 起始时间戳，单位秒 |
| end_timestamp_sec   | int64 | 否  | 结束时间戳，单位秒 |

### 调用示例

查询业务 `2` 的创建和更新事件。

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_biz_id": [
      2
    ],
    "type": [
      "create",
      "update"
    ]
  },
  "fuzzy_include_conditions": {
    "operator": [
      "admin"
    ]
  },
  "operate_time_range": {
    "start_timestamp_sec": 1744588800,
    "end_timestamp_sec": 1744675200
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123465",
  "data": {
    "total": 2,
    "items": [
      {
        "tenant_id": "default",
        "configpolicy_id": 10001,
        "configpolicy_name": "prod-agent-config",
        "type": "create",
        "configpolicy_type": "config_policy_agent",
        "version": 1,
        "operate_time": 1744670000000,
        "operator": "admin",
        "bk_biz_id": 2
      },
      {
        "tenant_id": "default",
        "configpolicy_id": 10001,
        "configpolicy_name": "prod-agent-config-v2",
        "type": "update",
        "configpolicy_type": "config_policy_agent",
        "version": 2,
        "operate_time": 1744671200000,
        "operator": "admin",
        "bk_biz_id": 2
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称       | 参数类型   | 描述           |
|------------|--------|--------------|
| code       | int32  | 状态码，`0` 表示成功 |
| message    | string | 请求信息         |
| request_id | string | 请求 ID        |
| error      | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息         |
| data       | object | 响应数据         |

#### data

| 参数名称  | 参数类型  | 描述             |
|-------|-------|----------------|
| total | int64 | 当前过滤条件命中的总记录条数 |
| items | array | 配置策略事件列表       |

#### data.items[n]

| 参数名称              | 参数类型   | 描述                                                                            |
|-------------------|--------|-------------------------------------------------------------------------------|
| tenant_id         | string | 租户 ID                                                                         |
| configpolicy_id   | int64  | 配置策略 ID                                                                       |
| configpolicy_name | string | 配置策略名称                                                                        |
| type              | string | 事件类型，可选值：`create`、`update`、`delete`、`enable`、`disable`、`reorder_priorities`   |
| configpolicy_type | string | 配置策略类型，可选值：`config_policy_agent`、`config_policy_proxy`、`config_policy_plugin` |
| version           | int64  | 配置策略版本                                                                        |
| operate_time      | int64  | 操作时间，Unix 毫秒时间戳                                                               |
| operator          | string | 操作人                                                                           |
| bk_biz_id         | int64  | 业务 ID                                                                         |
