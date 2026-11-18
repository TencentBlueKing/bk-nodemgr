### 描述

- 该接口提供版本：v3.0.1-alpha.114+。
- 该接口所需权限：system_access（系统访问）、plugin_history_view（查看插件历史）。
- 该接口功能描述：根据操作 ID 列表查询插件任务流操作实例，支持仅返回总数。建议每次传入一个操作 ID 定位实例。

### URL

POST /api/v3/open/plugin/workflow/operation/instance/list

### 输入参数

| 参数名称     | 参数类型     | 必选 | 描述                                                                        |
| ------------ | ------------ | ---- | --------------------------------------------------------------------------- |
| only_count   | bool         | 否   | 是否只返回总数，不返回操作实例详情，默认为 `false`                          |
| operation_id | string array | 是   | 操作 ID 列表，至少包含一个元素，且首个元素不可为空；建议每次只传一个操作 ID |

### 调用示例

```json
{
  "only_count": false,
  "operation_id": ["op-plugin-0001"]
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
    "oper_inst_data": [
      {
        "operation_id": "op-plugin-0001",
        "oper_inst_id": "inst-plugin-0001",
        "oper_inst_status": "running",
        "operation_def_name": "upgrade_plugin",
        "parent_operation_id": "",
        "action_names": [
          "fetch_plugin_process",
          "verify_plugin_availability",
          "check_plugin_process_alive",
          "stop_process",
          "inject_plugin_custom_deploy_config",
          "inject_main_config_detail",
          "render_plugin_deployment",
          "fetch_process_sub_config_into_deployment",
          "ensure_and_update_plugin_config_details",
          "render_plugin_config",
          "overwrite_plugin_config_for_compatibility",
          "transfer_plugin_pkg_to_node",
          "upgrade_plugin",
          "wait_plugin_installer_complete",
          "push_plugin_config",
          "start_process",
          "update_process"
        ],
        "life_cycle": {
          "state": "running",
          "create_time": 1719820000000,
          "start_time": 1719820010000,
          "end_time": 0,
          "stop_time": 0
        }
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

| 参数名称       | 参数类型 | 描述                       |
| -------------- | -------- | -------------------------- |
| total          | int64    | 当前查询命中的操作实例总数 |
| oper_inst_data | array    | 操作实例简要数据列表       |

#### data.oper_inst_data[n]

| 参数名称                      | 参数类型     | 描述                                                                                              |
| ----------------------------- | ------------ | ------------------------------------------------------------------------------------------------- |
| operation_id                  | string       | 操作 ID                                                                                           |
| oper_inst_id                  | string       | 操作实例 ID                                                                                       |
| oper_inst_status              | string       | 操作实例状态，可选值：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated` |
| operation_def_name            | string       | 操作定义名称                                                                                      |
| parent_operation_id           | string       | 父操作 ID；无父操作时为空字符串                                                                   |
| action_names                  | string array | 操作实例的动作名称列表，按操作定义中的动作顺序排列                                                |
| life_cycle                    | object       | 操作实例生命周期信息                                                                              |
| latest_action_inst_brief_data | object       | 最新动作实例摘要；当前响应不提供此字段                                                            |

#### data.oper_inst_data[n].life_cycle

| 参数名称    | 参数类型 | 描述                                      |
| ----------- | -------- | ----------------------------------------- |
| state       | string   | 生命周期状态，可选值同 `oper_inst_status` |
| create_time | int64    | 创建时间，Unix 毫秒时间戳                 |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳；未记录时为 `0` |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳；未记录时为 `0` |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳；未记录时为 `0` |

#### data.oper_inst_data[n].latest_action_inst_brief_data

| 参数名称 | 参数类型     | 描述         |
| -------- | ------------ | ------------ |
| name     | string       | 最新动作名称 |
| tags     | string array | 最新动作标签 |

### 处理规则说明

- 当前用户和租户由网关认证上下文确定，不能通过请求参数指定或切换。
- `operation_id` 至少包含一个元素，且首个元素不可为空。排障时每次传入一个 `operation_id`，不支持跨任务流批量排障。
- 当 `only_count=true` 时，返回 `data.total`，`data.oper_inst_data` 为空数组。
- 本接口没有分页、条件过滤或排序参数；不保证操作实例列表的返回顺序。
- 可按 `action_names` 的顺序展示日志查询接口返回的各动作日志，而非依赖日志对象的键顺序。
