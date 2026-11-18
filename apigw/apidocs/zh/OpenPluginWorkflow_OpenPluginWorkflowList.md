### 描述

- 该接口提供版本：v3.0.1-alpha.114+。
- 该接口所需权限：system_access（系统访问）、plugin_history_view（查看插件历史）。
- 该接口功能描述：查询插件任务流列表，支持分页、精确条件过滤和操作时间范围过滤。

### URL

POST /api/v3/open/plugin/workflow/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                                                   |
| ------------------------ | -------- | ---- | ------------------------------------------------------ |
| page                     | object   | 否   | 分页配置；当 `only_count=false` 或未传时需传入有效分页 |
| only_count               | bool     | 否   | 是否只返回总数，不返回任务流详情，默认为 `false`       |
| exact_include_conditions | object   | 是   | 精确匹配包含条件；无过滤条件时传 `{}`                  |
| fuzzy_include_conditions | object   | 否   | 模糊匹配包含条件，当前对象为空                         |
| operate_time_range       | object   | 否   | 操作时间范围，Unix 秒时间戳                            |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                            |
| -------- | -------- | ---- | ------------------------------- |
| offset   | int32    | 否   | 分页起始位置，非负，默认为 `0`  |
| limit    | int32    | 是   | 每页条数，取值范围为 `(0, 500]` |

#### exact_include_conditions

| 参数名称           | 参数类型     | 必选 | 描述                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ------------------ | ------------ | ---- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| workflow_id        | string array | 否   | 任务流 ID 列表                                                                                                                                                                                                                                                                                                                                                                                                                     |
| type               | string array | 否   | 插件任务流类型列表，可选值：`install_plugin`、`install_plugin_no_start`、`upgrade_plugin`、`uninstall_plugin`、`reconfig_plugin`、`apply_plugin_config`、`apply_plugin_subconfig`、`compose_plugin_subconfig`、`compose_projected_plugin_subconfigs`、`remove_plugin_subconfig`、`start_plugin`、`restart_plugin`、`stop_plugin`、`debug_plugin`、`stop_plugin_v2`、`ensure_plugin_v2`、`uninstall_plugin_v2`、`migrate_plugin_v2` |
| bk_host_id         | int64 array  | 否   | 主机 ID 列表                                                                                                                                                                                                                                                                                                                                                                                                                       |
| status             | string array | 否   | 任务流状态列表，可选值：`running`、`success`、`failed`、`partial_failed`                                                                                                                                                                                                                                                                                                                                                           |
| operator           | string array | 否   | 历史记录中的操作人列表；不改变当前请求的身份                                                                                                                                                                                                                                                                                                                                                                                       |
| bk_host_innerip    | string array | 否   | 主机内网 IPv4 列表                                                                                                                                                                                                                                                                                                                                                                                                                 |
| bk_host_innerip_v6 | string array | 否   | 主机内网 IPv6 列表                                                                                                                                                                                                                                                                                                                                                                                                                 |
| bk_biz_id          | int64 array  | 否   | 业务 ID 列表                                                                                                                                                                                                                                                                                                                                                                                                                       |
| deploy_policy_id   | int64 array  | 否   | 部署策略 ID 列表                                                                                                                                                                                                                                                                                                                                                                                                                   |

#### fuzzy_include_conditions

当前对象为空，传空对象 `{}` 或不传。

#### operate_time_range

| 参数名称            | 参数类型 | 必选 | 描述                                                     |
| ------------------- | -------- | ---- | -------------------------------------------------------- |
| start_timestamp_sec | int64    | 否   | 起始时间戳，单位秒，非负，默认为 `0`                     |
| end_timestamp_sec   | int64    | 否   | 结束时间戳，单位秒，非负，默认为 `0`；不得小于起始时间戳 |

### 调用示例

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "workflow_id": ["wf-plugin-0001"]
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
        "workflow_id": "wf-plugin-0001",
        "trigger_id": "trigger-plugin-0001",
        "type": "upgrade_plugin",
        "bk_host_id": [1001],
        "operator": "admin",
        "operate_time": 1719820000000,
        "finish_time": 0,
        "status": "running",
        "bk_biz_id": [2],
        "deploy_policy_ids": [10001]
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
| items    | array    | 插件任务流列表               |

#### data.items[n]

| 参数名称          | 参数类型    | 描述                                                     |
| ----------------- | ----------- | -------------------------------------------------------- |
| tenant_id         | string      | 租户 ID；当前响应不提供此字段，不应依赖它识别租户        |
| workflow_id       | string      | 任务流 ID                                                |
| trigger_id        | string      | 触发器 ID                                                |
| type              | string      | 插件任务流类型，枚举值同 `exact_include_conditions.type` |
| bk_host_id        | int64 array | 主机 ID 列表                                             |
| operator          | string      | 操作人                                                   |
| operate_time      | int64       | 操作时间，Unix 毫秒时间戳                                |
| finish_time       | int64       | 完成时间，Unix 毫秒时间戳；未记录时为 `0`                |
| status            | string      | 任务流状态，枚举值同 `exact_include_conditions.status`   |
| bk_biz_id         | int64 array | 业务 ID 列表                                             |
| deploy_policy_ids | int64 array | 部署策略 ID 列表                                         |

### 处理规则说明

- 当前用户和租户由网关认证上下文确定，不能通过请求参数指定或切换；`operator` 仅用于过滤历史记录。
- 查询结果按 `plugin_history_view` 权限收敛到当前用户可查询的业务范围。
- 无精确过滤条件时也应传入 `exact_include_conditions: {}`。
- 当 `only_count=false` 或未传时，`page.limit` 必须在 `(0, 500]` 范围内；当 `only_count=true` 时不需要传 `page`，返回 `data.total`，`data.items` 为空数组。
- 任务流按操作时间倒序返回，不支持自定义排序；当前模糊条件对象不支持任何过滤字段。
- 请求时间范围使用秒，响应中的时间使用毫秒。
