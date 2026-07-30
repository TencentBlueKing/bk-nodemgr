### 描述

- 该接口提供版本：v3.0.1-alpha.63+。
- 该接口所需权限：config_policy_view（查看配置策略）。
- 该接口功能描述：查询配置策略列表，支持分页、精确过滤和模糊过滤。

### URL

POST /api/v3/policy/config/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                       |
| ------------------------ | -------- | ---- | -------------------------- |
| page                     | object   | 否   | 分页配置                   |
| only_count               | bool     | 否   | 是否只返回总数，不返回详情 |
| exact_include_conditions | object   | 否   | 精确匹配包含条件           |
| fuzzy_include_conditions | object   | 否   | 模糊匹配包含条件           |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                            |
| -------- | -------- | ---- | ------------------------------- |
| offset   | int32    | 否   | 分页起始位置，起始值为 `0`      |
| limit    | int32    | 否   | 每页条数，协议层限制最大 `1000` |

#### exact_include_conditions

| 参数名称           | 参数类型     | 必选 | 描述                                                                                           |
| ------------------ | ------------ | ---- | ---------------------------------------------------------------------------------------------- |
| configpolicy_id    | int64 array  | 否   | 配置策略 ID 列表                                                                               |
| bk_biz_id          | int64 array  | 否   | 业务 ID 列表                                                                                   |
| configpolicy_type  | string array | 否   | 配置策略类型列表，可选值：`config_policy_agent`、`config_policy_proxy`、`config_policy_plugin` |
| enabled            | bool array   | 否   | 启用状态列表                                                                                   |
| target_plugin_name | string array | 否   | 目标插件名称列表                                                                               |

#### fuzzy_include_conditions

| 参数名称          | 参数类型     | 必选 | 描述                       |
| ----------------- | ------------ | ---- | -------------------------- |
| configpolicy_name | string array | 否   | 配置策略名称列表，模糊匹配 |
| operator          | string array | 否   | 操作人列表，模糊匹配       |

### 调用示例

查询业务 `2` 下启用状态的插件配置策略。

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "configpolicy_type": ["config_policy_plugin"],
    "enabled": [true],
    "target_plugin_name": ["bkmonitorbeat"]
  },
  "fuzzy_include_conditions": {
    "configpolicy_name": ["prod"]
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
    "total": 1,
    "items": [
      {
        "tenant_id": "default",
        "configpolicy_id": 10001,
        "configpolicy_name": "prod-plugin-config",
        "configpolicy_type": "config_policy_plugin",
        "bk_biz_id": 2,
        "remark": "生产环境插件配置",
        "scopes": [
          {
            "bk_networkarea_id": -1,
            "bk_networkunit_id": -1,
            "os_type": "",
            "cpu_arch": ""
          }
        ],
        "configs_string": {
          "bk_cloud_id": "0"
        },
        "configs_int": {
          "plugin.base.cpu_percent_limit": 10,
          "plugin.base.mem_percent_limit": 10
        },
        "configs_bool": {
          "enable_metrics": true
        },
        "enabled": true,
        "updated_time": 1744675200000,
        "operator": "admin",
        "version": 3,
        "priority": 1,
        "target_host_ids": [1001, 1002],
        "target_plugin_name": "bkmonitorbeat"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| code       | int32    | 状态码，`0` 表示成功     |
| message    | string   | 请求信息                 |
| request_id | string   | 请求 ID                  |
| error      | object   | 错误信息，成功时通常为空 |
| permission | object   | 权限信息                 |
| data       | object   | 响应数据                 |

#### data

| 参数名称 | 参数类型 | 描述                         |
| -------- | -------- | ---------------------------- |
| total    | int64    | 当前过滤条件命中的总记录条数 |
| items    | array    | 配置策略列表                 |

#### data.items[n]

| 参数名称           | 参数类型    | 描述                                                                                       |
| ------------------ | ----------- | ------------------------------------------------------------------------------------------ |
| tenant_id          | string      | 租户 ID                                                                                    |
| configpolicy_id    | int64       | 配置策略 ID                                                                                |
| configpolicy_name  | string      | 配置策略名称                                                                               |
| configpolicy_type  | string      | 配置策略类型，可选值：`config_policy_agent`、`config_policy_proxy`、`config_policy_plugin` |
| bk_biz_id          | int64       | 业务 ID                                                                                    |
| remark             | string      | 备注                                                                                       |
| scopes             | array       | 策略生效范围列表                                                                           |
| configs_string     | object      | 字符串类型配置键值对                                                                       |
| configs_int        | object      | 整型配置键值对                                                                             |
| configs_bool       | object      | 布尔配置键值对                                                                             |
| enabled            | bool        | 是否启用                                                                                   |
| updated_time       | int64       | 更新时间，Unix 毫秒时间戳                                                                  |
| operator           | string      | 操作人                                                                                     |
| version            | int64       | 配置策略版本                                                                               |
| priority           | int64       | 优先级，数值越小优先级越高                                                                 |
| target_host_ids    | int64 array | 定向目标主机 ID 列表                                                                       |
| target_plugin_name | string      | 目标插件名称                                                                               |

#### data.items[n].scopes[n]

| 参数名称          | 参数类型 | 描述                           |
| ----------------- | -------- | ------------------------------ |
| bk_networkarea_id | int64    | 管控区域 ID，`-1` 表示任意     |
| bk_networkunit_id | int64    | 管控单元 ID，`-1` 表示任意     |
| os_type           | string   | 操作系统类型，空字符串表示任意 |
| cpu_arch          | string   | CPU 架构，空字符串表示任意     |
