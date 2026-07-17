### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 该接口所需权限：无。
- 该接口功能描述：按过滤条件获取指定进程字段的去重候选值。

### URL

POST /api/v3/process/distinct

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| selector | object | 否 | 需要执行去重查询的字段选择器；值为 `true` 的字段会被查询 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |

#### selector

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| os_type | bool | 否 | 是否返回操作系统类型去重值 |
| cpu_arch | bool | 否 | 是否返回 CPU 架构去重值 |
| version | bool | 否 | 是否返回进程版本去重值 |
| status | bool | 否 | 是否返回进程状态去重值 |
| plugin_name | bool | 否 | 是否返回插件名称去重值 |
| plugin_group | bool | 否 | 是否返回插件分组去重值 |
| plugin_pkg_name | bool | 否 | 是否返回插件包名称去重值 |

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
- 只有 `selector` 中值为 `true` 的字段会执行去重查询；未选择的响应字段返回空数组。

### 调用示例

查询业务 `2` 中运行进程的操作系统类型、CPU 架构和插件名称候选值。

```json
{
  "selector": {
    "os_type": true,
    "cpu_arch": true,
    "plugin_name": true
  },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "status": ["running"]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260717-000004",
  "error": null,
  "permission": null,
  "data": {
    "os_type": ["linux", "windows"],
    "cpu_arch": ["amd64", "arm64"],
    "version": [],
    "status": [],
    "plugin_name": ["bk-monitor-agent", "bk-log-collector"],
    "plugin_group": [],
    "plugin_pkg_name": []
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
| data | object | 去重查询结果 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| os_type | string array | 操作系统类型去重值 |
| cpu_arch | string array | CPU 架构去重值 |
| version | string array | 进程版本去重值 |
| status | string array | 进程状态去重值，可选值：`init`、`running`、`stopped`、`unregister`、`unknown` |
| plugin_name | string array | 插件名称去重值 |
| plugin_group | string array | 插件分组去重值 |
| plugin_pkg_name | string array | 插件包名称去重值 |

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

- 响应始终包含 `data` 下的 7 个字段；未在 `selector` 中选择的字段返回空数组。
- 当前 handler 不执行业务可见性权限收敛逻辑。
