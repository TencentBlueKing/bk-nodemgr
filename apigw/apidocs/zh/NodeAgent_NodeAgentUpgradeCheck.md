### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：agent_operate（操作Agent）。
- 该接口功能描述：批量检查 Agent 是否满足升级条件。

### URL

POST /api/v3/node/agent/upgrade_check

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| host | object array | 是 | 主机信息列表 |
| target_version | object array | 否 | 目标版本列表 |

#### host[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 | 否 | 主机 ID |
| bk_networkunit_id | int64 | 否 | 管控单元 ID |
| cpu_arch | string | 否 | CPU 架构 |

#### target_version[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| version | string | 否 | 版本号 |
| cpu_arch | string | 否 | CPU 架构 |
| os_type | string | 否 | 操作系统类型 |

### 调用示例

```json
{
  "host": [
    {
      "bk_host_id": 1,
      "bk_networkunit_id": 1,
      "cpu_arch": "amd64"
    }
  ],
  "target_version": [
    {
      "version": "3.2.1",
      "cpu_arch": "amd64",
      "os_type": "linux"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "results": [
      {
        "status": "running",
        "matched": {
          "bk_host_id": 1,
          "bk_biz_id": 1,
          "bk_networkarea_id": 1,
          "bk_networkunit_id": 1,
          "os_type": "linux",
          "node_role": "agent",
          "bk_host_innerip_list": [
            "10.0.0.1"
          ],
          "bk_host_innerip_v6_list": [
            "2001:db8::1"
          ]
        }
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| results | object array | 结果列表 |

#### data.results[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| status | string | 状态 |
| matched | object | 匹配到的主机信息 |
