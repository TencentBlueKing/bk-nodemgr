### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 版本变更：`v3.0.1-alpha.13+` 增加权限响应字段；`v3.0.1-alpha.18+` 增加主机 IP 与业务 ID 过滤。
- 该接口所需权限：无。
- 该接口功能描述：按过滤条件获取插件任务流的类型、主机 ID、操作人和状态去重候选值。

### URL

POST /api/v3/plugin/workflow/distinct

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件，当前对象为空 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| workflow_id | string array | 否 | 任务流 ID 列表 |
| type | string array | 否 | 插件任务流类型列表，可选值：`install_plugin`、`upgrade_plugin`、`uninstall_plugin`、`reconfig_plugin`、`apply_plugin_subconfig`、`remove_plugin_subconfig`、`start_plugin`、`restart_plugin`、`stop_plugin`、`debug_plugin`、`stop_plugin_v2`、`migrate_plugin_v2` |
| bk_host_id | int64 array | 否 | 主机 ID 列表 |
| status | string array | 否 | 任务流状态列表，可选值：`running`、`success`、`failed`、`partial_failed` |
| operator | string array | 否 | 操作人列表 |
| bk_host_innerip | string array | 否 | 主机内网 IPv4 列表 |
| bk_host_innerip_v6 | string array | 否 | 主机内网 IPv6 列表 |
| bk_biz_id | int64 array | 否 | 业务 ID 列表 |
| deploy_policy_id | int64 array | 否 | 部署策略 ID 列表 |

#### fuzzy_include_conditions

当前对象为空，传空对象 `{}` 或不传。

### 调用示例

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "type": ["upgrade_plugin"],
    "status": ["running"]
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
    "type": ["install_plugin", "upgrade_plugin"],
    "bk_host_id": [1001, 1002],
    "operator": ["admin", "ops"],
    "status": ["running", "success"]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| type | string array | 插件任务流类型去重结果，枚举值同 `exact_include_conditions.type` |
| bk_host_id | int64 array | 主机 ID 去重结果 |
| operator | string array | 操作人去重结果 |
| status | string array | 任务流状态去重结果，枚举值同 `exact_include_conditions.status` |

### 说明

- 当前实现固定返回上述四组去重字段，不支持通过请求选择返回列。
- 当前 handler 未执行 IAM 权限检查。
