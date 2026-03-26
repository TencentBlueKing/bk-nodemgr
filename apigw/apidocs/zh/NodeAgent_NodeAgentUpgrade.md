### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：agent_operate（操作Agent）。
- 该接口功能描述：批量升级节点Agent到指定版本，支持强制重启和优雅重启超时配置。

### URL

POST /api/v3/node/agent/upgrade

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| host | object array | 是 | 待升级Agent的主机列表 |

#### host[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 | 是 | 主机ID |
| target_version | string | 是 | 目标Agent版本号 |
| force | bool | 否 | 是否强制重启，默认false。设为true时跳过优雅重启流程直接强制重启 |
| graceful_restart_timeout_sec | int64 | 否 | 优雅重启超时时间（秒），超时后将强制重启。默认0表示使用系统默认超时 |

### 调用示例

批量升级两台主机的Agent到2.1.0版本，其中一台使用强制重启。

```json
{
  "host": [
    {
      "bk_host_id": 1001,
      "target_version": "2.1.0",
      "force": false,
      "graceful_restart_timeout_sec": 60
    },
    {
      "bk_host_id": 1002,
      "target_version": "2.1.0",
      "force": true
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "workflow_id": "workflow-abc123def456"
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
| workflow_id | string | 工作流ID，可用于查询升级任务状态 |

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
