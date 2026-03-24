### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：。
- 该接口功能描述：批量卸载指定主机上的插件。

### URL

POST /api/v3/plugin/uninstall

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| plugin | object array | 是 | 待卸载插件列表，至少包含一个元素 |

#### plugin[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 | 否 | 目标主机ID；未传时默认按 -1 处理，但实际需为正整数 |
| plugin_name | string | 是 | 插件名称 |

**参数说明**：
- `plugin`：至少需要传入一个待处理对象，否则请求会校验失败。
- `bk_host_id`：虽然 Proto 字段为 optional，但未传时会被服务端补为 `-1`，随后校验失败；实际使用时应传入有效主机ID。
- `plugin_name`：不能为空。

### 调用示例

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent"
    },
    {
      "bk_host_id": 1002,
      "plugin_name": "bk-log-collector"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000005",
  "error": null,
  "data": {
    "workflow_id": "plugin-uninstall-workflow-123456"
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
| workflow_id | string | 工作流ID，可用于查询插件卸载任务状态 |

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
