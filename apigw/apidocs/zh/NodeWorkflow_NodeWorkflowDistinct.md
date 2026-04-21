### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：无。
- 该接口功能描述：按过滤条件获取节点任务流的去重字段候选值，返回任务流类型、业务 ID、操作人和任务流状态。

### URL

POST /api/v3/node/workflow/distinct

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                           |
| ------------------------ | -------- | ---- | ------------------------------ |
| exact_include_conditions | object   | 否   | 精确匹配包含条件               |
| fuzzy_include_conditions | object   | 否   | 模糊匹配包含条件，当前对象为空 |

#### exact_include_conditions

| 参数名称           | 参数类型     | 必选 | 描述                                                                                                                                                                                                   |
| ------------------ | ------------ | ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| workflow_id        | string array | 否   | 任务流 ID 列表                                                                                                                                                                                         |
| type               | string array | 否   | 任务流类型列表，可选值：`install_agent`、`install_proxy`、`upgrade_agent`、`upgrade_proxy`、`reconfig_agent`、`reconfig_proxy`、`restart_agent`、`restart_proxy`、`uninstall_agent`、`uninstall_proxy` |
| bk_biz_id          | int64 array  | 否   | 业务 ID 列表                                                                                                                                                                                           |
| status             | string array | 否   | 任务流状态列表，可选值：`running`、`success`、`failed`、`partial_failed`                                                                                                                               |
| operator           | string array | 否   | 操作人列表                                                                                                                                                                                             |
| bk_host_innerip    | string array | 否   | 主机内网 IPv4 列表                                                                                                                                                                                     |
| bk_host_innerip_v6 | string array | 否   | 主机内网 IPv6 列表                                                                                                                                                                                     |
| node_role          | string array | 否   | 节点角色列表，可选值：`blank`、`agent`、`proxy`                                                                                                                                                        |

#### fuzzy_include_conditions

当前对象为空，传空对象 `{}` 或不传。

### 调用示例

查询业务 `2` 下 `upgrade_agent` 且状态为 `running` 的任务流候选去重值。

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "type": ["upgrade_agent"],
    "status": ["running"],
    "node_role": ["agent"]
  },
  "fuzzy_include_conditions": {}
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "type": ["install_agent", "upgrade_agent"],
    "bk_biz_id": [2, 3],
    "operator": ["admin", "ops"],
    "status": ["running", "success", "failed"]
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

| 参数名称  | 参数类型     | 描述                                                                                                                                                                                                       |
| --------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| type      | string array | 任务流类型去重结果，可选值：`install_agent`、`install_proxy`、`upgrade_agent`、`upgrade_proxy`、`reconfig_agent`、`reconfig_proxy`、`restart_agent`、`restart_proxy`、`uninstall_agent`、`uninstall_proxy` |
| bk_biz_id | int64 array  | 业务 ID 去重结果                                                                                                                                                                                           |
| operator  | string array | 操作人去重结果                                                                                                                                                                                             |
| status    | string array | 任务流状态去重结果，可选值：`running`、`success`、`failed`、`partial_failed`                                                                                                                               |

### 说明

- 当前实现固定返回 `type`、`bk_biz_id`、`operator`、`status` 四组去重字段，不支持通过请求选择返回列。
- `fuzzy_include_conditions` 在当前协议中为空对象，且转换层不会读取其内容。
- 当前 handler `DistinctNodeWorkflow` 未执行业务可见性权限收敛逻辑，不同于 `ListNodeWorkflow`。
