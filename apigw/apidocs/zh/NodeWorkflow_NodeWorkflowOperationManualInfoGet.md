### 描述

- 该接口提供版本：v3.0.1-alpha.56+。
- 该接口所需权限：agent_operate（操作 Agent）、proxy_operate（操作 Proxy）。
- 该接口功能描述：获取指定操作的手动处理信息，包括网络策略和手动执行命令。

**权限说明**：后端根据 `workflow_id` 对应任务流的节点角色收敛 `agent_operate` 或 `proxy_operate` 权限。

### URL

POST /api/v3/node/workflow/operation/manual/info/get

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| workflow_id | string | 是 | 任务流 ID |
| operation_id | string | 是 | 操作 ID |

### 调用示例

```json
{
  "workflow_id": "wf-20240601-0001",
  "operation_id": "op-20240601-0001"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "network_policies": [],
    "commands": [
      {
        "type": "bash",
        "command": "bash install.sh"
      },
      {
        "type": "bat",
        "command": "install.bat"
      }
    ]
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
| network_policies | object array | 网络策略列表 |
| commands | object array | 手动执行命令列表 |

#### data.network_policies[]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| name | string | 网络策略名称 |
| description_en | string | 英文描述 |
| description_zh | string | 中文描述 |
| source | object | 源端点 |
| target | object | 目标端点 |
| service | object | 服务配置 |

#### source / target

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| name | string | 端点名称 |
| type | string | 端点类型 |
| values | string array | 端点取值列表 |
| description_en | string | 英文描述 |
| description_zh | string | 中文描述 |

#### service

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| protocol | string | 协议 |
| ports | string array | 端口列表 |
| description_en | string | 英文描述 |
| description_zh | string | 中文描述 |

#### data.commands[]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| type | string | 命令类型，可选值：`bash`、`bat` |
| command | string | 命令内容 |
