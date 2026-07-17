### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 该接口所需权限：无。
- 该接口功能描述：按过滤条件统计每台主机上的进程数量。

### URL

POST /api/v3/process/get_distribution_by_host_id

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 array | 否 | 主机 ID 列表 |
| bk_biz_id | int64 array | 否 | 业务 ID 列表 |
| plugin_group | string array | 否 | 插件分组列表，常见值为 `default`，但不是固定枚举 |
| generation | int64 array | 否 | 进程代际列表 |
| platform_os | string array | 否 | 操作系统类型列表 |
| platform_arch | string array | 否 | CPU 架构列表 |
| status | string array | 否 | 进程状态列表，可选值：`init`、`running`、`stopped`、`unregister`、`unknown` |
| agent_id | string array | 否 | Agent ID 列表 |
| version | string array | 否 | 进程版本列表 |
| plugin_name | string array | 否 | 插件名称列表 |
| plugin_pkg_name | string array | 否 | 插件包名称列表 |

#### fuzzy_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| name | string array | 否 | 进程二进制名称列表，按模糊条件匹配 |
| plugin_pkg_name | string array | 否 | 插件包名称列表，按模糊条件匹配 |

**参数说明**：

- 同一条件对象中的不同字段按 AND 关系组合；同一数组内的值作为该字段的候选集合。
- `exact_include_conditions` 与 `fuzzy_include_conditions` 均为包含型条件，当前接口未暴露排除型条件。

### 调用示例

统计业务 `2` 中运行中的 `bk-monitor-agent` 进程在各主机上的数量。

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "plugin_name": ["bk-monitor-agent"],
    "status": ["running"]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260717-000002",
  "error": null,
  "permission": null,
  "data": {
    "101": 2,
    "102": 1
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
| permission | object | 权限申请信息 |
| data | object | 主机进程数量分布，键为主机 ID，值为匹配的进程数量 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| {bk_host_id} | int64 | 以主机 ID 作为键的动态字段 |
| {bk_host_id} 对应的值 | int64 | 该主机上匹配过滤条件的进程数量 |

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误系统标识 |
| message | string | 错误消息 |
| details | object array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误代码 |
| message | string | 错误消息 |

### 说明

- `data` 在 Proto 中定义为 `map<int64, int64>`。实际 JSON 返回时，对象键会序列化为字符串形式的主机 ID。
- 当前 handler 不执行业务可见性权限收敛逻辑。
