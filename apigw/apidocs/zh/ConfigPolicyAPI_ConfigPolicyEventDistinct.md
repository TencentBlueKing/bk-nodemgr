### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：config_policy_history_view（查看配置策略历史）。
- 该接口功能描述：按过滤条件查询配置策略事件的去重字段集合。

### URL

POST /api/v3/policy/config/event/distinct

### 输入参数

| 参数名称                     | 参数类型   | 必选 | 描述       |
|--------------------------|--------|----|----------|
| exact_include_conditions | object | 否  | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否  | 模糊匹配包含条件 |
| operate_time_range       | object | 否  | 操作时间范围   |

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

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [
      2
    ],
    "type": [
      "create",
      "update",
      "enable"
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
  "request_id": "req-123466",
  "data": {
    "configpolicy_id": [
      10001,
      10002
    ],
    "configpolicy_name": [
      "prod-agent-config",
      "prod-proxy-config"
    ],
    "configpolicy_type": [
      "config_policy_agent",
      "config_policy_proxy"
    ],
    "type": [
      "create",
      "update",
      "enable"
    ],
    "version": [
      1,
      2,
      3
    ],
    "operator": [
      "admin",
      "ops"
    ],
    "bk_biz_id": [
      2
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
| data       | object | 去重结果         |

#### data

| 参数名称              | 参数类型         | 描述                                                                                |
|-------------------|--------------|-----------------------------------------------------------------------------------|
| configpolicy_id   | int64 array  | 配置策略 ID 去重结果                                                                      |
| configpolicy_name | string array | 配置策略名称去重结果                                                                        |
| configpolicy_type | string array | 配置策略类型去重结果，可选值：`config_policy_agent`、`config_policy_proxy`、`config_policy_plugin` |
| type              | string array | 事件类型去重结果，可选值：`create`、`update`、`delete`、`enable`、`disable`、`reorder_priorities`   |
| version           | int64 array  | 配置策略版本去重结果                                                                        |
| operator          | string array | 操作人去重结果                                                                           |
| bk_biz_id         | int64 array  | 业务 ID 去重结果                                                                        |
