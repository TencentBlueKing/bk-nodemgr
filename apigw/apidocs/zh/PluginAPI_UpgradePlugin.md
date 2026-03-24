### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：。
- 该接口功能描述：批量升级指定主机上的插件，并可指定目标版本、子配置文件和自定义渲染上下文。

### URL

POST /api/v3/plugin/upgrade

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| plugin | object array | 是 | 待升级插件列表，至少包含一个元素 |

#### plugin[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 | 否 | 目标主机ID；未传时默认按 -1 处理，但实际需为正整数 |
| plugin_name | string | 是 | 插件名称 |
| version | string | 是 | 目标插件版本 |
| config_name | string array | 否 | 需要下发的子配置文件名称列表；取值来自对应插件版本发布包中的配置模板名称 |
| custom_config_context | object | 否 | 自定义配置渲染上下文；为自由结构对象，但禁止使用系统保留顶层键 |

**参数说明**：
- `plugin`：至少需要传入一个待处理对象，否则请求会校验失败。
- `bk_host_id`：虽然 Proto 字段为 optional，但未传时会被服务端补为 `-1`，随后校验失败；实际使用时应传入有效主机ID。
- `version`：不能为空，用于指定插件升级目标版本。
- `config_name`：不是固定枚举，实际可用值取决于目标插件版本发布包中的配置模板名称。
- `custom_config_context`：可传入模板渲染所需的自定义变量。禁止使用保留顶层键，包括 `PluginInfo`、`NodeInfo`、`PreDefinitionConstants`、`CustomContext`、`plugin_path`、`nodeman`、`cmdb_instance`、`target`、`control_info`。

### 调用示例

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent",
      "version": "2.5.0",
      "config_name": [
        "bkmonitoragent.conf"
      ],
      "custom_config_context": {
        "cluster_id": "prod-ap-guangzhou"
      }
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000006",
  "error": null,
  "data": {
    "workflow_id": "plugin-upgrade-workflow-123456"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求ID |
| error | object | 错误信息，成功时为null |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| workflow_id | string | 工作流ID，可用于查询插件升级任务状态 |

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误系统标识 |
| message | string | 错误消息 |
| details | array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误代码 |
| message | string | 错误消息 |
